package controllers

import (
	"time"

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
		Level string `json:"level" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.memberService.UpdateMemberLevel(id, req.Level); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "等级更新成功", nil)
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

	db.DB.Model(&models.Member{}).Count(&total)
	db.DB.Model(&models.Member{}).Where("status = ?", "active").Count(&active)
	db.DB.Model(&models.Member{}).Where("status IN ?", []string{"pending_review", "pending_payment", "registering"}).Count(&pending)
	db.DB.Model(&models.Member{}).Where("status = ?", "rejected").Count(&rejected)
	db.DB.Model(&models.Member{}).Where("status = ?", "pending_payment").Count(&pendingPayment)
	db.DB.Model(&models.Member{}).Where("created_at >= ?", todayStart).Count(&todayNew)

	response.Success(c, gin.H{
		"total":           total,
		"active":          active,
		"pending":         pending,
		"rejected":        rejected,
		"pending_payment": pendingPayment,
		"today_new":       todayNew,
	})
}
