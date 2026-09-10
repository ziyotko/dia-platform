package controllers

import (
	"application/internal/service"
	"application/pkg/response"

	"github.com/gin-gonic/gin"
)

type ExpertController struct {
	service service.ExpertService
}

func (ctrl *ExpertController) List(c *gin.Context) {
	page, size := getPage(c)
	keyword := c.Query("keyword")
	status := c.Query("status")
	list, total, err := ctrl.service.List(page, size, keyword, status)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctrl *ExpertController) Create(c *gin.Context) {
	var req service.ExpertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.Create(req); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "专家库", "新增评审专家", req.Name)
	response.OkWithMessage(c, "新增成功", nil)
}

func (ctrl *ExpertController) Update(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req service.ExpertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.Update(id, req); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "专家库", "编辑评审专家", req.Name)
	response.OkWithMessage(c, "更新成功", nil)
}

func (ctrl *ExpertController) Delete(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.service.Delete(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "专家库", "删除评审专家", "id="+c.Param("id"))
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctrl *ExpertController) SetStatus(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.SetStatus(id, req.Status); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "状态已更新", nil)
}
