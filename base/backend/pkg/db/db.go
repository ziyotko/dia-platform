package db

import (
	"fmt"
	"log"
	"strings"

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
	// 统一硬删：先物理清掉旧库遗留的软删除数据。
	// 模型已不再内嵌 gorm.DeletedAt（查询不过滤 deleted_at），不清理这些行它们会重新"出现"。
	if purged := purgeLegacySoftDeletedRows(); purged > 0 {
		if err := cleanupOrphanUserRelations(); err != nil {
			log.Printf("[migrate] 清理残留角色关联失败: %s", err)
		}
	}
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

// purgeLegacySoftDeletedRows 清理历史软删除残留，返回被物理删除的行数。
//
// 本项目已统一为物理删除（硬删）：模型不再内嵌 gorm.DeletedAt，所有查询也不再带
// `deleted_at IS NULL` 过滤。若不清理，旧库里 `deleted_at IS NOT NULL` 的行会因为过滤条件
// 消失而重新出现在列表里（还会继续占用唯一索引），因此启动时逐表物理删除。
//
// 幂等且开销很小：先一次性查出带 `deleted_at` 列的表并统计残留行数，没有残留（含新库）直接返回；
// 只有确实有残留时才逐表 DELETE。
//
// 只记日志、不阻断启动：可能被其它表的外键引用（如 `base_app` 被 `base_app_instance` 引用），
// 这类表先跳过、下一轮（子表被清空后）再试，最多 3 轮，仍失败则告警 —— 宁可不清理，
// 也不制造孤儿数据或让服务起不来。
func purgeLegacySoftDeletedRows() int64 {
	var tables []string
	err := DB.Raw("SELECT TABLE_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND COLUMN_NAME = 'deleted_at'").
		Scan(&tables).Error
	if err != nil {
		log.Printf("[migrate] 读取软删除列信息失败，跳过历史软删除残留清理: %s", err)
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
	var softDeleted int64
	err = DB.Raw("SELECT COALESCE(SUM(n), 0) FROM (" + strings.Join(parts, " UNION ALL ") + ") AS t").Scan(&softDeleted).Error
	if err != nil {
		log.Printf("[migrate] 统计历史软删除残留失败，跳过清理: %s", err)
		return 0
	}
	if softDeleted == 0 {
		return 0
	}

	purged := int64(0)
	remaining := tables
	// 最多 3 轮：外键约束下父子表有先后依赖（父表要等子表先被清空），每轮清掉一批、重试剩下的。
	// 某一轮完全没进展就停止（互相引用之类）；仍失败的表只告警，不阻断启动。
	for round := 1; round <= 3 && len(remaining) > 0; round++ {
		var retry []string
		for _, table := range remaining {
			res := DB.Exec("DELETE FROM `" + table + "` WHERE deleted_at IS NOT NULL")
			if res.Error != nil {
				retry = append(retry, table)
				continue
			}
			if res.RowsAffected > 0 {
				log.Printf("[migrate] %s 表物理删除 %d 条历史软删除记录（已统一为硬删）", table, res.RowsAffected)
				purged += res.RowsAffected
			}
		}
		if len(retry) == 0 {
			break
		}
		if round == 3 || len(retry) == len(remaining) {
			for _, table := range retry {
				log.Printf("[migrate] %s 表的历史软删除记录未能清理（可能被其它表的外键引用），请人工处理", table)
			}
			break
		}
		remaining = retry
	}
	return purged
}

// dedupeUserUsernames 为 base_user 的 (tenant_id, username) 唯一索引做数据准备。
//
// 背景：早期版本没有唯一索引（并发写入会产生同租户同名账号），(tenant_id, username) 上出现
// 重复行时建索引必然失败，所以启动建索引前先把历史重复数据整理成「一组一行」：
// 保留 id 最小的一条，其余按 <原名>_dup<id> 改名并打日志（需人工在用户管理里确认）。
// 旧版软删除行已由 purgeLegacySoftDeletedRows 在更早的启动阶段物理删除，这里不再区分存活状态。
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

	for _, g := range groups {
		// COALESCE 必需：SQL 的 MIN(id) 在空集时返回 NULL，直接扫进 uint64 会报
		// "converting NULL to uint64 is unsupported" 而让服务起不来。
		var keepID uint64
		if err := DB.Raw("SELECT COALESCE(MIN(id), 0) FROM base_user WHERE tenant_id = ? AND username = ?",
			g.TenantID, g.Username).Scan(&keepID).Error; err != nil {
			return err
		}

		// LEFT(username, 40) + '_dup' + id(最多 20 位) = 最多 64 字符，不会超出列长度
		res := DB.Exec("UPDATE base_user SET username = CONCAT(LEFT(username, 40), '_dup', id) WHERE tenant_id = ? AND username = ? AND id <> ?",
			g.TenantID, g.Username, keepID)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected > 0 {
			log.Printf("[migrate] 租户 %d 存在同名的重复账号，已保留 id=%d，其余 %d 条改名为 %s_dup<id>，请人工确认",
				g.TenantID, keepID, res.RowsAffected, g.Username)
		}
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
