package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/services"
	"server/utils"
)

type AdController struct {
	adService   *services.AdService
	userService *services.UserService
}

func NewAdController() *AdController {
	return &AdController{
		adService:   &services.AdService{},
		userService: &services.UserService{},
	}
}

func (c *AdController) GetAds(ctx *gin.Context) {
	name := ctx.Query("name")
	templateIDStr := ctx.Query("templateId")
	columnIDStr := ctx.Query("columnId")
	statusStr := ctx.Query("status")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("pageSize", "10")

	templateID := 0
	if templateIDStr != "" {
		if id, err := strconv.Atoi(templateIDStr); err == nil {
			templateID = id
		}
	}
	columnID := 0
	if columnIDStr != "" {
		if id, err := strconv.Atoi(columnIDStr); err == nil {
			columnID = id
		}
	}
	status := -1
	if statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			status = s
		}
	}
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize < 1 {
		pageSize = 10
	}

	ads, total, err := c.adService.GetAds(name, templateID, columnID, status, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取广告列表失败"))
		return
	}

	// 补充模板和栏目名称
	var templateIDs []uint
	var columnIDs []uint
	for _, a := range ads {
		if a.TemplateID > 0 {
			templateIDs = append(templateIDs, a.TemplateID)
		}
		if a.ColumnID > 0 {
			columnIDs = append(columnIDs, a.ColumnID)
		}
	}

	templateMap := make(map[uint]string)
	columnMap := make(map[uint]string)
	if len(templateIDs) > 0 {
		var templates []models.Template
		utils.DB.Where("id IN ?", templateIDs).Find(&templates)
		for _, t := range templates {
			templateMap[t.ID] = t.Name
		}
	}
	if len(columnIDs) > 0 {
		var columns []models.Column
		utils.DB.Where("id IN ?", columnIDs).Find(&columns)
		for _, col := range columns {
			columnMap[col.ID] = col.Name
		}
	}

	var list []gin.H
	for _, a := range ads {
		list = append(list, gin.H{
			"id":           a.ID,
			"name":         a.Name,
			"templateId":   a.TemplateID,
			"templateName": templateMap[a.TemplateID],
			"columnId":     a.ColumnID,
			"columnName":   columnMap[a.ColumnID],
			"image":        a.Image,
			"link":         a.Link,
			"sort":         a.Sort,
			"status":       a.Status,
			"startTime":    formatTime(a.StartTime),
			"endTime":      formatTime(a.EndTime),
			"author":       a.Author,
			"authorCode":   a.AuthorCode,
			"createdAt":    a.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	ctx.JSON(http.StatusOK, utils.Success("获取广告列表成功", utils.PageData(list, total, page, pageSize)))
}

func (c *AdController) GetAdByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "广告ID无效"))
		return
	}
	ad, err := c.adService.GetAdByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取广告失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("获取广告成功", gin.H{
		"id":         ad.ID,
		"name":       ad.Name,
		"templateId": ad.TemplateID,
		"columnId":   ad.ColumnID,
		"image":      ad.Image,
		"link":       ad.Link,
		"sort":       ad.Sort,
		"status":     ad.Status,
		"startTime":  formatTime(ad.StartTime),
		"endTime":    formatTime(ad.EndTime),
		"author":     ad.Author,
		"authorCode": ad.AuthorCode,
	}))
}

func (c *AdController) CreateAd(ctx *gin.Context) {
	var req struct {
		Name       string `json:"name"`
		TemplateID uint   `json:"templateId"`
		ColumnID   uint   `json:"columnId"`
		Image      string `json:"image"`
		Link       string `json:"link"`
		Sort       int    `json:"sort"`
		Status     int    `json:"status"`
		StartTime  string `json:"startTime"`
		EndTime    string `json:"endTime"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}
	if err := services.ValidateTemplateColumn(req.TemplateID, req.ColumnID); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SafeErrText(err)))
		return
	}
	startTime, err := parseTime(req.StartTime)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "开始时间格式错误"))
		return
	}
	endTime, err := parseTime(req.EndTime)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "结束时间格式错误"))
		return
	}
	ad := &models.Ad{
		Name:       req.Name,
		TemplateID: req.TemplateID,
		ColumnID:   req.ColumnID,
		Image:      req.Image,
		Link:       req.Link,
		Sort:       req.Sort,
		Status:     req.Status,
		StartTime:  startTime,
		EndTime:    endTime,
	}
	userID := ctx.GetUint("userID")
	user, err := c.userService.GetUserByID(userID)
	if err == nil && user != nil {
		ad.Author = user.Username
		ad.AuthorCode = strconv.FormatUint(uint64(user.ID), 10)
	}
	err = c.adService.CreateAd(ad)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("创建广告失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("创建广告成功", nil))
}

func (c *AdController) UpdateAd(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "广告ID无效"))
		return
	}
	var req struct {
		Name       string `json:"name"`
		TemplateID uint   `json:"templateId"`
		ColumnID   uint   `json:"columnId"`
		Image      string `json:"image"`
		Link       string `json:"link"`
		Sort       int    `json:"sort"`
		Status     int    `json:"status"`
		StartTime  string `json:"startTime"`
		EndTime    string `json:"endTime"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}
	if err := services.ValidateTemplateColumn(req.TemplateID, req.ColumnID); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SafeErrText(err)))
		return
	}
	startTime, err := parseTime(req.StartTime)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "开始时间格式错误"))
		return
	}
	endTime, err := parseTime(req.EndTime)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "结束时间格式错误"))
		return
	}
	ad := &models.Ad{
		Name:       req.Name,
		TemplateID: req.TemplateID,
		ColumnID:   req.ColumnID,
		Image:      req.Image,
		Link:       req.Link,
		Sort:       req.Sort,
		Status:     req.Status,
		StartTime:  startTime,
		EndTime:    endTime,
	}
	userID := ctx.GetUint("userID")
	user, err := c.userService.GetUserByID(userID)
	if err == nil && user != nil {
		ad.Author = user.Username
		ad.AuthorCode = strconv.FormatUint(uint64(user.ID), 10)
	}
	err = c.adService.UpdateAd(uint(id), ad)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("更新广告失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新广告成功", nil))
}

func (c *AdController) UpdateAdStatus(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "广告ID无效"))
		return
	}
	var req struct {
		Status int `json:"status"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误"))
		return
	}
	err = c.adService.UpdateAdStatus(uint(id), req.Status)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "更新状态失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新状态成功", nil))
}

func (c *AdController) DeleteAd(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "广告ID无效"))
		return
	}
	err = c.adService.DeleteAd(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("删除广告失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("删除广告成功", nil))
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func parseTime(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
