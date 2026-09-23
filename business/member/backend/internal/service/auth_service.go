package service

import (
	"errors"
	"fmt"
	"member/config"
	"member/internal/models"
	"member/pkg/captcha"
	"member/pkg/db"
	mjwt "member/pkg/jwt"
	"member/pkg/utils"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// 登录失败锁定策略：连续失败 maxLoginFailCount 次后锁定 loginLockDuration
const (
	maxLoginFailCount = 5
	loginLockDuration = 30 * time.Minute
)

type AuthService struct{}

// Register registers a new member and returns a token
func (s *AuthService) Register(req RegisterRequest) (*LoginResponse, error) {
	// Verify captcha
	if !captcha.Verify(req.CaptchaID, req.CaptchaCode) {
		return nil, errors.New("验证码错误")
	}

	// 会员类型白名单 + 手机/邮箱规范化与格式校验（与后台「新增会员」同口径）：
	// 原先不校验 member_type，可注册出 member_type='x' 的会员，
	// 导致后台筛选/详情分支/类型标签全部落到异常分支。
	req.MemberType = strings.TrimSpace(req.MemberType)
	if req.MemberType == "" {
		req.MemberType = models.MemberTypeUnit
	}
	if req.MemberType != models.MemberTypeUnit && req.MemberType != models.MemberTypePersonal {
		return nil, errors.New("会员类型不正确")
	}
	req.Mobile = strings.TrimSpace(req.Mobile)
	req.Email = strings.TrimSpace(req.Email)
	if req.Mobile != "" && !utils.IsValidMobile(req.Mobile) {
		return nil, errors.New("请输入合法的手机号")
	}
	if req.Email != "" && !utils.IsValidEmail(req.Email) {
		return nil, errors.New("邮箱格式不正确")
	}

	// Check if username already exists
	var exist models.Member
	if err := db.DB.Where("username = ?", req.Username).First(&exist).Error; err == nil {
		return nil, errors.New("用户名已存在")
	}

	// Check mobile/email uniqueness (backend-level safeguard)
	if req.Mobile != "" {
		var m models.Member
		if err := db.DB.Where("mobile = ?", req.Mobile).First(&m).Error; err == nil {
			return nil, errors.New("手机号已被注册")
		}
	}
	if req.Email != "" {
		var e models.Member
		if err := db.DB.Where("email = ?", req.Email).First(&e).Error; err == nil {
			return nil, errors.New("邮箱已被注册")
		}
	}

	// 单位会员：公司名称/统一社会信用代码查重
	req.CompanyName = strings.TrimSpace(req.CompanyName)
	req.CreditCode = strings.ToUpper(strings.TrimSpace(req.CreditCode))
	if req.MemberType == models.MemberTypeUnit {
		if req.CompanyName != "" {
			var cn models.Member
			if err := db.DB.Where("company_name = ?", req.CompanyName).First(&cn).Error; err == nil {
				return nil, errors.New("该公司名称已存在")
			}
		}
		if req.CreditCode != "" {
			var cc models.Member
			if err := db.DB.Where("credit_code = ?", req.CreditCode).First(&cc).Error; err == nil {
				return nil, errors.New("该统一社会信用代码已存在")
			}
		}
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	member := models.Member{
		Username:    req.Username,
		Password:    string(hashed),
		Mobile:      req.Mobile,
		Email:       req.Email,
		MemberType:  req.MemberType,
		MemberLevel: "",
		Status:      models.MemberStatusRegistering,
	}

	// Save company info for unit members
	if req.MemberType == models.MemberTypeUnit {
		member.CompanyName = req.CompanyName
		member.CreditCode = req.CreditCode
		member.LegalPerson = req.LegalPerson
		member.ContactPerson = req.ContactPerson
		member.ContactMobile = req.ContactMobile
		member.Address = req.Address
		member.CertFile = req.CertFile
	}

	if err := db.DB.Create(&member).Error; err != nil {
		return nil, fmt.Errorf("注册失败: %w", err)
	}

	// Generate JWT token right after registration
	token, err := mjwt.GenerateToken(member.ID, member.Username, member.IsAdmin)
	if err != nil {
		return nil, errors.New("token生成失败")
	}

	return &LoginResponse{
		Token: token,
		Member: &MemberInfo{
			ID:          member.ID,
			Username:    member.Username,
			Mobile:      member.Mobile,
			Email:       member.Email,
			MemberType:  member.MemberType,
			MemberLevel: member.MemberLevel,
			Status:      member.Status,
			IsAdmin:     member.IsAdmin,
			Avatar:      member.Avatar,
			CompanyName: member.CompanyName,
		},
	}, nil
}

// CheckExists checks if the given field value is already in use
// 支持 username/mobile/email/company_name/credit_code（单位会员信息查重）
func (s *AuthService) CheckExists(field, value string) (bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return false, nil
	}
	var count int64
	query := db.DB.Model(&models.Member{})
	switch field {
	case "username", "mobile", "email", "company_name":
		query = query.Where(field+" = ?", value)
	case "credit_code":
		query = query.Where("credit_code = ?", strings.ToUpper(value))
	default:
		return false, errors.New("不支持的字段")
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ensureProfileUnique 校验资料变更后的手机号/邮箱/单位名称/统一社会信用代码
// 是否已被其他会员占用（排除自己）
func ensureProfileUnique(memberID uint64, req UpdateProfileRequest) error {
	type uniqueCheck struct {
		column  string
		value   string
		message string
	}
	checks := []uniqueCheck{
		{"mobile", req.Mobile, "该手机号已被其他用户使用"},
		{"email", req.Email, "该邮箱已被其他用户使用"},
		{"company_name", req.CompanyName, "该公司名称已被其他会员使用"},
		{"credit_code", req.CreditCode, "该统一社会信用代码已被其他会员使用"},
	}
	for _, c := range checks {
		if c.value == "" {
			continue
		}
		var count int64
		err := db.DB.Model(&models.Member{}).
			Where(c.column+" = ? AND id <> ?", c.value, memberID).
			Count(&count).Error
		if err != nil {
			return err
		}
		if count > 0 {
			return errors.New(c.message)
		}
	}
	return nil
}

// Login authenticates a member
func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
	// Verify captcha
	if !captcha.Verify(req.CaptchaID, req.CaptchaCode) {
		return nil, errors.New("验证码错误")
	}

	var member models.Member
	if err := db.DB.Where("username = ?", req.Username).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}

	// Check lock
	if member.LockedUntil != nil && member.LockedUntil.After(time.Now()) {
		return nil, fmt.Errorf("账户已锁定，请于 %s 后再试", member.LockedUntil.Format("15:04:05"))
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(member.Password), []byte(req.Password)); err != nil {
		// 原子自增后回读判断：原先 `member.LoginFailCount++` 后整行 Save，
		// 并发爆破会丢更新（绕过 5 次锁定），且整行保存会覆盖其它并发写入的字段。
		if err := db.DB.Model(&models.Member{}).Where("id = ?", member.ID).
			UpdateColumn("login_fail_count", gorm.Expr("login_fail_count + 1")).Error; err != nil {
			return nil, errors.New("用户名或密码错误")
		}
		var fresh models.Member
		if err := db.DB.Select("login_fail_count").First(&fresh, member.ID).Error; err == nil &&
			fresh.LoginFailCount >= maxLoginFailCount {
			lockUntil := &models.LocalTime{Time: time.Now().Add(loginLockDuration)}
			db.DB.Model(&models.Member{}).Where("id = ?", member.ID).Update("locked_until", lockUntil)
		}
		return nil, errors.New("用户名或密码错误")
	}

	// Reset fail count on success（只更新这两列，不整行保存）
	if member.LoginFailCount > 0 || member.LockedUntil != nil {
		db.DB.Model(&models.Member{}).Where("id = ?", member.ID).Updates(map[string]any{
			"login_fail_count": 0,
			"locked_until":     nil,
		})
	}

	// Generate JWT
	token, err := mjwt.GenerateToken(member.ID, member.Username, member.IsAdmin)
	if err != nil {
		return nil, errors.New("token生成失败")
	}

	return &LoginResponse{
		Token: token,
		Member: &MemberInfo{
			ID:          member.ID,
			Username:    member.Username,
			Mobile:      member.Mobile,
			Email:       member.Email,
			MemberType:  member.MemberType,
			MemberLevel: member.MemberLevel,
			Status:      member.Status,
			IsAdmin:     member.IsAdmin,
			Avatar:      member.Avatar,
			CompanyName: member.CompanyName,
		},
	}, nil
}

// GetProfile returns the member profile
func (s *AuthService) GetProfile(memberID uint64) (*models.Member, error) {
	var member models.Member
	if err := db.DB.First(&member, memberID).Error; err != nil {
		return nil, errors.New("会员不存在")
	}
	return &member, nil
}

// UpdateProfile updates the member profile and records a profile change
// (资料变更记录) for every tracked field that changed. Passwords are never recorded.
func (s *AuthService) UpdateProfile(memberID uint64, req UpdateProfileRequest) error {
	var member models.Member
	if err := db.DB.First(&member, memberID).Error; err != nil {
		return errors.New("会员不存在")
	}

	// 规范化后查重（排除自己）：手机号/邮箱/单位名称/统一社会信用代码
	req.Mobile = strings.TrimSpace(req.Mobile)
	req.Email = strings.TrimSpace(req.Email)
	req.CompanyName = strings.TrimSpace(req.CompanyName)
	req.CreditCode = strings.ToUpper(strings.TrimSpace(req.CreditCode))
	if err := ensureProfileUnique(memberID, req); err != nil {
		return err
	}

	updates := map[string]interface{}{
		"mobile":             req.Mobile,
		"email":              req.Email,
		"name":               req.Name,
		"id_card":            req.IDCard,
		"company_name":       req.CompanyName,
		"credit_code":        req.CreditCode,
		"legal_person":       req.LegalPerson,
		"contact_person":     req.ContactPerson,
		"contact_title":      req.ContactTitle,
		"contact_mobile":     req.ContactMobile,
		"industry":           req.Industry,
		"founded_date":       req.FoundedDate,
		"registered_capital": req.RegisteredCapital,
		"employee_count":     req.EmployeeCount,
		"business_scope":     req.BusinessScope,
		"postal_code":        req.PostalCode,
		"address":            req.Address,
		"website":            req.Website,
		"description":        req.Description,
		"cert_file":          req.CertFile,
		"avatar":             req.Avatar,
	}

	// Build the "after" state from the request to snapshot what was changed.
	newMember := member
	newMember.Mobile = req.Mobile
	newMember.Email = req.Email
	newMember.Name = req.Name
	newMember.IDCard = req.IDCard
	newMember.CompanyName = req.CompanyName
	newMember.CreditCode = req.CreditCode
	newMember.LegalPerson = req.LegalPerson
	newMember.ContactPerson = req.ContactPerson
	newMember.ContactTitle = req.ContactTitle
	newMember.ContactMobile = req.ContactMobile
	newMember.Industry = req.Industry
	newMember.FoundedDate = req.FoundedDate
	newMember.RegisteredCapital = req.RegisteredCapital
	newMember.EmployeeCount = req.EmployeeCount
	newMember.BusinessScope = req.BusinessScope
	newMember.PostalCode = req.PostalCode
	newMember.Address = req.Address
	newMember.Website = req.Website
	newMember.Description = req.Description
	newMember.CertFile = req.CertFile
	newMember.Avatar = req.Avatar

	oldSnapshot := profileSnapshot(&member)
	newSnapshot := profileSnapshot(&newMember)

	changes := s.collectProfileChanges(oldSnapshot, newSnapshot)
	if len(changes) == 0 {
		// Nothing changed; still apply (idempotent) so the API behaves as before.
		return db.DB.Model(&models.Member{}).Where("id = ?", memberID).Updates(updates).Error
	}

	return db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Member{}).Where("id = ?", memberID).Updates(updates).Error; err != nil {
			return err
		}
		for _, c := range changes {
			rec := models.ProfileChange{
				MemberID:   member.ID,
				Username:   member.Username,
				Name:       c.Name,
				OldContent: c.OldValue,
				NewContent: c.NewValue,
				Operator:   member.Username,
			}
			if err := tx.Create(&rec).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// profileFields lists the Chinese titles of tracked profile fields, in a stable order.
var profileFields = []string{
	"手机号",
	"邮箱",
	"姓名",
	"身份证号",
	"单位名称",
	"组织机构代码证",
	"法定代表人",
	"联系人",
	"联系人职务",
	"联系人手机号",
	"所属行业",
	"成立日期",
	"注册资本",
	"员工规模",
	"经营范围",
	"邮编",
	"单位地址",
	"网站",
	"简介",
	"组织机构证",
}

// profileSnapshot returns the tracked profile fields (excluding password and
// non-profile fields) keyed by their Chinese display names.
func profileSnapshot(m *models.Member) map[string]interface{} {
	return map[string]interface{}{
		"手机号":     m.Mobile,
		"邮箱":      m.Email,
		"姓名":      m.Name,
		"身份证号":    m.IDCard,
		"单位名称":    m.CompanyName,
		"组织机构代码证": m.CreditCode,
		"法定代表人":   m.LegalPerson,
		"联系人":     m.ContactPerson,
		"联系人职务":   m.ContactTitle,
		"联系人手机号":  m.ContactMobile,
		"所属行业":    m.Industry,
		"成立日期":    m.FoundedDate,
		"注册资本":    m.RegisteredCapital,
		"员工规模":    m.EmployeeCount,
		"经营范围":    m.BusinessScope,
		"邮编":      m.PostalCode,
		"单位地址":    m.Address,
		"网站":      m.Website,
		"简介":      m.Description,
		"组织机构证":   m.CertFile,
	}
}

// profileChange is one changed profile field.
type profileChange struct {
	Name     string
	OldValue string
	NewValue string
}

// collectProfileChanges returns one entry per tracked field whose value changed.
func (s *AuthService) collectProfileChanges(oldS, newS map[string]interface{}) []profileChange {
	var changes []profileChange
	for _, name := range profileFields {
		oldV := profileValueString(oldS[name])
		newV := profileValueString(newS[name])
		if oldV != newV {
			changes = append(changes, profileChange{Name: name, OldValue: oldV, NewValue: newV})
		}
	}
	return changes
}

// profileValueString converts a profile field value to its stored string form.
func profileValueString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}

// ChangePassword changes the member password
func (s *AuthService) ChangePassword(memberID uint64, oldPwd, newPwd string) error {
	var member models.Member
	if err := db.DB.First(&member, memberID).Error; err != nil {
		return errors.New("会员不存在")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(member.Password), []byte(oldPwd)); err != nil {
		return errors.New("原密码错误")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}
	return db.DB.Model(&member).Update("password", string(hashed)).Error
}

// --- Request/Response types ---

type CheckExistsRequest struct {
	Field string `json:"field" binding:"required,oneof=username mobile email company_name credit_code"`
	Value string `json:"value" binding:"required"`
}

type RegisterRequest struct {
	Username      string `json:"username" binding:"required"`
	Password      string `json:"password" binding:"required,min=6"`
	Mobile        string `json:"mobile"`
	Email         string `json:"email"`
	MemberType    string `json:"member_type"`
	CaptchaID     string `json:"captcha_id" binding:"required"`
	CaptchaCode   string `json:"captcha_code" binding:"required"`
	CompanyName   string `json:"company_name"`
	CreditCode    string `json:"credit_code"`
	LegalPerson   string `json:"legal_person"`
	ContactPerson string `json:"contact_person"`
	ContactMobile string `json:"contact_mobile"`
	Address       string `json:"address"`
	CertFile      string `json:"cert_file"`
}

type LoginRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	CaptchaID   string `json:"captcha_id" binding:"required"`
	CaptchaCode string `json:"captcha_code" binding:"required"`
}

type LoginResponse struct {
	Token  string      `json:"token"`
	Member *MemberInfo `json:"member"`
}

type MemberInfo struct {
	ID            uint64 `json:"id"`
	Username      string `json:"username"`
	Mobile        string `json:"mobile"`
	Email         string `json:"email"`
	MemberType    string `json:"member_type"`
	MemberLevel   string `json:"member_level"`
	Status        string `json:"status"`
	IsAdmin       bool   `json:"is_admin"`
	Avatar        string `json:"avatar"`
	CompanyName   string `json:"company_name"`
	ContactPerson string `json:"contact_person"`
}

type UpdateProfileRequest struct {
	Mobile            string `json:"mobile"`
	Email             string `json:"email"`
	Name              string `json:"name"`
	IDCard            string `json:"id_card"`
	CompanyName       string `json:"company_name"`
	CreditCode        string `json:"credit_code"`
	LegalPerson       string `json:"legal_person"`
	ContactPerson     string `json:"contact_person"`
	ContactTitle      string `json:"contact_title"`
	ContactMobile     string `json:"contact_mobile"`
	Industry          string `json:"industry"`
	FoundedDate       string `json:"founded_date"`
	RegisteredCapital string `json:"registered_capital"`
	EmployeeCount     int    `json:"employee_count"`
	BusinessScope     string `json:"business_scope"`
	PostalCode        string `json:"postal_code"`
	Address           string `json:"address"`
	Website           string `json:"website"`
	Description       string `json:"description"`
	CertFile          string `json:"cert_file"`
	Avatar            string `json:"avatar"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// publicSiteConfigKeys 允许通过公开接口 GET /site-info 暴露的配置项白名单。
// 采用白名单而非黑名单：后台后续新增的配置（如邮件密钥、内部参数）默认不会被公开；
// charter_content / charter_file 属内容型配置，已有各自的公开接口，不在其中。
var publicSiteConfigKeys = map[string]bool{
	"site_name":         true,
	"site_description":  true,
	"copyright_name":    true,
	"icp_no":            true,
	"beian_no":          true,
	"contact_phone":     true,
	"contact_email":     true,
	"bank_name":         true,
	"bank_account":      true,
	"bank_account_name": true,
	"fee_amount":        true,
}

func (s *AuthService) GetSiteConfig() map[string]string {
	var configs []models.SystemConfig
	db.DB.Find(&configs)
	result := make(map[string]string)
	for _, c := range configs {
		if publicSiteConfigKeys[c.Key] {
			result[c.Key] = c.Value
		}
	}
	// Defaults
	if result["site_name"] == "" {
		result["site_name"] = "中国电器工业协会会员系统"
	}
	if result["bank_name"] == "" {
		result["bank_name"] = "中国工商银行"
	}
	if result["bank_account"] == "" {
		result["bank_account"] = "0200 0041 0920 1234 567"
	}
	if result["bank_account_name"] == "" {
		result["bank_account_name"] = "中国电器工业协会"
	}
	if result["fee_amount"] == "" {
		result["fee_amount"] = "2000"
	}
	if result["copyright_name"] == "" {
		result["copyright_name"] = "中国电器工业协会"
	}
	if result["icp_no"] == "" {
		result["icp_no"] = "京ICP备09041796号-1"
	}
	if result["beian_no"] == "" {
		result["beian_no"] = "京公网安备11010502000000号"
	}
	return result
}

func init() {
	// Ensure config defaults
	cfg := config.Cfg
	_ = cfg
}
