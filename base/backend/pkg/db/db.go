package db

import (
	"fmt"
	"log"

	"base/config"
	"base/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg *config.MySQL) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.Charset,
	)
	var logMode logger.LogLevel
	if config.Cfg.Server.Mode == "debug" {
		logMode = logger.Info
	} else {
		logMode = logger.Silent
	}
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logMode),
	})
	if err != nil {
		return err
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	if err := migrate(); err != nil {
		return err
	}
	return normalizeLegacyData()
}

// normalizeLegacyData 兼容历史数据的幂等修正。
// 早期消息表用 status 同时表示已读状态（0 未读 / 1 已读）与发送状态，现已拆分为 is_read + status(2 草稿 / 3 已发送)。
func normalizeLegacyData() error {
	if err := DB.Model(&models.Message{}).Where("status = ?", 1).Update("is_read", true).Error; err != nil {
		return err
	}
	return DB.Model(&models.Message{}).Where("status IN ?", []int{0, 1}).Update("status", 3).Error
}

func migrate() error {
	// 必须先清理历史重复用户名，否则 AutoMigrate 建 (tenant_id, username) 唯一索引会直接失败，
	// 结果是服务根本起不来（不是慢一步的问题）。
	if err := dedupeUserUsernames(); err != nil {
		return err
	}
	if err := DB.AutoMigrate(
		&models.Tenant{},
		&models.App{},
		&models.AppInstance{},
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.RolePermission{},
		&models.WorkflowRole{},
		&models.WorkflowRoleUser{},
		&models.Workflow{},
		&models.WorkflowNode{},
		&models.WorkflowInstance{},
		&models.WorkflowTask{},
		&models.WorkflowLog{},
		&models.Menu{},
		&models.MessageRead{},
		&models.OperationLog{},
		&models.Setting{},
		&models.LoginLog{},
		&models.Organization{},
		&models.Message{},
		&models.MessageTemplate{},
		&models.Dict{},
		&models.DictItem{},
		&models.UploadedFile{},
	); err != nil {
		return err
	}
	return dropLegacyUserUsernameIndex()
}

// dedupeUserUsernames 为 base_user 的 (tenant_id, username) 唯一索引做数据准备。
//
// 背景：唯一索引在软删除后仍然占位，(tenant_id, username) 上出现重复行时建索引必然失败，
// 所以启动建索引前先把历史重复数据整理成「一组一行」：
//   - 组内若有存活账号 → 保留 id 最小的存活账号；软删除的重复行物理删除（逻辑上已删除，只占唯一键位）；
//     其余存活账号改名 <原名>_dup<id> 并打日志（历史并发写入才会出现，需人工在用户管理里确认）。
//   - 整组都已软删除 → 保留 id 最小的一条（保留软删除状态，Create 时走「恢复」分支），其余物理删除。
//
// 幂等：清理后每组只剩一行，再次启动时 GROUP BY ... HAVING COUNT(*) > 1 查不到数据，直接返回。
func dedupeUserUsernames() error {
	if !DB.Migrator().HasTable(&models.User{}) {
		return nil
	}
	var groups []struct {
		TenantID uint64
		Username string
	}
	if err := DB.Raw("SELECT tenant_id, username FROM base_user GROUP BY tenant_id, username HAVING COUNT(*) > 1").
		Scan(&groups).Error; err != nil {
		return err
	}
	if len(groups) == 0 {
		return nil
	}

	purged := int64(0)
	for _, g := range groups {
		var keepID uint64
		// COALESCE 必需：SQL 的 MIN(id) 在「没有存活行」时返回 NULL，直接扫进 uint64 会报
		// "converting NULL to uint64 is unsupported" 而让服务起不来。
		if err := DB.Raw("SELECT COALESCE(MIN(id), 0) FROM base_user WHERE tenant_id = ? AND username = ? AND deleted_at IS NULL",
			g.TenantID, g.Username).Scan(&keepID).Error; err != nil {
			return err
		}
		if keepID == 0 {
			if err := DB.Raw("SELECT COALESCE(MIN(id), 0) FROM base_user WHERE tenant_id = ? AND username = ?",
				g.TenantID, g.Username).Scan(&keepID).Error; err != nil {
				return err
			}
		}

		res := DB.Exec("DELETE FROM base_user WHERE tenant_id = ? AND username = ? AND id <> ? AND deleted_at IS NOT NULL",
			g.TenantID, g.Username, keepID)
		if res.Error != nil {
			return res.Error
		}
		purged += res.RowsAffected

		res = DB.Exec("UPDATE base_user SET username = CONCAT(LEFT(username, 40), '_dup', id) WHERE tenant_id = ? AND username = ? AND id <> ? AND deleted_at IS NULL",
			g.TenantID, g.Username, keepID)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected > 0 {
			log.Printf("[migrate] 租户 %d 存在同名的存活账号，已保留 id=%d，其余 %d 条改名为 %s_dup<id>，请人工确认",
				g.TenantID, keepID, res.RowsAffected, g.Username)
		}
	}

	if purged > 0 {
		// LEFT(username, 40) + '_dup' + id(最多 20 位) = 最多 64 字符，不会超出列长度
		log.Printf("[migrate] 物理删除 %d 条软删除的重复用户名记录（(tenant_id, username) 唯一索引不能有重复行）", purged)
		return cleanupOrphanUserRelations()
	}
	return nil
}

// cleanupOrphanUserRelations 清理被物理删除用户残留的角色关联，避免脏关联数据。
func cleanupOrphanUserRelations() error {
	for _, table := range []string{"base_user_role", "base_workflow_role_user"} {
		if !DB.Migrator().HasTable(table) {
			continue
		}
		if err := DB.Exec("DELETE FROM " + table + " WHERE user_id NOT IN (SELECT id FROM base_user)").Error; err != nil {
			return err
		}
	}
	return nil
}

// dropLegacyUserUsernameIndex 删除历史遗留的单列非唯一索引 idx_base_user_username。
// 该索引已被 (tenant_id, username) 唯一索引取代（按租户查用户的场景走更左前缀即可），
// 留着只会让「用户名到底唯不唯一」看起来仍不确定。
func dropLegacyUserUsernameIndex() error {
	const legacy = "idx_base_user_username"
	if !DB.Migrator().HasIndex(&models.User{}, legacy) {
		return nil
	}
	return DB.Migrator().DropIndex(&models.User{}, legacy)
}
