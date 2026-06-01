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
	menus, err := c.menuService.GetMenuList()
	if err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "获取菜单树失败"))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("获取菜单树成功", menus))
}

func (c *MenuController) CreateMenu(ctx *gin.Context) {
	var req models.Menu
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: "+err.Error()))
		return
	}
	if err := c.menuService.CreateMenu(&req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "创建菜单失败: "+err.Error()))
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
		ctx.JSON(http.StatusOK, utils.Error(1, "参数错误: "+err.Error()))
		return
	}

	if err := c.menuService.UpdateMenu(uint(id), &req); err != nil {
		ctx.JSON(http.StatusOK, utils.Error(1, "更新菜单失败: "+err.Error()))
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
		ctx.JSON(http.StatusOK, utils.Error(1, "删除菜单失败: "+err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, utils.Success("删除菜单成功", nil))
}
