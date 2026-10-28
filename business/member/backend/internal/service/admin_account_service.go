package service

import (
	"errors"
	"regexp"
	"strings"

	"member/internal/models"
	"member/pkg/db"
	"member/pkg/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var adminUsernamePattern = regexp.MustCompile(`^[a-z0-9_-]{1,20}$`)

type AdminAccountService struct{}

type AdminAccount struct {
	ID        uint64 `json:"id"`
	Username  string `json:"username"`
	Mobile    string `json:"mobile"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type AdminAccountRequest struct {
	Username string `json:"username" binding:"required"`
	Mobile   string `json:"mobile"`
	Email    string `json:"email"`
}

type AdminAccountResult struct {
	Account  AdminAccount `json:"account"`
	Password string       `json:"password,omitempty"`
}

func (s *AdminAccountService) List() ([]AdminAccount, error) {
	var members []models.Member
	if err := db.DB.Select("id", "username", "mobile", "email", "created_at").
		Where("is_admin = ? AND is_super_admin = ?", true, false).
		Order("id DESC").Find(&members).Error; err != nil {
		return nil, err
	}
	accounts := make([]AdminAccount, 0, len(members))
	for _, member := range members {
		accounts = append(accounts, adminAccountFromMember(member))
	}
	return accounts, nil
}

func (s *AdminAccountService) Create(req AdminAccountRequest) (*AdminAccountResult, error) {
	req.Username = strings.TrimSpace(req.Username)
	req.Mobile = strings.TrimSpace(req.Mobile)
	req.Email = strings.TrimSpace(req.Email)
	if !adminUsernamePattern.MatchString(req.Username) {
		return nil, errors.New("用户名只能包含小写字母、数字、- 和 _，且不能超过20位")
	}
	if err := validateAdminContact(req.Mobile, req.Email); err != nil {
		return nil, err
	}
	if err := ensureAdminContactUnique(req.Mobile, req.Email, 0); err != nil {
		return nil, err
	}

	password, err := utils.RandomPassword(0)
	if err != nil {
		return nil, errors.New("密码生成失败")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}
	member := models.Member{
		Username:     req.Username,
		Password:     string(hashed),
		Mobile:       req.Mobile,
		Email:        req.Email,
		MemberType:   models.MemberTypeUnit,
		Status:       models.MemberStatusActive,
		IsAdmin:      true,
		IsSuperAdmin: false,
	}
	if err := db.DB.Create(&member).Error; err != nil {
		if isAdminAccountDuplicate(err) {
			return nil, errors.New("用户名已存在")
		}
		return nil, err
	}
	return &AdminAccountResult{Account: adminAccountFromMember(member), Password: password}, nil
}

func (s *AdminAccountService) Update(id uint64, req AdminAccountRequest) error {
	req.Username = strings.TrimSpace(req.Username)
	req.Mobile = strings.TrimSpace(req.Mobile)
	req.Email = strings.TrimSpace(req.Email)
	if !adminUsernamePattern.MatchString(req.Username) {
		return errors.New("用户名只能包含小写字母、数字、- 和 _，且不能超过20位")
	}
	if err := validateAdminContact(req.Mobile, req.Email); err != nil {
		return err
	}
	var member models.Member
	if err := db.DB.Where("id = ? AND is_admin = ? AND is_super_admin = ?", id, true, false).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("下属管理员不存在")
		}
		return err
	}
	var count int64
	if err := db.DB.Model(&models.Member{}).Where("username = ? AND id <> ?", req.Username, id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("用户名已存在")
	}
	if err := ensureAdminContactUnique(req.Mobile, req.Email, id); err != nil {
		return err
	}
	if err := db.DB.Model(&member).Updates(map[string]any{
		"username": req.Username,
		"mobile":   req.Mobile,
		"email":    req.Email,
	}).Error; err != nil {
		if isAdminAccountDuplicate(err) {
			return errors.New("用户名已存在")
		}
		return err
	}
	return nil
}

func (s *AdminAccountService) ResetPassword(id uint64) (string, error) {
	var member models.Member
	if err := db.DB.Where("id = ? AND is_admin = ? AND is_super_admin = ?", id, true, false).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("下属管理员不存在")
		}
		return "", err
	}
	password, err := utils.RandomPassword(0)
	if err != nil {
		return "", errors.New("密码生成失败")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("密码加密失败")
	}
	if err := db.DB.Model(&member).Updates(map[string]any{
		"password":            string(hashed),
		"password_changed_at": models.NewInvalidBefore(),
	}).Error; err != nil {
		return "", err
	}
	return password, nil
}

func (s *AdminAccountService) Delete(id uint64, currentAdminID uint64) error {
	if id == currentAdminID {
		return errors.New("不能删除当前登录账号")
	}
	result := db.DB.Where("id = ? AND is_admin = ? AND is_super_admin = ?", id, true, false).Delete(&models.Member{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("下属管理员不存在")
	}
	return nil
}

func validateAdminContact(mobile, email string) error {
	if mobile != "" && !utils.IsValidMobile(mobile) {
		return errors.New("请输入合法的手机号")
	}
	if email != "" && !utils.IsValidEmail(email) {
		return errors.New("邮箱格式不正确")
	}
	return nil
}

func ensureAdminContactUnique(mobile, email string, excludeID uint64) error {
	if mobile != "" {
		var count int64
		if err := db.DB.Model(&models.Member{}).Where("mobile = ? AND id <> ?", mobile, excludeID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("手机号已被注册")
		}
	}
	if email != "" {
		var count int64
		if err := db.DB.Model(&models.Member{}).Where("email = ? AND id <> ?", email, excludeID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("邮箱已被注册")
		}
	}
	return nil
}

func adminAccountFromMember(member models.Member) AdminAccount {
	return AdminAccount{
		ID:        member.ID,
		Username:  member.Username,
		Mobile:    member.Mobile,
		Email:     member.Email,
		CreatedAt: member.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func isAdminAccountDuplicate(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "duplicate")
}
