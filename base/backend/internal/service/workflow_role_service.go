package service

import (
	"errors"
	"fmt"

	"base/internal/models"
	"base/pkg/db"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// WorkflowRoleService 流程角色（审批角色）服务。
//
// 租户口径与 user/role 保持一致：读、写一律严格按租户隔离，
// 平台超管（tenantID == models.PlatformTenantID）可跨租户查看与管理。
type WorkflowRoleService struct{}

func (s WorkflowRoleService) Create(r *models.WorkflowRole) error {
	var count int64
	if err := db.DB.Model(&models.WorkflowRole{}).
		Where("tenant_id = ? AND code = ?", r.TenantID, r.Code).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该租户下流程角色编码已存在")
	}
	// 必须 Omit 关联：WorkflowRole.Users 是 many2many，GORM 的 Create 会把请求体里的 users 一并 upsert 进 base_user。
	// 否则持有 base:workflow-role:create 的用户可以凭空造出一条 tenant_id=0 / is_admin=true 的账号（提权为平台超管）。
	// 成员统一走 POST /workflow-roles/:id/users。
	return db.DB.Omit(clause.Associations).Create(r).Error
}

func (s WorkflowRoleService) Update(r *models.WorkflowRole, tenantID uint64) error {
	query := db.DB.Model(&models.WorkflowRole{}).Where("id = ?", r.ID)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("流程角色不存在")
	}

	// 编码为幂等键，租户内不允许重复（排除自身）
	var dup int64
	if err := db.DB.Model(&models.WorkflowRole{}).
		Where("tenant_id = ? AND code = ? AND id <> ?", r.TenantID, r.Code, r.ID).
		Count(&dup).Error; err != nil {
		return err
	}
	if dup > 0 {
		return errors.New("该租户下流程角色编码已存在")
	}

	return db.DB.Model(&models.WorkflowRole{}).Where("id = ?", r.ID).Updates(map[string]interface{}{
		"name":        r.Name,
		"code":        r.Code,
		"description": r.Description,
		"status":      r.Status,
	}).Error
}

// Delete 删除流程角色，同时清理成员关联，避免残留脏关联数据。
// 仍被流程节点引用（approver_type = role 且 approver_id = 本角色）时拒绝删除：
// 否则该节点会解析不到审批人而静默「自动通过」，等于无声跳过一道审批。
func (s WorkflowRoleService) Delete(id uint64, tenantID uint64) error {
	query := db.DB.Where("id = ?", id)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	var role models.WorkflowRole
	if err := query.First(&role).Error; err != nil {
		return err
	}
	var refs int64
	if err := db.DB.Model(&models.WorkflowNode{}).
		Where("approver_type = ? AND approver_id = ?", models.ApproverTypeRole, role.ID).
		Count(&refs).Error; err != nil {
		return err
	}
	if refs > 0 {
		return fmt.Errorf("该流程角色仍被 %d 个流程节点引用，请先在「流程定义」中调整节点后再删除", refs)
	}
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("workflow_role_id = ?", role.ID).Delete(&models.WorkflowRoleUser{}).Error; err != nil {
			return err
		}
		return tx.Delete(&role).Error
	})
}

func (s WorkflowRoleService) GetByID(id uint64, tenantID uint64) (*models.WorkflowRole, error) {
	var r models.WorkflowRole
	query := db.DB.Preload("Users").Where("id = ?", id)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&r).Error; err != nil {
		return &r, err
	}
	r.UserCount = int64(len(r.Users))
	return &r, nil
}

// List 分页查询流程角色。tenantID 为当前登录用户所属租户：
//   - 普通租户用户（tenantID > 0）只能看到本租户；
//   - 平台超管（tenantID == 0）不传 filterTenantID 时查看全部租户，传则只看指定租户。
func (s WorkflowRoleService) List(tenantID, filterTenantID uint64, page, size int, keyword string, status *int) ([]models.WorkflowRole, int64, error) {
	var list []models.WorkflowRole
	var total int64
	query := db.DB.Model(&models.WorkflowRole{})
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	} else if filterTenantID > 0 {
		query = query.Where("tenant_id = ?", filterTenantID)
	}
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	// 成员数量单独聚合一次，避免逐行统计
	if len(list) > 0 {
		ids := make([]uint64, 0, len(list))
		for _, item := range list {
			ids = append(ids, item.ID)
		}
		type row struct {
			WorkflowRoleID uint64
			Total          int64
		}
		var rows []row
		if err := db.DB.Model(&models.WorkflowRoleUser{}).
			Select("workflow_role_id, COUNT(*) AS total").
			Where("workflow_role_id IN ?", ids).
			Group("workflow_role_id").
			Scan(&rows).Error; err != nil {
			return nil, 0, err
		}
		counts := make(map[uint64]int64, len(rows))
		for _, r := range rows {
			counts[r.WorkflowRoleID] = r.Total
		}
		for i := range list {
			list[i].UserCount = counts[list[i].ID]
		}
	}
	return list, total, nil
}

// AssignUsers 覆盖式设置流程角色成员。
// 成员必须与流程角色同租户，避免把其他租户的用户塞进本租户的审批角色。
func (s WorkflowRoleService) AssignUsers(roleID uint64, userIDs []uint64, tenantID uint64) error {
	var role models.WorkflowRole
	query := db.DB.Where("id = ?", roleID)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&role).Error; err != nil {
		return err
	}

	var users []models.User
	if len(userIDs) > 0 {
		if err := db.DB.Where("id IN ? AND tenant_id = ?", userIDs, role.TenantID).Find(&users).Error; err != nil {
			return err
		}
		if len(users) != len(userIDs) {
			return errors.New("存在不属于该流程角色所属租户的用户")
		}
	}
	return db.DB.Model(&role).Association("Users").Replace(users)
}

// ListUserOptions 成员选择器使用的用户选项（不分页，仅返回必要字段）。
// 范围与流程角色一致：普通租户用户只看本租户；平台超管可指定租户，不指定则全部租户。
func (s WorkflowRoleService) ListUserOptions(tenantID, filterTenantID uint64, keyword string, limit int) ([]models.User, error) {
	if limit <= 0 {
		limit = 100
	}
	var users []models.User
	query := db.DB.Model(&models.User{})
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	} else if filterTenantID > 0 {
		query = query.Where("tenant_id = ?", filterTenantID)
	}
	if keyword != "" {
		query = query.Where("username LIKE ? OR real_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	err := query.Order("status DESC, username ASC").Limit(limit).Find(&users).Error
	return users, err
}
