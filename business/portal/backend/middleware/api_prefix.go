package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"server/config"
	"server/models"
	"server/services"
	"server/utils"
)

// 登录用户路由中无需 api_prefix 授权校验的路径。
// 这些接口属于个人账户/工具类接口或没有对应菜单的只读接口，任意已认证用户均可用；
// 若后续需要收紧，可在此处调整或移除对应条目。
var memberAPIPrefixExemptPaths = map[string]bool{
	"/logout":           true,
	"/profile":          true,
	"/profile/password": true,
	"/menus/user":       true,
	"/upload":           true,
	"/pages":            true,
	"/static-pages":     true,
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
// 需在 AuthMiddleware 之后挂载（依赖上下文中的 userID）。
func MenuAPIPrefixMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 去除全局 API 前缀，得到路由级路径（如 /caamm/api/articles -> /articles）
		path := strings.TrimPrefix(c.Request.URL.Path, config.AppConfig.Server.ApiPrefix)
		path = strings.TrimSuffix(path, "/")
		if path == "" {
			path = "/"
		}

		// 账户/工具类及无菜单只读接口直接放行
		if memberAPIPrefixExemptPaths[path] {
			c.Next()
			return
		}

		userID := c.GetUint("userID")
		if userID == 0 {
			c.JSON(http.StatusOK, utils.Error(1, "未授权"))
			c.Abort()
			return
		}

		prefixes, err := getUserMenuAPIPrefixes(userID)
		if err != nil {
			c.JSON(http.StatusOK, utils.Error(1, "服务异常，请稍后重试"))
			c.Abort()
			return
		}

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

// getUserMenuAPIPrefixes 收集用户已授权菜单（含子菜单）中所有非空的 api_prefix。
func getUserMenuAPIPrefixes(userID uint) ([]string, error) {
	menuService := &services.MenuService{}
	menus, err := menuService.GetUserMenus(userID)
	if err != nil {
		return nil, err
	}

	var prefixes []string
	var collect func([]models.Menu)
	collect = func(items []models.Menu) {
		for _, m := range items {
			if m.APIPrefix != "" {
				prefixes = append(prefixes, m.APIPrefix)
			}
			if len(m.Children) > 0 {
				collect(m.Children)
			}
		}
	}
	collect(menus)
	return prefixes, nil
}
