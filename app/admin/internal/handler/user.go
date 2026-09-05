package handler

import (
	"nunu-monorepo/app/admin/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

// UserHandler 是脚手架演示的用户接口处理器。
type UserHandler struct {
	*Handler
	userService service.UserService
}

// NewUserHandler 构造演示处理器，由注入容器调用。
func NewUserHandler(i do.Injector) (*UserHandler, error) {
	return &UserHandler{
		Handler:     do.MustInvoke[*Handler](i),
		userService: do.MustInvoke[service.UserService](i),
	}, nil
}

// GetUsers 是演示占位接口，未实现查询逻辑。
func (h *UserHandler) GetUsers(ctx *gin.Context) {

}
