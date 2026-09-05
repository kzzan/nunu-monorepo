package handler

import (
	"net/http"
	v1 "nunu-monorepo/app/admin/api/v1"

	"github.com/gin-gonic/gin"
)

// GetApis godoc
// @Summary 获取API列表
// @Schemes
// @Description 获取API列表
// @Tags API模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int true "页码"
// @Param pageSize query int true "每页数量"
// @Param group query string false "API分组"
// @Param name query string false "API名称"
// @Param path query string false "API路径"
// @Param method query string false "请求方法"
// @Success 200 {object} v1.GetApisResponse
// @Router /v1/admin/apis [get]
// GetApis 处理分页查询 API 列表请求。
func (h *AdminHandler) GetApis(ctx *gin.Context) {
	var req v1.GetApisRequest
	if err := ctx.ShouldBind(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	data, err := h.adminService.GetApis(ctx, &req)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, nil)
		return
	}

	v1.HandleSuccess(ctx, data)
}

// ApiCreate 处理创建 API 资源请求。
func (h *AdminHandler) ApiCreate(ctx *gin.Context) {
	var req v1.ApiCreateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	if err := h.adminService.ApiCreate(ctx, &req); err != nil {
		handleAdminServiceError(ctx, err)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// ApiUpdate godoc
// @Summary 更新API
// @Schemes
// @Description 更新API信息
// @Tags API模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.ApiUpdateRequest true "参数"
// @Success 200 {object} v1.Response
// @Router /v1/admin/api [put]
func (h *AdminHandler) ApiUpdate(ctx *gin.Context) {
	var req v1.ApiUpdateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	if err := h.adminService.ApiUpdate(ctx, &req); err != nil {
		handleAdminServiceError(ctx, err)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// ApiDelete godoc
// @Summary 删除API
// @Schemes
// @Description 删除指定API
// @Tags API模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param id query uint true "API ID"
// @Success 200 {object} v1.Response
// @Router /v1/admin/api [delete]
func (h *AdminHandler) ApiDelete(ctx *gin.Context) {
	var req v1.ApiDeleteRequest
	if err := ctx.ShouldBind(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
		return
	}
	if err := h.adminService.ApiDelete(ctx, req.ID); err != nil {
		handleAdminServiceError(ctx, err)
		return
	}
	v1.HandleSuccess(ctx, nil)
}
