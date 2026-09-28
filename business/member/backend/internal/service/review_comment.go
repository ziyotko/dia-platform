package service

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// reviewCommentMaxRunes 与 models.Application.ReviewComment /
// models.Article.ReviewComment 列的 size(500) 保持一致。
const reviewCommentMaxRunes = 500

// normalizeReviewComment 统一「入会审核」与「文章审核」的审核意见口径：
//   - 去掉首尾空白；
//   - 拒绝时必填（不能只填空格），否则返回「请填写拒绝理由」；
//   - 长度不超过 500 字（按字符数计，避免中文被列宽截断）。
//
// 前端弹窗只做提示性校验，真正的兜底以本函数为准。
func normalizeReviewComment(approved bool, comment string) (string, error) {
	comment = strings.TrimSpace(comment)
	if !approved && comment == "" {
		return "", errors.New("请填写拒绝理由")
	}
	if utf8.RuneCountInString(comment) > reviewCommentMaxRunes {
		return "", errors.New("审核意见不能超过 500 个字")
	}
	return comment, nil
}
