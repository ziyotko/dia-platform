package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/services"
	"server/utils"
)

type RoleController struct {
	roleService *services.RoleService
}

func NewRoleController() *RoleController {
	return &RoleController{
		roleService: &services.RoleService{},
	}
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

	ctx.JSON(http.StatusOK, utils.Success("获取角色列表成功", gin.H{
		"list":  result.List,
		"total": result.Total,
	}))
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
		ctx.JSON(http.StatusOK, utils.Error(1, err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取角色成功", role))
}

func (c *RoleController) CreateRole(ctx *gin.Context) {
	var req models.Role
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: "+err.Error()))
		return
	}
	if err := c.roleService.CreateRole(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "创建角色失败: "+err.Error()))
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

	var req models.Role
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: "+err.Error()))
		return
	}

	if err := c.roleService.UpdateRole(uint(id), &req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "更新角色失败: "+err.Error()))
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

	if err := c.roleService.DeleteRole(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "删除角色失败: "+err.Error()))
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
		ctx.JSON(http.StatusOK, utils.Error(1, "获取角色权限失败: "+err.Error()))
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

	var req struct {
		Permissions []uint `json:"permissions"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: "+err.Error()))
		return
	}

	if err := c.roleService.UpdateRolePermissions(uint(id), req.Permissions); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "更新角色权限失败: "+err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新角色权限成功", nil))
}
