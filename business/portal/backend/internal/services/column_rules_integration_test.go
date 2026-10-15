package services

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"portal/internal/models"
	"portal/pkg/utils"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestColumnRulesWithMySQL(t *testing.T) {
	dsn := os.Getenv("DIA_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set DIA_TEST_MYSQL_DSN to run the column integration test")
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
	var template models.Template
	if err := tx.First(&template).Error; err != nil {
		t.Fatal(err)
	}
	stamp := time.Now().UnixNano()
	service := &ColumnService{}
	parent := models.Column{Name: fmt.Sprintf("栏目规则测试父%d", stamp), Code: fmt.Sprintf("testparent%d", stamp), TemplateID: template.ID, Status: 1, DisplayType: 2}
	if err := service.CreateColumn(&parent); err != nil {
		t.Fatal(err)
	}
	article := models.Article{Title: fmt.Sprintf("栏目规则测试文章%d", stamp), Status: models.ArticleStatusPublished, Type: 1}
	if err := tx.Create(&article).Error; err != nil {
		t.Fatal(err)
	}
	for _, model := range []any{
		&models.ArticleColumn{ArticleID: article.ID, ColumnID: parent.ID},
		&models.ArticleColumnPublish{ArticleID: article.ID, ColumnID: parent.ID, TemplateID: template.ID, ArticleTitle: article.Title},
		&models.ArticleColumnAudit{ArticleID: article.ID, ColumnID: parent.ID, Status: 1},
		&models.ArticleColumnAuditHistory{ArticleID: article.ID, ColumnID: parent.ID, Action: 1},
		&models.Ad{Name: "栏目规则测试广告", ColumnID: parent.ID, TemplateID: template.ID, Status: 1},
		&models.Link{Name: "栏目规则测试友链", Url: "https://example.org", ColumnID: parent.ID, TemplateID: template.ID, Status: 1},
	} {
		if err := tx.Create(model).Error; err != nil {
			t.Fatal(err)
		}
	}
	child := models.Column{Name: fmt.Sprintf("栏目规则测试子%d", stamp), Code: fmt.Sprintf("testchild%d", stamp), TemplateID: template.ID, ParentID: parent.ID, Status: 1, DisplayType: 2}
	if err := service.CreateColumn(&child); err != nil {
		t.Fatal(err)
	}
	parentCounts, err := service.GetColumnDataCounts(parent.ID)
	if err != nil || parentCounts.HasData {
		t.Fatalf("parent still has direct data: %+v, %v", parentCounts, err)
	}
	childCounts, err := service.GetColumnDataCounts(child.ID)
	if err != nil || childCounts.Articles != 1 || childCounts.Publishes != 1 || childCounts.Audits != 1 || childCounts.AuditHistory != 1 || childCounts.Ads != 1 || childCounts.Links != 1 {
		t.Fatalf("migration incomplete: %+v, %v", childCounts, err)
	}
	articles, total, err := (&ArticleService{}).GetArticles("", 0, 0, int(parent.ID), 0, -1, -1, 0, "", "", "", 1, 10)
	if err != nil || total != 1 || len(articles) != 1 || articles[0].ID != article.ID {
		t.Fatalf("parent article query failed: total=%d, err=%v", total, err)
	}
	_, publishedTotal, err := (&ArticleService{}).GetArticleColumnPublishes("", parent.ID, 1, 10)
	if err != nil || publishedTotal != 1 {
		t.Fatalf("parent publish query failed: total=%d, err=%v", publishedTotal, err)
	}
	ads, adTotal, err := (&AdService{}).GetAds("", 0, int(parent.ID), -1, 1, 10)
	if err != nil || adTotal != 1 || len(ads) != 1 {
		t.Fatalf("parent ad query failed: total=%d, err=%v", adTotal, err)
	}
	links, linkTotal, err := (&LinkService{}).GetLinks("", 0, int(parent.ID), -1, 1, 10)
	if err != nil || linkTotal != 1 || len(links) != 1 {
		t.Fatalf("parent link query failed: total=%d, err=%v", linkTotal, err)
	}
	staticColumns, err := service.GetColumnPublishes()
	if err != nil {
		t.Fatal(err)
	}
	foundParent := false
	for _, column := range staticColumns {
		foundParent = foundParent || column.ID == parent.ID
	}
	if !foundParent {
		t.Fatal("parent column missing from static publish list")
	}
	if err := (&ArticleService{}).SetArticleColumns(article.ID, []uint{parent.ID}, "test"); err == nil || !strings.Contains(err.Error(), "父栏目") {
		t.Fatalf("parent publication should be rejected, got %v", err)
	}
	if err := ValidateTemplateColumn(template.ID, parent.ID); err != ErrParentColumn {
		t.Fatalf("parent ad/link placement should be rejected, got %v", err)
	}
	if err := ValidateTemplateColumn(template.ID, child.ID); err != nil {
		t.Fatalf("leaf ad/link placement should be accepted, got %v", err)
	}
	if err := service.DeleteColumn(child.ID); err == nil || !strings.Contains(err.Error(), "请先处理数据") {
		t.Fatalf("nonempty column deletion should be rejected, got %v", err)
	}
	empty := models.Column{Name: fmt.Sprintf("栏目规则测试空%d", stamp), Code: fmt.Sprintf("testempty%d", stamp), TemplateID: template.ID, ParentID: parent.ID, Status: 1, DisplayType: 2}
	if err := service.CreateColumn(&empty); err != nil {
		t.Fatal(err)
	}
	if err := service.DeleteColumn(empty.ID); err != nil {
		t.Fatalf("empty column deletion failed: %v", err)
	}
	grandchild := models.Column{Name: fmt.Sprintf("栏目规则测试孙%d", stamp), Code: fmt.Sprintf("testgrandchild%d", stamp), TemplateID: template.ID, ParentID: child.ID, Status: 1, DisplayType: 2}
	if err := service.CreateColumn(&grandchild); err != nil {
		t.Fatal(err)
	}
	grandchildCounts, err := service.GetColumnDataCounts(grandchild.ID)
	if err != nil || grandchildCounts.Articles != 1 {
		t.Fatalf("second-level migration failed: %+v, %v", grandchildCounts, err)
	}
	_, total, err = (&ArticleService{}).GetArticles("", 0, 0, int(parent.ID), 0, -1, -1, 0, "", "", "", 1, 10)
	if err != nil || total != 1 {
		t.Fatalf("nested parent query failed: total=%d, err=%v", total, err)
	}
}
