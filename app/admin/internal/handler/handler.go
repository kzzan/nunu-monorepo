// Package handler 是 admin 应用的传输层：解析 HTTP 请求、调用 service、
// 统一写出响应；本层不包含业务规则。
package handler

import (
	"nunu-monorepo/pkg/jwt"
	"nunu-monorepo/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

// Handler 是 handler 层公共依赖。
type Handler struct {
	logger *log.Logger
}

// Package registers all handler-layer providers into the injector.
var Package = do.Package(
	do.Lazy(New),
	do.Lazy(NewUserHandler),
	do.Lazy(NewAdminHandler),
)

// New 构造 handler 公共依赖，由注入容器调用。
func New(i do.Injector) (*Handler, error) {
	return &Handler{
		logger: do.MustInvoke[*log.Logger](i),
	}, nil
}

// GetUserIdFromCtx 从 gin 上下文取出 JWT 中间件写入的当前用户 ID；
// 未认证时返回 0。
func GetUserIdFromCtx(ctx *gin.Context) uint {
	v, exists := ctx.Get("claims")
	if !exists {
		return 0
	}
	return v.(*jwt.MyCustomClaims).UserId
}
