package models

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
		&WorkflowRole{},
		&WorkflowRoleUser{},
		&ArticleColumnAudit{},
		&ArticleColumnAuditHistory{},
		&ArticleColumnPublish{},
		&StaticLog{},
		&ArticleAttachment{},
	}
}
