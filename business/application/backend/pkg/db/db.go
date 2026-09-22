package db

import (
	"fmt"
	"log"
	"strings"
	"time"

	"application/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg *config.MySQLConfig) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.Charset)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("Failed to connect database: " + err.Error())
	}

	sqlDB, err := DB.DB()
	if err != nil {
		panic("Failed to get sql.DB: " + err.Error())
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if purged := purgeLegacySoftDeletedRows(); purged > 0 {
		log.Printf("[migrate] 共物理删除 %d 条历史软删除记录", purged)
	}
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
// 只记日志、不阻断启动：可能被其它表的外键引用，这类表先跳过、下一轮（子表被清空后）再试，
// 最多 3 轮，仍失败则告警 —— 宁可不清理，也不制造孤儿数据或让服务起不来。
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
	var pending int64
	err = DB.Raw("SELECT COALESCE(SUM(n), 0) FROM (" + strings.Join(parts, " UNION ALL ") + ") AS t").Scan(&pending).Error
	if err != nil {
		log.Printf("[migrate] 统计历史软删除残留失败，跳过清理: %s", err)
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
