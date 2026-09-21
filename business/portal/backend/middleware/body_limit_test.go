package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// BodyLimitMiddleware 的单元测试（本仓库目前唯一的单元测试文件）。
// 覆盖三条路径：Content-Length 预检、chunked（无 Content-Length）兜底、multipart 放行。

func newBodyLimitRouter(maxMB int) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(BodyLimitMiddleware(maxMB))
	r.POST("/json", func(c *gin.Context) {
		// 下游控制器会读取 body：这里回显长度，用于验证「被读过的 body 已恢复」与「超限时未进入 handler」
		body, _ := io.ReadAll(c.Request.Body)
		c.String(http.StatusOK, "ok:%d", len(body))
	})
	return r
}

// doPost 发送请求；chunked=true 时把 Content-Length 置为 -1，模拟分块传输
func doPost(r *gin.Engine, contentType, body string, chunked bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/json", bytes.NewBufferString(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if chunked {
		req.ContentLength = -1
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestBodyLimitAllowsSmallJSON(t *testing.T) {
	r := newBodyLimitRouter(1) // 上限 1MB
	w := doPost(r, "application/json", `{"a":1}`, false)

	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "ok:7") {
		t.Fatalf("正常请求应放行并把 body 交给 handler，got code=%d body=%q", w.Code, w.Body.String())
	}
}

func TestBodyLimitRejectsOversizedJSON(t *testing.T) {
	r := newBodyLimitRouter(1)
	// 2MB 的 body（base64 内联图片的极端情况）：应被拒绝且不进入 handler
	w := doPost(r, "application/json", strings.Repeat("x", 2<<20), false)

	body := w.Body.String()
	if !strings.Contains(body, "请求体过大") {
		t.Fatalf("超限请求应返回「请求体过大」，got %q", body)
	}
	if strings.Contains(body, "ok:") {
		t.Fatalf("超限请求不应进入 handler，got %q", body)
	}
}

func TestBodyLimitRejectsOversizedChunkedRequest(t *testing.T) {
	r := newBodyLimitRouter(1)
	// Content-Length 未知（chunked）：走 LimitReader 兜底判定
	w := doPost(r, "application/json", strings.Repeat("x", 2<<20), true)

	if !strings.Contains(w.Body.String(), "请求体过大") {
		t.Fatalf("分块传输超限应被拒绝，got %q", w.Body.String())
	}
}

func TestBodyLimitAllowsChunkedRequestAndRestoresBody(t *testing.T) {
	r := newBodyLimitRouter(1)
	w := doPost(r, "application/json", `{"a":1}`, true)

	// 关键：兜底读取后必须把 body 还原，否则下游控制器拿不到参数
	if !strings.Contains(w.Body.String(), "ok:7") {
		t.Fatalf("分块传输未超限应放行且 body 可被 handler 读取，got %q", w.Body.String())
	}
}

func TestBodyLimitExemptsMultipart(t *testing.T) {
	r := newBodyLimitRouter(1)
	// 文件上传（multipart）不受该上限约束，由 upload_controller 与 Nginx 限制
	w := doPost(r, "multipart/form-data; boundary=xxx", strings.Repeat("x", 2<<20), false)

	if !strings.Contains(w.Body.String(), "ok:") {
		t.Fatalf("multipart 请求应放行，got %q", w.Body.String())
	}
}
