package service

import (
	"errors"
	"fmt"

	"base/internal/models"
	"base/pkg/db"
	"base/pkg/utils"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type TenantService struct{}

// Create 创建租户。租户编码唯一（重名直接报错），删除是物理删除，
// 编码删除后可以直接重新使用。
func (s TenantService) Create(t *models.Tenant) error {
	if t.Code == "" {
		t.Code = "T" + utils.RandomDigit(8)
	}
	if err := validateTenantFields(t); err != nil {
		return err
	}

	var count int64
	if err := db.DB.Model(&models.Tenant{}).Where("code = ?", t.Code).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("租户编码已存在")
	}
	if err := db.DB.Create(t).Error; err != nil {
		// 并发下两个请求可能同时通过预检查，唯一索引兜底
		if isDuplicateEntry(err) {
			return errors.New("租户编码已存在")
		}
		return err
	}
	return nil
}

func (s TenantService) Update(t *models.Tenant) error {
	if err := validateTenantFields(t); err != nil {
		return err
	}
	if err := ensureRecordExists(db.DB.Model(&models.Tenant{}).Where("id = ?", t.ID), "租户不存在"); err != nil {
		return err
	}
	return db.DB.Model(t).Updates(map[string]interface{}{
		"name":          t.Name,
		"status":        t.Status,
		"contact_name":  t.ContactName,
		"contact_phone": t.ContactPhone,
		"description":   t.Description,
	}).Error
}

// validateTenantFields 校验租户各字段长度（对应 base_tenant 的定长列）。
func validateTenantFields(t *models.Tenant) error {
	return validateLengths(
		fieldLen{"租户编码", t.Code, 64},
		fieldLen{"租户名称", t.Name, 128},
		fieldLen{"联系人", t.ContactName, 64},
		fieldLen{"联系电话", t.ContactPhone, 32},
		fieldLen{"描述", t.Description, 512},
	)
}

// Delete 删除租户。
//
// 分三类处理（旧实现只删租户行，其余租户级数据会永久残留成「查不到也删不掉」的孤儿数据）：
//   - 需要人工确认的实体（用户/角色/应用实例/流程角色/流程定义/流程实例）→ 拒绝删除，提示先清理；
//   - 其余租户级业务数据（消息与已读记录/消息模板/字典与字典项/机构/租户自建菜单/上传文件）
//     → 在同一事务内级联删除；磁盘上的上传文件在事务提交后尽力删除；
//   - 操作日志与登录日志 → **有意保留**（审计留痕优先）：平台超管仍可在「审计日志/登录日志」中追溯；
//     它们是历史事实记录，随租户一起抹掉属于不可逆的审计损失。若确需彻底清空，请用日志页的「清理」功能。
//     残留行只有平台超管能看到（其列表不按租户过滤），MySQL 8 的自增 id 不复用，不会串给新建租户。
func (s TenantService) Delete(id uint64) error {
	var fileKeys []string
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Tenant{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errors.New("租户不存在")
		}

		refs := []struct {
			model interface{}
			name  string
		}{
			{&models.User{}, "用户"},
			{&models.Role{}, "角色"},
			{&models.AppInstance{}, "应用实例"},
			{&models.WorkflowRole{}, "流程角色"},
			{&models.Workflow{}, "流程定义"},
			{&models.WorkflowInstance{}, "流程实例"},
		}
		for _, ref := range refs {
			var n int64
			if err := tx.Model(ref.model).Where("tenant_id = ?", id).Count(&n).Error; err != nil {
				return err
			}
			if n > 0 {
				return fmt.Errorf("该租户下还有 %d 条%s数据，请先清理后再删除租户", n, ref.name)
			}
		}

		// 先记录待删文件，事务提交后再删磁盘，避免磁盘失败导致业务回滚
		if err := tx.Model(&models.UploadedFile{}).Where("tenant_id = ?", id).Pluck("file_key", &fileKeys).Error; err != nil {
			return err
		}

		// 关联表先清（子查询指向即将删除的父表）
		for _, sql := range []string{
			"DELETE FROM base_message_read WHERE message_id IN (SELECT id FROM base_message WHERE tenant_id = ?)",
			"DELETE FROM base_role_menu WHERE menu_id IN (SELECT id FROM base_menu WHERE tenant_id = ?)",
			"DELETE FROM base_dict_item WHERE dict_id IN (SELECT id FROM base_dict WHERE tenant_id = ?)",
		} {
			if err := tx.Exec(sql, id).Error; err != nil {
				return err
			}
		}
		for _, sql := range []string{
			"DELETE FROM base_message WHERE tenant_id = ?",
			"DELETE FROM base_message_template WHERE tenant_id = ?",
			"DELETE FROM base_dict WHERE tenant_id = ?",
			"DELETE FROM base_organization WHERE tenant_id = ?",
			"DELETE FROM base_menu WHERE tenant_id = ?",
			"DELETE FROM base_uploaded_file WHERE tenant_id = ?",
		} {
			if err := tx.Exec(sql, id).Error; err != nil {
				return err
			}
		}
		return tx.Where("id = ?", id).Delete(&models.Tenant{}).Error
	})
	if err != nil {
		return err
	}

	// 磁盘文件清理：失败只告警（数据库行已删，残留文件不影响使用）
	if len(fileKeys) > 0 {
		fs := NewFileService()
		for _, key := range fileKeys {
			if err := fs.RemoveStoredFile(key); err != nil {
				logrus.WithError(err).Warnf("删除租户 %d 的上传文件失败: %s", id, key)
			}
		}
	}
	return nil
}

func (s TenantService) GetByID(id uint64) (*models.Tenant, error) {
	var t models.Tenant
	err := db.DB.First(&t, id).Error
	return &t, err
}

func (s TenantService) List(page, size int, keyword string) ([]models.Tenant, int64, error) {
	var list []models.Tenant
	var total int64
	query := db.DB.Model(&models.Tenant{})
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}
