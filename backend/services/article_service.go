package services

import (
	"fmt"
	"server/models"
	"server/utils"

	"gorm.io/gorm"
)

type ArticleService struct{}

func (s *ArticleService) GetArticles(title string, categoryID int, status int, auditStatus int, page int, pageSize int) ([]models.Article, int64, error) {
	var articles []models.Article
	var total int64
	query := utils.DB.Model(&models.Article{})
	if title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	if auditStatus >= 0 {
		query = query.Where("audit_status = ?", auditStatus)
	}
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err = query.Preload("Category").Preload("Tags").Preload("Columns").Order("is_top DESC, id DESC").Limit(pageSize).Offset(offset).Find(&articles).Error
	return articles, total, err
}

func (s *ArticleService) GetArticleByID(id uint) (*models.Article, error) {
	var article models.Article
	err := utils.DB.Preload("Category").Preload("Tags").Preload("Columns").First(&article, id).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (s *ArticleService) CreateArticle(article *models.Article, tagIDs []uint) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(article).Error; err != nil {
			return err
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
		return nil
	})
}

func (s *ArticleService) UpdateArticle(id uint, article *models.Article, tagIDs []uint) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		var old models.Article
		if err := tx.First(&old, id).Error; err != nil {
			return err
		}
		updates := map[string]interface{}{
			"title":         article.Title,
			"category_id":   article.CategoryID,
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
		}
		if err := tx.Model(&old).Updates(updates).Error; err != nil {
			return err
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
		return nil
	})
}

func (s *ArticleService) UpdateArticleStatus(id uint, status int) error {
	return utils.DB.Model(&models.Article{}).Where("id = ?", id).Update("status", status).Error
}

func (s *ArticleService) UpdateAuditStatus(id uint, auditStatus int) error {
	return utils.DB.Model(&models.Article{}).Where("id = ?", id).Update("audit_status", auditStatus).Error
}

func (s *ArticleService) SetArticleColumns(id uint, columnIDs []uint) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		var article models.Article
		if err := tx.First(&article, id).Error; err != nil {
			return err
		}
		if len(columnIDs) > 0 {
			var columns []models.Column
			for _, cid := range columnIDs {
				columns = append(columns, models.Column{ID: cid})
			}
			if err := tx.Model(&article).Association("Columns").Replace(&columns); err != nil {
				return err
			}
		} else {
			if err := tx.Model(&article).Association("Columns").Clear(); err != nil {
				return err
			}
		}
		if err := tx.Model(&article).Update("column_count", len(columnIDs)).Error; err != nil {
			return err
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
		if err := tx.Model(&article).Association("Tags").Clear(); err != nil {
			return err
		}
		if err := tx.Model(&article).Association("Columns").Clear(); err != nil {
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

func (s *ArticleService) GetArticleCount() int64 {
	var count int64
	utils.DB.Model(&models.Article{}).Count(&count)
	return count
}

// StartArticleAudit 提交文章审核，为每个绑定了工作流的栏目创建审核记录
func (s *ArticleService) StartArticleAudit(articleID uint) error {
	var article models.Article
	if err := utils.DB.Preload("Columns").First(&article, articleID).Error; err != nil {
		return err
	}
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		// 更新文章状态为审核中
		if err := tx.Model(&article).Update("audit_status", 1).Error; err != nil {
			return err
		}
		// 清除旧的审核记录
		if err := tx.Where("article_id = ?", articleID).Delete(&models.ArticleColumnAudit{}).Error; err != nil {
			return err
		}
		// 为每个绑定了工作流的栏目创建审核记录
		for _, col := range article.Columns {
			if col.WorkflowID == nil || *col.WorkflowID == 0 {
				continue
			}
			var firstNode models.WorkflowNode
			err := tx.Where("workflow_id = ?", *col.WorkflowID).Order("sort_order ASC").First(&firstNode).Error
			if err != nil {
				continue // 流程没有节点，跳过
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
}

// GetArticleAuditProgress 获取文章在各栏目的审核进度
func (s *ArticleService) GetArticleAuditProgress(articleID uint) ([]models.ArticleColumnAudit, error) {
	var audits []models.ArticleColumnAudit
	err := utils.DB.Where("article_id = ?", articleID).Find(&audits).Error
	return audits, err
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
	if currentNode.ApproverID != 0 && currentNode.ApproverID != userID {
		return fmt.Errorf("当前节点审批人不是您，无权操作")
	}
	// 查找下一个节点
	var nextNode models.WorkflowNode
	err := utils.DB.Where("workflow_id = ? AND sort_order > ?", audit.WorkflowID, currentNode.SortOrder).Order("sort_order ASC").First(&nextNode).Error
	if err != nil {
		// 没有下一个节点，标记为已通过
		return utils.DB.Model(&audit).Updates(map[string]interface{}{
			"status":          1,
			"current_node_id": 0,
			"approve_remark":  remark,
		}).Error
	}
	// 推进到下一个节点
	return utils.DB.Model(&audit).Updates(map[string]interface{}{
		"current_node_id": nextNode.ID,
		"approve_remark":  remark,
	}).Error
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
	if currentNode.ApproverID != 0 && currentNode.ApproverID != userID {
		return fmt.Errorf("当前节点审批人不是您，无权操作")
	}
	return utils.DB.Model(&audit).Updates(map[string]interface{}{
		"status":        2,
		"reject_remark": remark,
	}).Error
}

// CompleteArticleAudit 完成文章审核（所有栏目通过后调用）
func (s *ArticleService) CompleteArticleAudit(articleID uint) error {
	return utils.DB.Model(&models.Article{}).Where("id = ?", articleID).Update("audit_status", 2).Error
}
