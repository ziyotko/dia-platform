package controllers

import (
	"strconv"
	"strings"

	"base/internal/models"
	"base/internal/service"
	"base/pkg/notifier"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type MessageController struct {
	service service.MessageService
}

// Create 新建草稿（草稿不会直接投递，发送请用 /messages/send 或 /messages/:id/send）
func (ctl *MessageController) Create(c *gin.Context) {
	in, ok := bindDraftInput(c)
	if !ok {
		return
	}
	draft, err := ctl.service.CreateDraft(c.GetUint64("userID"), c.GetString("username"), c.GetUint64("tenantID"), in)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, draft)
}

// Update 编辑草稿（仅本人的草稿）
func (ctl *MessageController) Update(c *gin.Context) {
	in, ok := bindDraftInput(c)
	if !ok {
		return
	}
	if err := ctl.service.UpdateDraft(uint64(parseID(c)), c.GetUint64("userID"), c.GetUint64("tenantID"), in); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "草稿已保存", nil)
}

// SendDraft 发送已有草稿
func (ctl *MessageController) SendDraft(c *gin.Context) {
	if err := ctl.service.SendDraft(uint64(parseID(c)), c.GetUint64("userID"), c.GetUint64("tenantID")); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "发送成功", nil)
}

// bindDraftInput 解析草稿请求体
func bindDraftInput(c *gin.Context) (service.DraftInput, bool) {
	var req struct {
		ReceiverID uint64 `json:"receiverId"`
		Title      string `json:"title"`
		Content    string `json:"content"`
		Type       string `json:"type"`
		Priority   string `json:"priority"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return service.DraftInput{}, false
	}
	if req.Type == "" {
		req.Type = "system"
	}
	if req.Priority == "" {
		req.Priority = "normal"
	}
	return service.DraftInput{
		ReceiverID: req.ReceiverID,
		Title:      req.Title,
		Content:    req.Content,
		Type:       req.Type,
		Priority:   req.Priority,
	}, true
}

func (ctl *MessageController) Send(c *gin.Context) {
	var req struct {
		ReceiverIDs  []uint64          `json:"receiverIds"`
		ReceiverType string            `json:"receiverType"`
		Title        string            `json:"title"`
		Content      string            `json:"content"`
		Type         string            `json:"type"`
		Priority     string            `json:"priority"`
		Channel      string            `json:"channel"`
		TemplateCode string            `json:"templateCode"`
		Vars         map[string]string `json:"vars"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	senderID := c.GetUint64("userID")
	senderName := c.GetString("username")
	tenantID := c.GetUint64("tenantID")

	if req.Type == "" {
		req.Type = "system"
	}
	if req.Priority == "" {
		req.Priority = "normal"
	}

	// 按模板发送：未显式提供标题/渠道/内容时取模板，并用 vars 渲染 {{占位符}}
	if req.TemplateCode != "" {
		tpl, err := (service.MessageTemplateService{}).GetByCode(req.TemplateCode, tenantID)
		if err != nil {
			response.FailWithCode(c, response.CodeBadRequest, "消息模板不存在或已停用："+req.TemplateCode)
			return
		}
		if req.Channel == "" {
			req.Channel = tpl.Channel
		}
		if req.Title == "" {
			req.Title = service.RenderMessageTemplate(tpl.Subject, req.Vars)
		}
		if req.Content == "" {
			req.Content = service.RenderMessageTemplate(tpl.Content, req.Vars)
		}
	}
	if req.Channel == "" {
		req.Channel = "in-app"
	}
	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.Content) == "" {
		response.FailWithCode(c, response.CodeBadRequest, "请填写标题与内容")
		return
	}

	// 站外渠道：必须是已注册（已配置）的发送器，暂不支持的渠道直接拒绝而不是静默忽略
	if req.Channel != "in-app" {
		if _, err := notifier.Get(req.Channel); err != nil {
			response.FailWithCode(c, response.CodeBadRequest,
				"暂不支持的发送渠道："+req.Channel+"（请在「系统设置 → 通知渠道」中完成配置）")
			return
		}
	}
	// 邮件/短信需要具体的接收人（手机号/邮箱），不支持全员广播；企业微信为群机器人，本身就是群发
	if req.ReceiverType == "all" && (req.Channel == "email" || req.Channel == "sms") {
		response.FailWithCode(c, response.CodeBadRequest, "全员广播暂不支持该渠道，请指定接收用户或改用站内信/企业微信")
		return
	}

	var err error
	if req.ReceiverType == "all" {
		err = ctl.service.Broadcast(senderID, senderName, tenantID, req.Title, req.Content, req.Type, req.Priority)
	} else {
		err = ctl.service.SendToUsers(senderID, senderName, tenantID, req.ReceiverIDs, req.Title, req.Content, req.Type, req.Priority)
	}
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	// 站外渠道：站内信写入后额外发送（失败只记日志，不影响主流程）
	if req.Channel != "in-app" {
		go ctl.sendExternal(tenantID, req.Channel, req.ReceiverIDs, req.ReceiverType, req.Title, req.Content)
	}

	response.OkWithMessage(c, "发送成功", nil)
}

// sendExternal 按渠道把消息推送到站外（邮件用邮箱、短信用手机号、企业微信用群机器人）。
// 接收人范围与站内信同一口径：只能触达本租户的用户（平台超管 tenantID=0 时不限租户）。
func (ctl *MessageController) sendExternal(tenantID uint64, channel string, receiverIDs []uint64, receiverType, subject, content string) {
	// 企业微信群机器人：一个 webhook 对应一个群，整条消息只推一次
	if channel == "wechat" {
		if err := ctl.service.SendExternal(channel, "", subject, content); err != nil {
			logrus.WithError(err).Warn("推送企业微信消息失败")
		}
		return
	}

	userSvc := service.UserService{}
	for _, uid := range receiverIDs {
		user, err := userSvc.GetByID(uid, tenantID)
		if err != nil {
			continue
		}
		var to string
		switch channel {
		case "email":
			to = user.Email
		case "sms":
			to = user.Phone
		}
		if to == "" {
			continue
		}
		if err := ctl.service.SendExternal(channel, to, subject, content); err != nil {
			logrus.WithError(err).Warnf("发送%s通知失败: userID=%d to=%s", channel, uid, to)
		}
	}
}

// Channels 当前已接入（已配置）的发送渠道，供前端渲染发送渠道下拉。
func (ctl *MessageController) Channels(c *gin.Context) {
	response.Ok(c, gin.H{"channels": notifier.Names()})
}

func (ctl *MessageController) Delete(c *gin.Context) {
	id := uint64(parseID(c))
	tenantID := c.GetUint64("tenantID")
	// 广播消息在库中只有一行（租户内共享），只允许发送者或管理员删除
	admin := models.IsPlatformTenant(tenantID) || isAdminUser(c.GetUint64("userID"))
	if err := ctl.service.Delete(id, c.GetUint64("userID"), admin, tenantID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctl *MessageController) Get(c *gin.Context) {
	id := uint64(parseID(c))
	m, err := ctl.service.GetByID(id, c.GetUint64("userID"), c.GetUint64("tenantID"))
	if err != nil {
		response.Fail(c, "消息不存在")
		return
	}
	response.Ok(c, m)
}

func (ctl *MessageController) List(c *gin.Context) {
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	isRead, _ := strconv.Atoi(c.DefaultQuery("isRead", "-1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	list, total, err := ctl.service.List(service.MessageListQuery{
		TenantID: c.GetUint64("tenantID"),
		UserID:   c.GetUint64("userID"),
		Box:      c.DefaultQuery("box", "inbox"),
		Status:   status,
		IsRead:   isRead,
		Page:     page,
		Size:     size,
	})
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctl *MessageController) UnreadCount(c *gin.Context) {
	receiverID := c.GetUint64("userID")
	count, err := ctl.service.GetUnreadCount(receiverID, c.GetUint64("tenantID"))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, count)
}

func (ctl *MessageController) MarkRead(c *gin.Context) {
	id := uint64(parseID(c))
	receiverID := c.GetUint64("userID")
	if err := ctl.service.MarkRead(id, receiverID, c.GetUint64("tenantID")); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "已标记为已读", nil)
}

// MarkAllRead 将当前用户的全部未读消息标为已读
func (ctl *MessageController) MarkAllRead(c *gin.Context) {
	count, err := ctl.service.MarkAllRead(c.GetUint64("userID"), c.GetUint64("tenantID"))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "已全部标为已读", gin.H{"count": count})
}
