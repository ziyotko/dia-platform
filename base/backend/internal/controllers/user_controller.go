package controllers

import (
	"strconv"

	"base/internal/models"
	"base/internal/service"
	"base/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	service service.UserService
}

// CreateUserReq 新增用户请求。
// 密码不能直接绑定 models.User（其 Password 字段为 json:"-"，绑定时会被忽略），故单独定义请求体。
type CreateUserReq struct {
	Username string `json:"username" binding:"required,min=2,max=64"`
	Password string `json:"password" binding:"omitempty,min=6,max=64"`
	RealName string `json:"realName" binding:"max=64"`
	Phone    string `json:"phone" binding:"max=32"`
	Email    string `json:"email" binding:"omitempty,email"`
	Avatar   string `json:"avatar"`
	IsAdmin  bool   `json:"isAdmin"`
	Status   *int   `json:"status"`
	TenantID uint64 `json:"tenantId"`
}

func (ctl *UserController) Create(c *gin.Context) {
	var req CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	status := 1
	if req.Status != nil {
		status = *req.Status
	}
	u := models.User{
		TenantID: resolveTenantID(c, req.TenantID),
		Username: req.Username,
		Password: req.Password,
		RealName: req.RealName,
		Phone:    req.Phone,
		Email:    req.Email,
		Avatar:   req.Avatar,
		Status:   status,
		IsAdmin:  req.IsAdmin,
	}
	if err := ctl.service.Create(&u); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, u)
}

func (ctl *UserController) Update(c *gin.Context) {
	var u models.User
	if err := c.ShouldBindJSON(&u); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	u.ID = uint64(parseID(c))
	tenantID := c.GetUint64("tenantID")
	if err := ctl.service.Update(&u, tenantID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Ok(c, u)
}

func (ctl *UserController) Delete(c *gin.Context) {
	id := uint64(parseID(c))
	tenantID := c.GetUint64("tenantID")
	if err := ctl.service.Delete(id, tenantID); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "删除成功", nil)
}

func (ctl *UserController) Get(c *gin.Context) {
	id := uint64(parseID(c))
	tenantID := c.GetUint64("tenantID")
	u, err := ctl.service.GetByID(id, tenantID)
	if err != nil {
		response.Fail(c, "用户不存在")
		return
	}
	response.Ok(c, u)
}

func (ctl *UserController) List(c *gin.Context) {
	tenantID := c.GetUint64("tenantID")
	filterTenantID, _ := strconv.ParseUint(c.Query("tenantId"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	keyword := c.Query("keyword")
	list, total, err := ctl.service.List(tenantID, filterTenantID, page, size, keyword)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Page(c, list, total)
}

func (ctl *UserController) AssignRoles(c *gin.Context) {
	id := uint64(parseID(c))
	var req struct {
		RoleIDs []uint64 `json:"roleIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithCode(c, response.CodeBadRequest, "参数错误")
		return
	}
	if err := ctl.service.AssignRoles(id, req.RoleIDs, c.GetUint64("tenantID")); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "分配成功", nil)
}

func (ctl *UserController) ResetPassword(c *gin.Context) {
	id := uint64(parseID(c))
	if err := ctl.service.ResetPassword(id, c.GetUint64("tenantID")); err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.OkWithMessage(c, "密码已重置为 123456", nil)
}
