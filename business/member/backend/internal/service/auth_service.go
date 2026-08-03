package service

import (
	"errors"
	"fmt"
	"member/config"
	"member/internal/models"
	"member/pkg/captcha"
	"member/pkg/db"
	mjwt "member/pkg/jwt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct{}

// Register registers a new member and returns a token
func (s *AuthService) Register(req RegisterRequest) (*LoginResponse, error) {
	// Verify captcha
	if !captcha.Verify(req.CaptchaID, req.CaptchaCode) {
		return nil, errors.New("验证码错误")
	}

	// Check if username already exists
	var exist models.Member
	if err := db.DB.Where("username = ?", req.Username).First(&exist).Error; err == nil {
		return nil, errors.New("用户名已存在")
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
func (s *AuthService) CheckExists(field, value string) (bool, error) {
	var count int64
	query := db.DB.Model(&models.Member{})
	switch field {
	case "username", "mobile", "email":
		query = query.Where(field+" = ?", value)
	default:
		return false, errors.New("不支持的字段")
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
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
		member.LoginFailCount++
		if member.LoginFailCount >= 5 {
			t := time.Now().Add(30 * time.Minute)
			member.LockedUntil = &models.LocalTime{Time: t}
		}
		db.DB.Save(&member)
		return nil, errors.New("用户名或密码错误")
	}

	// Reset fail count on success
	if member.LoginFailCount > 0 {
		member.LoginFailCount = 0
		member.LockedUntil = nil
		db.DB.Save(&member)
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

// UpdateProfile updates the member profile
func (s *AuthService) UpdateProfile(memberID uint64, req UpdateProfileRequest) error {
	updates := map[string]interface{}{
		"mobile":         req.Mobile,
		"email":          req.Email,
		"name":           req.Name,
		"id_card":        req.IDCard,
		"company_name":   req.CompanyName,
		"credit_code":    req.CreditCode,
		"legal_person":   req.LegalPerson,
		"contact_person": req.ContactPerson,
		"address":        req.Address,
		"website":        req.Website,
		"description":    req.Description,
		"cert_file":      req.CertFile,
	}
	return db.DB.Model(&models.Member{}).Where("id = ?", memberID).Updates(updates).Error
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

// RequestPasswordReset creates a reset token
func (s *AuthService) RequestPasswordReset(email string) error {
	var member models.Member
	if err := db.DB.Where("email = ?", email).First(&member).Error; err != nil {
		return errors.New("该邮箱未注册")
	}
	token := uuid.New().String()
	expire := time.Now().Add(1 * time.Hour)
	reset := models.PasswordReset{
		MemberID: member.ID,
		Token:    token,
		ExpireAt: &models.LocalTime{Time: expire},
	}
	db.DB.Create(&reset)
	// In production, send email with token
	return nil
}

// ResetPassword resets password with token
func (s *AuthService) ResetPassword(token, newPwd string) error {
	var reset models.PasswordReset
	if err := db.DB.Where("token = ? AND used = ?", token, false).First(&reset).Error; err != nil {
		return errors.New("无效的重置链接")
	}
	if reset.ExpireAt.Before(time.Now()) {
		return errors.New("重置链接已过期")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}
	db.DB.Model(&models.Member{}).Where("id = ?", reset.MemberID).Update("password", string(hashed))
	db.DB.Model(&reset).Update("used", true)
	return nil
}

// --- Request/Response types ---

type CheckExistsRequest struct {
	Field string `json:"field" binding:"required,oneof=username mobile email"`
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
	ID          uint64 `json:"id"`
	Username    string `json:"username"`
	Mobile      string `json:"mobile"`
	Email       string `json:"email"`
	MemberType  string `json:"member_type"`
	MemberLevel string `json:"member_level"`
	Status      string `json:"status"`
	IsAdmin     bool   `json:"is_admin"`
	Avatar      string `json:"avatar"`
	CompanyName string `json:"company_name"`
}

type UpdateProfileRequest struct {
	Mobile        string `json:"mobile"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	IDCard        string `json:"id_card"`
	CompanyName   string `json:"company_name"`
	CreditCode    string `json:"credit_code"`
	LegalPerson   string `json:"legal_person"`
	ContactPerson string `json:"contact_person"`
	Address       string `json:"address"`
	Website       string `json:"website"`
	Description   string `json:"description"`
	CertFile      string `json:"cert_file"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (s *AuthService) GetSiteConfig() map[string]string {
	var configs []models.SystemConfig
	db.DB.Find(&configs)
	result := make(map[string]string)
	for _, c := range configs {
		result[c.Key] = c.Value
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
