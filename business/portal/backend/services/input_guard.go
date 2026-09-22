package services

import (
	"fmt"
	"strings"

	"server/models"
	"server/utils"
)

// 本文件集中存放「写入前的输入校验」，避免各 service 散落重复的查重/上限逻辑。

// uniqueScope 额外的作用域条件（例如「同一模板内唯一」）
type uniqueScope struct {
	Field string
	Value uint
}

// ensureValueUnique 单字段唯一性校验：value 为空则跳过（可空字段）；excludeID > 0 时排除自身（更新场景）。
func ensureValueUnique(model any, field, value, entity, label string, excludeID uint, scope *uniqueScope) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	query := utils.DB.Model(model).Where(field+" = ?", value)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if scope != nil {
		query = query.Where(scope.Field+" = ?", scope.Value)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%s%s「%s」已存在", entity, label, value)
	}
	return nil
}

// ensureNameCodeUnique 名称 + 编码双字段唯一性校验（编码可空，仅非空时校验）。
// 背景：category / tag / column 表都没有唯一索引，重名会让统计页按名称聚合时互相混淆、
// 单页静态化按栏目名称定位时生成同一个静态文件，因此在写入前显式拦截并给出可操作文案。
func ensureNameCodeUnique(model any, entity, name, code string, excludeID uint, scope *uniqueScope) error {
	if err := ensureValueUnique(model, "name", name, entity, "名称", excludeID, scope); err != nil {
		return err
	}
	return ensureValueUnique(model, "code", code, entity, "编码", excludeID, scope)
}

// maxArticleAttachments 单篇文章的附件数量上限（前端 el-upload 也是 10，这里做服务端兜底）
const maxArticleAttachments = 10

// validateAttachmentLimit 附件数量上限校验（按去重后的 URL 计，与落库口径一致）
func validateAttachmentLimit(attachments []models.ArticleAttachment) error {
	seen := make(map[string]bool)
	count := 0
	for _, att := range attachments {
		if strings.TrimSpace(att.URL) == "" || seen[att.URL] {
			continue
		}
		seen[att.URL] = true
		count++
	}
	if count > maxArticleAttachments {
		return fmt.Errorf("附件数量不能超过 %d 个", maxArticleAttachments)
	}
	return nil
}
