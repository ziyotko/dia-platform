package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestContentSourceRequestBindsSecretEnvAndLegacySecret(t *testing.T) {
	var req contentSourceRequest
	if err := json.Unmarshal([]byte(`{"name":"测试公众号","appId":"wx-test","secretEnv":"PORTAL_WECHAT_TEST_SECRET","appSecret":"private-test-secret"}`), &req); err != nil {
		t.Fatal(err)
	}
	if req.Name != "测试公众号" || req.AppID != "wx-test" || req.SecretEnv != "PORTAL_WECHAT_TEST_SECRET" || req.AppSecret != "private-test-secret" {
		t.Fatalf("content source request binding failed: %+v", req.ContentSource)
	}
}

func TestRejectLegacyAppSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)

	empty := httptest.NewRecorder()
	emptyCtx, _ := gin.CreateTestContext(empty)
	emptyCtx.Request = httptest.NewRequest(http.MethodPost, "/content-sources", nil)
	if rejectLegacyAppSecret(emptyCtx, "   ") {
		t.Fatal("空白 appSecret 不应被拒绝")
	}
	if empty.Body.Len() != 0 {
		t.Fatalf("空白 appSecret 不应写入响应: %s", empty.Body.String())
	}

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/content-sources", nil)
	if !rejectLegacyAppSecret(ctx, "private-test-secret") {
		t.Fatal("非空 appSecret 应被拒绝")
	}
	var resp struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if resp.Code == 0 || !strings.Contains(resp.Msg, "旧版本") {
		t.Fatalf("响应未提示前端版本过旧: %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "private-test-secret") {
		t.Fatal("响应中不应回显 appSecret")
	}
}
