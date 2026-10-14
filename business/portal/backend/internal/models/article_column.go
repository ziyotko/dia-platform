package models

// ArticleColumn 文章-栏目关联表（同时也是 Article.Columns 的 many2many 连接表）。
//
// 显式建模的意义同 ArticleTag：主键 (article_id, column_id) 只支撑「按文章查栏目」，
// 图文管理的「发布栏目」筛选（`WHERE column_id = ?`）需要 column_id 反向索引，
// 否则每次筛选都全表扫描 article_column。
type ArticleColumn struct {
	ArticleID uint `gorm:"primaryKey" json:"articleId"`
	ColumnID  uint `gorm:"primaryKey;index" json:"columnId"`
}
