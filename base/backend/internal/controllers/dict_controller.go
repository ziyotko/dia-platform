package controllers

import (
	"strconv"

	"base/internal/models"
	"base/internal/service"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

type DictController struct {
	service service.DictService
}

func (ctl *DictController) Create(c *gin.Context) {
	var d models.Dict
	if err := c.ShouldBindJSON(&d); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	d.TenantID = c.GetUint64("tenantID")
	if err := ctl.service.Create(&d); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, d)
}

func (ctl *DictController) Update(c *gin.Context) {
	var d models.Dict
	if err := c.ShouldBindJSON(&d); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	d.ID = uint64(parseID(c))
	if err := ctl.service.Update(&d, c.GetUint64("tenantID")); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, d)
}

func (ctl *DictController) Delete(c *gin.Context) {
	id := uint64(parseID(c))
	if err := ctl.service.Delete(id, c.GetUint64("tenantID")); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctl *DictController) Get(c *gin.Context) {
	id := uint64(parseID(c))
	d, err := ctl.service.GetByID(id, c.GetUint64("tenantID"))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, d)
}

func (ctl *DictController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))

	q := service.DictListQuery{
		TenantID: c.GetUint64("tenantID"),
		Code:     c.Query("code"),
		Name:     c.Query("name"),
		Status:   status,
		Page:     page,
		Size:     size,
	}
	list, total, err := ctl.service.List(q)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctl *DictController) GetByCode(c *gin.Context) {
	code := c.Param("code")
	d, err := ctl.service.GetByCode(code, c.GetUint64("tenantID"))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, d)
}

func (ctl *DictController) SaveItems(c *gin.Context) {
	id := uint64(parseID(c))
	var req struct {
		Items []models.DictItem `json:"items" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	if err := ctl.service.SaveItems(id, req.Items); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "保存成功", nil)
}
