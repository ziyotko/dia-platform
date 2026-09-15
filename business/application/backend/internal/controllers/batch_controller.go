package controllers

import (
	"application/internal/service"
	"application/pkg/response"

	"github.com/gin-gonic/gin"
)

type BatchController struct {
	service service.BatchService
}

func (ctrl *BatchController) Create(c *gin.Context) {
	// Bound as a plain object so the date pickers' "YYYY-MM-DD HH:mm:ss" values
	// are accepted (the model's *time.Time fields only bind RFC3339).
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	batch, err := ctrl.service.CreateFromPayload(payload)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "批次管理", "新增申报批次", batch.Title)
	response.Ok(c, batch)
}

func (ctrl *BatchController) Update(c *gin.Context) {
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
	recordAudit(c, "批次管理", "编辑申报批次", "id="+c.Param("id"))
	response.OkWithMessage(c, "更新成功", nil)
}

func (ctrl *BatchController) Delete(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.service.Delete(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "批次管理", "删除申报批次", "id="+c.Param("id"))
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctrl *BatchController) Publish(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.service.Publish(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "批次管理", "发布申报批次", "id="+c.Param("id"))
	response.OkWithMessage(c, "批次已发布", nil)
}

func (ctrl *BatchController) Close(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.service.Close(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "批次管理", "结束申报批次", "id="+c.Param("id"))
	response.OkWithMessage(c, "批次已结束", nil)
}

func (ctrl *BatchController) StartReview(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := ctrl.service.StartReview(id); err != nil {
		response.Fail(c, err.Error())
		return
	}
	recordAudit(c, "批次管理", "进入评审阶段", "id="+c.Param("id"))
	response.OkWithMessage(c, "已进入评审阶段", nil)
}

func (ctrl *BatchController) GetByID(c *gin.Context) {
	id := parseUint(c.Param("id"))
	batch, err := ctrl.service.GetByID(id)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, batch)
}

func (ctrl *BatchController) List(c *gin.Context) {
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

// ListVisible lists the batches an applicant may see (申报中 / 评审中 / 已结束)
func (ctrl *BatchController) ListVisible(c *gin.Context) {
	page, size := getPage(c)
	keyword := c.Query("keyword")
	list, total, err := ctrl.service.ListVisible(page, size, keyword)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}
