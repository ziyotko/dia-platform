package controllers

import (
	"encoding/csv"
	"time"

	"application/internal/service"
	"application/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuditController struct {
	service service.AuditService
}

func (ctrl *AuditController) List(c *gin.Context) {
	page, size := getPage(c)
	module := c.Query("module")
	keyword := c.Query("keyword")
	list, total, err := ctrl.service.List(page, size, module, keyword)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

// Export streams the filtered log as CSV. 与其它接口不同，这里返回的不是统一 JSON
// 结构而是文件流（前端用带 token 的 blob 请求下载）；UTF-8 BOM 是为了 Excel
// 直接双击打开时中文不乱码。最多导出 maxAuditExportRows 条。
func (ctrl *AuditController) Export(c *gin.Context) {
	module := c.Query("module")
	keyword := c.Query("keyword")
	list, err := ctrl.service.Export(module, keyword)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	filename := "audit-logs-" + time.Now().Format("20060102-150405") + ".csv"
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("X-Content-Type-Options", "nosniff")
	_, _ = c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})
	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{"ID", "时间", "操作人", "模块", "操作", "详情", "IP"})
	for _, row := range list {
		_ = w.Write([]string{
			itoa64(row.ID),
			row.CreatedAt.Format("2006-01-02 15:04:05"),
			row.Admin,
			row.Module,
			row.Action,
			row.Detail,
			row.IP,
		})
	}
	w.Flush()
	recordAudit(c, "系统日志", "导出日志", "module="+module+" keyword="+keyword)
}
