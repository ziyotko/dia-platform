package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"portal/internal/services"
	"portal/pkg/utils"
)

type MemberStatisticsController struct {
	service *services.MemberStatisticsService
}

func NewMemberStatisticsController() *MemberStatisticsController {
	return &MemberStatisticsController{service: services.NewMemberStatisticsService(utils.DB)}
}

// GetPublishedDashboard returns only aggregates from a published Data Center dataset.
func (c *MemberStatisticsController) GetPublishedDashboard(ctx *gin.Context) {
	year := 0
	if raw := strings.TrimSpace(ctx.Query("year")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 2000 || parsed > 2100 {
			ctx.JSON(http.StatusOK, utils.Error(1, "统计年份无效"))
			return
		}
		year = parsed
	}
	query := services.MemberStatisticsDashboardQuery{
		Year: year, Region: strings.TrimSpace(ctx.Query("region")),
		Province: strings.TrimSpace(ctx.Query("province")), SubField: strings.TrimSpace(ctx.Query("subField")),
	}
	var valid bool
	if query.IsBelt, valid = memberStatisticsFlagFilter(ctx.Query("isBelt")); !valid {
		ctx.JSON(http.StatusOK, utils.Error(1, "一带筛选条件无效"))
		return
	}
	if query.IsAxis, valid = memberStatisticsFlagFilter(ctx.Query("isAxis")); !valid {
		ctx.JSON(http.StatusOK, utils.Error(1, "一轴筛选条件无效"))
		return
	}
	if len([]rune(query.Region)) > 50 || len([]rune(query.Province)) > 50 || len([]rune(query.SubField)) > 100 {
		ctx.JSON(http.StatusOK, utils.Error(1, "筛选条件过长"))
		return
	}
	dashboard, err := c.service.GetPublishedDashboard(query)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(http.StatusOK, utils.Error(1, "该年度暂无已发布统计数据"))
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取数据中心统计信息失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取数据中心统计信息成功", dashboard))
}

func memberStatisticsFlagFilter(raw string) (*int, bool) {
	switch strings.TrimSpace(raw) {
	case "":
		return nil, true
	case "0", "1":
		value, _ := strconv.Atoi(strings.TrimSpace(raw))
		return &value, true
	default:
		return nil, false
	}
}

func memberStatisticsID(ctx *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil || id == 0 {
		ctx.JSON(http.StatusOK, utils.Error(1, "数据集 ID 无效"))
		return 0, false
	}
	return uint(id), true
}

func (c *MemberStatisticsController) GetDatasets(ctx *gin.Context) {
	datasets, err := c.service.ListDatasets()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取年度统计数据集失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取年度统计数据集成功", utils.AllData(datasets, int64(len(datasets)))))
}

func (c *MemberStatisticsController) GetDataset(ctx *gin.Context) {
	id, ok := memberStatisticsID(ctx)
	if !ok {
		return
	}
	detail, err := c.service.GetDatasetDetail(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(http.StatusOK, utils.Error(1, "数据集不存在"))
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取年度统计详情失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取年度统计详情成功", detail))
}

func (c *MemberStatisticsController) GetRows(ctx *gin.Context) {
	id, ok := memberStatisticsID(ctx)
	if !ok {
		return
	}
	query := services.MemberStatisticsRowQuery{
		EnterpriseName: strings.TrimSpace(ctx.Query("enterpriseName")),
		Province:       strings.TrimSpace(ctx.Query("province")),
		Region:         strings.TrimSpace(ctx.Query("region")),
		SubField:       strings.TrimSpace(ctx.Query("subField")),
		MissingRD:      ctx.Query("missingRd") == "1",
		Page:           memberZoneQueryPage(ctx.DefaultQuery("page", "1"), 1),
		PageSize:       memberZoneQueryPageSize(ctx.DefaultQuery("pageSize", "20"), 20),
	}
	rows, total, err := c.service.GetRows(id, query)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(http.StatusOK, utils.Error(1, "数据集不存在"))
		return
	}
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "查询统计明细失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("查询统计明细成功", utils.PageData(rows, total, query.Page, query.PageSize)))
}

func (c *MemberStatisticsController) UpdateStatus(ctx *gin.Context) {
	id, ok := memberStatisticsID(ctx)
	if !ok {
		return
	}
	var req struct {
		Status *int `json:"status"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil || req.Status == nil || (*req.Status != 1 && *req.Status != 2) {
		ctx.JSON(http.StatusOK, utils.Error(1, "状态只能是已发布或已下线"))
		return
	}
	if err := c.service.UpdateStatus(id, *req.Status); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusOK, utils.Error(1, "数据集不存在"))
			return
		}
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("更新数据集状态失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新数据集状态成功", nil))
}
