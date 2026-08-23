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

/*
"%重点工作%", "%首页-轮播%", "%行业要闻%", "%协会活动%", "%文件公告%", "%协会文件%", "%协调合作%", "%行业动态%", "%品牌建设%", "%展会信息%", "%法规标准%", "%企业新闻%", "%国际合作%", "%最新政策%", "%标准法规%",
	"%协会简介%", "%协会章程%", "%组织机构%", "%主要职责%", "%协会荣誉%", "5231531", "5234908", "5231532", "5231533", "5235950", "5113238",
"%党建专栏轮播图%","%党建要闻%","%学习教育%","%工作动态%",
"%会员工作%","%会员风采%","5236416","5236415","5022079","5236414","5228799","5223238","5235804",
"%协会领导%","%协会工作-轮播图%", "%协会动态%", "%分支机构动态%", "%国际合作%", "%展会信息%", "%协调合作%",
	"%国内数据%", "%国外数据%", "%产销%", "%进出口%"

-----------------------------------------------------------------------------------------------------------------------------------------------------

"首页头条", "首页轮播", "首页行业要闻", "首页协会活动", "首页通知公告", "首页协会文件", "首页行业发展", "首页智能网联", "首页品牌服务", "首页展会信息", "首页标准法规", "首页企业新闻", "首页国际合作", "首页行业政策", "首页法律法规",
	"协会概况简介", "协会概况章程", "协会概况组织", "协会概况职责", "协会概况荣誉", "协会概况轮值会长", "协会概况副会长", "协会概况常务理事", "协会概况理事", "协会概况会员代表", "协会概况普通会员",
"党建专区轮播","党建专区党建要闻","党建专区学习教育","党建专区工作动态",
"会员专区会员工作","会员专区会员风采","会员专区分支机构介绍","会员专区分支机构管理办法","会员专区分支机构服务","会员专区分支机构名单","会员专区会员管理","会员专区协会章程","会员专区会费缴纳标准与方法"
"协会概况领导团队","协会工作头条", "协会工作协会动态", "协会工作分支机构动态", "协会工作国际合作", "协会工作展会信息", "协会工作行业发展",
	"统计数据国内数据", "统计数据国外数据", "统计数据产销", "统计数据进出口"
*/

// 每行对应每个首页的栏目
var oldColumn = []string{"%会员工作%", "%会员风采%", "5236416", "5236415", "5022079", "5236414", "5228799", "5223238", "5235804"}
var newColumn = []string{"会员专区会员工作", "会员专区会员风采", "会员专区分支机构介绍", "会员专区分支机构管理办法", "会员专区分支机构服务", "会员专区分支机构名单", "会员专区会员管理", "会员专区协会章程", "会员专区会费缴纳标准与方法"}

func main() {
	createNewDBConnection()
	createHistoryDBConnection()

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
		"260s",
		"260s",
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
	err := newDB.Where("name like ? and display_type<>7", newColumn[index]).First(&col).Error
	return col, err
}

func migrateHistoryArchives(index int) (int, int) {
	var archives []CmsArchives
	if strings.Contains(oldColumn[index], "%") { //模糊查询
		if err := historyDB.Where("category_name like ?", oldColumn[index]).Find(&archives).Error; err != nil {
			log.Fatalf("查询历史文章失败: %s", err)
		}
	} else { //精确查询
		id, err := strconv.Atoi(oldColumn[index])
		if err != nil {
			log.Fatalf("转换ID失败: %s", err)
		}
		if err := historyDB.Where("id = ?", id).Find(&archives).Error; err != nil {
			log.Fatalf("查询历史文章失败: %s", err)
		}
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
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Title:       a.Title,
		Type:        1,
		Summary:     a.Des,
		Content:     pathToContent(a.Content),
		Status:      1,
		AuditStatus: 2,
		Cover:       pathToURL(a.Thumbnail),
		Author:      authorCode,
		AuthorCode:  authorCode,
		Source:      a.Source,
		PublishTime: publishTime,
		URL:         a.LinkURL,
		ColumnCount: 1,
	}
}

func pathToContent(path string) string {
	if strings.Contains(path, "/uploads") {
		path = "/caamm/uploads" + path
	}
	if strings.Contains(path, "http://file.caam.org.cn/") {
		path = strings.ReplaceAll(path, "http://file.caam.org.cn/", "/caamm/uploads")
	}
	return path
}

func pathToURL(path string) string {
	if strings.Contains(path, "http://file.caam.org.cn/") {
		path = strings.ReplaceAll(path, "http://file.caam.org.cn/", "/caamm/uploads")
	} else {
		if strings.Contains(path, "http://") {
			return path
		}
		if strings.Contains(path, "https://") {
			return path
		}
		if path != "" {
			path = "/caamm/uploads" + path
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
