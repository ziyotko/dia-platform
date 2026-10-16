package services

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"portal/internal/models"
	"portal/pkg/utils"
)

func TestContentSourceViewDoesNotExposeSecret(t *testing.T) {
	t.Setenv("PORTAL_WECHAT_TEST_SECRET", "private-test-secret")
	view := contentSourceView(&models.ContentSource{
		AppID: "wx-test", AppSecret: "legacy-db-secret", SecretEnv: "PORTAL_WECHAT_TEST_SECRET",
	})
	encoded, err := json.Marshal(view)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"private-test-secret", "legacy-db-secret", "appSecret"} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("content source response exposed %q", forbidden)
		}
	}
	if !strings.Contains(string(encoded), `"secretConfigured":true`) {
		t.Fatalf("content source response does not indicate configured secret: %s", encoded)
	}
	if !strings.Contains(string(encoded), `"secretEnv":"PORTAL_WECHAT_TEST_SECRET"`) {
		t.Fatalf("content source response does not include environment variable name: %s", encoded)
	}
}

func TestContentSourceRequiresSecretEnvAndDiscardsPlaintext(t *testing.T) {
	source := &models.ContentSource{
		Name: "测试公众号", Provider: models.ContentSourceProviderWeChat,
		AccountType: models.ContentSourceAccountService, AppID: "wx-test",
		AppSecret: "must-not-be-stored", SecretEnv: "PORTAL_WECHAT_TEST_SECRET",
		SyncMode: models.ContentSourceSyncPublished, Status: 1,
	}
	if err := normalizeAndValidateContentSource(source); err != nil {
		t.Fatal(err)
	}
	if source.AppSecret != "" {
		t.Fatal("plaintext secret was retained")
	}
	source.SecretEnv = ""
	if err := normalizeAndValidateContentSource(source); err == nil {
		t.Fatal("expected missing environment variable name to be rejected")
	}
}

func TestContentSourceSecretEnvPersistenceWithMySQL(t *testing.T) {
	dsn := os.Getenv("DIA_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set DIA_TEST_MYSQL_DSN to run the content source integration test")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{NamingStrategy: schema.NamingStrategy{SingularTable: true}})
	if err != nil {
		t.Fatal(err)
	}
	if !db.Migrator().HasColumn(&models.ContentSource{}, "secret_env") {
		t.Fatal("content_source.secret_env column is missing")
	}
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	previousDB := utils.DB
	utils.DB = tx
	t.Cleanup(func() {
		tx.Rollback()
		utils.DB = previousDB
	})

	service := NewContentSourceService()
	stamp := time.Now().UnixNano()
	source := &models.ContentSource{
		Name: "公众号密钥测试", Provider: models.ContentSourceProviderWeChat,
		AccountType: models.ContentSourceAccountService,
		AppID:       fmt.Sprintf("wx-secret-test-%d", stamp), SecretEnv: "PORTAL_WECHAT_TEST_SECRET", AppSecret: "must-not-be-stored",
		SyncMode: models.ContentSourceSyncPublished, Status: 1,
	}
	if err := service.Create(source); err != nil {
		t.Fatal(err)
	}
	stored, err := service.Get(source.ID)
	if err != nil || stored.AppSecret != "" || stored.SecretEnv != "PORTAL_WECHAT_TEST_SECRET" {
		t.Fatalf("only environment variable name should be saved: err=%v", err)
	}

	update := *source
	update.SecretEnv = "PORTAL_WECHAT_NEW_SECRET"
	update.AppSecret = "another-plaintext-secret"
	if err := service.Update(source.ID, &update); err != nil {
		t.Fatal(err)
	}
	stored, err = service.Get(source.ID)
	if err != nil || stored.AppSecret != "" || stored.SecretEnv != "PORTAL_WECHAT_NEW_SECRET" {
		t.Fatalf("environment variable update or legacy column cleanup failed: err=%v", err)
	}
}
