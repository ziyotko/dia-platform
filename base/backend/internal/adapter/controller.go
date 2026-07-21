package adapter

import (
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

// AdapterController 子应用统一代理入口
// 路由：/api/v1/app/:appCode/*path
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
	if app.Type == "iframe" {
		response.FailWithCode(c, response.CodeBadRequest, "IFrame 应用不支持 API 代理")
		return
	}
	ProxyToApp(c, app.BackendURL)
}
