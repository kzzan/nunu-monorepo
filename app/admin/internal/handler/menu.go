package handler

import (
	"net/http"
	v1 "nunu-monorepo/app/admin/api/v1"

	"github.com/gin-gonic/gin"
)

// GetMenus godoc
// @Summary 获取用户菜单
// @Schemes
// @Description 获取当前用户的菜单列表
// @Tags 菜单模块
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} v1.GetMenuResponse
// @Router /v1/menus [get]
func (h *AdminHandler) GetMenus(ctx *gin.Context) {
	data, err := h.adminService.GetMenus(ctx, GetUserIdFromCtx(ctx))
	if err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	// 过滤权限菜单
	v1.HandleSuccess(ctx, data)
}

// GetAdminMenus godoc
// @Summary 获取管理员菜单
// @Schemes
// @Description 获取管理员菜单列表
// @Tags 菜单模块
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} v1.GetMenuResponse
// @Router /v1/admin/menus [get]
func (h *AdminHandler) GetAdminMenus(ctx *gin.Context) {
	data, err := h.adminService.GetAdminMenus(ctx)
	if err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	// 过滤权限菜单
	v1.HandleSuccess(ctx, data)
}

// MenuUpdate godoc
// @Summary 更新菜单
// @Schemes
// @Description 更新菜单信息
// @Tags 菜单模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.MenuUpdateRequest true "参数"
// @Success 200 {object} v1.Response
// @Router /v1/admin/menu [put]
func (h *AdminHandler) MenuUpdate(ctx *gin.Context) {
	var req v1.MenuUpdateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	if err := h.adminService.MenuUpdate(ctx, &req); err != nil {
		handleAdminServiceError(ctx, err)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// MenuCreate godoc
// @Summary 创建菜单
// @Schemes
// @Description 创建新的菜单
// @Tags 菜单模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.MenuCreateRequest true "参数"
// @Success 200 {object} v1.Response
// @Router /v1/admin/menu [post]
func (h *AdminHandler) MenuCreate(ctx *gin.Context) {
	var req v1.MenuCreateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	if err := h.adminService.MenuCreate(ctx, &req); err != nil {
		handleAdminServiceError(ctx, err)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// MenuDelete godoc
// @Summary 删除菜单
// @Schemes
// @Description 删除指定菜单
// @Tags 菜单模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param id query uint true "菜单ID"
// @Success 200 {object} v1.Response
// @Router /v1/admin/menu [delete]
func (h *AdminHandler) MenuDelete(ctx *gin.Context) {
	var req v1.MenuDeleteRequest
	if err := ctx.ShouldBind(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	if err := h.adminService.MenuDelete(ctx, req.ID); err != nil {
		handleAdminServiceError(ctx, err)
		return

	}
	v1.HandleSuccess(ctx, nil)
}
