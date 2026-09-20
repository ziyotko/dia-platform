package services

import (
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"server/models"
	"server/utils"

	"github.com/xuri/excelize/v2"
)

type UserService struct{}

type UserListResult struct {
	Total int64         `json:"total"`
	List  []models.User `json:"list"`
}

func (s *UserService) Login(email, account, mobile, password, captchaID, captchaCode string) (*models.User, string, string, error) {
	settingsService := SettingsService{}
	settings, err := settingsService.GetSettings()
	if err != nil {
		return nil, "", "", errors.New("获取系统设置失败")
	}
	// 登录验证码开关来自「系统设置-安全设置」的 captchaEnabled（关闭后不再校验）
	if settings.CaptchaEnabled && !utils.VerifyCaptcha(captchaID, captchaCode) {
		return nil, "", "", errors.New("验证码错误")
	}

	var user models.User
	if email != "" {
		err = utils.DB.Where("email = ?", email).First(&user).Error
	} else if account != "" {
		err = utils.DB.Where("account = ?", account).First(&user).Error
	} else if mobile != "" {
		err = utils.DB.Where("mobile = ?", mobile).First(&user).Error
	} else {
		return nil, "", "", errors.New("获取用户失败")
	}

	if err != nil {
		return nil, "", "", errors.New("获取用户失败")
	}

	if user.Status != 1 {
		return nil, "", "", errors.New("用户已禁用")
	}

	if settings.LockEnabled {
		if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
			remaining := int(time.Until(*user.LockedUntil).Minutes()) + 1
			return nil, "", "", fmt.Errorf("登录失败次数过多，请 %d 分钟后重试", remaining)
		}

		if !user.ComparePassword(password) {
			// 原子自增：并发失败登录不会相互覆盖（原先 "读-加一-写" 会丢计数，导致锁定阈值被拖长）
			if err := utils.DB.Model(&models.User{}).Where("id = ?", user.ID).
				UpdateColumn("login_fail_count", gorm.Expr("login_fail_count + 1")).Error; err != nil {
				utils.Logger.Warnf("累加用户[%d]登录失败次数失败: %s", user.ID, err)
			}

			var fresh models.User
			if err := utils.DB.Select("login_fail_count").First(&fresh, user.ID).Error; err == nil &&
				fresh.LoginFailCount >= settings.MaxFailCount {
				lockUntil := time.Now().Add(time.Duration(settings.LockDuration) * time.Minute)
				if err := utils.DB.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
					"login_fail_count": 0,
					"locked_until":     lockUntil,
				}).Error; err != nil {
					utils.Logger.Warnf("锁定用户[%d]失败: %s", user.ID, err)
				}
				return nil, "", "", fmt.Errorf("登录失败次数过多，账号已锁定 %d 分钟", settings.LockDuration)
			}
			return nil, "", "", errors.New("密码错误")
		}

		if user.LoginFailCount > 0 || user.LockedUntil != nil {
			utils.DB.Model(&user).Updates(map[string]any{
				"login_fail_count": 0,
				"locked_until":     nil,
			})
		}
	} else {
		if !user.ComparePassword(password) {
			return nil, "", "", errors.New("密码错误")
		}
	}

	expiresHour := 0
	if settings != nil && settings.TokenExpire > 0 {
		expiresHour = settings.TokenExpire
	}
	token, err := utils.GenerateToken(user.ID, user.Email, expiresHour)
	if err != nil {
		return nil, "", "", err
	}

	// 签名密钥由 Token 派生，登录不再单独生成 / 存储 signKey（见 utils/sign.go 的 DeriveSignKey）。
	// 保留第 3 个返回值为空，供旧接口兼容；前端已不再使用。
	return &user, token, "", nil
}

func (s *UserService) Logout(token string) error {
	claims, err := utils.ParseToken(token)
	if err != nil {
		return err
	}

	// 黑名单 TTL 取 Token 剩余有效期，避免 key 永久驻留导致 Redis 无限增长
	if claims.ExpiresAt == nil {
		return utils.Redis.Set(utils.Ctx, "blacklist:"+claims.ID, token, 24*time.Hour).Err()
	}
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		// Token 本身已过期，无需再拉黑
		return nil
	}
	return utils.Redis.Set(utils.Ctx, "blacklist:"+claims.ID, token, ttl).Err()
}

func (s *UserService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := utils.DB.First(&user, userID).Error; err != nil {
		return nil, errors.New("获取用户失败")
	}
	return &user, nil
}

func (s *UserService) GetUserList(page, pageSize int, username, account string, status *int) (*UserListResult, error) {
	var users []models.User
	var total int64

	query := utils.DB.Model(&models.User{})

	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if account != "" {
		query = query.Where("account LIKE ?", "%"+account+"%")
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	// pageSize <= 0 表示不限制（用于下拉选项等需要全量的场景），避免数据量超过分页上限时被静默截断
	listQuery := query.Order("id DESC")
	if pageSize > 0 {
		listQuery = listQuery.Offset((page - 1) * pageSize).Limit(pageSize)
	}
	err = listQuery.Find(&users).Error
	if err != nil {
		return nil, err
	}

	return &UserListResult{
		Total: total,
		List:  users,
	}, nil
}

func (s *UserService) CreateUser(username, account, email, password, phone string, status int, sex int, roleIds []int, orgIds []uint) (string, error) {
	if password == "" {
		// 未指定密码时生成随机强密码，由调用方展示一次，避免固定弱默认密码
		password = utils.GenerateRandomPassword(12)
	} else if err := s.validatePasswordLength(password); err != nil {
		return "", err
	}

	user := &models.User{
		Username: username,
		Account:  account,
		Email:    email,
		Password: password,
		Mobile:   phone,
		Status:   status,
		Sex:      sex,
	}

	if len(roleIds) > 0 {
		roleIdsStr := make([]string, len(roleIds))
		for i, id := range roleIds {
			roleIdsStr[i] = strconv.Itoa(id)
		}
		user.RoleIds = strings.Join(roleIdsStr, ",")
	}

	if err := utils.DB.Create(user).Error; err != nil {
		return "", err
	}

	if len(orgIds) > 0 {
		orgService := OrganizationService{}
		for _, orgId := range orgIds {
			if orgId > 0 {
				if err := orgService.AddUserToOrganization(orgId, user.ID); err != nil {
					return "", err
				}
			}
		}
	}

	return password, nil
}

func (s *UserService) CheckFieldUnique(field, value string, excludeID uint) (bool, error) {
	columnMap := map[string]string{
		"account": "account",
		"email":   "email",
		"phone":   "mobile",
	}
	column, ok := columnMap[field]
	if !ok {
		return false, errors.New("无效的字段")
	}

	var count int64
	query := utils.DB.Model(&models.User{}).Where(column+" = ?", value)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

type ImportUserResult struct {
	SuccessCount       int               `json:"successCount"`
	FailCount          int               `json:"failCount"`
	FailDetails        []string          `json:"failDetails"`
	GeneratedPasswords map[string]string `json:"generatedPasswords,omitempty"` // 账号 -> 初始随机密码
}

func (s *UserService) ImportUsers(file multipart.File, fileSize int64) (*ImportUserResult, error) {
	f, err := excelize.OpenReader(file, excelize.Options{})
	if err != nil {
		return nil, fmt.Errorf("读取 Excel 失败: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, errors.New("Excel 工作表为空")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取工作表失败: %w", err)
	}

	if len(rows) < 2 {
		return nil, errors.New("Excel 数据行数不足")
	}

	result := &ImportUserResult{
		SuccessCount: 0,
		FailCount:    0,
		FailDetails:  make([]string, 0),
	}

	for i, row := range rows[1:] {
		lineNum := i + 2
		if len(row) < 4 {
			result.FailCount++
			result.FailDetails = append(result.FailDetails, fmt.Sprintf("第 %d 行: 字段数量不足", lineNum))
			continue
		}

		username := strings.TrimSpace(row[0])
		account := strings.TrimSpace(row[1])
		email := strings.TrimSpace(row[2])
		phone := strings.TrimSpace(row[3])

		if username == "" || account == "" || email == "" || phone == "" {
			result.FailCount++
			result.FailDetails = append(result.FailDetails, fmt.Sprintf("第 %d 行: 存在空字段", lineNum))
			continue
		}

		pwd := utils.GenerateRandomPassword(12)
		if _, err := s.CreateUser(username, account, email, pwd, phone, 1, 0, []int{defaultImportedUserRoleID}, nil); err != nil {
			result.FailCount++
			result.FailDetails = append(result.FailDetails, fmt.Sprintf("第 %d 行 (%s): %s", lineNum, account, err.Error()))
			continue
		}

		if result.GeneratedPasswords == nil {
			result.GeneratedPasswords = make(map[string]string)
		}
		result.GeneratedPasswords[account] = pwd
		result.SuccessCount++
	}

	return result, nil
}

func (s *UserService) UpdateUser(id uint, username, account, email, password, phone string, status int, sex int, roleIds []int, orgIds []uint) error {
	updates := map[string]any{
		"username": username,
		"sex":      sex,
	}

	if id == builtinSuperAdminUserID {
		// 内置管理员：账号/邮箱/手机号是登录凭据，密码同样是凭据，一律不允许通过用户管理接口修改
		// （否则有用户管理权限的角色改掉管理员密码即可接管账号）；状态与角色同样保持原值。
		// 凭据只能由管理员本人通过「个人中心」的修改资料/修改密码接口变更。
		updates["status"] = 1
	} else {
		updates["account"] = account
		updates["email"] = email
		updates["mobile"] = phone
		updates["status"] = status
		// 角色始终写入（空数组表示清空角色），修复“无法清空角色”的问题
		roleIdStrs := make([]string, len(roleIds))
		for i, rid := range roleIds {
			roleIdStrs[i] = strconv.Itoa(rid)
		}
		updates["role_ids"] = strings.Join(roleIdStrs, ",")

		// 密码需显式做 SM3 加盐哈希后再入库：map 更新不会触发模型 BeforeUpdate 钩子（否则会写入明文）
		if password != "" {
			if err := s.validatePasswordLength(password); err != nil {
				return err
			}
			updates["password"] = utils.SM3HashPassword(password)
		}
	}

	if err := utils.DB.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return err
	}

	orgService := OrganizationService{}
	oldOrgs, _ := orgService.GetOrganizationsByUserId(id)
	oldOrgIdMap := make(map[uint]bool)
	for _, org := range oldOrgs {
		oldOrgIdMap[org.ID] = true
	}
	newOrgIdMap := make(map[uint]bool)
	for _, orgId := range orgIds {
		if orgId > 0 {
			newOrgIdMap[orgId] = true
		}
	}

	for _, org := range oldOrgs {
		if !newOrgIdMap[org.ID] {
			if err := orgService.RemoveUserFromOrganization(org.ID, id); err != nil {
				return err
			}
		}
	}
	for _, orgId := range orgIds {
		if orgId > 0 && !oldOrgIdMap[orgId] {
			if err := orgService.AddUserToOrganization(orgId, id); err != nil {
				return err
			}
		}
	}

	return nil
}

// 内置管理员用户 ID（与前端 users.vue 及 models/seed.go 保持一致），不可删除、不可禁用
const builtinSuperAdminUserID uint = 1

// 导入用户默认角色 ID（内容作者）。系统仅内置角色 1-4，不可引用不存在的角色。
const defaultImportedUserRoleID = 4

func (s *UserService) DeleteUser(id uint) error {
	if id == builtinSuperAdminUserID {
		return errors.New("内置管理员不可删除")
	}
	orgService := OrganizationService{}
	if err := orgService.RemoveUserFromAllOrganizations(id); err != nil {
		return err
	}
	// 部门成员关系同样需要清理，否则 department.user_ids/user_count 会残留已删除用户
	deptService := DepartmentService{}
	if err := deptService.RemoveUserFromAllDepartments(id); err != nil {
		return err
	}
	return utils.DB.Unscoped().Delete(&models.User{}, id).Error
}

func (s *UserService) UpdateUserStatus(id uint, status int) error {
	if id == builtinSuperAdminUserID && status != 1 {
		return errors.New("内置管理员不可禁用")
	}
	return utils.DB.Model(&models.User{}).Where("id = ?", id).Update("status", status).Error
}

func (s *UserService) UpdateProfile(id uint, email, phone, bio, avatar string) error {
	return utils.DB.Model(&models.User{}).Where("id = ?", id).Updates(map[string]any{
		"email":  email,
		"mobile": phone,
		"bio":    bio,
		"avatar": avatar,
	}).Error
}

func (s *UserService) ChangePassword(id uint, oldPassword, newPassword string) error {
	user, err := s.GetUserByID(id)
	if err != nil {
		return err
	}
	if !user.ComparePassword(oldPassword) {
		return errors.New("原密码错误")
	}
	hashedPassword := utils.SM3HashPassword(newPassword)
	return utils.DB.Model(&models.User{}).Where("id = ?", id).Update("password", hashedPassword).Error
}

// validatePasswordLength 按系统设置（minPasswordLength，默认 8）校验密码长度；空密码表示不修改，直接放行。
// 在管理员重置密码与手动创建用户时调用，与「个人中心-修改密码」保持同一口径。
func (s *UserService) validatePasswordLength(password string) error {
	if password == "" {
		return nil
	}
	settingsService := SettingsService{}
	settings, err := settingsService.GetMinPasswordLengthSettings()
	if err != nil || settings == nil || settings.MinPasswordLength <= 0 {
		// 设置不可用时不做额外限制，避免阻断正常操作
		return nil
	}
	if len([]rune(password)) < settings.MinPasswordLength {
		return fmt.Errorf("密码长度不能少于%d位", settings.MinPasswordLength)
	}
	return nil
}

func (s *UserService) GetUserRoleIds(userId uint) ([]int, error) {
	var user models.User
	if err := utils.DB.First(&user, userId).Error; err != nil {
		return nil, err
	}
	if user.RoleIds == "" {
		return []int{}, nil
	}

	roleIdStrs := strings.Split(user.RoleIds, ",")
	roleIds := make([]int, 0, len(roleIdStrs))
	for _, idStr := range roleIdStrs {
		if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
			roleIds = append(roleIds, id)
		}
	}

	return roleIds, nil
}

// MustGetUserRoleIds 用于调用方不关心错误的场景（查询失败视为无角色），避免各处重复忽略 err。
func (s *UserService) MustGetUserRoleIds(userId uint) []int {
	roleIds, err := s.GetUserRoleIds(userId)
	if err != nil {
		return nil
	}
	return roleIds
}
