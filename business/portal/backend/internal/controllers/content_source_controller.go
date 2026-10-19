package controllers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"portal/internal/models"
	"portal/internal/services"
	"portal/pkg/utils"
)

type ContentSourceController struct {
	service *services.ContentSourceService
}

type contentSourceRequest struct {
	models.ContentSource
	AppSecret string `json:"appSecret"`
}

func NewContentSourceController() *ContentSourceController {
	return &ContentSourceController{service: services.NewContentSourceService()}
}

func (c *ContentSourceController) List(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	status := -1
	if raw := ctx.Query("status"); raw != "" {
		status, _ = strconv.Atoi(raw)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	list, total, err := c.service.List(ctx.Query("name"), status, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("获取内容源失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取成功", utils.PageData(list, total, page, pageSize)))
}

func (c *ContentSourceController) Create(ctx *gin.Context) {
	var req contentSourceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}
	if rejectLegacyAppSecret(ctx, req.AppSecret) {
		return
	}
	source := req.ContentSource
	source.CreatedBy = ctx.GetUint("userID")
	if err := c.service.Create(&source); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("创建内容源失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("创建成功", gin.H{"id": source.ID}))
}

func (c *ContentSourceController) Update(ctx *gin.Context) {
	id, ok := parseContentSourceID(ctx)
	if !ok {
		return
	}
	var req contentSourceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}
	if rejectLegacyAppSecret(ctx, req.AppSecret) {
		return
	}
	source := req.ContentSource
	if err := c.service.Update(id, &source); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("更新内容源失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新成功", nil))
}

func (c *ContentSourceController) Delete(ctx *gin.Context) {
	id, ok := parseContentSourceID(ctx)
	if !ok {
		return
	}
	if err := c.service.Delete(id); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("删除内容源失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("删除成功", nil))
}

func (c *ContentSourceController) TestConnection(ctx *gin.Context) {
	id, ok := parseContentSourceID(ctx)
	if !ok {
		return
	}
	if err := c.service.TestConnection(id); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("连接测试失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("连接成功，AppID 与 AppSecret 可用", nil))
}

func (c *ContentSourceController) StartSync(ctx *gin.Context) {
	id, ok := parseContentSourceID(ctx)
	if !ok {
		return
	}
	log, err := c.service.StartSync(id, ctx.GetUint("userID"), currentContentSyncOperator(ctx))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("启动同步失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("同步任务已提交", log))
}

func (c *ContentSourceController) ListLogs(ctx *gin.Context) {
	id, ok := parseContentSourceID(ctx)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	logs, total, err := c.service.ListSyncLogs(id, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("获取同步记录失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取成功", utils.PageData(logs, total, page, pageSize)))
}

func (c *ContentSourceController) GetLog(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil || id == 0 {
		ctx.JSON(http.StatusOK, utils.Error(1, "同步记录ID无效"))
		return
	}
	log, err := c.service.GetSyncLog(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("获取同步记录失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取成功", log))
}

// 只保存环境变量名的版本发布前，前端会直接提交 appSecret；
// 收到该字段说明管理后台仍是旧版页面，日志留痕便于区分「环境变量没配」和「前端没更新」。
func rejectLegacyAppSecret(ctx *gin.Context, appSecret string) bool {
	if strings.TrimSpace(appSecret) == "" {
		return false
	}
	log.Printf("[content-source] 拒绝旧版页面提交 appSecret（userID=%d, ip=%s, ua=%s），请重新部署管理后台前端",
		ctx.GetUint("userID"), ctx.ClientIP(), ctx.Request.UserAgent())
	ctx.JSON(http.StatusOK, utils.Error(1,
		"AppSecret 不在后台保存，请填写环境变量名并在后端配置密钥"+
			"（本次请求仍提交了 appSecret，说明管理后台前端是旧版本，请部署最新前端后强制刷新页面）"))
	return true
}

func parseContentSourceID(ctx *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil || id == 0 {
		ctx.JSON(http.StatusOK, utils.Error(1, "内容源ID无效"))
		return 0, false
	}
	return uint(id), true
}

func currentContentSyncOperator(ctx *gin.Context) string {
	if current, ok := ctx.Get("currentUser"); ok {
		if user, ok := current.(*models.User); ok {
			for _, candidate := range []string{user.Username, user.Account, user.Email} {
				if strings.TrimSpace(candidate) != "" {
					return candidate
				}
			}
		}
	}
	return "admin"
}
