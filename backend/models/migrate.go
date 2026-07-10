package models

import "server/utils"

// MigrateArticleCategoryNullable 将 article.category_id 改为可空并移除外键约束
func MigrateArticleCategoryNullable() {
	if utils.DB == nil {
		return
	}
	// MySQL：移除外键约束（不存在时忽略错误）
	_ = utils.DB.Exec("ALTER TABLE article DROP FOREIGN KEY fk_article_category")
	// 修改列为可空（GORM AutoMigrate 也会处理，但显式执行更可靠）
	_ = utils.DB.Exec("ALTER TABLE article MODIFY category_id bigint unsigned NULL")
}

// MigrateArticleCategory 将 article.category_id 的旧数据迁移到 article_category 关联表
func MigrateArticleCategory() {
	if utils.DB == nil {
		return
	}
	_ = utils.DB.Exec(`
		INSERT IGNORE INTO article_category (article_id, category_id)
		SELECT id, category_id FROM article WHERE category_id IS NOT NULL
	`)
}

// AllModels 返回所有需要自动迁移的数据库模型
func AllModels() []interface{} {
	return []interface{}{
		&User{},
		&Menu{},
		&Role{},
		&OperationLog{},
		&LoginLog{},
		&VisitAnalytics{},
		&Setting{},
		&Template{},
		&Page{},
		&Column{},
		&Category{},
		&Tag{},
		&Article{},
		&ArticleCategory{},
		&Ad{},
		&Link{},
		&Department{},
		&Organization{},
		&Workflow{},
		&WorkflowNode{},
		&ArticleColumnAudit{},
		&ArticleColumnAuditHistory{},
		&ArticleColumnPublish{},
		&StaticLog{},
		&ArticleAttachment{},
	}
}
