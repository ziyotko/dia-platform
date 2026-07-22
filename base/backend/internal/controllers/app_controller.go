package controllers

import (
	"strconv"

	"base/internal/adapter"
	"base/internal/models"
	"base/internal/service"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

type AppController struct {
	service service.AppService
}

func (ctl *AppController) Create(c *gin.Context) {
	var a models.App
	if err := c.ShouldBindJSON(&a); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	if err := ctl.service.Create(&a); err != nil {
		response.Fail(c, err.Error())
		return
	}
	go adapter.DefaultRegistry.Reload()
	response.Ok(c, a)
}

func (ctl *AppController) Update(c *gin.Context) {
	var a models.App
	if err := c.ShouldBindJSON(&a); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	a.ID = uint64(parseID(c))
	if err := ctl.service.Update(&a); err != nil {
		response.Fail(c, err.Error())
		return
	}
	go adapter.DefaultRegistry.Reload()
	response.Ok(c, a)
}

func (ctl *AppController) Delete(c *gin.Context) {
	id := uint64(parseID(c))
	if err := ctl.service.Delete(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	go adapter.DefaultRegistry.Reload()
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctl *AppController) Get(c *gin.Context) {
	id := uint64(parseID(c))
	a, err := ctl.service.GetByID(id)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, a)
}

func (ctl *AppController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	keyword := c.Query("keyword")
	list, total, err := ctl.service.List(page, size, keyword)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}
