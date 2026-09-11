package db

import (
	"fmt"
	"log"
	"member/config"
	"os"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg *config.MySQLConfig) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.Charset,
	)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:                                   newLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connection established")
}

// PurgeSoftDeleted 完成「软删 -> 硬删」的收尾迁移：
//  1. 物理删除历史遗留的软删记录（deleted_at IS NOT NULL），避免它们在去掉软删过滤后“复活”；
//  2. 删除已无用的 deleted_at 列。
//
// 该操作幂等、非致命：任何一步失败只记录日志，不影响服务启动。
func PurgeSoftDeleted() {
	var rows []struct {
		TableName string `gorm:"column:table_name"`
	}
	if err := DB.Raw(`SELECT TABLE_NAME AS table_name FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND COLUMN_NAME = 'deleted_at'`).Scan(&rows).Error; err != nil {
		log.Printf("PurgeSoftDeleted: 查询 deleted_at 列失败: %v", err)
		return
	}
	if len(rows) == 0 {
		return
	}

	processed := 0
	for _, r := range rows {
		name := strings.ReplaceAll(r.TableName, "`", "``")
		quoted := "`" + name + "`"

		if err := DB.Exec("DELETE FROM " + quoted + " WHERE deleted_at IS NOT NULL").Error; err != nil {
			log.Printf("PurgeSoftDeleted: 清理 %s 软删记录失败: %v", r.TableName, err)
			continue
		}
		if err := DB.Exec("ALTER TABLE " + quoted + " DROP COLUMN deleted_at").Error; err != nil {
			log.Printf("PurgeSoftDeleted: 删除 %s.deleted_at 列失败: %v", r.TableName, err)
			continue
		}
		processed++
	}
	log.Printf("PurgeSoftDeleted: 已处理 %d/%d 张表（软删记录已物理清除）", processed, len(rows))
}
