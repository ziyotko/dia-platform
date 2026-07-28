package controllers

import (
	"member/internal/service"
	"member/pkg/response"

	"github.com/gin-gonic/gin"
)

type OrganizationController struct {
	orgService service.OrganizationService
}

func (ctrl *OrganizationController) GetTree(c *gin.Context) {
	tree, err := ctrl.orgService.GetOrganizationTree()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, tree)
}

func (ctrl *OrganizationController) GetOrganization(c *gin.Context) {
	id := parseUint(c.Param("id"))
	org, err := ctrl.orgService.GetOrganization(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, org)
}

func (ctrl *OrganizationController) CreateOrganization(c *gin.Context) {
	var req service.CreateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	org, err := ctrl.orgService.CreateOrganization(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "创建成功", org)
}

func (ctrl *OrganizationController) UpdateOrganization(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req service.UpdateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.orgService.UpdateOrganization(id, req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "更新成功", nil)
}

func (ctrl *OrganizationController) DeleteOrganization(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.orgService.DeleteOrganization(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}

// ---- Member Organizations ----

type MemberOrgController struct {
	memberOrgService service.MemberOrgService
}

func (ctrl *MemberOrgController) GetMyOrgs(c *gin.Context) {
	memberID := getMemberID(c)
	orgs, err := ctrl.memberOrgService.GetMyOrgs(memberID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, orgs)
}

func (ctrl *MemberOrgController) JoinOrg(c *gin.Context) {
	memberID := getMemberID(c)
	var req struct {
		OrgID uint64 `json:"org_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.memberOrgService.JoinOrg(memberID, req.OrgID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "加入成功", nil)
}

func (ctrl *MemberOrgController) LeaveOrg(c *gin.Context) {
	memberID := getMemberID(c)
	id := parseUint(c.Param("id"))
	if err := ctrl.memberOrgService.LeaveOrgByID(memberID, id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "退出成功", nil)
}

func getMemberID(c *gin.Context) uint64 {
	return c.GetUint64("memberID")
}
