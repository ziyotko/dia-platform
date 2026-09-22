package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"server/config"
	"server/services"
	"server/utils"
)

// 无需 api_prefix 授权校验的路径（member 与 admin 两组共用）。
// 这些接口属于个人账户/工具类接口、或跨模块复用的只读基础数据（分类/标签/栏目/页面选项），
// 任意已认证用户均可用；若后续需要收紧，可在此处调整或移除对应条目。
//
// 注意：此处按「路径」**精确**匹配（不做前缀展开），不区分 HTTP 方法。因此 /templates、/columns 这类同时存在只读（member）
// 与写（admin）路由的路径，仅豁免集合路径本身（GET /templates、GET /columns）与挂在其上的写操作（POST /templates、POST /columns），
// 子路径（如 PUT/DELETE /templates/1）**不在**豁免范围内，须由菜单 api_prefix 覆盖（「栏目管理」= /columns,/templates）。
// 「模板管理」菜单自带 /templates 前缀，因此广告/友链等跨模块读取模板下拉时也需要该豁免。
var apiPrefixExemptPaths = map[string]bool{
	"/logout":                    true,
	"/profile":                   true,
	"/profile/password":          true,
	"/menus/user":                true,
	"/menus/tree":                true,
	"/upload":                    true,
	"/user-options":              true,
	"/workflow-role-options":     true,
	"/templates":                 true,
	"/static-pages":              true,
	"/categories/all":            true,
	"/tags/all":                  true,
	"/columns":                   true,
	"/minPasswordLengthSettings": true,
}

// 按「路径前缀」豁免的只读子路径：用于 /xxx/{id} 这类精确匹配覆盖不到的动态路径。
// 仅影响成员组中的只读接口；同名的写接口都在 admin 组，需先过 AdminMiddleware（仅角色 1），不受本表影响。
var apiPrefixExemptPrefixes = []string{
	// 审核弹窗读取流程定义与节点（GET /workflows/:id、GET /workflows/:id/nodes）。
	// 审核角色（内容审核/内容作者）默认只被授予 /articles、/dashboard 前缀，若不豁免，
	// 打开审核弹窗时每个栏目都会报「没有授权」且节点为空，审核无法进行。
	"/workflows/",
}

// isExemptPath 判断路径是否属于豁免范围（先精确匹配，再前缀匹配）
func isExemptPath(path string) bool {
	if apiPrefixExemptPaths[path] {
		return true
	}
	for _, prefix := range apiPrefixExemptPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// pathMatchesAPIPrefix 判断请求路径是否命中某个 api_prefix。
// 支持精确匹配（path == prefix）与子路径匹配（如 /articles 命中 /articles/123）。
func pathMatchesAPIPrefix(path, prefix string) bool {
	if prefix == "" {
		return false
	}
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}

// MenuAPIPrefixMiddleware 校验当前请求路径是否命中用户已授权菜单的 api_prefix。
// 若请求路径未被任一已授权 api_prefix 覆盖，则返回“没有授权”。
// 需在 AuthMiddleware 之后挂载（依赖上下文中的 userID）；member 与 admin 两组均使用，
// 使「菜单可见范围」与「接口可调用范围」保持一致。
func MenuAPIPrefixMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 去除全局 API 前缀，得到路由级路径（如 /caamm/api/articles -> /articles）
		path := strings.TrimPrefix(c.Request.URL.Path, config.AppConfig.Server.ApiPrefix)
		path = strings.TrimSuffix(path, "/")
		if path == "" {
			path = "/"
		}

		// 账户/工具类及无菜单只读接口直接放行
		if isExemptPath(path) {
			c.Next()
			return
		}

		userID := c.GetUint("userID")
		if userID == 0 {
			c.JSON(http.StatusOK, utils.Error(1, "未授权"))
			c.Abort()
			return
		}

		// 菜单前缀 + 管理员标记按用户缓存（15s TTL）：原先每个请求都要「角色→权限→菜单→前缀」查 4~6 次
		access, err := services.GetUserAccess(userID)
		if err != nil {
			c.JSON(http.StatusOK, utils.Error(1, "服务异常，请稍后重试"))
			c.Abort()
			return
		}
		prefixes := access.Prefixes

		for _, prefix := range prefixes {
			if pathMatchesAPIPrefix(path, prefix) {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusOK, utils.Error(1, "没有授权"))
		c.Abort()
	}
}

// collectMenuAPIPrefixes（含多前缀拆分）与缓存已下沉到 services/user_access_cache.go，
// 供菜单授权校验与管理判定共用。
