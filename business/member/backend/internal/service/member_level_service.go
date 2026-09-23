package service

import (
	"errors"
	"member/internal/models"
	"member/pkg/db"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type MemberLevelService struct{}

// List returns all member levels ordered by level asc
func (s *MemberLevelService) List() ([]models.MemberLevel, error) {
	var list []models.MemberLevel
	if err := db.DB.Order("level ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// Get returns a single member level
func (s *MemberLevelService) Get(id uint64) (*models.MemberLevel, error) {
	var l models.MemberLevel
	if err := db.DB.First(&l, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("会员等级不存在")
		}
		return nil, err
	}
	return &l, nil
}

// Create creates a new member level
func (s *MemberLevelService) Create(req MemberLevelRequest) (*models.MemberLevel, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("等级名称不能为空")
	}
	// 等级名称唯一：重名会让「按名称反查等级」的历史数据与前端展示产生歧义
	var dup int64
	if err := db.DB.Model(&models.MemberLevel{}).Where("name = ?", name).Count(&dup).Error; err != nil {
		return nil, err
	}
	if dup > 0 {
		return nil, errors.New("该等级名称已存在")
	}

	// Auto-assign next level number
	var max models.MemberLevel
	db.DB.Order("level DESC").First(&max)
	nextLevel := max.Level + 1

	l := models.MemberLevel{
		Name:        name,
		Level:       nextLevel,
		Description: req.Description,
	}
	if err := db.DB.Create(&l).Error; err != nil {
		return nil, err
	}
	return &l, nil
}

// Update updates a member level
func (s *MemberLevelService) Update(id uint64, req MemberLevelRequest) error {
	var l models.MemberLevel
	if err := db.DB.First(&l, id).Error; err != nil {
		return errors.New("会员等级不存在")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errors.New("等级名称不能为空")
	}
	var dup int64
	if err := db.DB.Model(&models.MemberLevel{}).Where("name = ? AND id <> ?", name, id).Count(&dup).Error; err != nil {
		return err
	}
	if dup > 0 {
		return errors.New("该等级名称已存在")
	}
	return db.DB.Model(&models.MemberLevel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":        name,
		"description": req.Description,
	}).Error
}

// Delete deletes a member level.
// 会员等级 ID 以字符串形式保存在 member_users.member_level 中，
// 会费记录/证书/机构等级配置/证书模板也通过 level_id 引用该等级，
// 删除前必须校验是否仍被占用，避免产生悬空引用。
func (s *MemberLevelService) Delete(id uint64) error {
	var l models.MemberLevel
	if err := db.DB.First(&l, id).Error; err != nil {
		return errors.New("会员等级不存在")
	}

	// 1) 会员正在使用该等级（member_level 存的是等级 ID 字符串）
	var memberCount int64
	if err := db.DB.Model(&models.Member{}).
		Where("member_level = ?", strconv.FormatUint(l.ID, 10)).
		Count(&memberCount).Error; err != nil {
		return err
	}
	if memberCount > 0 {
		return errors.New("该等级下有会员，无法删除")
	}

	// 2) 其它引用该等级的记录
	refs := []struct {
		model interface{}
		msg   string
	}{
		{&models.FeeRecord{}, "该等级已被会费记录使用，无法删除"},
		{&models.Certificate{}, "该等级已被会员证书使用，无法删除"},
		{&models.MemberOrgLevel{}, "该等级已被机构等级配置使用，无法删除"},
		{&models.MemberCertificateTemplate{}, "该等级已被证书模板使用，无法删除"},
	}
	for _, r := range refs {
		var cnt int64
		if err := db.DB.Model(r.model).Where("level_id = ?", l.ID).Count(&cnt).Error; err != nil {
			return err
		}
		if cnt > 0 {
			return errors.New(r.msg)
		}
	}

	return db.DB.Delete(&models.MemberLevel{}, id).Error
}

// MoveUp decreases the level order (moves up in rank)
func (s *MemberLevelService) MoveUp(id uint64) error {
	var current models.MemberLevel
	if err := db.DB.First(&current, id).Error; err != nil {
		return errors.New("会员等级不存在")
	}
	if current.Level <= 0 {
		return errors.New("已是最高等级")
	}

	var above models.MemberLevel
	if err := db.DB.Where("level < ?", current.Level).Order("level DESC").First(&above).Error; err != nil {
		return errors.New("已是最高等级")
	}

	tx := db.DB.Begin()
	if err := tx.Model(&current).Update("level", above.Level).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&above).Update("level", current.Level).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// MoveDown increases the level order (moves down in rank)
func (s *MemberLevelService) MoveDown(id uint64) error {
	var current models.MemberLevel
	if err := db.DB.First(&current, id).Error; err != nil {
		return errors.New("会员等级不存在")
	}

	var below models.MemberLevel
	if err := db.DB.Where("level > ?", current.Level).Order("level ASC").First(&below).Error; err != nil {
		return errors.New("已是最低等级")
	}

	tx := db.DB.Begin()
	if err := tx.Model(&current).Update("level", below.Level).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Model(&below).Update("level", current.Level).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

type MemberLevelRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}
