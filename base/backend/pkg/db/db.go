package db

import (
	"fmt"

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
	return DB.AutoMigrate(
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
		&models.OperationLog{},
		&models.Setting{},
		&models.LoginLog{},
		&models.Organization{},
		&models.Message{},
		&models.MessageTemplate{},
		&models.Dict{},
		&models.DictItem{},
		&models.UploadedFile{},
	)
}
