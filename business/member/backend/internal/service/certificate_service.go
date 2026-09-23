package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"member/config"
	"member/internal/models"
	"member/pkg/certpdf"
	"member/pkg/db"
	"member/pkg/utils"

	"gorm.io/gorm"
)

type CertificateService struct{}

// GetMyCertificates returns the member's certificates
func (s *CertificateService) GetMyCertificates(memberID uint64) ([]models.Certificate, error) {
	var certs []models.Certificate
	if err := db.DB.Where("member_id = ?", memberID).Order("created_at DESC").Find(&certs).Error; err != nil {
		return nil, err
	}
	return certs, nil
}

// GetCertificate returns a certificate by ID (owner only)
func (s *CertificateService) GetCertificate(id, memberID uint64) (*models.Certificate, error) {
	var cert models.Certificate
	if err := db.DB.Preload("Member").Where("id = ? AND member_id = ?", id, memberID).First(&cert).Error; err != nil {
		return nil, errors.New("证书不存在")
	}
	return &cert, nil
}

// CreateCertificate creates a certificate (admin)
func (s *CertificateService) CreateCertificate(req CreateCertRequest) (*models.Certificate, error) {
	cert := models.Certificate{
		MemberID: req.MemberID,
		CertNo:   req.CertNo,
		IssuedAt: req.IssuedAt,
		ExpireAt: req.ExpireAt,
		FilePath: req.FilePath,
		Status:   models.CertStatusActive,
	}
	if err := db.DB.Create(&cert).Error; err != nil {
		return nil, err
	}
	return &cert, nil
}

// UpdateCertificate updates a certificate (admin).
// filePath/status 使用指针区分“未提供”（nil）与“清空”（指向空字符串）。
func (s *CertificateService) UpdateCertificate(id uint64, filePath, status *string) error {
	updates := map[string]interface{}{}
	if filePath != nil {
		updates["file_path"] = *filePath
	}
	if status != nil {
		updates["status"] = *status
	}
	if len(updates) == 0 {
		return nil
	}
	return db.DB.Model(&models.Certificate{}).Where("id = ?", id).Updates(updates).Error
}

// GenerateCertificateForMember generates a certificate for a member (admin/renew):
// 作废原生效证书 → 建新证书 → 生成 PDF 并写入 file_path。
// 等级取会员当前等级；没有等级时证书仍会发出，只是等级为空。
func (s *CertificateService) GenerateCertificateForMember(memberID uint64) (*models.Certificate, error) {
	return s.CreateCertificateForMember(memberID, 0, "")
}

// CreateCertificateForMember 创建并落盘会员证书。
// levelID/levelName 由调用方确定（如新增会员/审批通过时选定的等级）；为 0 时回退会员当前等级。
// 返回的证书已含 file_path（PDF 生成失败时 file_path 为空，同时通过日志告警）。
func (s *CertificateService) CreateCertificateForMember(memberID, levelID uint64, levelName string) (*models.Certificate, error) {
	var member models.Member
	if err := db.DB.First(&member, memberID).Error; err != nil {
		return nil, errors.New("会员不存在")
	}
	if levelID == 0 {
		levelID, levelName = resolveMemberLevel(member.MemberLevel)
	}
	if levelName == "" {
		levelName = memberLevelNameByID(levelID)
	}

	cert, err := createCertificateRow(db.DB, memberID, levelID, levelName)
	if err != nil {
		return nil, err
	}

	if err := s.GenerateFileForCertificate(cert); err != nil {
		// PDF 生成失败不阻断发证：证书记录仍可用，管理员可在「证书管理」重新生成
		utils.LogWarn("生成证书 PDF 失败（cert_id=%d）：%v", cert.ID, err)
	}
	return cert, nil
}

// createCertificateRow 在给定连接/事务中作废旧生效证书并写入新证书行（不生成 PDF）。
// 会员证书的 DB 写入统一走这里，保证「同一会员同时只有一张生效证书」。
// q 可以是 db.DB（独立发证），也可以是事务对象（新增会员/审批通过时与其它写入同事务）。
func createCertificateRow(q *gorm.DB, memberID, levelID uint64, levelName string) (*models.Certificate, error) {
	now := time.Now()
	if err := q.Model(&models.Certificate{}).
		Where("member_id = ? AND status = ?", memberID, models.CertStatusActive).
		Update("status", models.CertStatusExpired).Error; err != nil {
		return nil, err
	}

	cert := models.Certificate{
		MemberID:       memberID,
		CertNo:         generateCertNo(memberID),
		IssuedAt:       &models.LocalTime{Time: now},
		ExpireAt:       &models.LocalTime{Time: time.Date(now.Year(), 12, 31, 23, 59, 59, 0, now.Location())},
		Status:         models.CertStatusActive,
		LevelID:        levelID,
		LevelName:      levelName,
		CertTemplateID: certTemplateIDForLevel(q, levelID),
	}
	if err := q.Create(&cert).Error; err != nil {
		return nil, err
	}
	return &cert, nil
}

// GenerateFileForCertificate 渲染证书 PDF 并把相对路径回写到 member_certificates.file_path。
// 供发证、审批通过、续证、管理员重新生成等入口统一调用。
func (s *CertificateService) GenerateFileForCertificate(cert *models.Certificate) error {
	if cert == nil || cert.ID == 0 {
		return errors.New("证书不存在")
	}
	var member models.Member
	if err := db.DB.First(&member, cert.MemberID).Error; err != nil {
		return errors.New("会员不存在")
	}

	issuedAt := time.Now()
	if cert.IssuedAt != nil && !cert.IssuedAt.Time.IsZero() {
		issuedAt = cert.IssuedAt.Time
	}
	expireAt := time.Date(issuedAt.Year(), 12, 31, 23, 59, 59, 0, issuedAt.Location())
	if cert.ExpireAt != nil && !cert.ExpireAt.Time.IsZero() {
		expireAt = cert.ExpireAt.Time
	}

	fontPath := ""
	if config.Cfg != nil {
		fontPath = config.Cfg.Certificate.FontPath
	}

	renderData := certpdf.Data{
		Title:      certificatePDFTitle(cert),
		OrgName:    certificateOrgName(&member),
		IssuerName: certificateIssuerName(),
		MemberName: memberDisplayName(&member),
		MemberType: memberTypeLabel(member.MemberType),
		LevelName:  cert.LevelName,
		CertNo:     cert.CertNo,
		IssuedAt:   issuedAt,
		ExpireAt:   expireAt,
	}

	// 优先按该等级上传的 PDF 模板套打；模板缺失/无法解析时回退到默认版式
	var (
		data     []byte
		renderEr error
	)
	if tplFile, ok := certificateTemplateFile(cert.CertTemplateID); ok {
		data, renderEr = certpdf.GenerateWithTemplate(tplFile, renderData, fontPath)
		if renderEr != nil {
			utils.LogWarn("按模板套打证书失败，回退默认版式（cert_id=%d, template=%s）：%v", cert.ID, tplFile, renderEr)
			data = nil
		}
	}
	if data == nil {
		data, renderEr = certpdf.Generate(renderData, fontPath)
		if renderEr != nil {
			return renderEr
		}
	}

	path, err := utils.SaveBytes("certificates", certificateFileName(cert), data)
	if err != nil {
		return err
	}
	if err := db.DB.Model(&models.Certificate{}).Where("id = ?", cert.ID).
		Update("file_path", path).Error; err != nil {
		return err
	}
	cert.FilePath = path
	return nil
}

// certificateTemplateFile 返回该证书样式模板的本地文件路径；
// 未配置样式、文件不存在或路径非法时返回 ok=false（调用方回退默认版式）。
func certificateTemplateFile(templateID uint64) (string, bool) {
	if templateID == 0 {
		return "", false
	}
	var tpl models.MemberCertificateTemplate
	if err := db.DB.First(&tpl, templateID).Error; err != nil {
		return "", false
	}
	return utils.LocalUploadPath(tpl.TemplateFile)
}

// ListCertificates 分页返回证书发放记录（admin）。
func (s *CertificateService) ListCertificates(page, size int, keyword, status string) ([]models.Certificate, int64, error) {
	var list []models.Certificate
	var total int64

	query := db.DB.Model(&models.Certificate{}).Preload("Member")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		kw := "%" + keyword + "%"
		query = query.Where(
			"cert_no LIKE ? OR level_name LIKE ? OR member_id IN (SELECT id FROM member_users WHERE username LIKE ? OR company_name LIKE ? OR name LIKE ?)",
			kw, kw, kw, kw, kw)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// certificatePDFTitle 证书标题：优先使用该等级配置的证书样式名称。
func certificatePDFTitle(cert *models.Certificate) string {
	if cert.CertTemplateID > 0 {
		var tpl models.MemberCertificateTemplate
		if err := db.DB.First(&tpl, cert.CertTemplateID).Error; err == nil && strings.TrimSpace(tpl.Name) != "" {
			return tpl.Name
		}
	}
	return "会员证书"
}

// certificateOrgName 证书上的入会机构：会员的主入会机构。
func certificateOrgName(member *models.Member) string {
	if _, name := (&MemberService{}).primaryOrg(member); name != "" {
		return name
	}
	return certificateIssuerName()
}

// certificateIssuerName 发证机构：总会（parent_id = 0 的顶级机构）名称。
func certificateIssuerName() string {
	var root models.Organization
	if err := db.DB.Where("parent_id = 0").Order("id ASC").First(&root).Error; err == nil {
		return root.Name
	}
	return ""
}

// certTemplateIDForLevel 取该等级配置的证书样式，未配置时回退到最低等级样式。
// q 与写入方保持一致（事务内外均可）。
func certTemplateIDForLevel(q *gorm.DB, levelID uint64) uint64 {
	var tpl models.MemberCertificateTemplate
	if levelID > 0 {
		if err := q.Where("level_id = ?", levelID).First(&tpl).Error; err == nil {
			return tpl.ID
		}
	}
	q.Joins("JOIN member_levels ml ON ml.id = member_certificate_templates.level_id").
		Order("ml.level ASC").
		First(&tpl)
	return tpl.ID
}

// certificateFileName 证书 PDF 文件名：证书编号（去掉不安全字符）+ 证书 ID，保证唯一。
func certificateFileName(cert *models.Certificate) string {
	base := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			return r
		default:
			return -1
		}
	}, cert.CertNo)
	if base == "" {
		base = "cert"
	}
	return fmt.Sprintf("%s-%d.pdf", base, cert.ID)
}

// RenewMyCertificate marks old certificates as expired and creates a new one (member self-service).
// 仅正式会员且当年已缴费方可续证，防止任意会员刷证。
func (s *CertificateService) RenewMyCertificate(memberID uint64) (*models.Certificate, error) {
	var member models.Member
	if err := db.DB.First(&member, memberID).Error; err != nil {
		return nil, errors.New("会员不存在")
	}
	if member.Status != models.MemberStatusActive {
		return nil, errors.New("仅正式会员可续证")
	}

	// 校验当年已缴费
	var paidCount int64
	if err := db.DB.Model(&models.FeeRecord{}).
		Where("member_id = ? AND year = ? AND status = ?", memberID, time.Now().Year(), models.FeeStatusPaid).
		Count(&paidCount).Error; err != nil {
		return nil, err
	}
	if paidCount == 0 {
		return nil, errors.New("当年尚未缴费，暂不能续证")
	}

	// 频次限制：已有生效且文件正常时不允许重复续证（前端也只在无生效证书时展示入口）。
	// 例外：生效证书的 PDF 未生成（生成失败）时允许重试，避免会员拿不到文件。
	var activeCert models.Certificate
	if err := db.DB.Where("member_id = ? AND status = ?", memberID, models.CertStatusActive).
		Order("id DESC").First(&activeCert).Error; err == nil {
		if strings.TrimSpace(activeCert.FilePath) != "" {
			return nil, errors.New("当前已有生效证书，无需重复续证")
		}
	}

	// 重新发证：作废原生效证书 + 生成带等级/样式的证书与 PDF（由 CreateCertificateForMember 完成）
	cert, err := s.CreateCertificateForMember(memberID, 0, "")
	if err != nil {
		return nil, err
	}

	// 续证成功即写入“证书续期”会籍记录（年度延续，等级不变）。
	// 续证接口可重复调用（每次都会重发证书），但会籍记录按“每会员每年一条”幂等，
	// 避免同一年度重复续证时刷出多条相同的会籍记录。
	year := time.Now().Year()
	if !membershipChangeExists(memberID, models.ReasonCertRenew, year) {
		levelID, levelName := resolveMemberLevel(member.MemberLevel)
		orgID, orgName := (&MemberService{}).primaryOrg(&member)
		_ = writeMembershipChange(&member, year, orgID, orgName,
			levelID, levelName, levelID, levelName, models.ReasonCertRenew, member.Username)
	}

	return cert, nil
}

// RegenerateMissingCertificates 批量补生成 file_path 为空的证书 PDF（历史存量数据）。
// limit 为单次处理上限（<=0 或 >500 时取 200，避免一次请求处理过多）；
// 返回：成功数、失败数、失败明细。单张失败不影响其它证书。
func (s *CertificateService) RegenerateMissingCertificates(limit int) (int, int, []string, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	var list []models.Certificate
	if err := db.DB.Where("file_path IS NULL OR file_path = ''").Order("id ASC").Limit(limit).Find(&list).Error; err != nil {
		return 0, 0, nil, err
	}

	ok, failed := 0, 0
	failures := make([]string, 0)
	for i := range list {
		if err := s.GenerateFileForCertificate(&list[i]); err != nil {
			failed++
			failures = append(failures, fmt.Sprintf("#%d %s：%v", list[i].ID, list[i].CertNo, err))
			continue
		}
		ok++
	}
	return ok, failed, failures, nil
}

type CreateCertRequest struct {
	MemberID uint64            `json:"member_id" binding:"required"`
	CertNo   string            `json:"cert_no" binding:"required"`
	IssuedAt *models.LocalTime `json:"issued_at"`
	ExpireAt *models.LocalTime `json:"expire_at"`
	FilePath string            `json:"file_path"`
}
