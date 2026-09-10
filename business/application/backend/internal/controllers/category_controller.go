package controllers

import (
	"application/internal/models"
	"application/internal/service"
	"application/pkg/response"

	"github.com/gin-gonic/gin"
)

type CategoryController struct {
	service service.CategoryService
}

func (ctrl *CategoryController) Create(c *gin.Context) {
	var category models.ProjectCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.Create(&category); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "类别管理", "新增类别", category.Name)
	response.Ok(c, category)
}

func (ctrl *CategoryController) Update(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if err := ctrl.service.Update(id, updates); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "类别管理", "编辑类别", "id="+c.Param("id"))
	response.OkWithMessage(c, "更新成功", nil)
}

func (ctrl *CategoryController) Delete(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.service.Delete(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "类别管理", "删除类别", "id="+c.Param("id"))
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctrl *CategoryController) List(c *gin.Context) {
	list, err := ctrl.service.List()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, list)
}
