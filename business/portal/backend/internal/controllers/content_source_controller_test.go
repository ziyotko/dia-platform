package controllers

import (
	"encoding/json"
	"testing"
)

func TestContentSourceRequestBindsAppSecret(t *testing.T) {
	var req contentSourceRequest
	if err := json.Unmarshal([]byte(`{"name":"测试公众号","appId":"wx-test","appSecret":"private-test-secret"}`), &req); err != nil {
		t.Fatal(err)
	}
	if req.Name != "测试公众号" || req.AppID != "wx-test" || req.AppSecret != "private-test-secret" {
		t.Fatalf("content source request binding failed: %+v", req.ContentSource)
	}
}
