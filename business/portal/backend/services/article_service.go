package services

import (
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

func (s *ArticleService) buildArticleListQuery(title string, categoryID int, tagID int, columnID int, status int, auditStatus int, articleType int, author string, source string) *gorm.DB {
	query := utils.DB.Model(&models.Article{})
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

func (s *ArticleService) GetArticles(title string, categoryID int, tagID int, columnID int, status int, auditStatus int, articleType int, author string, source string, page int, pageSize int) ([]models.Article, int64, error) {
	var total int64
	countQuery := s.buildArticleListQuery(title, categoryID, tagID, columnID, status, auditStatus, articleType, author, source)
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	idQuery := s.buildArticleListQuery(title, categoryID, tagID, columnID, status, auditStatus, articleType, author, source)
	var ids []uint
	if err := idQuery.Select("article.id").Order("article.is_top DESC, article.created_at DESC").Limit(pageSize).Offset(offset).Scan(&ids).Error; err != nil {
		return nil, 0, err
	}

	if len(ids) == 0 {
		return []models.Article{}, total, nil
	}

	var articles []models.Article
	err := utils.DB.Where("id IN ?", ids).
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
		// 已下线文章重新编辑时，转为草稿（编辑状态），并清理栏目、审核、发布等关联数据
		isOffline := old.Status == 2 || article.Status == 2
		if isOffline {
			article.Status = 0
			article.AuditStatus = 0
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
		if isOffline {
			updates["column_count"] = 0
		}
		if err := tx.Model(&old).Updates(updates).Error; err != nil {
			return err
		}
		// 已下线文章清理关联数据
		if isOffline {
			if err := tx.Model(&old).Association("Columns").Clear(); err != nil {
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
			if err := tx.Model(&old).Association("Categories").Clear(); err != nil {
				return err
			}
			if err := tx.Model(&old).Association("Tags").Clear(); err != nil {
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

func (s *ArticleService) UpdateArticleStatus(id uint, status int) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Article{}).Where("id = ?", id).Update("status", status).Error; err != nil {
			return err
		}
		if status == 2 {
			if err := tx.Where("article_id = ?", id).Unscoped().Delete(&models.ArticleColumnPublish{}).Error; err != nil {
				return err
			}
		}
		return nil
	})
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
		if err := tx.Model(&article).Association("Categories").Clear(); err != nil {
			return err
		}
		if err := tx.Model(&article).Association("Tags").Clear(); err != nil {
			return err
		}
		if err := tx.Model(&article).Association("Columns").Clear(); err != nil {
			return err
		}
		if err := tx.Where("article_id = ?", id).Unscoped().Delete(&models.ArticleColumnAudit{}).Error; err != nil {
			return err
		}
		if err := tx.Where("article_id = ?", id).Unscoped().Delete(&models.ArticleColumnAuditHistory{}).Error; err != nil {
			return err
		}
		if err := tx.Where("article_id = ?", id).Unscoped().Delete(&models.ArticleColumnPublish{}).Error; err != nil {
			return err
		}
		if err := tx.Where("article_id = ?", id).Unscoped().Delete(&models.ArticleAttachment{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Delete(&article).Error
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

type ArticleAuthorStat struct {
	Author     string `json:"author"`
	AuthorCode string `json:"authorCode"`
	Count      int64  `json:"count"`
}

func (s *ArticleService) GetArticleAuthorStats(period string) ([]ArticleAuthorStat, int64, error) {
	var results []ArticleAuthorStat
	var total int64
	now := time.Now()
	loc := now.Location()

	query := utils.DB.Model(&models.Article{}).Where("status = ?", 1)

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

// RestartArticleAudit 重新提交文章审核（清空旧记录后重新走提交流程）
func (s *ArticleService) RestartArticleAudit(articleID uint) error {
	var article models.Article
	if err := utils.DB.First(&article, articleID).Error; err != nil {
		return err
	}
	if article.Status != 0 {
		return fmt.Errorf("只有草稿状态的文章可以重新提交审核")
	}
	if article.AuditStatus != 2 {
		return fmt.Errorf("只有已审核状态的文章可以重新提交审核")
	}
	return s.StartArticleAudit(articleID)
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
		if err := tx.Where("article_id = ?", articleID).Unscoped().Delete(&models.ArticleColumnAudit{}).Error; err != nil {
			return err
		}
		if err := tx.Where("article_id = ?", articleID).Unscoped().Delete(&models.ArticleColumnAuditHistory{}).Error; err != nil {
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
	err := utils.DB.Transaction(func(tx *gorm.DB) error {
		// 更新文章状态为审核中
		if err := tx.Model(&article).Update("audit_status", 1).Error; err != nil {
			return err
		}
		// 清除旧的审核记录
		if err := tx.Where("article_id = ?", articleID).Unscoped().Delete(&models.ArticleColumnAudit{}).Error; err != nil {
			return err
		}
		// 清除旧的审核历史记录
		if err := tx.Where("article_id = ?", articleID).Unscoped().Delete(&models.ArticleColumnAuditHistory{}).Error; err != nil {
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
			var firstNode models.WorkflowNode
			err := tx.Where("workflow_id = ?", *col.WorkflowID).Order("sort_order ASC").First(&firstNode).Error
			if err != nil {
				// 流程没有节点，视为直接通过
				audit := models.ArticleColumnAudit{
					ArticleID:     articleID,
					ColumnID:      col.ID,
					WorkflowID:    *col.WorkflowID,
					CurrentNodeID: 0,
					Status:        1,
				}
				if err := tx.Create(&audit).Error; err != nil {
					return err
				}
				continue
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
	s.tryCompleteArticleAudit(articleID)
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
		if node.ApproverID == 0 {
			return true, nil
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
		var headDepartments []models.Department
		if err := utils.DB.Where("leader_code = ?", currentUser.ID).Find(&headDepartments).Error; err != nil {
			return false, err
		}
		for _, hd := range headDepartments {
			if slices.Contains(authorDeptIDs, hd.ID) {
				return true, nil
			}
		}
		return false, nil
	case "user", "":
		if node.ApproverID == 0 || node.ApproverID == userID {
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
	if err := utils.DB.Create(&history).Error; err != nil {
		return err
	}
	// 查找下一个节点
	var nextNode models.WorkflowNode
	err = utils.DB.Where("workflow_id = ? AND sort_order > ?", audit.WorkflowID, currentNode.SortOrder).Order("sort_order ASC").First(&nextNode).Error
	if err != nil {
		// 没有下一个节点，标记为已通过，记录通过人信息
		now := time.Now()
		if err := utils.DB.Model(&audit).Updates(map[string]any{
			"status":            1,
			"current_node_id":   0,
			"approve_remark":    remark,
			"approve_user_id":   userID,
			"approve_user_name": userName,
			"approve_time":      now,
		}).Error; err != nil {
			return err
		}
		s.tryCompleteArticleAudit(articleID)
		return nil
	}
	// 推进到下一个节点
	return utils.DB.Model(&audit).Updates(map[string]any{
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
	var article models.Article
	if err := utils.DB.First(&article, articleID).Error; err != nil {
		return err
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
	if err := utils.DB.Create(&history).Error; err != nil {
		return err
	}
	if err := utils.DB.Model(&audit).Updates(map[string]any{
		"status":        2,
		"reject_remark": remark,
	}).Error; err != nil {
		return err
	}
	s.tryCompleteArticleAudit(articleID)
	return nil
}

// GetArticleAuditHistory 获取文章指定栏目的审核历史
func (s *ArticleService) GetArticleAuditHistory(articleID uint, columnID uint) ([]models.ArticleColumnAuditHistory, error) {
	var histories []models.ArticleColumnAuditHistory
	err := utils.DB.Where("article_id = ? AND column_id = ?", articleID, columnID).Order("created_at ASC").Find(&histories).Error
	return histories, err
}

// tryCompleteArticleAudit 检查文章所有栏目流程是否都已结束（通过或驳回），若是则自动完成文章审核
func (s *ArticleService) tryCompleteArticleAudit(articleID uint) {
	var pendingCount int64
	utils.DB.Model(&models.ArticleColumnAudit{}).
		Where("article_id = ? AND status = ?", articleID, 0).
		Count(&pendingCount)
	if pendingCount == 0 {
		s.CompleteArticleAudit(articleID)
	}
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

// GetDetailPageRoutePath 获取详情页统一访问路径
func (s *ArticleService) GetDetailPageRoutePath() (string, string, error) {
	type Page struct {
		RoutePath string `gorm:"column:route_path"`
		Name      string `gorm:"column:name"`
	}
	var page Page
	err := utils.DB.Model(&models.Page{}).
		Select("route_path,name").
		Where("page_type = ? AND status = ?", "detail", 1).
		Order("id ASC").
		Limit(1).
		Scan(&page).Error
	return page.RoutePath, page.Name, err
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

// GetMyAuditArticles 获取当前用户需要审核的文章列表
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
		//Where("article.author_code != ?", strconv.FormatUint(uint64(userID), 10)).
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
			continue
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

// CompleteArticleAudit 完成文章审核（所有栏目通过后调用）
func (s *ArticleService) CompleteArticleAudit(articleID uint) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
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
		if err := tx.Where("article_id = ?", articleID).Unscoped().Delete(&models.ArticleColumnPublish{}).Error; err != nil {
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
			if err := tx.First(&col, audit.ColumnID).Error; err != nil {
				continue // 栏目不存在则跳过
			}
			publish := models.ArticleColumnPublish{
				PageID:       col.PageID,
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
			if err := tx.Model(&article).Updates(map[string]any{"status": 1}).Error; err != nil {
				return err
			}

		}
		return nil
	})
}
