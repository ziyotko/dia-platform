package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/services"
	"server/utils"
)

type RoleController struct {
	roleService *services.RoleService
	userService *services.UserService
}

func NewRoleController() *RoleController {
	return &RoleController{
		roleService: &services.RoleService{},
		userService: &services.UserService{},
	}
}

// canModifyRole 判断操作者能否修改/删除目标角色：
// 角色序号越小角色越高，只能操作序号 >= 自身最低序号的角色；超级管理员（ID=1）一律不可改。
func (c *RoleController) canModifyRole(operatorID, targetRoleID uint) bool {
	if targetRoleID == models.RoleIDSuperAdmin {
		return false
	}
	minRoleID, ok := getMinRoleID(c.userService.MustGetUserRoleIds(operatorID))
	if !ok {
		return false
	}
	return targetRoleID >= uint(minRoleID)
}

func (c *RoleController) GetRoles(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	name := ctx.Query("name")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	result, err := c.roleService.GetRoleList(page, pageSize, name)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取角色列表失败"))
		return
	}

	ctx.JSON(http.StatusOK, utils.Success("获取角色列表成功", utils.PageData(result.List, result.Total, page, pageSize)))
}

func (c *RoleController) GetAllRoles(ctx *gin.Context) {
	roles, err := c.roleService.GetAllRoles()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取角色列表失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取角色列表成功", roles))
}

func (c *RoleController) GetRoleByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "角色ID无效"))
		return
	}

	role, err := c.roleService.GetRoleByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SafeErrText(err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取角色成功", role))
}

func (c *RoleController) CreateRole(ctx *gin.Context) {
	var req models.Role
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}
	if err := c.roleService.CreateRole(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("创建角色失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("创建角色成功", nil))
}

func (c *RoleController) UpdateRole(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "角色ID无效"))
		return
	}

	if !c.canModifyRole(ctx.GetUint("userID"), uint(id)) {
		ctx.JSON(http.StatusOK, utils.Error(1, "无权修改该角色"))
		return
	}

	// 读取原始请求体：既用于绑定角色字段，也用于判断请求体是否**显式携带**了 permissions
	// （未携带时不得写入，否则会把该角色已有的菜单权限清空）。
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误"))
		return
	}
	var req models.Role
	if err := json.Unmarshal(body, &req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}
	var raw map[string]json.RawMessage
	_ = json.Unmarshal(body, &raw)
	_, hasPermissions := raw["permissions"]

	if err := c.roleService.UpdateRole(uint(id), &req, hasPermissions); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("更新角色失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新角色成功", nil))
}

func (c *RoleController) DeleteRole(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "角色ID无效"))
		return
	}

	if !c.canModifyRole(ctx.GetUint("userID"), uint(id)) {
		ctx.JSON(http.StatusOK, utils.Error(1, "无权删除该角色"))
		return
	}

	if err := c.roleService.DeleteRole(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("删除角色失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("删除角色成功", nil))
}

func (c *RoleController) GetRolePermissions(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "角色ID无效"))
		return
	}

	perms, err := c.roleService.GetRolePermissions(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("获取角色权限失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取角色权限成功", perms))
}

func (c *RoleController) UpdateRolePermissions(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "角色ID无效"))
		return
	}

	// 权限约束：内置角色 1 一律不可改；只能修改角色序号 >= 自身最低序号的角色
	if !c.canModifyRole(ctx.GetUint("userID"), uint(id)) {
		ctx.JSON(http.StatusOK, utils.Error(1, "无权修改该角色的权限"))
		return
	}

	var req struct {
		Permissions []uint `json:"permissions"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}

	if err := c.roleService.UpdateRolePermissions(uint(id), req.Permissions); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("更新角色权限失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新角色权限成功", nil))
}
