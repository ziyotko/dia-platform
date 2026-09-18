package service

import (
	"encoding/json"
	"errors"
	"strings"

	"member/internal/models"
	"member/pkg/db"
	"member/pkg/utils"

	"gorm.io/gorm"
)

// 协会章程相关配置键（均存放在 member_system_configs 中）。
// 章程是单篇文档，用系统配置表的键值存储即可，无需为单篇文档单独建表。
// 注意：system_configs.value 列类型为 text（64KB），足以容纳章程正文。
const (
	// CharterConfigKey 章程正文（富文本 HTML）
	CharterConfigKey = "charter_content"
	// CharterConfigDesc 正文配置项说明（后台「系统管理」列表中可见）
	CharterConfigDesc = "协会章程（富文本 HTML，请在后台「协会章程」页面维护）"
	// CharterFileConfigKey 章程 PDF 附件元信息（JSON）
	CharterFileConfigKey = "charter_file"
)

// CharterFileMeta 章程 PDF 附件元信息，序列化为 JSON 存进配置值。
type CharterFileMeta struct {
	Path       string `json:"path"`        // 相对仓库根的路径，形如 uploads/charter/2026-09/xxx.pdf
	Name       string `json:"name"`        // 上传时的原始文件名
	Size       int64  `json:"size"`        // 字节数
	UploadedAt string `json:"uploaded_at"` // 上传时间（RFC3339）
}

type CharterService struct{}

// GetCharter 返回协会章程的富文本正文；尚未维护过时返回空串（由前端展示空状态）。
func (s *CharterService) GetCharter() (string, error) {
	cfg, err := s.getConfig(CharterConfigKey)
	if err != nil || cfg == nil {
		return "", err
	}
	return cfg.Value, nil
}

// SaveCharter 保存协会章程正文：配置项不存在时自动创建（upsert）。
// 正文为富文本 HTML，入库前消毒（公开页面会用 v-html 渲染）。
func (s *CharterService) SaveCharter(content string) error {
	return s.upsertConfig(CharterConfigKey, utils.SanitizeRichText(content), CharterConfigDesc)
}

// GetCharterFile 返回章程 PDF 附件元信息。
// 从未上传过、配置为空或 JSON 损坏时返回 nil（不报错，由调用方回退到历史固定路径）。
func (s *CharterService) GetCharterFile() (*CharterFileMeta, error) {
	cfg, err := s.getConfig(CharterFileConfigKey)
	if err != nil || cfg == nil {
		return nil, err
	}
	if strings.TrimSpace(cfg.Value) == "" {
		return nil, nil
	}
	var meta CharterFileMeta
	if err := json.Unmarshal([]byte(cfg.Value), &meta); err != nil {
		return nil, nil
	}
	return &meta, nil
}

// SaveCharterFile 保存（覆盖）章程 PDF 附件元信息。
func (s *CharterService) SaveCharterFile(meta CharterFileMeta) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return s.upsertConfig(CharterFileConfigKey, string(data), "协会章程 PDF 附件（请在后台「协会章程」页面上传）")
}

// ClearCharterFile 清空章程 PDF 附件元信息（磁盘文件由调用方负责删除）。
func (s *CharterService) ClearCharterFile() error {
	return db.DB.Where("`key` = ?", CharterFileConfigKey).Delete(&models.SystemConfig{}).Error
}

// getConfig 按键取配置，不存在时返回 (nil, nil)。
func (s *CharterService) getConfig(key string) (*models.SystemConfig, error) {
	var cfg models.SystemConfig
	err := db.DB.Where("`key` = ?", key).First(&cfg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

// upsertConfig 按 key 写入配置值，不存在则创建。
func (s *CharterService) upsertConfig(key, value, desc string) error {
	cfg, err := s.getConfig(key)
	if err != nil {
		return err
	}
	if cfg == nil {
		return db.DB.Create(&models.SystemConfig{Key: key, Value: value, Description: desc}).Error
	}
	return db.DB.Model(&models.SystemConfig{}).Where("id = ?", cfg.ID).
		Update("value", value).Error
}
