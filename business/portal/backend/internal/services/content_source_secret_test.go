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
	view := contentSourceView(&models.ContentSource{
		AppID: "wx-test", AppSecret: "private-test-secret", SecretEnv: "LEGACY_SECRET_ENV",
	})
	encoded, err := json.Marshal(view)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"private-test-secret", "LEGACY_SECRET_ENV", "appSecret", "secretEnv"} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("content source response exposed %q", forbidden)
		}
	}
	if !strings.Contains(string(encoded), `"secretConfigured":true`) {
		t.Fatalf("content source response does not indicate configured secret: %s", encoded)
	}
}

func TestContentSourceSecretPersistenceWithMySQL(t *testing.T) {
	dsn := os.Getenv("DIA_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set DIA_TEST_MYSQL_DSN to run the content source integration test")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{NamingStrategy: schema.NamingStrategy{SingularTable: true}})
	if err != nil {
		t.Fatal(err)
	}
	if !db.Migrator().HasColumn(&models.ContentSource{}, "app_secret") {
		t.Fatal("content_source.app_secret column is missing")
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
		AppID:       fmt.Sprintf("wx-secret-test-%d", stamp), AppSecret: "first-test-secret",
		SyncMode: models.ContentSourceSyncPublished, Status: 1,
	}
	if err := service.Create(source); err != nil {
		t.Fatal(err)
	}
	stored, err := service.Get(source.ID)
	if err != nil || stored.AppSecret != "first-test-secret" {
		t.Fatalf("saved secret was not read back: err=%v", err)
	}

	update := *source
	update.AppSecret = ""
	if err := service.Update(source.ID, &update); err != nil {
		t.Fatal(err)
	}
	stored, err = service.Get(source.ID)
	if err != nil || stored.AppSecret != "first-test-secret" {
		t.Fatalf("blank edit did not preserve secret: err=%v", err)
	}

	update.AppSecret = "replacement-test-secret"
	if err := service.Update(source.ID, &update); err != nil {
		t.Fatal(err)
	}
	stored, err = service.Get(source.ID)
	if err != nil || stored.AppSecret != "replacement-test-secret" || stored.SecretEnv != "" {
		t.Fatalf("replacement secret or legacy field was not saved: err=%v", err)
	}
}
