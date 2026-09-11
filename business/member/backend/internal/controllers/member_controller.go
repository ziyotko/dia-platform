package controllers

import (
	"time"

	"member/internal/middleware"
	"member/internal/models"
	"member/internal/service"
	"member/pkg/db"
	"member/pkg/response"

	"github.com/gin-gonic/gin"
)

type MemberController struct {
	memberService service.MemberService
}

// ListMembers lists all members (admin)
func (ctrl *MemberController) ListMembers(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 10)
	keyword := c.Query("keyword")
	status := c.Query("status")
	memberType := c.Query("member_type")

	members, total, err := ctrl.memberService.ListMembers(page, size, keyword, status, memberType)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"list":  members,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// GetMember returns a member detail (admin)
func (ctrl *MemberController) GetMember(c *gin.Context) {
	id := parseUint(c.Param("id"))
	member, err := ctrl.memberService.GetMember(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, member)
}

// UpdateMemberStatus updates member status (admin)
func (ctrl *MemberController) UpdateMemberStatus(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.memberService.UpdateMemberStatus(id, req.Status); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "状态更新成功", nil)
}

// UpdateMemberLevel updates member level (admin)
func (ctrl *MemberController) UpdateMemberLevel(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req struct {
		LevelID uint64 `json:"level_id" binding:"required"`
		Reason  string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.Reason == "" {
		response.BadRequest(c, "请填写变更原因")
		return
	}
	if err := ctrl.memberService.UpdateMemberLevel(id, req.LevelID, middleware.GetUsername(c), req.Reason); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "等级更新成功", nil)
}

// ResetMemberPassword resets a member's password to the default (admin)
func (ctrl *MemberController) ResetMemberPassword(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.memberService.ResetMemberPassword(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "密码已重置为默认密码", nil)
}

// CreateMember creates a new member directly (admin)
func (ctrl *MemberController) CreateMember(c *gin.Context) {
	var req service.CreateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整信息")
		return
	}
	member, err := ctrl.memberService.CreateMember(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "新增会员成功", member)
}

// GetMemberLevelChanges lists a single member's level change history (admin)
func (ctrl *MemberController) GetMemberLevelChanges(c *gin.Context) {
	id := parseUint(c.Param("id"))
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 10)

	list, total, err := ctrl.memberService.GetMemberLevelChanges(id, page, size)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// ListLevelChanges lists membership change records (admin)
func (ctrl *MemberController) ListLevelChanges(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 10)
	keyword := c.Query("keyword")
	memberType := c.Query("member_type")

	list, total, err := ctrl.memberService.ListLevelChanges(page, size, keyword, memberType)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// ListProfileChanges lists profile change records (资料变更记录, admin)
func (ctrl *MemberController) ListProfileChanges(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	size := parseIntDefault(c.Query("size"), 10)
	keyword := c.Query("keyword")

	list, total, err := ctrl.memberService.ListProfileChanges(page, size, keyword)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// GetMemberLevelOptions returns the levels this member can be changed to (from paid memberships)
func (ctrl *MemberController) GetMemberLevelOptions(c *gin.Context) {
	id := parseUint(c.Param("id"))
	levels, err := ctrl.memberService.GetMemberAvailableLevels(id)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, levels)
}

// GetMemberJoinedOrgs returns the organizations a member has joined (admin)
func (ctrl *MemberController) GetMemberJoinedOrgs(c *gin.Context) {
	id := parseUint(c.Param("id"))
	info, err := ctrl.memberService.GetMemberJoinedOrgs(id)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, info)
}

// DeleteMember deletes a member (admin)
func (ctrl *MemberController) DeleteMember(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.memberService.DeleteMember(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}

// GetMemberStats returns member statistics (admin)
func (ctrl *MemberController) GetMemberStats(c *gin.Context) {
	var total, active, pending, rejected, pendingPayment, todayNew int64

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 会员统计不包含管理员
	db.DB.Model(&models.Member{}).Where("is_admin = ?", false).Count(&total)
	db.DB.Model(&models.Member{}).Where("is_admin = ? AND status = ?", false, "active").Count(&active)
	db.DB.Model(&models.Member{}).Where("is_admin = ? AND status IN ?", false, []string{"pending_review", "pending_payment", "registering"}).Count(&pending)
	db.DB.Model(&models.Member{}).Where("is_admin = ? AND status = ?", false, "rejected").Count(&rejected)
	db.DB.Model(&models.Member{}).Where("is_admin = ? AND status = ?", false, "pending_payment").Count(&pendingPayment)
	db.DB.Model(&models.Member{}).Where("is_admin = ? AND created_at >= ?", false, todayStart).Count(&todayNew)

	response.Success(c, gin.H{
		"total":           total,
		"active":          active,
		"pending":         pending,
		"rejected":        rejected,
		"pending_payment": pendingPayment,
		"today_new":       todayNew,
	})
}
