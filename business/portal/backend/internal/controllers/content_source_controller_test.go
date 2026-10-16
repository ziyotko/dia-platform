package controllers

import (
	"encoding/json"
	"testing"
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
