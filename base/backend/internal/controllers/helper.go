package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// parseID 解析路径参数中的 id。
func parseID(c *gin.Context) int {
	id, _ := strconv.Atoi(c.Param("id"))
	return id
}

// resolveTenantID 解析本次操作的目标租户。
//   - 平台超级管理员（tenantID == 0）：可采用请求中指定的租户，未指定则为平台级（0）；
//   - 普通租户用户：一律强制为自身所属租户，忽略请求中的 tenantId，避免跨租户写入。
func resolveTenantID(c *gin.Context, requested uint64) uint64 {
	callerTenantID := c.GetUint64("tenantID")
	if callerTenantID != 0 {
		return callerTenantID
	}
	return requested
}
