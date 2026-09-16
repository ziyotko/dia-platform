package adapter

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"base/internal/models"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

// setBaseUserHeaders 向子应用注入底座用户上下文。
// 注意：userID / tenantID 在 gin 上下文里是 uint64，不能用 GetString 读取（会得到空串）。
func setBaseUserHeaders(req *http.Request, c *gin.Context) {
	req.Header.Set("X-Base-User-ID", strconv.FormatUint(c.GetUint64("userID"), 10))
	req.Header.Set("X-Base-Username", c.GetString("username"))
	req.Header.Set("X-Base-Tenant-ID", strconv.FormatUint(c.GetUint64("tenantID"), 10))
}

// ProxyToApp 将请求代理到子应用后端。
// 转发路径 = target 路径 + App.ApiPrefix（子应用自己的 API 前缀，如 /caamm/api）+ 通配符路径，
// 这样前端只需调 `/business_base/api/app/{appCode}/{业务路径}`，子应用无需感知底座前缀。
func ProxyToApp(c *gin.Context, app *models.App) {
	target, err := url.Parse(app.BackendURL)
	if err != nil {
		response.Fail(c, "子应用地址配置错误")
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
		// 将 API 前缀与通配符路径拼接到目标后端地址
		req.URL.Path = singleJoiningSlash(singleJoiningSlash(target.Path, app.ApiPrefix), c.Param("path"))
		req.URL.RawQuery = c.Request.URL.RawQuery
		// 注入底座用户信息
		setBaseUserHeaders(req, c)
	}
	proxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Del("Access-Control-Allow-Origin")
		return nil
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
