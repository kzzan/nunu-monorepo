package handler

import (
	"net/http"
	v1 "nunu-monorepo/app/admin/api/v1"

	"github.com/gin-gonic/gin"
)

// GetUserPermissions godoc
// @Summary 获取用户权限
// @Schemes
// @Description 获取当前用户的权限列表
// @Tags 权限模块
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} v1.GetUserPermissionsData
// @Router /v1/admin/user/permissions [get]
func (h *AdminHandler) GetUserPermissions(ctx *gin.Context) {
	data, err := h.adminService.GetUserPermissions(ctx, GetUserIdFromCtx(ctx))
	if err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	// 过滤权限菜单
	v1.HandleSuccess(ctx, data)
}

// GetRolePermissions godoc
// @Summary 获取角色权限
// @Schemes
// @Description 获取指定角色的权限列表
// @Tags 权限模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param role query string true "角色名称"
// @Success 200 {object} v1.GetRolePermissionsData
// @Router /v1/admin/role/permissions [get]
func (h *AdminHandler) GetRolePermissions(ctx *gin.Context) {
	var req v1.GetRolePermissionsRequest
	if err := ctx.ShouldBind(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	data, err := h.adminService.GetRolePermissions(ctx, req.Role)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, nil)
		return
	}
	v1.HandleSuccess(ctx, data)
}

// UpdateRolePermission godoc
// @Summary 更新角色权限
// @Schemes
// @Description 更新指定角色的权限列表
// @Tags 权限模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.UpdateRolePermissionRequest true "参数"
// @Success 200 {object} v1.Response
// @Router /v1/admin/role/permissions [put]
func (h *AdminHandler) UpdateRolePermission(ctx *gin.Context) {
	var req v1.UpdateRolePermissionRequest
	if err := ctx.ShouldBind(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	err := h.adminService.UpdateRolePermission(ctx, &req)
	if err != nil {
		handleAdminServiceError(ctx, err)
		return
	}
	v1.HandleSuccess(ctx, nil)
}
