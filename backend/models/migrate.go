package models

// AllModels 返回所有需要自动迁移的数据库模型
func AllModels() []interface{} {
	return []interface{}{
		&User{},
		&Menu{},
		&Role{},
		&OperationLog{},
		&LoginLog{},
		&SiteAnalytics{},
		&Settings{},
		&Template{},
		&Page{},
		&Column{},
		&Category{},
		&Tag{},
		&Article{},
		&Ad{},
		&Link{},
		&Department{},
		&Workflow{},
		&WorkflowNode{},
		&ArticleColumnAudit{},
		&ArticleColumnAuditHistory{},
		&ArticleColumnPublish{},
		&StaticLog{},
	}
}
