package services

import (
	"context"
	"errors"
	"fmt"
	"server/models"
	"server/utils"
	"slices"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ArticleService struct{}

func (s *ArticleService) buildArticleListQuery(title string, categoryID int, tagID int, columnID int, status int, auditStatus int, articleType int, author string, authorCode string, source string) *gorm.DB {
	query := utils.DB.Model(&models.Article{})
	// authorCode：归属过滤（非管理员只能看到自己的文章），author：按作者名搜索
	if authorCode != "" {
		query = query.Where("article.author_code = ?", authorCode)
	}
	if title != "" {
		query = query.Where("MATCH(title) AGAINST (? IN BOOLEAN MODE) OR title LIKE ?", title, "%"+title+"%")
	}
	if categoryID > 0 {
		query = query.Where("article.id IN (SELECT article_id FROM article_category WHERE category_id = ?)", categoryID)
	}
	if tagID > 0 {
		query = query.Where("article.id IN (SELECT article_id FROM article_tag WHERE tag_id = ?)", tagID)
	}
	if columnID > 0 {
		query = query.Where("article.id IN (SELECT article_id FROM article_column WHERE column_id = ?)", columnID)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	if auditStatus >= 0 {
		query = query.Where("audit_status = ?", auditStatus)
	}
	if articleType > 0 {
		query = query.Where("type = ?", articleType)
	}
	if author != "" {
		query = query.Where("MATCH(author) AGAINST (? IN BOOLEAN MODE) OR author LIKE ?", author, "%"+author+"%")
	}
	if source != "" {
		query = query.Where("MATCH(source) AGAINST (? IN BOOLEAN MODE) OR source LIKE ?", source, "%"+source+"%")
	}
	return query
}

func (s *ArticleService) GetArticles(title string, categoryID int, tagID int, columnID int, status int, auditStatus int, articleType int, author string, authorCode string, source string, page int, pageSize int) ([]models.Article, int64, error) {
	var total int64
	countQuery := s.buildArticleListQuery(title, categoryID, tagID, columnID, status, auditStatus, articleType, author, authorCode, source)
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	idQuery := s.buildArticleListQuery(title, categoryID, tagID, columnID, status, auditStatus, articleType, author, authorCode, source)
	var ids []uint
	if err := idQuery.Select("article.id").Order("article.is_top DESC, article.created_at DESC").Limit(pageSize).Offset(offset).Scan(&ids).Error; err != nil {
		return nil, 0, err
	}

	if len(ids) == 0 {
		return []models.Article{}, total, nil
	}

	var articles []models.Article
	// 列表不返回正文（longtext）：Omit 避免逐行加载大字段，详情接口再单独加载
	err := utils.DB.Omit("Content").
		Where("id IN ?", ids).
		Order("is_top DESC, created_at DESC").
		Preload("Categories").Preload("Tags").Preload("Columns").Preload("Attachments").
		Find(&articles).Error
	return articles, total, err
}

// SearchPublishedArticles 开放搜索（无需认证）：仅返回已发布文章（status=1）
// 支持标题模糊搜索及分类/标签/类型/作者/来源过滤，带分页；不返回 content 字段
func (s *ArticleService) SearchPublishedArticles(title string, categoryID int, tagID int, articleType int, author string, source string, page int, pageSize int) ([]models.Article, int64, error) {
	query := utils.DB.Model(&models.Article{}).Where("status = ?", 1)
	if title != "" {
		query = query.Where("MATCH(title) AGAINST (? IN BOOLEAN MODE) OR title LIKE ?", title, "%"+title+"%")
	}
	if categoryID > 0 {
		query = query.Where("article.id IN (SELECT article_id FROM article_category WHERE category_id = ?)", categoryID)
	}
	if tagID > 0 {
		query = query.Where("article.id IN (SELECT article_id FROM article_tag WHERE tag_id = ?)", tagID)
	}
	if articleType > 0 {
		query = query.Where("type = ?", articleType)
	}
	if author != "" {
		query = query.Where("MATCH(author) AGAINST (? IN BOOLEAN MODE) OR author LIKE ?", author, "%"+author+"%")
	}
	if source != "" {
		query = query.Where("MATCH(source) AGAINST (? IN BOOLEAN MODE) OR source LIKE ?", source, "%"+source+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var ids []uint
	if err := query.Select("article.id").
		Order("article.is_top DESC, article.created_at DESC").
		Limit(pageSize).Offset(offset).
		Scan(&ids).Error; err != nil {
		return nil, 0, err
	}

	if len(ids) == 0 {
		return []models.Article{}, total, nil
	}

	var articles []models.Article
	// 公开搜索不返回正文：Omit Content 避免加载 longtext 大字段
	err := utils.DB.Omit("Content").
		Where("id IN ?", ids).
		Order("is_top DESC, created_at DESC").
		Preload("Categories").Preload("Tags").Preload("Columns").
		Find(&articles).Error
	return articles, total, err
}

func (s *ArticleService) GetArticleByID(id uint) (*models.Article, error) {
	var article models.Article
	err := utils.DB.Preload("Categories").Preload("Tags").Preload("Columns").Preload("Attachments").First(&article, id).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (s *ArticleService) CreateArticle(article *models.Article, tagIDs []uint, categoryIDs []uint) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		// 发布/审核状态只能由审核流程或状态接口变更，忽略请求体携带的值，
		// 防止作者通过 POST /articles {"status":1} 直接发布、绕过审核。
		article.Status = models.ArticleStatusDraft
		article.AuditStatus = 0
		// 类型白名单：未传/非法值统一按「图文」处理（类型决定编辑弹窗与可选栏目）
		if !models.IsValidArticleType(article.Type) {
			article.Type = models.ArticleTypeGraphic
		}
		// 先暂存附件，避免 GORM Create 自动关联插入导致重复
		attachments := article.Attachments
		article.Attachments = nil

		if err := tx.Create(article).Error; err != nil {
			return err
		}
		if len(categoryIDs) > 0 {
			var categories []models.Category
			for _, id := range categoryIDs {
				categories = append(categories, models.Category{ID: id})
			}
			if err := tx.Model(article).Association("Categories").Append(&categories); err != nil {
				return err
			}
		}
		if len(tagIDs) > 0 {
			var tags []models.Tag
			for _, id := range tagIDs {
				tags = append(tags, models.Tag{ID: id})
			}
			if err := tx.Model(article).Association("Tags").Append(&tags); err != nil {
				return err
			}
		}
		if len(attachments) > 0 {
			seen := make(map[string]bool)
			unique := make([]models.ArticleAttachment, 0, len(attachments))
			for _, att := range attachments {
				if att.URL == "" || seen[att.URL] {
					continue
				}
				seen[att.URL] = true
				att.ArticleID = article.ID
				att.ID = 0
				unique = append(unique, att)
			}
			if len(unique) > 0 {
				if err := tx.Create(&unique).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (s *ArticleService) UpdateArticle(id uint, article *models.Article, tagIDs []uint, categoryIDs []uint) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		var old models.Article
		if err := tx.First(&old, id).Error; err != nil {
			return err
		}
		// 审核中/已发布（audit_status != 0）的文章不允许通过编辑接口改内容：
		// 否则作者可在审核人查看后替换正文（上线的是未审内容），或对已发布文章做无需审核的修改。
		// 已下线文章（takeOffline 已把 audit_status 归零）不受此限制。
		if old.AuditStatus != 0 && old.Status != models.ArticleStatusOffline {
			return fmt.Errorf("文章正在审核或已发布，不能编辑；如需修改请先撤回审核或将文章下线")
		}
		// 发布/审核状态不由本接口（编辑内容）变更，防止作者用 PUT {"status":1} 绕过审核：
		//   - 文章当前为「已下线」：编辑动作将其转回草稿，并清理栏目/审核/发布等关联数据；
		//   - 其余情况：一律沿用数据库中的当前值（请求体中的 status/auditStatus 被忽略）。
		isOffline := old.Status == models.ArticleStatusOffline
		if isOffline {
			article.Status = models.ArticleStatusDraft
			article.AuditStatus = 0
		} else {
			article.Status = old.Status
			article.AuditStatus = old.AuditStatus
		}
		updates := map[string]any{
			"title":         article.Title,
			"summary":       article.Summary,
			"content":       article.Content,
			"status":        article.Status,
			"audit_status":  article.AuditStatus,
			"is_top":        article.IsTop,
			"is_bold":       article.IsBold,
			"default_color": article.DefaultColor,
			"cover":         article.Cover,
			"author":        article.Author,
			"author_code":   article.AuthorCode,
			"source":        article.Source,
			"publish_time":  article.PublishTime,
			"url":           article.URL,
		}
		// 类型：仅接受白名单内的值，非法/未传时保持原值（前端各类编辑弹窗均会显式携带自己的类型）
		if models.IsValidArticleType(article.Type) {
			updates["type"] = article.Type
		}
		if err := tx.Model(&old).Updates(updates).Error; err != nil {
			return err
		}
		// 已下线文章清理关联数据（与「下线」接口共用 takeOffline，保证两条路径状态一致）
		if isOffline {
			if err := takeOffline(tx, &old); err != nil {
				return err
			}
		}
		if len(categoryIDs) > 0 {
			var categories []models.Category
			for _, cid := range categoryIDs {
				categories = append(categories, models.Category{ID: cid})
			}
			if err := tx.Model(&old).Association("Categories").Replace(&categories); err != nil {
				return err
			}
		} else {
			if err := tx.Model(&old).Association("Categories").Clear(); err != nil {
				return err
			}
		}
		if len(tagIDs) > 0 {
			var tags []models.Tag
			for _, tid := range tagIDs {
				tags = append(tags, models.Tag{ID: tid})
			}
			if err := tx.Model(&old).Association("Tags").Replace(&tags); err != nil {
				return err
			}
		} else {
			if err := tx.Model(&old).Association("Tags").Clear(); err != nil {
				return err
			}
		}
		// 更新附件：删除旧附件，创建新附件（去重）
		if err := tx.Where("article_id = ?", id).Delete(&models.ArticleAttachment{}).Error; err != nil {
			return err
		}
		if len(article.Attachments) > 0 {
			seen := make(map[string]bool)
			unique := make([]models.ArticleAttachment, 0, len(article.Attachments))
			for _, att := range article.Attachments {
				if att.URL == "" || seen[att.URL] {
					continue
				}
				seen[att.URL] = true
				att.ArticleID = id
				att.ID = 0
				unique = append(unique, att)
			}
			if len(unique) > 0 {
				if err := tx.Create(&unique).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// takeOffline 文章下线时统一清理发布/审核/栏目关联数据：
// 下线后不应再残留 article_column_publish（否则列表仍会按栏目命中）、进行中的栏目审核，
// 以及栏目绑定（否则重新上线时与实际栏目不符）。审核状态重置为未提交，便于作者修改后重新送审。
func takeOffline(tx *gorm.DB, article *models.Article) error {
	if err := tx.Model(article).Association("Columns").Clear(); err != nil {
		return err
	}
	if err := tx.Where("article_id = ?", article.ID).Delete(&models.ArticleColumnAudit{}).Error; err != nil {
		return err
	}
	if err := tx.Where("article_id = ?", article.ID).Delete(&models.ArticleColumnAuditHistory{}).Error; err != nil {
		return err
	}
	if err := tx.Where("article_id = ?", article.ID).Delete(&models.ArticleColumnPublish{}).Error; err != nil {
		return err
	}
	return tx.Model(article).Update("audit_status", 0).Error
}

func (s *ArticleService) UpdateArticleStatus(id uint, status int) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		var article models.Article
		if err := tx.First(&article, id).Error; err != nil {
			return err
		}
		if err := tx.Model(&article).Update("status", status).Error; err != nil {
			return err
		}
		// 草稿(0)/下线(2) 都代表不再对外发布，统一清理发布/审核/栏目关联：
		// 否则会出现 (status=0, audit_status=2) 这类非法组合——前端显示「已审核」但文章是草稿，
		// 且栏目绑定/发布记录残留、列表仍会按栏目命中。
		if status != models.ArticleStatusPublished {
			return takeOffline(tx, &article)
		}
		return nil
	})
}

func (s *ArticleService) SetArticleColumns(id uint, columnIDs []uint) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		var article models.Article
		if err := tx.First(&article, id).Error; err != nil {
			return err
		}
		// 审核中/已审核（已发布）时禁止改栏目：
		// 否则会残留已解绑栏目的审核记录、新栏目无审核记录，后续“全部通过”时会错误发布。
		if article.AuditStatus != 0 {
			return fmt.Errorf("文章正在审核或已发布，不能修改栏目；如需调整请先撤回审核或将文章下线")
		}
		if len(columnIDs) > 0 {
			// 先去重（保持顺序）：前端下拉只会给启用栏目，但直调 API 可传任意 ID；
			// 关联表 FK 只保证「存在」，禁用栏目不应再被投放（否则会排到静态化列表里）。
			unique := make([]uint, 0, len(columnIDs))
			seen := make(map[uint]bool, len(columnIDs))
			for _, cid := range columnIDs {
				if cid == 0 || seen[cid] {
					continue
				}
				seen[cid] = true
				unique = append(unique, cid)
			}
			if len(unique) == 0 {
				return fmt.Errorf("请选择至少一个有效栏目")
			}
			var columns []models.Column
			if err := tx.Where("id IN ?", unique).Find(&columns).Error; err != nil {
				return err
			}
			if len(columns) != len(unique) {
				return fmt.Errorf("所选栏目不存在或已被删除，请刷新后重试")
			}
			for i := range columns {
				if columns[i].Status != 1 {
					return fmt.Errorf("栏目「%s」已禁用，不能投放文章", columns[i].Name)
				}
			}
			toBind := make([]models.Column, 0, len(unique))
			for _, cid := range unique {
				toBind = append(toBind, models.Column{ID: cid})
			}
			if err := tx.Model(&article).Association("Columns").Replace(&toBind); err != nil {
				return err
			}
		} else {
			if err := tx.Model(&article).Association("Columns").Clear(); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *ArticleService) DeleteArticle(id uint) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		var article models.Article
		if err := tx.First(&article, id).Error; err != nil {
			return err
		}
		if err := tx.Model(&article).Association("Categories").Clear(); err != nil {
			return err
		}
		if err := tx.Model(&article).Association("Tags").Clear(); err != nil {
			return err
		}
		if err := tx.Model(&article).Association("Columns").Clear(); err != nil {
			return err
		}
		if err := tx.Where("article_id = ?", id).Delete(&models.ArticleColumnAudit{}).Error; err != nil {
			return err
		}
		if err := tx.Where("article_id = ?", id).Delete(&models.ArticleColumnAuditHistory{}).Error; err != nil {
			return err
		}
		if err := tx.Where("article_id = ?", id).Delete(&models.ArticleColumnPublish{}).Error; err != nil {
			return err
		}
		if err := tx.Where("article_id = ?", id).Delete(&models.ArticleAttachment{}).Error; err != nil {
			return err
		}
		return tx.Delete(&article).Error
	})
}

func (s *ArticleService) GetArticleCountByAuthor(authorCode string) int64 {
	var count int64
	utils.DB.Model(&models.Article{}).Where("author_code = ?", authorCode).Count(&count)
	return count
}

// GetArticleCount 全站「已发布」文章数。
// 口径与其它文章统计一致（只计 status=1）：原先无任何状态过滤，把草稿/已下线也算进去，
// 与仪表盘卡片「已发布文章」的语义不符。
func (s *ArticleService) GetArticleCount() int64 {
	var count int64
	utils.DB.Model(&models.Article{}).Where("status = ?", models.ArticleStatusPublished).Count(&count)
	return count
}

type ArticleAuthorStat struct {
	Author     string `json:"author"`
	AuthorCode string `json:"authorCode"`
	Count      int64  `json:"count"`
}

func (s *ArticleService) GetArticleAuthorStats(period string, authorCode string) ([]ArticleAuthorStat, int64, error) {
	var results []ArticleAuthorStat
	var total int64
	now := time.Now()
	loc := now.Location()

	query := utils.DB.Model(&models.Article{}).Where("status = ?", 1)
	// 归属过滤：与文章列表（GetArticles）同一口径，非管理员只能统计自己的文章
	if authorCode != "" {
		query = query.Where("author_code = ?", authorCode)
	}

	switch period {
	case "week":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		monday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, -(weekday - 1))
		nextMonday := monday.AddDate(0, 0, 7)
		query = query.Where("created_at >= ? AND created_at < ?", monday, nextMonday)
	case "month":
		startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		endOfMonth := startOfMonth.AddDate(0, 1, 0)
		query = query.Where("created_at >= ? AND created_at < ?", startOfMonth, endOfMonth)
	case "year":
		startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, loc)
		endOfYear := startOfYear.AddDate(1, 0, 0)
		query = query.Where("created_at >= ? AND created_at < ?", startOfYear, endOfYear)
	default:
		return nil, 0, fmt.Errorf("无效的 period 参数")
	}

	err := query.Select("author, author_code, COUNT(*) as count").
		Group("author, author_code").
		Order("count DESC, author ASC").
		Scan(&results).Error
	if err != nil {
		return nil, 0, err
	}

	for _, r := range results {
		total += r.Count
	}

	return results, total, nil
}

// WithdrawArticleAudit 撤回文章审核
func (s *ArticleService) WithdrawArticleAudit(articleID uint) error {
	var article models.Article
	if err := utils.DB.First(&article, articleID).Error; err != nil {
		return err
	}
	if article.AuditStatus != 1 {
		return fmt.Errorf("只有审核中的文章可以撤回审核")
	}
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&article).Update("audit_status", 0).Error; err != nil {
			return err
		}
		// 只清「当前审核记录」，与 StartArticleAudit 同口径保留审核历史：
		// 撤回后重新送审时，之前各节点的通过/驳回记录应仍可追溯（历史接口按轮次累积展示）。
		if err := tx.Where("article_id = ?", articleID).Delete(&models.ArticleColumnAudit{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// StartArticleAudit 提交文章审核，为每个绑定了工作流的栏目创建审核记录
func (s *ArticleService) StartArticleAudit(articleID uint) error {
	var article models.Article
	if err := utils.DB.Preload("Columns").First(&article, articleID).Error; err != nil {
		return err
	}
	if len(article.Columns) == 0 {
		return fmt.Errorf("请先为文章选择栏目后再提交审核")
	}
	// 仅草稿可以提交审核：已发布（1）无需再送审，已下线（2）需先编辑转草稿，
	// 否则可通过审核流程把下线/已发布文章直接推到发布态。
	if article.Status != models.ArticleStatusDraft {
		return fmt.Errorf("仅草稿状态的文章可以提交审核，请先编辑文章将其转为草稿")
	}
	// 审核中不允许重复提交（需先撤回），避免静默清空当前审核进度
	if article.AuditStatus == 1 {
		return fmt.Errorf("文章正在审核中，请先撤回审核后再重新提交")
	}
	err := utils.DB.Transaction(func(tx *gorm.DB) error {
		// 更新文章状态为审核中
		if err := tx.Model(&article).Update("audit_status", 1).Error; err != nil {
			return err
		}
		// 清除旧的「当前审核记录」（按栏目重建），但**保留审核历史**：
		// 历史接口按 (article_id, column_id) 正序返回，多轮送审的记录会自然累积，便于追溯每一轮的驳回/通过。
		if err := tx.Where("article_id = ?", articleID).Delete(&models.ArticleColumnAudit{}).Error; err != nil {
			return err
		}
		// 为每个栏目创建审核记录：有流程的走审核，无流程的直接通过
		for _, col := range article.Columns {
			if col.WorkflowID == nil || *col.WorkflowID == 0 {
				// 未配置流程的栏目直接通过
				audit := models.ArticleColumnAudit{
					ArticleID:     articleID,
					ColumnID:      col.ID,
					WorkflowID:    0,
					CurrentNodeID: 0,
					Status:        1,
				}
				if err := tx.Create(&audit).Error; err != nil {
					return err
				}
				continue
			}
			// 绑定了审核流程的栏目必须先校验流程节点可用（节点存在且审批人可解析）：
			// 流程无节点或审批人为空都会让审核永久卡住，这里直接返回可定位的原因。
			workflowService := WorkflowService{}
			if err := workflowService.ValidateWorkflowResolvable(*col.WorkflowID); err != nil {
				// 用 %w 保留错误链：否则 controller 的 SanitizeError 无法识别 DB 错误，会把 SQL/表名透传给前端
				return fmt.Errorf("栏目「%s」的审核流程不可用: %w", col.Name, err)
			}
			// 「部门负责人」节点需作者已归属部门且该部门已配置负责人，否则无人可审
			if err := s.validateDeptHeadApprovers(*col.WorkflowID, article.AuthorCode); err != nil {
				return fmt.Errorf("栏目「%s」的审核流程不可用: %w", col.Name, err)
			}
			var firstNode models.WorkflowNode
			if err := tx.Where("workflow_id = ?", *col.WorkflowID).Order("sort_order ASC").First(&firstNode).Error; err != nil {
				return fmt.Errorf("栏目「%s」的审核流程没有可用节点", col.Name)
			}
			audit := models.ArticleColumnAudit{
				ArticleID:     articleID,
				ColumnID:      col.ID,
				WorkflowID:    *col.WorkflowID,
				CurrentNodeID: firstNode.ID,
				Status:        0,
			}
			if err := tx.Create(&audit).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	// 提交成功后若「全部栏目都免审」会立即发布；发布失败必须让调用方看到，
	// 否则作者以为已提交（界面停在「审核中」），实际永远不会上线。此时可先「撤回」再重试。
	if err := s.tryCompleteArticleAudit(articleID); err != nil {
		return fmt.Errorf("审核已提交，但文章自动发布失败: %w", err)
	}
	return nil
}

// GetArticleAuditProgress 获取文章在各栏目的审核进度
func (s *ArticleService) GetArticleAuditProgress(articleID uint) ([]models.ArticleColumnAudit, error) {
	var audits []models.ArticleColumnAudit
	err := utils.DB.Where("article_id = ?", articleID).Find(&audits).Error
	return audits, err
}

func (s *ArticleService) getUserName(userID uint) string {
	var user models.User
	if err := utils.DB.First(&user, userID).Error; err != nil {
		return ""
	}
	return user.Username
}

func parseIntIDs(s string) []int {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	ids := make([]int, 0, len(parts))
	for _, p := range parts {
		if id, err := strconv.Atoi(strings.TrimSpace(p)); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

func containsInt(ids []int, target int) bool {
	return slices.Contains(ids, target)
}

// getUserDepartmentIDs 返回用户所属的所有部门 ID
func (s *ArticleService) getUserDepartmentIDs(userID uint) ([]uint, error) {
	var departments []models.Department
	uidStr := strconv.Itoa(int(userID))
	err := utils.DB.Where("user_ids LIKE ? OR user_ids LIKE ? OR user_ids LIKE ?", "%"+uidStr+"%", "%"+uidStr+",%", "%,"+uidStr+"%").Find(&departments).Error
	if err != nil {
		return nil, err
	}
	uid := int(userID)
	var result []uint
	for _, dept := range departments {
		if containsInt(parseIntIDs(dept.UserIds), uid) {
			result = append(result, dept.ID)
		}
	}
	return result, nil
}

// canUserApproveNode 判断指定用户是否有权限审批当前节点
func (s *ArticleService) canUserApproveNode(node *models.WorkflowNode, userID uint, authorCode string) (bool, error) {
	switch node.ApproverType {
	case "role":
		// 未指定审批角色 = 无人可审，必须拒绝：否则任意登录用户（含作者本人）都能审批通过
		if node.ApproverID == 0 {
			return false, nil
		}
		workflowRoleService := WorkflowRoleService{}
		roleIDs, err := workflowRoleService.GetUserWorkflowRoleIds(userID)
		if err != nil {
			return false, err
		}
		if slices.Contains(roleIDs, node.ApproverID) {
			return true, nil
		}
		return false, nil
	case "dept_head":
		var currentUser models.User
		if err := utils.DB.First(&currentUser, userID).Error; err != nil {
			return false, err
		}
		authorID, err := strconv.ParseUint(authorCode, 10, 32)
		if err != nil {
			return false, nil
		}
		var author models.User
		if err := utils.DB.First(&author, authorID).Error; err != nil {
			return false, nil
		}
		authorDeptIDs, err := s.getUserDepartmentIDs(author.ID)
		if err != nil {
			return false, err
		}
		// leader_code 的当前口径是「用户 ID」（部门管理/机构管理均按下拉选中用户写入）；
		// 机构管理页早期版本写入的是用户账号，这里兼容两种取值，避免历史数据导致审批人永远匹配不上。
		leaderCodes := []string{strconv.FormatUint(uint64(currentUser.ID), 10)}
		if currentUser.Account != "" {
			leaderCodes = append(leaderCodes, currentUser.Account)
		}
		var headDepartments []models.Department
		if err := utils.DB.Where("leader_code IN ?", leaderCodes).Find(&headDepartments).Error; err != nil {
			return false, err
		}
		for _, hd := range headDepartments {
			if slices.Contains(authorDeptIDs, hd.ID) {
				return true, nil
			}
		}
		return false, nil
	case "user", "":
		// 未指定审批人一律拒绝（原因同上），仅显式指定为当前用户时放行
		if node.ApproverID != 0 && node.ApproverID == userID {
			return true, nil
		}
		return false, nil
	default:
		return false, fmt.Errorf("未知的审批人类型: %s", node.ApproverType)
	}
}

// CanApproveArticleColumn 判断指定用户是否能审批文章指定栏目的当前节点
func (s *ArticleService) CanApproveArticleColumn(articleID uint, columnID uint, userID uint) (bool, error) {
	var audit models.ArticleColumnAudit
	if err := utils.DB.Where("article_id = ? AND column_id = ? AND status = ?", articleID, columnID, 0).First(&audit).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	var currentNode models.WorkflowNode
	if err := utils.DB.First(&currentNode, audit.CurrentNodeID).Error; err != nil {
		return false, err
	}
	var article models.Article
	if err := utils.DB.First(&article, articleID).Error; err != nil {
		return false, err
	}
	return s.canUserApproveNode(&currentNode, userID, article.AuthorCode)
}

// GetApproverName 根据节点和文章作者解析审批人显示名称（调试用）
func (s *ArticleService) GetApproverName(node *models.WorkflowNode, authorCode string) string {
	if node == nil {
		return ""
	}
	approverType := node.ApproverType
	if approverType == "" {
		approverType = "user"
	}
	switch approverType {
	case "user":
		if node.ApproverID == 0 {
			return "指定成员（未指定）"
		}
		var user models.User
		if err := utils.DB.First(&user, node.ApproverID).Error; err != nil {
			return "指定成员（未知）"
		}
		return user.Username
	case "role":
		if node.ApproverID == 0 {
			return "流程角色（未指定）"
		}
		var role models.WorkflowRole
		if err := utils.DB.First(&role, node.ApproverID).Error; err != nil {
			return "流程角色（未知）"
		}
		return role.Name
	case "dept_head":
		authorID, err := strconv.ParseUint(authorCode, 10, 32)
		if err != nil {
			return "部门负责人（作者ID无效）"
		}
		var author models.User
		if err := utils.DB.First(&author, authorID).Error; err != nil {
			return "部门负责人（作者未知）"
		}
		deptIDs, err := s.getUserDepartmentIDs(author.ID)
		if err != nil || len(deptIDs) == 0 {
			return "部门负责人（未找到作者部门）"
		}
		var departments []models.Department
		if err := utils.DB.Where("id IN ?", deptIDs).Find(&departments).Error; err != nil {
			return "部门负责人（查询失败）"
		}
		names := make([]string, 0, len(departments))
		for _, d := range departments {
			if d.Leader != "" {
				names = append(names, d.Leader)
			}
		}
		if len(names) == 0 {
			return "部门负责人（未设置负责人）"
		}
		return strings.Join(names, "、")
	default:
		return ""
	}
}

// AdvanceArticleAudit 推进指定文章栏目的审核到下一节点
func (s *ArticleService) AdvanceArticleAudit(articleID uint, columnID uint, userID uint, remark string) error {
	var audit models.ArticleColumnAudit
	if err := utils.DB.Where("article_id = ? AND column_id = ?", articleID, columnID).First(&audit).Error; err != nil {
		return err
	}
	if audit.Status != 0 {
		return nil // 已结束
	}
	// 查找当前节点并校验权限
	var currentNode models.WorkflowNode
	if err := utils.DB.First(&currentNode, audit.CurrentNodeID).Error; err != nil {
		return err
	}
	var article models.Article
	if err := utils.DB.First(&article, articleID).Error; err != nil {
		return err
	}
	// 只有「审核中」(audit_status=1) 的文章可以推进：被驳回后文章回到未提交态（audit_status=0）并清空发布记录，
	// 此时同一文章其它栏目残留的待审记录不应再被推进（否则界面显示「待审核」却还能继续审批）。
	if article.AuditStatus != 1 {
		return fmt.Errorf("该文章当前不在审核中，无法推进审核，请刷新后重试")
	}
	canApprove, err := s.canUserApproveNode(&currentNode, userID, article.AuthorCode)
	if err != nil {
		return err
	}
	if !canApprove {
		return fmt.Errorf("当前节点审批人不是您，无权操作")
	}
	userName := s.getUserName(userID)
	// 记录当前节点的通过历史
	history := models.ArticleColumnAuditHistory{
		ArticleID:    articleID,
		ColumnID:     columnID,
		WorkflowID:   audit.WorkflowID,
		NodeID:       currentNode.ID,
		NodeName:     currentNode.Name,
		ApproverType: currentNode.ApproverType,
		Action:       1,
		OperatorID:   userID,
		OperatorName: userName,
		Remark:       remark,
	}
	// 查找下一个节点：按 (sort_order, id) 全序取当前节点的紧接着的一条。
	// 原实现用 `sort_order > 当前节点的 sort_order`，一旦存在排序值重复的节点（历史数据），
	// 会「找不到下一个节点」→ 直接标记为审核通过并触发自动发布，跳过后面的审批人。
	var nodes []models.WorkflowNode
	if err := utils.DB.Where("workflow_id = ?", audit.WorkflowID).
		Order("sort_order ASC, id ASC").Find(&nodes).Error; err != nil {
		return err
	}
	var nextNode models.WorkflowNode
	hasNext := false
	found := false
	for i := range nodes {
		if nodes[i].ID != currentNode.ID {
			continue
		}
		found = true
		if i+1 < len(nodes) {
			nextNode = nodes[i+1]
			hasNext = true
		}
		break
	}
	if !found {
		// 当前节点已不属于该流程（例如节点被重建/删除）：宁可报错也不能静默判定审核通过
		return fmt.Errorf("当前审核节点已失效，请联系管理员重新配置该流程的审批节点")
	}

	// 状态变更与历史记录放在同一事务，并用「条件更新」保证并发/重复点击时只有一个请求生效：
	// WHERE status = 0（推进时再限定 current_node_id），受影响行数为 0 说明该节点已被处理。
	// 否则同一节点会被推进两次、写入重复审核历史（原先先写历史再改状态，第一步失败还会留下分叉）。
	where, args := "id = ? AND status = 0", []any{audit.ID}
	updates := map[string]any{"approve_remark": remark}
	if hasNext {
		where += " AND current_node_id = ?"
		args = append(args, currentNode.ID)
		updates["current_node_id"] = nextNode.ID
	} else {
		// 没有下一个节点：标记为已通过，记录通过人信息
		updates["status"] = 1
		updates["current_node_id"] = 0
		updates["approve_user_id"] = userID
		updates["approve_user_name"] = userName
		updates["approve_time"] = time.Now()
	}

	err = utils.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&models.ArticleColumnAudit{}).Where(where, args...).Updates(updates)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("该审核节点已被处理，请刷新后重试")
		}
		return tx.Create(&history).Error
	})
	if err != nil {
		return err
	}
	if !hasNext {
		// 末节点通过后立即尝试发布：失败必须上报，否则审批人看到「已通过」而文章永不发布。
		if err := s.tryCompleteArticleAudit(articleID); err != nil {
			return fmt.Errorf("当前节点已通过，但文章发布失败，请联系管理员重试: %w", err)
		}
	}
	return nil
}

// RejectArticleAudit 驳回指定文章栏目的审核
func (s *ArticleService) RejectArticleAudit(articleID uint, columnID uint, userID uint, remark string) error {
	var audit models.ArticleColumnAudit
	if err := utils.DB.Where("article_id = ? AND column_id = ?", articleID, columnID).First(&audit).Error; err != nil {
		return err
	}
	if audit.Status != 0 {
		return nil
	}
	// 查找当前节点并校验权限
	var currentNode models.WorkflowNode
	if err := utils.DB.First(&currentNode, audit.CurrentNodeID).Error; err != nil {
		return err
	}
	var article models.Article
	if err := utils.DB.First(&article, articleID).Error; err != nil {
		return err
	}
	// 只有「审核中」(audit_status=1) 的文章可以驳回（口径同 AdvanceArticleAudit）
	if article.AuditStatus != 1 {
		return fmt.Errorf("该文章当前不在审核中，无法驳回，请刷新后重试")
	}
	canApprove, err := s.canUserApproveNode(&currentNode, userID, article.AuthorCode)
	if err != nil {
		return err
	}
	if !canApprove {
		return fmt.Errorf("当前节点审批人不是您，无权操作")
	}
	userName := s.getUserName(userID)
	// 记录驳回历史
	history := models.ArticleColumnAuditHistory{
		ArticleID:    articleID,
		ColumnID:     columnID,
		WorkflowID:   audit.WorkflowID,
		NodeID:       currentNode.ID,
		NodeName:     currentNode.Name,
		ApproverType: currentNode.ApproverType,
		Action:       2,
		OperatorID:   userID,
		OperatorName: userName,
		Remark:       remark,
	}
	if err := utils.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&models.ArticleColumnAudit{}).
			Where("id = ? AND status = 0", audit.ID).
			Updates(map[string]any{
				"status":        2,
				"reject_remark": remark,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil // 已被其他审批人处理，不再重复记录历史
		}
		return tx.Create(&history).Error
	}); err != nil {
		return err
	}
	// 驳回后可能触发「状态回退」（清发布记录 + audit_status 归零），失败同样需上报
	if err := s.tryCompleteArticleAudit(articleID); err != nil {
		return fmt.Errorf("驳回已记录，但文章状态回退失败: %w", err)
	}
	return nil
}

// validateDeptHeadApprovers 校验流程中的「部门负责人」节点对指定作者是否可解析：
// 作者需归属至少一个部门，且该部门配置了负责人（leader_code），否则审核将永远无法完成。
func (s *ArticleService) validateDeptHeadApprovers(workflowID uint, authorCode string) error {
	var count int64
	if err := utils.DB.Model(&models.WorkflowNode{}).
		Where("workflow_id = ? AND approver_type = ?", workflowID, "dept_head").
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return nil
	}
	authorID, err := strconv.ParseUint(authorCode, 10, 32)
	if err != nil {
		return fmt.Errorf("作者信息异常，无法解析部门负责人")
	}
	deptIDs, err := s.getUserDepartmentIDs(uint(authorID))
	if err != nil {
		return err
	}
	if len(deptIDs) == 0 {
		return fmt.Errorf("该流程包含「部门负责人」审批节点，但作者未归属任何部门")
	}
	var leaderCount int64
	if err := utils.DB.Model(&models.Department{}).
		Where("id IN ? AND leader_code IS NOT NULL AND leader_code <> ''", deptIDs).
		Count(&leaderCount).Error; err != nil {
		return err
	}
	if leaderCount == 0 {
		return fmt.Errorf("该流程包含「部门负责人」审批节点，但作者所属部门未设置负责人")
	}
	return nil
}

// CanApproveAnyColumn 判断用户是否为文章任一「进行中」栏目审核的当前审批人。
// 用于文章详情的可见性判断（非作者的管理者/审批人也需要查看正文）。
func (s *ArticleService) CanApproveAnyColumn(articleID, userID uint) (bool, error) {
	var audits []models.ArticleColumnAudit
	if err := utils.DB.Where("article_id = ? AND status = ?", articleID, 0).Find(&audits).Error; err != nil {
		return false, err
	}
	for _, audit := range audits {
		ok, err := s.CanApproveArticleColumn(articleID, audit.ColumnID, userID)
		if err != nil {
			// 不能把 DB 错误当成「无权限」：否则审批人会看不到本该由他审批的文章，且无任何提示
			utils.Logger.Warnf("判断文章[%d]栏目[%d]的审批权限失败: %s", articleID, audit.ColumnID, err)
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

// GetArticleAuditHistory 获取文章指定栏目的审核历史（**只返回当前一轮**）。
// 撤回/重新送审不再删除历史（便于审计留痕），但界面按「每个节点最新的处置」展示当前进度，
// 若把往轮记录一并下发，节点会显示上一轮的「已通过/已驳回」，与实际进度不符。
// 判定当前轮的起点：该 (article, column) 当前审核记录（article_column_audit）的创建时间——
// 撤回/重新送审会重建审核记录，因此早于它的历史属于上一轮。
// 若当前没有审核记录（如已撤回、已下线），则返回该栏目全部历史，供追溯参考。
func (s *ArticleService) GetArticleAuditHistory(articleID uint, columnID uint) ([]models.ArticleColumnAuditHistory, error) {
	query := utils.DB.Where("article_id = ? AND column_id = ?", articleID, columnID)
	var current models.ArticleColumnAudit
	if err := utils.DB.Where("article_id = ? AND column_id = ?", articleID, columnID).First(&current).Error; err == nil {
		query = query.Where("created_at >= ?", current.CreatedAt)
	}
	var histories []models.ArticleColumnAuditHistory
	err := query.Order("created_at ASC").Find(&histories).Error
	return histories, err
}

// tryCompleteArticleAudit 检查文章各栏目的审核流程是否都已结束：
//   - 仍有栏目在审核中：不做处理；
//   - 存在被驳回的栏目：不发布，回退为未提交状态，供作者修改后重新送审；
//   - 全部栏目通过：完成审核并发布。
//
// 返回值必须由调用方上抛：发布/回退失败若被吞掉，审批人会看到「已通过」而文章永不发布（或状态停在审核中），
// 前端也无从得知需要重试。
func (s *ArticleService) tryCompleteArticleAudit(articleID uint) error {
	var pendingCount, rejectedCount int64
	if err := utils.DB.Model(&models.ArticleColumnAudit{}).
		Where("article_id = ? AND status = ?", articleID, 0).
		Count(&pendingCount).Error; err != nil {
		return err
	}
	if pendingCount > 0 {
		return nil
	}
	if err := utils.DB.Model(&models.ArticleColumnAudit{}).
		Where("article_id = ? AND status = ?", articleID, 2).
		Count(&rejectedCount).Error; err != nil {
		return err
	}
	if rejectedCount > 0 {
		return s.rejectArticleAudit(articleID)
	}
	return s.CompleteArticleAudit(articleID)
}

// rejectArticleAudit 存在被驳回栏目时回退文章状态：不发布，标记为未提交（audit_status=0），
// 并清理可能残留的发布记录，供作者修改后重新送审（走「提交审核」）。
func (s *ArticleService) rejectArticleAudit(articleID uint) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Article{}).Where("id = ?", articleID).
			Update("audit_status", 0).Error; err != nil {
			return err
		}
		return tx.Where("article_id = ?", articleID).Delete(&models.ArticleColumnPublish{}).Error
	})
}

// ArticleColumnPublishItem 文章栏目发布列表项
type ArticleColumnPublishItem struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Source    string    `json:"source"`
	RoutePath string    `json:"routePath"`
}

// GetDetailPageRoutePath 获取详情页统一访问路径（取启用的详情页模板）
func (s *ArticleService) GetDetailPageRoutePath() (string, string, error) {
	type DetailTemplate struct {
		RoutePath string `gorm:"column:route_path"`
		Name      string `gorm:"column:name"`
	}
	var tpl DetailTemplate
	err := utils.DB.Model(&models.Template{}).
		Select("route_path,name").
		Where("type = ? AND status = ?", "detail", 1).
		Order("id ASC").
		Limit(1).
		Scan(&tpl).Error
	return tpl.RoutePath, tpl.Name, err
}

// GetArticleColumnPublishes 获取文章栏目发布（静态化）列表
func (s *ArticleService) GetArticleColumnPublishes(articleTitle string, columnID uint, page int, pageSize int) ([]ArticleColumnPublishItem, int64, error) {
	var items []ArticleColumnPublishItem
	var total int64
	query := utils.DB.Model(&models.Article{}).
		Select("article.id, article.created_at, article.updated_at, article.title, article.author, article.source").
		Where("article.id IN (SELECT DISTINCT article_id FROM article_column_publish where column_id in (SELECT id FROM `column` where display_type<>7))").
		Where("article.status = ?", 1).
		Order("article.updated_at DESC")
	if articleTitle != "" {
		query = query.Where("article.title LIKE ?", "%"+articleTitle+"%")
	}
	if columnID > 0 {
		query = query.Where("article.id IN (SELECT article_id FROM article_column_publish WHERE column_id = ?)", columnID)
	}
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err = query.Limit(pageSize).Offset(offset).Find(&items).Error
	return items, total, err
}

// GetMyAuditArticles 获取当前用户需要审核的文章列表。
// 语义说明：不排除「作者＝审批人」的情况——审批人由流程节点显式指定（user/role）或按部门负责人解析（dept_head），
// 若作者本人正好是节点审批人（例如部门负责人提交自己的文章），该文章应出现在其待办中；
// 是否能审批由 canUserApproveNode 统一判定，此处不再做额外过滤。
func (s *ArticleService) GetMyAuditArticles(userID uint, page, pageSize int) ([]models.Article, int64, error) {
	type auditItem struct {
		ArticleID    uint
		AuthorCode   string
		ApproverType string
		ApproverID   uint
	}
	var items []auditItem
	err := utils.DB.Table("article_column_audit aca").
		Select("aca.article_id, article.author_code, wn.approver_type, wn.approver_id").
		Joins("JOIN article ON article.id = aca.article_id").
		Joins("JOIN workflow_node wn ON wn.id = aca.current_node_id").
		Where("aca.status = ?", 0).
		Scan(&items).Error
	if err != nil {
		return nil, 0, err
	}

	seen := make(map[uint]bool)
	var allowedArticleIDs []uint
	for _, item := range items {
		node := &models.WorkflowNode{
			ApproverType: item.ApproverType,
			ApproverID:   item.ApproverID,
		}

		ok, err := s.canUserApproveNode(node, userID, item.AuthorCode)
		if err != nil {
			// 待办列表宁可报错也不能静默少显示（调用方会把错误返回给前端提示重试）
			utils.Logger.Warnf("判断用户[%d]对文章[%d]的审批权限失败: %s", userID, item.ArticleID, err)
			return nil, 0, err
		}
		if ok && !seen[item.ArticleID] {
			seen[item.ArticleID] = true
			allowedArticleIDs = append(allowedArticleIDs, item.ArticleID)
		}
	}

	total := int64(len(allowedArticleIDs))
	start := (page - 1) * pageSize
	if start >= len(allowedArticleIDs) {
		return []models.Article{}, total, nil
	}
	end := min(start+pageSize, len(allowedArticleIDs))
	pageIDs := allowedArticleIDs[start:end]

	var articles []models.Article
	if len(pageIDs) > 0 {
		err = utils.DB.Where("id IN ?", pageIDs).Order("created_at DESC").Find(&articles).Error
	}
	return articles, total, err
}

// CompleteArticleAudit 完成文章审核（所有栏目审核通过后调用）。
// 前置校验：必须存在栏目审核记录且全部为「已通过」，否则拒绝，避免管理员强制发布未审完的文章。
func (s *ArticleService) CompleteArticleAudit(articleID uint) error {
	err := utils.DB.Transaction(func(tx *gorm.DB) error {
		var totalAudits, passedAudits int64
		if err := tx.Model(&models.ArticleColumnAudit{}).Where("article_id = ?", articleID).Count(&totalAudits).Error; err != nil {
			return err
		}
		if totalAudits == 0 {
			return fmt.Errorf("该文章没有栏目审核记录，无法完成审核")
		}
		if err := tx.Model(&models.ArticleColumnAudit{}).
			Where("article_id = ? AND status = ?", articleID, 1).
			Count(&passedAudits).Error; err != nil {
			return err
		}
		if passedAudits != totalAudits {
			return fmt.Errorf("存在未完成的栏目审核，无法完成审核")
		}
		// 更新文章审核状态为已审核
		if err := tx.Model(&models.Article{}).Where("id = ?", articleID).Update("audit_status", 2).Error; err != nil {
			return err
		}
		// 获取文章基本信息
		var article models.Article
		if err := tx.First(&article, articleID).Error; err != nil {
			return err
		}
		// 清除该文章旧的发布记录
		if err := tx.Where("article_id = ?", articleID).Delete(&models.ArticleColumnPublish{}).Error; err != nil {
			return err
		}
		// 只查询审核通过的栏目记录
		var audits []models.ArticleColumnAudit
		if err := tx.Where("article_id = ? AND status = ?", articleID, 1).Find(&audits).Error; err != nil {
			return err
		}
		// 为每个通过的栏目创建发布记录
		for _, audit := range audits {
			var col models.Column
			// 只有「栏目确实不存在」才跳过；其它错误（DB 抖动/超时）必须中止事务，
			// 否则文章已置为已发布但 article_column_publish 缺行，栏目列表/静态化都会漏掉这篇文章。
			if err := tx.First(&col, audit.ColumnID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return err
			}
			publish := models.ArticleColumnPublish{
				TemplateID:   col.TemplateID,
				ColumnID:     audit.ColumnID,
				ArticleID:    article.ID,
				ArticleTitle: article.Title,
				Author:       article.Author,
				Source:       article.Source,
				IsTop:        article.IsTop,
				IsBold:       article.IsBold,
				Color:        article.DefaultColor,
			}
			if err := tx.Create(&publish).Error; err != nil {
				return err
			}
		}
		// 所有栏目都创建完成后统一更新文章状态（避免重复更新同一记录）
		if err := tx.Model(&article).Updates(map[string]any{"status": 1}).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}
	// 审核通过（完成审核/文章发布）时，调用静态化功能生成该文章的静态页（尽力而为，不影响主流程）
	// 静态化程序接口：POST /api/static/article?id={文章ID}&path={静态化输出路径}
	// 代理处理见 static_job_controller.go ArticleStatic / services.StaticJobService.GenerateArticleStaticByID
	NewStaticJobService().GenerateArticleStaticByID(context.Background(), strconv.FormatUint(uint64(articleID), 10))

	return nil
}
