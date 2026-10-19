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
	otherTemplate := models.Template{
		Name:   fmt.Sprintf("栏目唯一性测试模板%d", stamp),
		Code:   fmt.Sprintf("column-unique-test-%d", stamp),
		Type:   templateTypeSpecial,
		Status: 0,
	}
	if err := tx.Create(&otherTemplate).Error; err != nil {
		t.Fatal(err)
	}
	crossTemplateDuplicate := models.Column{Name: parent.Name, Code: parent.Code, TemplateID: otherTemplate.ID, Status: 1, DisplayType: 2}
	if err := service.CreateColumn(&crossTemplateDuplicate); err != nil {
		t.Fatalf("same name/code in another template should be accepted: %v", err)
	}
	sameTemplateName := models.Column{Name: parent.Name, Code: fmt.Sprintf("test-name-conflict-%d", stamp), TemplateID: template.ID, Status: 1, DisplayType: 2}
	if err := service.CreateColumn(&sameTemplateName); err == nil || !strings.Contains(err.Error(), "栏目名称") {
		t.Fatalf("same name in one template should be rejected, got %v", err)
	}
	sameTemplateCode := models.Column{Name: fmt.Sprintf("栏目编码冲突测试%d", stamp), Code: parent.Code, TemplateID: template.ID, Status: 1, DisplayType: 2}
	if err := service.CreateColumn(&sameTemplateCode); err == nil || !strings.Contains(err.Error(), "栏目编码") {
		t.Fatalf("same code in one template should be rejected, got %v", err)
	}
	updateSource := models.Column{Name: fmt.Sprintf("栏目更新来源%d", stamp), Code: fmt.Sprintf("test-update-source-%d", stamp), TemplateID: template.ID, Status: 1, DisplayType: 2}
	if err := service.CreateColumn(&updateSource); err != nil {
		t.Fatal(err)
	}
	otherTemplateOnly := models.Column{Name: fmt.Sprintf("栏目跨模板更新%d", stamp), Code: fmt.Sprintf("test-cross-template-update-%d", stamp), TemplateID: otherTemplate.ID, Status: 1, DisplayType: 2}
	if err := service.CreateColumn(&otherTemplateOnly); err != nil {
		t.Fatal(err)
	}
	updateSource.Name = otherTemplateOnly.Name
	updateSource.Code = otherTemplateOnly.Code
	if err := service.UpdateColumn(updateSource.ID, &updateSource); err != nil {
		t.Fatalf("updating to name/code used only by another template should be accepted: %v", err)
	}
	updateSource.Name = parent.Name
	updateSource.Code = parent.Code
	if err := service.UpdateColumn(updateSource.ID, &updateSource); err == nil || !strings.Contains(err.Error(), "已存在") {
		t.Fatalf("updating to name/code used in the same template should be rejected, got %v", err)
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
	if err != nil || parentCounts.Articles != 1 || parentCounts.Publishes != 1 || parentCounts.Audits != 1 || parentCounts.AuditHistory != 1 || parentCounts.Ads != 1 || parentCounts.Links != 1 {
		t.Fatalf("parent data changed after creating child: %+v, %v", parentCounts, err)
	}
	childCounts, err := service.GetColumnDataCounts(child.ID)
	if err != nil || childCounts.HasData {
		t.Fatalf("new child unexpectedly inherited parent data: %+v, %v", childCounts, err)
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
	if err := (&ArticleService{}).SetArticleColumns(article.ID, []uint{parent.ID}, "test"); err != nil {
		t.Fatalf("parent publication should be accepted, got %v", err)
	}
	if err := ValidateTemplateColumn(template.ID, parent.ID); err != nil {
		t.Fatalf("parent ad/link placement should be accepted, got %v", err)
	}
	if err := ValidateTemplateColumn(template.ID, child.ID); err != nil {
		t.Fatalf("child ad/link placement should be accepted, got %v", err)
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
	if err != nil || grandchildCounts.HasData {
		t.Fatalf("new grandchild unexpectedly inherited ancestor data: %+v, %v", grandchildCounts, err)
	}
	_, total, err = (&ArticleService{}).GetArticles("", 0, 0, int(parent.ID), 0, -1, -1, 0, "", "", "", 1, 10)
	if err != nil || total != 1 {
		t.Fatalf("nested parent query failed: total=%d, err=%v", total, err)
	}
}
