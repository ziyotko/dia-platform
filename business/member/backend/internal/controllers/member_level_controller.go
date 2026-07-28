package controllers

import (
	"member/internal/service"
	"member/pkg/response"

	"github.com/gin-gonic/gin"
)

type MemberLevelController struct {
	levelService service.MemberLevelService
}

func (ctrl *MemberLevelController) List(c *gin.Context) {
	list, err := ctrl.levelService.List()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.Success(c, list)
}

func (ctrl *MemberLevelController) Get(c *gin.Context) {
	id := parseUint(c.Param("id"))
	l, err := ctrl.levelService.Get(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, l)
}

func (ctrl *MemberLevelController) Create(c *gin.Context) {
	var req service.MemberLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写等级名称")
		return
	}
	l, err := ctrl.levelService.Create(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "创建成功", l)
}

func (ctrl *MemberLevelController) Update(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req service.MemberLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.levelService.Update(id, req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "更新成功", nil)
}

func (ctrl *MemberLevelController) Delete(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.levelService.Delete(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}

func (ctrl *MemberLevelController) MoveUp(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.levelService.MoveUp(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "上移成功", nil)
}

func (ctrl *MemberLevelController) MoveDown(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.levelService.MoveDown(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "下移成功", nil)
}
