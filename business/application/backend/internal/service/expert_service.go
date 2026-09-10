package service

import (
	"errors"

	"application/internal/models"
	"application/pkg/db"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ExpertService struct{}

type ExpertRequest struct {
	ID           uint64 `json:"id"`
	Name         string `json:"name"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Specialty    string `json:"specialty"`
	Title        string `json:"title"`
	Organization string `json:"organization"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Bio          string `json:"bio"`
}

func (s *ExpertService) List(page, size int, keyword, status string) ([]models.Expert, int64, error) {
	var list []models.Expert
	var total int64
	query := db.DB.Model(&models.Expert{})
	if keyword != "" {
		query = query.Where("name LIKE ? OR username LIKE ? OR specialty LIKE ? OR organization LIKE ? OR title LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status == "enabled" {
		query = query.Where("status = 1")
	} else if status == "disabled" {
		query = query.Where("status = 0")
	}
	query.Count(&total)
	err := query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	if err != nil {
		return list, total, err
	}
	s.fillReviewCounts(list)
	return list, total, nil
}

func (s *ExpertService) fillReviewCounts(list []models.Expert) {
	for i := range list {
		if list[i].AdminID == 0 {
			continue
		}
		db.DB.Model(&models.ReviewAssignment{}).Where("reviewer_id = ?", list[i].AdminID).Count(&list[i].ReviewCount)
		db.DB.Model(&models.ReviewAssignment{}).
			Where("reviewer_id = ? AND status = ?", list[i].AdminID, models.ReviewStatusScored).
			Count(&list[i].ScoredCount)
	}
}

// Create creates an expert profile and a linked reviewer login account
func (s *ExpertService) Create(req ExpertRequest) error {
	if req.Name == "" || req.Username == "" || req.Password == "" {
		return errors.New("请填写姓名、登录账号和密码")
	}
	var count int64
	db.DB.Model(&models.Expert{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return errors.New("该登录账号已存在")
	}
	db.DB.Model(&models.Admin{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return errors.New("该登录账号已存在")
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	admin := models.Admin{
		Username: req.Username,
		Password: string(hashed),
		RealName: req.Name,
		Phone:    req.Phone,
		Email:    req.Email,
		RoleCode: models.RoleReviewer,
		Status:   1,
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&admin).Error; err != nil {
			return err
		}
		expert := models.Expert{
			AdminID:      admin.ID,
			Name:         req.Name,
			Username:     req.Username,
			Specialty:    req.Specialty,
			Title:        req.Title,
			Organization: req.Organization,
			Phone:        req.Phone,
			Email:        req.Email,
			Bio:          req.Bio,
			Status:       1,
		}
		return tx.Create(&expert).Error
	})
}

func (s *ExpertService) Update(id uint64, req ExpertRequest) error {
	var expert models.Expert
	if err := db.DB.First(&expert, id).Error; err != nil {
		return errors.New("专家不存在")
	}

	updates := map[string]interface{}{
		"name":         req.Name,
		"specialty":    req.Specialty,
		"title":        req.Title,
		"organization": req.Organization,
		"phone":        req.Phone,
		"email":        req.Email,
		"bio":          req.Bio,
	}
	if err := db.DB.Model(&expert).Updates(updates).Error; err != nil {
		return err
	}

	adminUpdates := map[string]interface{}{
		"real_name": req.Name,
		"phone":     req.Phone,
		"email":     req.Email,
	}
	if req.Password != "" {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		adminUpdates["password"] = string(hashed)
	}
	return db.DB.Model(&models.Admin{}).Where("id = ?", expert.AdminID).Updates(adminUpdates).Error
}

func (s *ExpertService) Delete(id uint64) error {
	var expert models.Expert
	if err := db.DB.First(&expert, id).Error; err != nil {
		return errors.New("专家不存在")
	}
	var count int64
	db.DB.Model(&models.ReviewAssignment{}).Where("reviewer_id = ?", expert.AdminID).Count(&count)
	if count > 0 {
		return errors.New("该专家存在评审任务，无法删除")
	}
	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", expert.AdminID).Delete(&models.Admin{}).Error; err != nil {
			return err
		}
		return tx.Delete(&expert).Error
	})
}

func (s *ExpertService) SetStatus(id uint64, status int) error {
	var expert models.Expert
	if err := db.DB.First(&expert, id).Error; err != nil {
		return errors.New("专家不存在")
	}
	if err := db.DB.Model(&expert).Update("status", status).Error; err != nil {
		return err
	}
	return db.DB.Model(&models.Admin{}).Where("id = ?", expert.AdminID).Update("status", status).Error
}
