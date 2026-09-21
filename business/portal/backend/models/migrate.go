package models

import (
	"fmt"
	"strings"

	"server/utils"

	"gorm.io/gorm"
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

// 页面(page) 合并进模板(template) 的迁移（2026-09-21）。
//
// 背景：模板与页面原本 1:1 绑定，页面只是中间层；现改为「模板直接承载栏目」，
// column/ad/link/article_column_publish 一律引用 template_id，page 表整体移除。
//
// 迁移步骤（幂等，page 表与 page_id 列都不存在时直接返回）：
//  1. 页面未绑定模板（或绑定的模板已被删除）时，按页面信息补建模板，保留全部数据；
//  2. 把 page.code / page.route_path 回填到对应模板（仅在模板该字段为空时，不覆盖使用方维护的值）；
//  3. 把各表的 page_id 回填为对应页面的 template_id；
//  4. 删除各表的 page_id 列与 page 表。
//
// 注意：第 4 步为不可逆 DDL，上线前请先备份数据库。
func MigratePageLayerToTemplates() {
	migrator := utils.DB.Migrator()
	hasPageTable := migrator.HasTable("page")
	legacyTables := []string{"column", "ad", "link", "article_column_publish"}

	hasLegacyColumn := false
	for _, table := range legacyTables {
		if migrator.HasColumn(table, "page_id") {
			hasLegacyColumn = true
			break
		}
	}
	if !hasPageTable && !hasLegacyColumn {
		return // 已完成迁移
	}

	// 1. 读取页面（含已软删除的页面，供第 3 步回填；软删除页面不补建模板）
	type legacyPage struct {
		ID          uint
		Name        string
		Code        string
		PageType    string
		RoutePath   string
		Description string
		Status      int
		TemplateID  uint
	}
	var allPages []legacyPage
	if hasPageTable {
		if err := utils.DB.Table("page").Find(&allPages).Error; err != nil {
			utils.Logger.Errorf("读取 page 表失败，已跳过页面→模板合并迁移: %v", err)
			return
		}
	}

	createdTemplates := 0
	deletedPagesSkipped := 0
	for i := range allPages {
		page := &allPages[i]
		if page.TemplateID > 0 {
			var cnt int64
			if err := utils.DB.Model(&Template{}).Where("id = ?", page.TemplateID).Count(&cnt).Error; err == nil && cnt > 0 {
				continue
			}
		}
		// 软删除的页面不再补建模板（其关联数据视为历史脏数据，template_id 保持 0）
		var deletedCnt int64
		if err := utils.DB.Table("page").Where("id = ? AND deleted_at IS NOT NULL", page.ID).Count(&deletedCnt).Error; err == nil && deletedCnt > 0 {
			deletedPagesSkipped++
			continue
		}
		tpl := Template{
			Name:        page.Name,
			Code:        page.Code,
			Type:        page.PageType,
			RoutePath:   page.RoutePath,
			Description: page.Description,
			Status:      page.Status,
		}
		if err := utils.DB.Create(&tpl).Error; err != nil {
			utils.Logger.Warnf("为页面[%d:%s]补建模板失败: %v", page.ID, page.Name, err)
			continue
		}
		if err := utils.DB.Table("page").Where("id = ?", page.ID).Update("template_id", tpl.ID).Error; err != nil {
			utils.Logger.Warnf("回填页面[%d]的 template_id 失败: %v", page.ID, err)
			continue
		}
		page.TemplateID = tpl.ID
		createdTemplates++
	}

	// 2. 页面上的 code / route_path 是路由与静态化的取值来源，回填到模板
	filled := 0
	for i := range allPages {
		page := &allPages[i]
		if page.TemplateID == 0 {
			continue
		}
		if page.Code != "" {
			res := utils.DB.Exec("UPDATE template SET code = ? WHERE id = ? AND (code IS NULL OR code = '')", page.Code, page.TemplateID)
			filled += int(res.RowsAffected)
		}
		if page.RoutePath != "" {
			res := utils.DB.Exec("UPDATE template SET route_path = ? WHERE id = ? AND (route_path IS NULL OR route_path = '')", page.RoutePath, page.TemplateID)
			filled += int(res.RowsAffected)
		}
	}

	if createdTemplates > 0 {
		utils.Logger.Infof("页面→模板合并：为 %d 个未绑定模板的页面补建了模板", createdTemplates)
	}
	if deletedPagesSkipped > 0 {
		utils.Logger.Warnf("页面→模板合并：%d 个已软删除页面未补建模板，其关联栏目/广告/友链的 template_id 为 0，需人工处理", deletedPagesSkipped)
	}

	// 3 & 4. 回填引用 + 删列删表（同一连接内关闭外键检查，避免约束阻塞 DDL）
	err := utils.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
			return err
		}
		if hasPageTable {
			backfills := []string{
				"UPDATE `column` c JOIN `page` p ON c.page_id = p.id SET c.template_id = p.template_id",
				"UPDATE `ad` a JOIN `page` p ON a.page_id = p.id SET a.template_id = p.template_id",
				"UPDATE `link` l JOIN `page` p ON l.page_id = p.id SET l.template_id = p.template_id",
				"UPDATE `article_column_publish` acp JOIN `page` p ON acp.page_id = p.id SET acp.template_id = p.template_id",
			}
			for _, stmt := range backfills {
				if err := tx.Exec(stmt).Error; err != nil {
					return fmt.Errorf("%s: %w", stmt, err)
				}
			}
		}
		for _, table := range legacyTables {
			if !migrator.HasColumn(table, "page_id") {
				continue
			}
			stmt := fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN page_id", table)
			if err := tx.Exec(stmt).Error; err != nil {
				return fmt.Errorf("%s: %w", stmt, err)
			}
		}
		if hasPageTable {
			if err := tx.Exec("DROP TABLE `page`").Error; err != nil {
				return fmt.Errorf("DROP TABLE `page`: %w", err)
			}
		}
		return tx.Exec("SET FOREIGN_KEY_CHECKS = 1").Error
	})
	if err != nil {
		utils.Logger.Errorf("页面→模板合并迁移失败（page 表与 page_id 列可能只完成了一部分，请检查后重启）: %v", err)
		return
	}
	utils.Logger.Infof("页面→模板合并迁移完成：page 表已移除，column/ad/link/article_column_publish 改用 template_id（模板回填字段 %d 处）", filled)
}
