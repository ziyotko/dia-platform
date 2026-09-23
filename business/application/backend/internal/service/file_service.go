package service

import (
	"application/internal/models"
	"application/pkg/db"
)

type FileService struct{}

// CanUserRead reports whether the applicant may download a stored file: it must
// be a material of one of their own applications, or the attachment of a
// certificate issued to them. Uploads are no longer public (静态托管已移除),
// so every read goes through this check.
func (s *FileService) CanUserRead(userID uint64, fileURL string) bool {
	if userID == 0 || fileURL == "" {
		return false
	}
	var count int64
	db.DB.Model(&models.ApplicationMaterial{}).
		Joins("JOIN application_applications AS a ON a.id = application_materials.application_id").
		Where("application_materials.file_url = ? AND a.user_id = ?", fileURL, userID).
		Count(&count)
	if count > 0 {
		return true
	}
	count = 0
	db.DB.Model(&models.Certificate{}).
		Where("file_url = ? AND user_id = ?", fileURL, userID).
		Count(&count)
	return count > 0
}

// CanAdminRead reports whether the operator may download a stored file.
//
// 管理人/超管：任意文件（他们要处理全部申报与证书）。
// 评审人：只能看**分配给自己**的申报的材料——评审人已不再持有
// application:view 权限，所以这里按归属判断，而不是按权限码。
func (s *FileService) CanAdminRead(adminID uint64, roleCode, fileURL string) bool {
	if adminID == 0 || fileURL == "" {
		return false
	}
	if roleCode != models.RoleReviewer {
		return true
	}
	var count int64
	db.DB.Model(&models.ApplicationMaterial{}).
		Joins("JOIN application_review_assignments AS ra ON ra.application_id = application_materials.application_id").
		Where("application_materials.file_url = ? AND ra.reviewer_id = ?", fileURL, adminID).
		Count(&count)
	return count > 0
}
