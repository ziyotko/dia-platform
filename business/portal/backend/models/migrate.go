package models

import (
	"fmt"
	"strings"

	"server/utils"
)

// fulltextIndexes 声明必须存在的 FULLTEXT 索引。
//
// 背景：文章列表 / 公开搜索使用 `MATCH(col) AGAINST (? IN BOOLEAN MODE)`，
// 而 GORM AutoMigrate 不支持 FULLTEXT 索引（只能建普通索引），
// 因此新库、重建库或手工清理过的库会直接报错：
//
//	Can't find FULLTEXT index matching the column list
//
// 启动时按下面定义幂等补齐；索引名与既有环境保持一致，避免重复建索引。
var fulltextIndexes = []struct {
	table   string
	name    string
	columns []string
}{
	{table: "article", name: "idx_article_title_fulltext", columns: []string{"title"}},
	{table: "article", name: "idx_article_author_fulltext", columns: []string{"author"}},
	{table: "article", name: "idx_article_source_fulltext", columns: []string{"source"}},
}

// EnsureFulltextIndexes 幂等创建搜索所需的 FULLTEXT 索引。
// 同名索引或列完全相同的索引已存在时跳过；创建失败只记日志，不阻断启动。
func EnsureFulltextIndexes() {
	existing, err := existingFulltextIndexes()
	if err != nil {
		utils.Logger.Warnf("[migrate] 读取 FULLTEXT 索引信息失败，跳过补齐: %s", err)
		return
	}
	for _, ft := range fulltextIndexes {
		cols := strings.Join(ft.columns, ",")
		if current, ok := existing[ft.table][ft.name]; ok {
			if current == cols {
				continue
			}
			utils.Logger.Warnf("[migrate] FULLTEXT 索引 %s.%s 实际列为 (%s)，与预期 (%s) 不一致，已跳过", ft.table, ft.name, current, cols)
			continue
		}
		if name := findIndexByColumns(existing[ft.table], cols); name != "" {
			utils.Logger.Infof("[migrate] FULLTEXT 索引 %s(%s) 已存在（索引名 %s），无需创建", ft.table, cols, name)
			continue
		}
		ddl := fmt.Sprintf("ALTER TABLE `%s` ADD FULLTEXT INDEX `%s` (%s)",
			ft.table, ft.name, "`"+strings.Join(ft.columns, "`,`")+"`")
		if err := utils.DB.Exec(ddl).Error; err != nil {
			utils.Logger.Errorf("[migrate] 创建 FULLTEXT 索引 %s.%s 失败: %s", ft.table, ft.name, err)
			continue
		}
		utils.Logger.Infof("[migrate] 已创建 FULLTEXT 索引 %s.%s (%s)", ft.table, ft.name, cols)
	}
}

// existingFulltextIndexes 返回 表名 -> 索引名 -> 按序拼接的列名（如 "title"）。
func existingFulltextIndexes() (map[string]map[string]string, error) {
	type fulltextRow struct {
		TableName string `gorm:"column:table_name"`
		IndexName string `gorm:"column:index_name"`
		Columns   string `gorm:"column:columns"`
	}
	var rows []fulltextRow
	err := utils.DB.Raw(`SELECT TABLE_NAME AS table_name, INDEX_NAME AS index_name,
			GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX) AS columns
		FROM information_schema.STATISTICS
		WHERE TABLE_SCHEMA = DATABASE() AND INDEX_TYPE = 'FULLTEXT'
		GROUP BY TABLE_NAME, INDEX_NAME`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]map[string]string, len(rows))
	for _, row := range rows {
		if result[row.TableName] == nil {
			result[row.TableName] = map[string]string{}
		}
		result[row.TableName][row.IndexName] = row.Columns
	}
	return result, nil
}

// findIndexByColumns 在 索引名->列名 映射中查找列完全匹配的索引名，找不到返回空串。
func findIndexByColumns(indexes map[string]string, columns string) string {
	for name, cols := range indexes {
		if cols == columns {
			return name
		}
	}
	return ""
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

// 注：原「页面(page) 合并进模板(template)」的一次性启动迁移 `MigratePageLayerToTemplates()`
// 已于 2026-09-21 移除（各环境旧结构库均已迁移完成）。
// 如需从旧结构（存在 `page` 表与 `column`/`ad`/`link`/`article_column_publish.page_id`）升级，
// 请按 `business/portal/DEPLOY.md` 的「页面层合并迁移」章节手工执行 SQL（先备份，注意外键需先 DROP）。
