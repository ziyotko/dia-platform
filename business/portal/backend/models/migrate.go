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

// PurgeLegacySoftDeletedRows 清理历史软删除残留，返回被物理删除的行数。
//
// 本项目已统一为物理删除（硬删）：模型不再内嵌 `gorm.DeletedAt`，所有查询也不再带
// `deleted_at IS NULL` 过滤。若不清理，旧库里 `deleted_at IS NOT NULL` 的行会因为过滤条件
// 消失而重新出现在列表里（还会继续占用唯一索引），因此启动时逐表物理删除。
//
// 幂等且开销很小：先一次性查出带 `deleted_at` 列的表并统计残留行数，没有残留（含新库）直接返回；
// 只有确实有残留时才逐表 DELETE。
//
// 只记日志、不阻断启动：可能被其它表的外键引用（如 `template` 被 `column` 引用），
// 这类表先跳过、下一轮（子表被清空后）再试，最多 3 轮，仍失败则告警 —— 宁可不清理，
// 也不制造孤儿数据或让服务起不来。
func PurgeLegacySoftDeletedRows() int64 {
	var tables []string
	err := utils.DB.Raw("SELECT TABLE_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND COLUMN_NAME = 'deleted_at'").
		Scan(&tables).Error
	if err != nil {
		utils.Logger.Errorf("[migrate] 读取软删除列信息失败，跳过历史软删除残留清理: %s", err)
		return 0
	}
	if len(tables) == 0 {
		return 0
	}

	// 一次性统计是否真有残留：常规启动（已清干净）到此结束，不必逐表 DELETE
	parts := make([]string, 0, len(tables))
	for _, table := range tables {
		parts = append(parts, "SELECT COUNT(*) AS n FROM `"+table+"` WHERE deleted_at IS NOT NULL")
	}
	var pending int64
	err = utils.DB.Raw("SELECT COALESCE(SUM(n), 0) FROM (" + strings.Join(parts, " UNION ALL ") + ") AS t").Scan(&pending).Error
	if err != nil {
		utils.Logger.Errorf("[migrate] 统计历史软删除残留失败，跳过清理: %s", err)
		return 0
	}
	if pending == 0 {
		return 0
	}

	purged := int64(0)
	remaining := tables
	// 最多 3 轮：外键约束下父子表有先后依赖（父表要等子表先被清空），每轮清掉一批、重试剩下的。
	// 某一轮完全没进展就停止（互相引用之类）；仍失败的表只告警，不阻断启动。
	for round := 1; round <= 3 && len(remaining) > 0; round++ {
		var retry []string
		for _, table := range remaining {
			res := utils.DB.Exec("DELETE FROM `" + table + "` WHERE deleted_at IS NOT NULL")
			if res.Error != nil {
				retry = append(retry, table)
				continue
			}
			if res.RowsAffected > 0 {
				utils.Logger.Infof("[migrate] %s 表物理删除 %d 条历史软删除记录（已统一为硬删）", table, res.RowsAffected)
				purged += res.RowsAffected
			}
		}
		if len(retry) == 0 {
			break
		}
		if round == 3 || len(retry) == len(remaining) {
			for _, table := range retry {
				utils.Logger.Warnf("[migrate] %s 表的历史软删除记录未能清理（可能被其它表的外键引用），请人工处理", table)
			}
			break
		}
		remaining = retry
	}
	return purged
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
//
// 另：「固定静态化时间」（定时自动静态化）功能已整体移除（模型/设置页/文档均无该配置，
// AutoMigrate 不会删列），旧库 `setting` 表残留的 8 个 `*_static_time*` 列
// 请按 `business/portal/DEPLOY.md` 的「移除「固定静态化时间」遗留列」章节手工 DROP。
//
// 再另：本项目已统一为物理删除（硬删）——模型不再内嵌 `gorm.DeletedAt`，代码中不再有
// `Unscoped()` 与 `deleted_at IS NULL`，删除即 `DELETE`。启动时 `PurgeLegacySoftDeletedRows()`
// 会把旧库中 `deleted_at IS NOT NULL` 的历史软删数据物理删除（幂等，新库自动跳过），
// 否则这些行会因查询不再过滤而重新"出现"在列表里。AutoMigrate 不会删列/索引，
// 旧库残留的 `deleted_at` 列与 `idx_<表名>_deleted_at` 索引如需彻底清掉，
// 请按 `business/portal/DEPLOY.md` 的「移除软删除列 deleted_at」章节手工 DROP。
