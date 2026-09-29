package adapter

import (
	"base/internal/models"
	"base/internal/service"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

// AdapterController 子应用统一代理入口
// 路由：/business_base/api/app/:appCode/*path（前缀来自 config.yaml 的 server.api_prefix）
//
// 鉴权口径：入口挂 JWT + 租户开通校验，并按应用启用的接口权限点做二次校验
// （`service.PermissionService.HasAppAccess`：应用未登记权限点时放行，交由子应用自行控制；
// 登记了则未授权的调用返回 403「无权限访问该子应用接口」）。
// 平台超管（tenantID = 0）用于联调与管理，直接放行。
func AdapterController(c *gin.Context) {
	appCode := c.Param("appCode")
	app, ok := DefaultRegistry.Get(appCode)
	if !ok {
		response.FailWithCode(c, response.CodeNotFound, "应用未注册")
		return
	}
	if app.Status != 1 {
		response.Fail(c, "应用已禁用")
		return
	}
	// 当前仅支持 iframe（前端容器）与 proxy（API 代理）两种接入方式
	if app.Type == "iframe" {
		response.FailWithCode(c, response.CodeBadRequest, "IFrame 应用不支持 API 代理")
		return
	}

	tenantID := c.GetUint64("tenantID")
	userID := c.GetUint64("userID")
	if !models.IsPlatformTenant(tenantID) {
		enabled, err := (service.AppInstanceService{}).IsEnabled(tenantID, app.ID)
		if err != nil {
			response.Fail(c, err.Error())
			return
		}
		if !enabled {
			response.FailWithCode(c, response.CodeForbidden, "当前租户未开通该应用")
			return
		}

		// 子应用接口级权限：按应用启用——该 app_code 未登记权限点时交给子应用自行鉴权
		if !(service.UserService{}).IsAdmin(userID) {
			allowed, err := (service.PermissionService{}).HasAppAccess(userID, app.Code, c.Request.Method, c.Param("path"))
			if err != nil {
				response.Fail(c, err.Error())
				return
			}
			if !allowed {
				response.FailWithCode(c, response.CodeForbidden, "无权限访问该子应用接口")
				return
			}
		}
	}
	ProxyToApp(c, app)
}
