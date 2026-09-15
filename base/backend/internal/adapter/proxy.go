package adapter

import (
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

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

// ProxyToApp 将请求代理到子应用后端
func ProxyToApp(c *gin.Context, backendURL string) {
	target, err := url.Parse(backendURL)
	if err != nil {
		response.Fail(c, "子应用地址配置错误")
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
		// 将通配符路径拼接到目标后端地址
		req.URL.Path = singleJoiningSlash(target.Path, c.Param("path"))
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

// ForwardRequest 简单转发，用于适配器场景
func ForwardRequest(c *gin.Context, backendURL string, path string) (*http.Response, error) {
	targetURL := strings.TrimSuffix(backendURL, "/") + "/" + strings.TrimPrefix(path, "/")
	req, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		return nil, err
	}
	req.Header = c.Request.Header.Clone()
	setBaseUserHeaders(req, c)
	client := &http.Client{}
	return client.Do(req)
}

func CopyResponse(c *gin.Context, resp *http.Response) {
	defer resp.Body.Close()
	for k, v := range resp.Header {
		for _, vv := range v {
			c.Writer.Header().Add(k, vv)
		}
	}
	c.Writer.WriteHeader(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}
