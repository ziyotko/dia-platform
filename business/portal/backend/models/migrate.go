package models

import (
	"fmt"
	"server/utils"
)

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

// createIndexIfNotExists 在索引不存在时创建索引（兼容 MySQL 5.7/8.0）
func createIndexIfNotExists(table, index, columns string) {
	if utils.DB == nil {
		return
	}
	var count int64
	utils.DB.Raw(
		"SELECT COUNT(*) FROM information_schema.STATISTICS WHERE table_schema = DATABASE() AND table_name = ? AND index_name = ?",
		table, index,
	).Scan(&count)
	if count == 0 {
		utils.DB.Exec(fmt.Sprintf("CREATE INDEX %s ON %s(%s)", index, table, columns))
	}
}

// MigrateWorkflowNodeApproverType 为 workflow_node 表补充 approver_type 字段默认值
func MigrateWorkflowNodeApproverType() {
	if utils.DB == nil {
		return
	}
	_ = utils.DB.Exec("UPDATE workflow_node SET approver_type = 'user' WHERE approver_type = '' OR approver_type IS NULL")
}

// MigrateIndexes 创建 GORM AutoMigrate 不便表达或列顺序需要控制的辅助索引
func MigrateIndexes() {
	if utils.DB == nil {
		return
	}

	// 清理之前 GORM 按字段顺序创建的列顺序错误的复合索引，避免与下方正确顺序索引冲突
	_ = utils.DB.Exec("ALTER TABLE article DROP INDEX idx_article_top_created")
	_ = utils.DB.Exec("ALTER TABLE article DROP INDEX idx_article_status_audit_created")
	_ = utils.DB.Exec("ALTER TABLE article DROP INDEX idx_article_author_status_created")

	// many2many 关联表默认主键为 (article_id, xxx_id)，
	// 反向按 xxx_id 查询时需要单独索引
	createIndexIfNotExists("article_tag", "idx_article_tag_tag_id", "tag_id")
	createIndexIfNotExists("article_column", "idx_article_column_column_id", "column_id")

	// gorm.Model 中的 created_at 无法通过模型标签直接加索引，手动补充
	createIndexIfNotExists("operation_log", "idx_operation_log_created_at", "created_at")
	createIndexIfNotExists("static_log", "idx_static_log_created_at", "created_at")

	// 文章列表/统计常用复合索引（GORM 标签无法控制列顺序，手动创建）
	// 列表查询：WHERE deleted_at IS NULL ORDER BY is_top DESC, created_at DESC LIMIT ?
	createIndexIfNotExists("article", "idx_article_top_created", "is_top, created_at")
	// 列表查询：WHERE status=? AND audit_status=? ORDER BY created_at
	createIndexIfNotExists("article", "idx_article_status_audit_created", "status, audit_status, created_at")
	// 仪表盘：WHERE author_code=? AND status=?
	createIndexIfNotExists("article", "idx_article_author_status_created", "author_code, status, created_at")
}

// AllModels 返回所有需要自动迁移的数据库模型
func AllModels() []any {
	return []any{
		&User{},
		&Menu{},
		&Role{},
		&OperationLog{},
		&LoginLog{},
		&VisitAnalytics{},
		&LikeAnalytics{},
		&ShareAnalytics{},
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
