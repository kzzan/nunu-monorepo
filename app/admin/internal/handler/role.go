package handler

import (
	"net/http"
	v1 "nunu-monorepo/app/admin/api/v1"

	"github.com/gin-gonic/gin"
)

// GetRoles godoc
// @Summary 获取角色列表
// @Schemes
// @Description 获取角色列表
// @Tags 角色模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int true "页码"
// @Param pageSize query int true "每页数量"
// @Param sid query string false "角色ID"
// @Param name query string false "角色名称"
// @Success 200 {object} v1.GetRolesResponse
// @Router /v1/admin/roles [get]
// GetRoles 处理分页查询角色列表请求。
func (h *AdminHandler) GetRoles(ctx *gin.Context) {
	var req v1.GetRoleListRequest
	if err := ctx.ShouldBind(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	data, err := h.adminService.GetRoles(ctx, &req)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, nil)
		return
	}

	v1.HandleSuccess(ctx, data)
}

// RoleCreate godoc
// @Summary 创建角色
// @Schemes
// @Description 创建新的角色
// @Tags 角色模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.RoleCreateRequest true "参数"
// @Success 200 {object} v1.Response
// @Router /v1/admin/role [post]
// RoleCreate 处理创建角色请求。
func (h *AdminHandler) RoleCreate(ctx *gin.Context) {
	var req v1.RoleCreateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	if err := h.adminService.RoleCreate(ctx, &req); err != nil {
		handleAdminServiceError(ctx, err)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// RoleUpdate godoc
// @Summary 更新角色
// @Schemes
// @Description 更新角色信息
// @Tags 角色模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.RoleUpdateRequest true "参数"
// @Success 200 {object} v1.Response
// @Router /v1/admin/role [put]
func (h *AdminHandler) RoleUpdate(ctx *gin.Context) {
	var req v1.RoleUpdateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	if err := h.adminService.RoleUpdate(ctx, &req); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// RoleDelete godoc
// @Summary 删除角色
// @Schemes
// @Description 删除指定角色
// @Tags 角色模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param id query uint true "角色ID"
// @Success 200 {object} v1.Response
// @Router /v1/admin/role [delete]
func (h *AdminHandler) RoleDelete(ctx *gin.Context) {
	var req v1.RoleDeleteRequest
	if err := ctx.ShouldBind(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	if err := h.adminService.RoleDelete(ctx, req.ID); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}
