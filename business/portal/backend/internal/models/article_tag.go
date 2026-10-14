package models

// ArticleTag 文章-标签关联表（同时也是 Article.Tags 的 many2many 连接表）。
//
// 显式建模的意义：连接表主键是 (article_id, tag_id)，只支撑「按文章查标签」；
// 列表/公开搜索的「按标签筛文章」（`WHERE tag_id = ?`）需要反向索引，
// 否则每次筛选都全表扫描连接表。与 ArticleCategory 同一写法——
// 项目用 SingularTable 命名策略，模型 ArticleTag 正好落在连接表 article_tag 上。
type ArticleTag struct {
	ArticleID uint `gorm:"primaryKey" json:"articleId"`
	TagID     uint `gorm:"primaryKey;index" json:"tagId"`
}
