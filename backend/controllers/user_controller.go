package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/services"
	"server/utils"
)

type UserController struct {
	userService *services.UserService
}

type UserListItem struct {
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	Nickname   string `json:"nickname"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Status     int    `json:"status"`
	RoleIds    []int  `json:"roleIds"`
	CreateTime string `json:"createTime"`
}

func NewUserController() *UserController {
	return &UserController{
		userService: &services.UserService{},
	}
}

func (c *UserController) GetUsers(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	username := ctx.Query("username")
	statusStr := ctx.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	var status *int
	if statusStr != "" {
		s, _ := strconv.Atoi(statusStr)
		status = &s
	}

	result, err := c.userService.GetUserList(page, pageSize, username, status)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取用户列表失败"))
		return
	}

	list := make([]UserListItem, 0, len(result.List))
	for _, user := range result.List {
		item := UserListItem{
			ID:         user.ID,
			Username:   user.Username,
			Nickname:   user.Nickname,
			Email:      user.Email,
			Phone:      user.Mobile,
			Status:     user.Status,
			CreateTime: user.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		roleIds, err := c.userService.GetUserRoleIds(user.ID)
		if err == nil {
			item.RoleIds = roleIds
		} else {
			item.RoleIds = []int{}
		}

		list = append(list, item)
	}

	ctx.JSON(http.StatusOK, utils.Success("获取用户列表成功", gin.H{
		"list":  list,
		"total": result.Total,
	}))
}

func (c *UserController) CreateUser(ctx *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Nickname string `json:"nickname"`
		Email    string `json:"email" binding:"required,email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
		Status   int    `json:"status"`
		RoleIds  []int  `json:"roleIds"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: "+err.Error()))
		return
	}

	err := c.userService.CreateUser(req.Username, req.Nickname, req.Email, req.Password, req.Phone, req.Status, req.RoleIds)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "创建用户失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("创建用户成功", nil))
}

func (c *UserController) UpdateUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "用户ID无效"))
		return
	}

	var req struct {
		Username string `json:"username"`
		Nickname string `json:"nickname"`
		Email    string `json:"email" binding:"required,email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
		Status   int    `json:"status"`
		RoleIds  []int  `json:"roleIds"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: "+err.Error()))
		return
	}

	err = c.userService.UpdateUser(uint(id), req.Username, req.Nickname, req.Email, req.Password, req.Phone, req.Status, req.RoleIds)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "更新用户失败: "+err.Error()))
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

	err = c.userService.DeleteUser(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "删除用户失败: "+err.Error()))
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

	err = c.userService.UpdateUserStatus(uint(id), req.Status)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "更新状态失败: "+err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("更新状态成功", nil))
}

func (c *UserController) GetUserByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "用户ID无效"))
		return
	}

	user, err := c.userService.GetUserByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, err.Error()))
		return
	}

	item := UserListItem{
		ID:         user.ID,
		Username:   user.Username,
		Nickname:   user.Nickname,
		Email:      user.Email,
		Phone:      user.Mobile,
		Status:     user.Status,
		CreateTime: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	roleIds, err := c.userService.GetUserRoleIds(user.ID)
	if err == nil {
		item.RoleIds = roleIds
	} else {
		item.RoleIds = []int{}
	}

	ctx.JSON(http.StatusOK, utils.Success("获取用户信息成功", gin.H{"user": item}))
}

func (c *UserController) formatCreateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func (c *UserController) convertUsers(users []models.User) []UserListItem {
	list := make([]UserListItem, len(users))
	for i, user := range users {
		list[i] = UserListItem{
			ID:         user.ID,
			Username:   user.Username,
			Nickname:   user.Nickname,
			Email:      user.Email,
			Phone:      user.Mobile,
			Status:     user.Status,
			CreateTime: c.formatCreateTime(user.CreatedAt),
		}
	}
	return list
}
