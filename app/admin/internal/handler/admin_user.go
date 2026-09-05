package handler

import (
	"net/http"
	v1 "nunu-monorepo/app/admin/api/v1"

	"github.com/gin-gonic/gin"
)

// AdminUserUpdate godoc
// @Summary 更新管理员用户
// @Schemes
// @Description 更新管理员用户信息
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.AdminUserUpdateRequest true "参数"
// @Success 200 {object} v1.Response
// @Router /v1/admin/user [put]
func (h *AdminHandler) AdminUserUpdate(ctx *gin.Context) {
	var req v1.AdminUserUpdateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	if err := h.adminService.AdminUserUpdate(ctx, &req); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// AdminUserCreate godoc
// @Summary 创建管理员用户
// @Schemes
// @Description 创建新的管理员用户
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.AdminUserCreateRequest true "参数"
// @Success 200 {object} v1.Response
// @Router /v1/admin/user [post]
// AdminUserCreate 处理创建管理员请求。
func (h *AdminHandler) AdminUserCreate(ctx *gin.Context) {
	var req v1.AdminUserCreateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	if err := h.adminService.AdminUserCreate(ctx, &req); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, nil)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// AdminUserDelete godoc
// @Summary 删除管理员用户
// @Schemes
// @Description 删除指定管理员用户
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param id query uint true "用户ID"
// @Success 200 {object} v1.Response
// @Router /v1/admin/user [delete]
// AdminUserDelete 处理删除管理员请求。
func (h *AdminHandler) AdminUserDelete(ctx *gin.Context) {
	var req v1.AdminUserDeleteRequest
	if err := ctx.ShouldBind(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	if err := h.adminService.AdminUserDelete(ctx, req.ID); err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, nil)
		return

	}
	v1.HandleSuccess(ctx, nil)
}

// GetAdminUsers godoc
// @Summary 获取管理员用户列表
// @Schemes
// @Description 获取管理员用户列表
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int true "页码"
// @Param pageSize query int true "每页数量"
// @Param id query int false "用户ID"
// @Param username query string false "用户名"
// @Param nickname query string false "昵称"
// @Param phone query string false "手机号"
// @Param email query string false "邮箱"
// @Success 200 {object} v1.GetAdminUsersResponse
// @Router /v1/admin/users [get]
// GetAdminUsers 处理分页查询管理员列表请求。
func (h *AdminHandler) GetAdminUsers(ctx *gin.Context) {
	var req v1.GetAdminUsersRequest
	if err := ctx.ShouldBind(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	data, err := h.adminService.GetAdminUsers(ctx, &req)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, nil)
		return
	}

	v1.HandleSuccess(ctx, data)
}

// GetAdminUser godoc
// @Summary 获取管理用户信息
// @Schemes
// @Description
// @Tags 用户模块
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} v1.GetAdminUserResponse
// @Router /v1/admin/user [get]
// GetAdminUser 处理查询当前登录管理员详情的请求。
func (h *AdminHandler) GetAdminUser(ctx *gin.Context) {
	data, err := h.adminService.GetAdminUser(ctx, GetUserIdFromCtx(ctx))
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, nil)
		return
	}

	v1.HandleSuccess(ctx, data)
}
