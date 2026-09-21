package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"server/models"
	"server/services"
	"server/utils"
)

type MenuController struct {
	menuService *services.MenuService
}

func NewMenuController() *MenuController {
	return &MenuController{
		menuService: &services.MenuService{},
	}
}

func (c *MenuController) GetMenus(ctx *gin.Context) {
	menus, err := c.menuService.GetAllMenus()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取菜单列表失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取菜单列表成功", menus))
}

func (c *MenuController) GetMenuTree(ctx *gin.Context) {
	// 返回全部菜单（含已禁用）：本接口的调用方是「角色管理-分配权限」的权限树，
	// 原先只返回启用菜单，导致禁用菜单从树中消失、保存时被提交集合排除 → 该角色的这份权限被静默清除。
	menus, err := c.menuService.GetAllMenus()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取菜单树失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取菜单树成功", menus))
}

func (c *MenuController) GetUserMenus(ctx *gin.Context) {
	userID := ctx.GetUint("userID")
	menus, err := c.menuService.GetUserMenus(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取用户菜单失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取用户菜单成功", menus))
}

func (c *MenuController) CreateMenu(ctx *gin.Context) {
	var req models.Menu
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}
	if err := c.menuService.CreateMenu(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("创建菜单失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("创建菜单成功", nil))
}

func (c *MenuController) UpdateMenu(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "菜单ID无效"))
		return
	}

	var req models.Menu
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("参数错误", err)))
		return
	}

	if err := c.menuService.UpdateMenu(uint(id), &req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("更新菜单失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("更新菜单成功", nil))
}

func (c *MenuController) DeleteMenu(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "菜单ID无效"))
		return
	}

	if err := c.menuService.DeleteMenu(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("删除菜单失败", err)))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("删除菜单成功", nil))
}
