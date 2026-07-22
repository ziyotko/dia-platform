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
	return migrate()
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
