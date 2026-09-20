package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/services"
	"server/utils"
)

type UserController struct {
	userService *services.UserService
	orgService  *services.OrganizationService
}

type UserListItem struct {
	ID        uint     `json:"id"`
	Username  string   `json:"username"`
	Account   string   `json:"account"`
	OrgIds    []uint   `json:"orgIds"`
	OrgNames  []string `json:"orgNames"`
	Email     string   `json:"email"`
	Phone     string   `json:"phone"`
	Status    int      `json:"status"`
	Sex       int      `json:"sex"`
	RoleIds   []int    `json:"roleIds"`
	CreatedAt string   `json:"createdAt"`
}

func NewUserController() *UserController {
	return &UserController{
		userService: &services.UserService{},
		orgService:  &services.OrganizationService{},
	}
}

// getMinRoleID 返回角色 ID 列表中的最小角色 ID（角色序号越小角色越高）。
func getMinRoleID(roleIDs []int) (int, bool) {
	if len(roleIDs) == 0 {
		return 0, false
	}
	minRoleID := roleIDs[0]
	for _, id := range roleIDs {
		if id < minRoleID {
			minRoleID = id
		}
	}
	return minRoleID, true
}

// canAssignRoles 判断操作者能否为他人分配目标角色。
// 角色序号越小角色越高，操作者只能分配不高于自身最高角色的角色（目标角色序号 >= 操作者最高角色序号）。
func (c *UserController) canAssignRoles(operatorID uint, targetRoleIds []int) bool {
	operatorMin, ok := getMinRoleID(c.userService.MustGetUserRoleIds(operatorID))
	if !ok {
		return false
	}
	targetMin, hasTarget := getMinRoleID(targetRoleIds)
	if !hasTarget {
		return true // 不分配任何角色
	}
	return targetMin >= operatorMin
}

// canOperateUser 判断操作者能否修改/删除目标用户：只能操作角色序号不高于自己的用户。
// 避免低权限角色修改或删除管理员（角色 1）等更高权限账号；目标用户无角色时视为最低权限。
func (c *UserController) canOperateUser(operatorID uint, targetID uint) bool {
	if operatorID == targetID {
		return true
	}
	operatorMin, ok := getMinRoleID(c.userService.MustGetUserRoleIds(operatorID))
	if !ok {
		return false
	}
	targetMin, hasTarget := getMinRoleID(c.userService.MustGetUserRoleIds(targetID))
	if !hasTarget {
		return true
	}
	return targetMin >= operatorMin
}

func (c *UserController) GetUsers(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	username := ctx.Query("username")
	account := ctx.Query("account")
	statusStr := ctx.Query("status")

	// all=1 时不分页返回全量（供下拉选项使用，避免大数据量被分页截断）
	all := ctx.Query("all") == "1" || strings.EqualFold(ctx.Query("all"), "true")
	if all {
		page, pageSize = 1, 0
	} else {
		if page < 1 {
			page = 1
		}
		if pageSize < 1 {
			pageSize = 10
		}
	}

	var status *int
	if statusStr != "" {
		s, _ := strconv.Atoi(statusStr)
		status = &s
	}

	result, err := c.userService.GetUserList(page, pageSize, username, account, status)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取用户列表失败"))
		return
	}

	list := make([]UserListItem, 0, len(result.List))
	for _, user := range result.List {
		item := UserListItem{
			ID:        user.ID,
			Username:  user.Username,
			Account:   user.Account,
			Email:     user.Email,
			Phone:     user.Mobile,
			Status:    user.Status,
			Sex:       user.Sex,
			CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		roleIds, err := c.userService.GetUserRoleIds(user.ID)
		if err == nil {
			item.RoleIds = roleIds
		} else {
			item.RoleIds = []int{}
		}

		orgs, err := c.orgService.GetOrganizationsByUserId(user.ID)
		if err == nil {
			item.OrgIds = make([]uint, 0, len(orgs))
			item.OrgNames = make([]string, 0, len(orgs))
			for _, org := range orgs {
				item.OrgIds = append(item.OrgIds, org.ID)
				item.OrgNames = append(item.OrgNames, org.Name)
			}
		}

		list = append(list, item)
	}

	ctx.JSON(http.StatusOK, utils.Success("获取用户列表成功", utils.PageData(list, result.Total, page, pageSize)))
}

// GetUserOptions 登录用户可用的用户选项（仅 id/username，不含邮箱/手机号等敏感信息），
// 供内容编辑/审核页面展示审批人名称使用，避免为此开放管理端的用户列表接口。
func (c *UserController) GetUserOptions(ctx *gin.Context) {
	var users []models.User
	if err := utils.DB.Model(&models.User{}).
		Select("id, username").
		Where("status = ?", 1).
		Order("id ASC").
		Find(&users).Error; err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取用户选项失败"))
		return
	}
	type userOption struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
	}
	options := make([]userOption, 0, len(users))
	for _, user := range users {
		options = append(options, userOption{ID: user.ID, Username: user.Username})
	}
	ctx.JSON(http.StatusOK, utils.Success("获取成功", options))
}

func (c *UserController) CreateUser(ctx *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Account  string `json:"account"`
		Email    string `json:"email" binding:"required,email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
		Status   int    `json:"status"`
		Sex      int    `json:"sex"`
		RoleIds  []int  `json:"roleIds"`
		OrgIds   []uint `json:"orgIds"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}

	if !c.canAssignRoles(ctx.GetUint("userID"), req.RoleIds) {
		ctx.JSON(http.StatusOK, utils.Error(1, "不能分配高于当前用户最高角色的角色"))
		return
	}

	password, err := c.userService.CreateUser(req.Username, req.Account, req.Email, req.Password, req.Phone, req.Status, req.Sex, req.RoleIds, req.OrgIds)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("创建用户失败", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("创建用户成功", gin.H{"password": password}))
}

func (c *UserController) ImportUsers(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("上传文件失败", err)))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("打开文件失败", err)))
		return
	}
	defer file.Close()

	result, err := c.userService.ImportUsers(file, fileHeader.Size)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("导入失败", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("导入完成", result))
}

func (c *UserController) UpdateUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "用户ID无效"))
		return
	}

	if !c.canOperateUser(ctx.GetUint("userID"), uint(id)) {
		ctx.JSON(http.StatusOK, utils.Error(1, "不能操作权限高于自己的用户"))
		return
	}

	var req struct {
		Username string `json:"username"`
		Account  string `json:"account"`
		Email    string `json:"email" binding:"required,email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
		Status   int    `json:"status"`
		Sex      int    `json:"sex"`
		RoleIds  []int  `json:"roleIds"`
		OrgIds   []uint `json:"orgIds"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}

	if !c.canAssignRoles(ctx.GetUint("userID"), req.RoleIds) {
		ctx.JSON(http.StatusOK, utils.Error(1, "不能分配高于当前用户最高角色的角色"))
		return
	}

	err = c.userService.UpdateUser(uint(id), req.Username, req.Account, req.Email, req.Password, req.Phone, req.Status, req.Sex, req.RoleIds, req.OrgIds)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("更新用户失败", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("更新用户成功", nil))
}

func (c *UserController) DeleteUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "用户ID无效"))
		return
	}

	if !c.canOperateUser(ctx.GetUint("userID"), uint(id)) {
		ctx.JSON(http.StatusOK, utils.Error(1, "不能操作权限高于自己的用户"))
		return
	}

	err = c.userService.DeleteUser(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("删除用户失败", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("删除用户成功", nil))
}

func (c *UserController) UpdateUserStatus(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "用户ID无效"))
		return
	}

	var req struct {
		Status int `json:"status"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误"))
		return
	}

	if !c.canOperateUser(ctx.GetUint("userID"), uint(id)) {
		ctx.JSON(http.StatusOK, utils.Error(1, "不能操作权限高于自己的用户"))
		return
	}

	err = c.userService.UpdateUserStatus(uint(id), req.Status)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("更新状态失败", err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("更新状态成功", nil))
}

func (c *UserController) CheckFieldUnique(ctx *gin.Context) {
	field := ctx.Query("field")
	value := ctx.Query("value")
	excludeIDStr := ctx.Query("excludeId")

	if field == "" || value == "" {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误"))
		return
	}

	var excludeID uint
	if excludeIDStr != "" {
		id, err := strconv.ParseUint(excludeIDStr, 10, 32)
		if err == nil {
			excludeID = uint(id)
		}
	}

	unique, err := c.userService.CheckFieldUnique(field, value, excludeID)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SafeErrText(err)))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("", gin.H{
		"unique": unique,
	}))
}
