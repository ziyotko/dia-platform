package services

import (
	"fmt"
	"os"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"portal/internal/models"
	"portal/pkg/utils"
)

func TestContentSourceReimportsDeletedArticleWithMySQL(t *testing.T) {
	dsn := os.Getenv("DIA_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set DIA_TEST_MYSQL_DSN to run the content source integration test")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{NamingStrategy: schema.NamingStrategy{SingularTable: true}})
	if err != nil {
		t.Fatal(err)
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

	stamp := time.Now().UnixNano()
	source := &models.ContentSource{
		Name: "公众号重复导入测试", Provider: models.ContentSourceProviderWeChat,
		AccountType: models.ContentSourceAccountService,
		AppID:       fmt.Sprintf("wx-reimport-test-%d", stamp), SecretEnv: "PORTAL_WECHAT_TEST_SECRET",
		SyncMode: models.ContentSourceSyncPublished, Status: 1,
	}
	service := NewContentSourceService()
	if err := service.Create(source); err != nil {
		t.Fatal(err)
	}
	external := ExternalContentArticle{
		ExternalID: fmt.Sprintf("article-reimport-test-%d", stamp),
		Title:      "删除后重新导入测试", Content: "<p>测试正文</p>",
		URL: "https://mp.weixin.qq.com/s/example", ExternalUpdatedAt: time.Now(),
	}
	importAndCheck := func(wantCreated bool) uint {
		t.Helper()
		created, err := service.importArticle(source, external)
		if err != nil || created != wantCreated {
			t.Fatalf("importArticle() created=%t err=%v, want created=%t", created, err, wantCreated)
		}
		var mapping models.ContentImport
		if err := tx.Where("source_id = ? AND external_id = ?", source.ID, external.ExternalID).First(&mapping).Error; err != nil {
			t.Fatal(err)
		}
		return mapping.ArticleID
	}

	firstID := importAndCheck(true)
	var firstArticle models.Article
	if err := tx.First(&firstArticle, firstID).Error; err != nil || firstArticle.URL != "" {
		t.Fatalf("正文文章仍设置了跳转地址: url=%q err=%v", firstArticle.URL, err)
	}
	// 下一次同步修复旧版自动填入的公众号链接。
	if err := tx.Model(&models.Article{}).Where("id = ?", firstID).Update("url", external.URL).Error; err != nil {
		t.Fatal(err)
	}
	if id := importAndCheck(false); id != firstID {
		t.Fatalf("unchanged article was imported again: old=%d new=%d", firstID, id)
	}
	if err := tx.First(&firstArticle, firstID).Error; err != nil || firstArticle.URL != "" {
		t.Fatalf("同步未修复旧跳转地址: url=%q err=%v", firstArticle.URL, err)
	}
	// 用户后来手工设置的其他跳转地址不能被同步覆盖。
	if err := tx.Model(&models.Article{}).Where("id = ?", firstID).Update("url", "https://example.com/manual").Error; err != nil {
		t.Fatal(err)
	}
	importAndCheck(false)
	if err := tx.First(&firstArticle, firstID).Error; err != nil || firstArticle.URL != "https://example.com/manual" {
		t.Fatalf("同步覆盖了手工跳转地址: url=%q err=%v", firstArticle.URL, err)
	}
	if err := (&ArticleService{}).DeleteArticle(firstID); err != nil {
		t.Fatal(err)
	}
	var mappingCount int64
	if err := tx.Model(&models.ContentImport{}).Where("article_id = ?", firstID).Count(&mappingCount).Error; err != nil || mappingCount != 0 {
		t.Fatalf("article deletion left a deduplication mapping: count=%d err=%v", mappingCount, err)
	}
	secondID := importAndCheck(true)
	if secondID == firstID {
		t.Fatal("reimport reused the deleted article ID")
	}

	// 模拟旧版本留下的孤儿映射；下一次同步也应自动修复。
	if err := tx.Delete(&models.Article{}, secondID).Error; err != nil {
		t.Fatal(err)
	}
	thirdID := importAndCheck(true)
	if thirdID == secondID {
		t.Fatal("orphan mapping was not replaced")
	}
	importAndCheck(false)
}

func TestImportedArticleURL(t *testing.T) {
	if got := importedArticleURL(ExternalContentArticle{Content: "<p>正文</p>", URL: "https://mp.weixin.qq.com/s/example"}); got != "" {
		t.Fatalf("正文文章不应跳转外链: %q", got)
	}
	if got := importedArticleURL(ExternalContentArticle{Content: " \n ", URL: " https://mp.weixin.qq.com/s/example "}); got != "https://mp.weixin.qq.com/s/example" {
		t.Fatalf("无正文文章应保留原文链接: %q", got)
	}
}
