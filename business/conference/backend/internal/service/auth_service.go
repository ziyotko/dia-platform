package service

import (
	"errors"

	"conference/internal/models"
	"conference/pkg/captcha"
	"conference/pkg/db"
	"conference/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct{}

func (s *AuthService) GenerateCaptcha() (string, string, string, error) {
	captcha.Init()
	return captcha.Generate()
}

// --- Member ---

func (s *AuthService) MemberLogin(username, password, captchaID, captchaCode string) (map[string]interface{}, error) {
	if !captcha.Verify(captchaID, captchaCode) {
		return nil, errors.New("验证码错误")
	}

	var user models.User
	if err := db.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	if user.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	token, err := jwt.GenerateMemberToken(user.ID, user.Username, user.MemberLevel, user.Branch)
	if err != nil {
		return nil, errors.New("生成token失败")
	}

	return map[string]interface{}{
		"token": token,
		"user":  user,
	}, nil
}

// MemberRegisterRequest for member registration
type MemberRegisterRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	RealName    string `json:"realName"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Company     string `json:"company"`
	Branch      string `json:"branch"`
	MemberLevel string `json:"memberLevel"`
}

func (s *AuthService) MemberRegister(req MemberRegisterRequest) error {
	var count int64
	db.DB.Model(&models.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return errors.New("用户名已存在")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	user := models.User{
		Username:    req.Username,
		Password:    string(hashed),
		RealName:    req.RealName,
		Phone:       req.Phone,
		Email:       req.Email,
		Company:     req.Company,
		Branch:      req.Branch,
		MemberLevel: req.MemberLevel,
		IsValid:     true,
		Status:      1,
	}
	return db.DB.Create(&user).Error
}

func (s *AuthService) GetMemberProfile(userID uint64) (*models.User, error) {
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	return &user, nil
}

func (s *AuthService) UpdateMemberProfile(userID uint64, updates map[string]interface{}) error {
	delete(updates, "password")
	delete(updates, "username")
	delete(updates, "id")
	return db.DB.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error
}

func (s *AuthService) ChangeMemberPassword(userID uint64, oldPassword, newPassword string) error {
	var user models.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("原密码错误")
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	return db.DB.Model(&user).Update("password", string(hashed)).Error
}

// --- Admin ---

func (s *AuthService) AdminLogin(username, password, captchaID, captchaCode string) (map[string]interface{}, error) {
	if !captcha.Verify(captchaID, captchaCode) {
		return nil, errors.New("验证码错误")
	}

	var admin models.Admin
	if err := db.DB.Where("username = ?", username).First(&admin).Error; err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	if admin.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	token, err := jwt.GenerateAdminToken(admin.ID, admin.Username, admin.RoleCode)
	if err != nil {
		return nil, errors.New("生成token失败")
	}

	return map[string]interface{}{
		"token": token,
		"admin": admin,
	}, nil
}

func (s *AuthService) GetAdminProfile(adminID uint64) (*models.Admin, error) {
	var admin models.Admin
	if err := db.DB.First(&admin, adminID).Error; err != nil {
		return nil, errors.New("管理员不存在")
	}
	return &admin, nil
}

func (s *AuthService) UpdateAdminProfile(adminID uint64, updates map[string]interface{}) error {
	delete(updates, "password")
	delete(updates, "username")
	delete(updates, "id")
	delete(updates, "role_code")
	return db.DB.Model(&models.Admin{}).Where("id = ?", adminID).Updates(updates).Error
}

func (s *AuthService) ChangeAdminPassword(adminID uint64, oldPassword, newPassword string) error {
	var admin models.Admin
	if err := db.DB.First(&admin, adminID).Error; err != nil {
		return errors.New("管理员不存在")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(oldPassword)); err != nil {
		return errors.New("原密码错误")
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	return db.DB.Model(&admin).Update("password", string(hashed)).Error
}
