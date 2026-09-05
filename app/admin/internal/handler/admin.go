package handler

import (
	"errors"
	"net/http"

	v1 "nunu-monorepo/app/admin/api/v1"
	"nunu-monorepo/app/admin/internal/repository"
	"nunu-monorepo/app/admin/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

// AdminHandler 承载管理后台全部管理接口的 HTTP 入口。
type AdminHandler struct {
	*Handler
	adminService service.AdminService
}

// NewAdminHandler 构造管理接口处理器，由注入容器调用。
func NewAdminHandler(i do.Injector) (*AdminHandler, error) {
	return &AdminHandler{
		Handler:      do.MustInvoke[*Handler](i),
		adminService: do.MustInvoke[service.AdminService](i),
	}, nil
}

// handleAdminServiceError 把 service 层返回的业务错误映射为统一的
// HTTP 响应：参数/重复类错误回 400，未找到回 404，其余回 500。
func handleAdminServiceError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, v1.ErrBadRequest):
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, nil)
	case errors.Is(err, v1.ErrRoleAlreadyUse):
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrRoleAlreadyUse, nil)
	case errors.Is(err, repository.ErrNotFound):
		v1.HandleError(ctx, http.StatusNotFound, v1.ErrNotFound, nil)
	default:
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, nil)
	}
}

// ApiCreate godoc
// @Summary 创建API
// @Schemes
// @Description 创建新的API
// @Tags API模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.ApiCreateRequest true "参数"
// @Success 200 {object} v1.Response
// @Router /v1/admin/api [post]
