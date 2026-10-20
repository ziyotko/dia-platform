package service

import (
	"errors"
	"strconv"
	"time"

	"member/config"
	"member/internal/models"
	"member/pkg/db"
	"member/pkg/utils"

	"gorm.io/gorm"
)

// MigrateFeesDisabledData 将历史收费会籍投影到中性的会籍关系中。
// 所有写入都有存在性判断，可在每次启动时安全重复执行；历史费用和发票数据不做修改。
func MigrateFeesDisabledData() {
	if config.Cfg.Features.FeesEnabled {
		return
	}
	resolved := make(map[uint64]bool)

	var paid []models.FeeRecord
	if err := db.DB.Where("status = ? AND org_id > 0", models.FeeStatusPaid).Find(&paid).Error; err != nil {
		utils.LogWarn("会费关闭迁移：读取历史缴费记录失败：%v", err)
	} else {
		for _, fee := range paid {
			var member models.Member
			if err := db.DB.First(&member, fee.MemberID).Error; err != nil {
				continue
			}
			levelID, levelName := fee.LevelID, fee.LevelName
			if levelID == 0 {
				levelID, levelName = resolveMemberLevel(member.MemberLevel)
			}
			if levelID == 0 {
				utils.LogWarn("会费关闭迁移待处理：历史缴费 member_id=%d org_id=%d 缺少等级", fee.MemberID, fee.OrgID)
				continue
			}
			var newCert *models.Certificate
			if err := db.DB.Transaction(func(tx *gorm.DB) error {
				if err := ensureMemberOrganizationTx(tx, fee.MemberID, fee.OrgID, levelID, fee.CreatedAt); err != nil {
					return err
				}
				updates := map[string]interface{}{}
				if member.MemberLevel == "" {
					updates["member_level"] = strconv.FormatUint(levelID, 10)
				}
				if member.Status == models.MemberStatusPendingPay {
					updates["status"] = models.MemberStatusActive
				}
				if len(updates) > 0 {
					if err := tx.Model(&models.Member{}).Where("id = ?", member.ID).Updates(updates).Error; err != nil {
						return err
					}
				}
				var changeCount int64
				tx.Model(&models.MemberLevelChange{}).Where("member_id = ? AND reason = ?", member.ID, models.ReasonMemberCreate).Count(&changeCount)
				if changeCount == 0 {
					if err := writeMembershipChangeTx(tx, &member, fee.Year, fee.OrgID, fee.OrgName,
						0, "", levelID, levelName, models.ReasonMemberCreate, "系统迁移"); err != nil {
						return err
					}
				}
				if member.Status == models.MemberStatusActive || member.Status == models.MemberStatusPendingPay {
					cert, err := ensureMigrationCertificateTx(tx, member.ID, levelID, levelName)
					if err != nil {
						return err
					}
					newCert = cert
				}
				return nil
			}); err != nil {
				utils.LogWarn("会费关闭迁移：缴费会籍回填失败 member_id=%d org_id=%d：%v", fee.MemberID, fee.OrgID, err)
			} else {
				resolved[member.ID] = true
				if newCert != nil {
					if err := (&CertificateService{}).GenerateFileForCertificate(newCert); err != nil {
						utils.LogWarn("会费关闭迁移：证书 PDF 生成失败 cert_id=%d：%v", newCert.ID, err)
					}
				}
			}
		}
	}

	var apps []models.Application
	if err := db.DB.Where("status = ?", models.AppStatusApproved).Order("id ASC").Find(&apps).Error; err != nil {
		utils.LogWarn("会费关闭迁移：读取已通过申请失败：%v", err)
		return
	}
	for _, app := range apps {
		var member models.Member
		if err := db.DB.First(&member, app.MemberID).Error; err != nil {
			continue
		}
		var org models.Organization
		if err := db.DB.First(&org, app.OrgID).Error; err != nil {
			utils.LogWarn("会费关闭迁移待处理：member_id=%d 的申请机构 org_id=%d 不存在", member.ID, app.OrgID)
			continue
		}
		root := resolveOrgRoot(org)
		levelID, levelName := resolveMemberLevel(member.MemberLevel)
		if levelID == 0 {
			ol, err := findMinOrgLevel(root.ID)
			if err != nil && root.ID != org.ID {
				ol, err = findMinOrgLevel(org.ID)
			}
			if err != nil || ol.LevelID == 0 {
				utils.LogWarn("会费关闭迁移待处理：member_id=%d 无法确认会员等级（org_id=%d）", member.ID, org.ID)
				continue
			}
			levelID, levelName = ol.LevelID, ol.Level.Name
		}
		var newCert *models.Certificate
		err := db.DB.Transaction(func(tx *gorm.DB) error {
			if err := ensureMemberOrganizationTx(tx, member.ID, root.ID, levelID, app.CreatedAt); err != nil {
				return err
			}
			if org.ID != root.ID {
				if err := ensureMemberOrganizationTx(tx, member.ID, org.ID, levelID, app.CreatedAt); err != nil {
					return err
				}
			}
			updates := map[string]interface{}{}
			if member.MemberLevel == "" {
				updates["member_level"] = strconv.FormatUint(levelID, 10)
			}
			if member.Status == models.MemberStatusPendingPay {
				updates["status"] = models.MemberStatusActive
			}
			if len(updates) > 0 {
				if err := tx.Model(&models.Member{}).Where("id = ?", member.ID).Updates(updates).Error; err != nil {
					return err
				}
			}
			var changeCount int64
			tx.Model(&models.MemberLevelChange{}).Where("member_id = ? AND reason = ?", member.ID, models.ReasonMemberCreate).Count(&changeCount)
			if changeCount == 0 {
				if err := writeMembershipChangeTx(tx, &member, app.CreatedAt.Year(), org.ID, org.Name, 0, "", levelID, levelName, models.ReasonMemberCreate, "系统迁移"); err != nil {
					return err
				}
			}
			if member.Status == models.MemberStatusActive || member.Status == models.MemberStatusPendingPay {
				cert, err := ensureMigrationCertificateTx(tx, member.ID, levelID, levelName)
				if err != nil {
					return err
				}
				newCert = cert
			}
			return nil
		})
		if err != nil {
			utils.LogWarn("会费关闭迁移失败 member_id=%d：%v", member.ID, err)
			continue
		}
		resolved[member.ID] = true
		if newCert != nil {
			if err := (&CertificateService{}).GenerateFileForCertificate(newCert); err != nil {
				utils.LogWarn("会费关闭迁移：证书 PDF 生成失败 cert_id=%d：%v", newCert.ID, err)
			}
		}
	}

	var unresolved []models.Member
	db.DB.Where("status = ?", models.MemberStatusPendingPay).Find(&unresolved)
	for _, member := range unresolved {
		if !resolved[member.ID] {
			utils.LogWarn("会费关闭迁移待处理：member_id=%d username=%s 缺少可确认的机构或等级，保持待缴费状态", member.ID, member.Username)
		}
	}
	utils.LogInfo("会费关闭兼容迁移完成：已核对 %d 条已通过申请，时间=%s", len(apps), time.Now().Format(time.RFC3339))
}

// ensureMigrationCertificateTx 返回需要补生成 PDF 的证书；有效且文件存在时返回 nil。
func ensureMigrationCertificateTx(tx *gorm.DB, memberID, levelID uint64, levelName string) (*models.Certificate, error) {
	var cert models.Certificate
	err := tx.Where("member_id = ? AND status = ?", memberID, models.CertStatusActive).
		Order("id DESC").First(&cert).Error
	if err == nil {
		if cert.ExpireAt != nil && cert.ExpireAt.Time.Before(time.Now()) {
			return createCertificateRow(tx, memberID, levelID, levelName)
		}
		if cert.FilePath == "" {
			return &cert, nil
		}
		return nil, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return createCertificateRow(tx, memberID, levelID, levelName)
}
