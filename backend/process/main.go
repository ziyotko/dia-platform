package main

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"server/config"
	"server/models"
)

var (
	historyDB *gorm.DB
	newDB     *gorm.DB
)

// CmsArchives 对应历史表 cms_archives 的结构
type CmsArchives struct {
	ID           int       `gorm:"column:id"`
	Title        string    `gorm:"column:title"`
	Thumbnail    string    `gorm:"column:thumbnail"`
	Des          string    `gorm:"column:des"`
	Content      string    `gorm:"column:content"`
	Source       string    `gorm:"column:source"`
	CreateUserID int       `gorm:"column:create_user_id"`
	CreateTime   time.Time `gorm:"column:create_time"`
	PublishTime  time.Time `gorm:"column:publish_time"`
	LinkURL      string    `gorm:"column:link_url"`
	LastUpdate   time.Time `gorm:"column:last_update_time"`
}

func (CmsArchives) TableName() string {
	return "cms_archives"
}

var oldColumn = []string{"%首页-轮播%", "%行业要闻%", "%协会活动%"}
var newColumn = []string{"首页轮播", "首页行业要闻", "首页协会活动"}

func main() {
	createNewDBConnection()
	createHistoryDBConnection()

	newDB.Unscoped().Table("article_category").Where("article_id < 10000000").Delete(nil)
	newDB.Unscoped().Table("article_tag").Where("article_id < 10000000").Delete(nil)
	newDB.Unscoped().Table("article_column").Where("article_id < 10000000").Delete(nil)
	newDB.Unscoped().Where("article_id < 10000000").Delete(&models.ArticleColumnAudit{})
	newDB.Unscoped().Where("article_id < 10000000").Delete(&models.ArticleColumnPublish{})
	newDB.Unscoped().Where("id < 10000000").Delete(&models.Article{})
	var totalInserted, totalSkipped int
	for i := range oldColumn {
		inserted, skipped := migrateHistoryArchives(i)
		totalInserted += inserted
		totalSkipped += skipped
	}
	fmt.Printf("总计：成功迁移 %d 条，跳过 %d 条\n", totalInserted, totalSkipped)
}

func createNewDBConnection() {
	config.InitConfig()
	conf := config.AppConfig.Database

	parseTime := "false"
	if conf.ParseTime {
		parseTime = "true"
	}

	// 迁移可能出现慢查询，读写超时统一放大到 60s
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%s&loc=%s&timeout=%s&readTimeout=%s&writeTimeout=%s",
		conf.Username,
		conf.Password,
		conf.Host,
		conf.Port,
		conf.DBName,
		conf.Charset,
		parseTime,
		url.QueryEscape(conf.Loc),
		conf.Timeout,
		"60s",
		"60s",
	)
	println(dsn)
	var err error

	newDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		log.Fatalf("Failed to connect database: %s", err)
	}

	sqlDB, err := newDB.DB()
	if err != nil {
		log.Fatalf("Failed to get DB: %s", err)
	}

	err = sqlDB.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %s", err)
	}

	if conf.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(conf.MaxIdleConns)
	} else {
		sqlDB.SetMaxIdleConns(10)
	}

	if conf.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(conf.MaxOpenConns)
	} else {
		sqlDB.SetMaxOpenConns(100)
	}

	if conf.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetime) * time.Second)
	}

	if conf.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(conf.ConnMaxIdleTime) * time.Second)
	}
}

func createHistoryDBConnection() {
	conf := config.AppConfig.Database

	parseTime := "false"
	if conf.ParseTime {
		parseTime = "true"
	}

	// 迁移可能出现慢查询，读写超时统一放大到 60s
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%s&loc=%s&timeout=%s&readTimeout=%s&writeTimeout=%s",
		conf.Username,
		conf.Password,
		conf.Host,
		conf.Port,
		"bmsite_v3",
		conf.Charset,
		parseTime,
		url.QueryEscape(conf.Loc),
		conf.Timeout,
		"60s",
		"60s",
	)
	println(dsn)
	var err error

	historyDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		log.Fatalf("Failed to connect database: %s", err)
	}

	sqlDB, err := historyDB.DB()
	if err != nil {
		log.Fatalf("Failed to get DB: %s", err)
	}

	err = sqlDB.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %s", err)
	}

	if conf.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(conf.MaxIdleConns)
	} else {
		sqlDB.SetMaxIdleConns(10)
	}

	if conf.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(conf.MaxOpenConns)
	} else {
		sqlDB.SetMaxOpenConns(100)
	}

	if conf.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(conf.ConnMaxLifetime) * time.Second)
	}

	if conf.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(conf.ConnMaxIdleTime) * time.Second)
	}
}

func getTargetColumn(index int) (models.Column, error) {
	var col models.Column
	err := newDB.Where("name like ?", newColumn[index]).First(&col).Error
	return col, err
}

func migrateHistoryArchives(index int) (int, int) {
	var archives []CmsArchives
	if err := historyDB.Where("category_name like ?", oldColumn[index]).Where("is_publish=?", 1).Find(&archives).Error; err != nil {
		log.Fatalf("查询历史文章失败: %s", err)
	}

	fmt.Printf("共查询到 %d 条历史文章数据\n", len(archives))
	if len(archives) == 0 {
		return 0, 0
	}

	col, err := getTargetColumn(index)
	if err != nil {
		log.Fatalf("获取目标栏目失败: %s", err)
	}
	fmt.Printf("目标栏目: id=%d, page_id=%d\n", col.ID, col.PageID)

	var inserted, skipped int
	for _, a := range archives {

		article := mapArchiveToArticle(a)
		err := newDB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&article).Error; err != nil {
				log.Printf("迁移文章 %d 失败: %s", a.ID, err)
				//return err
			}
			if err := tx.Exec("INSERT INTO article_column (article_id, column_id) VALUES (?, ?)", article.ID, col.ID).Error; err != nil {
				return err
			}
			audit := models.ArticleColumnAudit{
				ArticleID:     article.ID,
				ColumnID:      col.ID,
				WorkflowID:    0,
				ApproveUserID: 0,
				CurrentNodeID: 0,
				Status:        1, // 已通过
			}
			if err := tx.Create(&audit).Error; err != nil {
				return err
			}
			publish := models.ArticleColumnPublish{
				PageID:       col.PageID,
				ColumnID:     col.ID,
				ArticleID:    article.ID,
				ArticleTitle: article.Title,
				Author:       article.Author,
				Source:       article.Source,
			}
			if err := tx.Create(&publish).Error; err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			log.Fatalf("迁移文章 %d 失败: %s", a.ID, err)
		}
		inserted++
	}

	fmt.Printf("成功迁移 %d 条，跳过 %d 条\n", inserted, skipped)
	return inserted, skipped
}

func mapArchiveToArticle(a CmsArchives) models.Article {

	createdAt := a.CreateTime
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	updatedAt := a.LastUpdate
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}

	var publishTime *models.LocalTime
	if !a.PublishTime.IsZero() {
		publishTime = &models.LocalTime{Time: a.PublishTime}
	}

	authorCode := ""
	if a.CreateUserID > 0 {
		authorCode = strconv.Itoa(a.CreateUserID)
	}

	return models.Article{
		ID:          uint(a.ID),
		Title:       a.Title,
		Type:        1,
		Summary:     a.Des,
		Content:     pathToURL(a.Content),
		Status:      1,
		AuditStatus: 2,
		Cover:       pathToURL(a.Thumbnail),
		Author:      authorCode,
		AuthorCode:  authorCode,
		Source:      a.Source,
		PublishTime: publishTime,
		URL:         a.LinkURL,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		ColumnCount: 1,
	}
}

func pathToURL(path string) string {
	if strings.Contains(path, "http://file.caam.org.cn/") {
		path = strings.ReplaceAll(path, "http://file.caam.org.cn/", "/uploads")
	} else {
		if strings.Contains(path, "http://") {
			return path
		}
		if strings.Contains(path, "https://") {
			return path
		}
		if path != "" {
			path = "/uploads" + path
		}
	}
	return path
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	// 避免截断中文字符出现乱码，按 rune 截断
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}
