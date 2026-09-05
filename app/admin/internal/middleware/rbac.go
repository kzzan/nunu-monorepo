package middleware

import (
	"net/http"
	v1 "nunu-monorepo/app/admin/api/v1"
	"nunu-monorepo/app/admin/internal/model"
	"nunu-monorepo/pkg/jwt"

	"github.com/casbin/casbin/v3"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware 是基于 Casbin 的接口鉴权中间件：以 "用户 ID" 为主语、
// "api:路径" 为资源、HTTP 方法为动作做 Enforce 判定；
// 内置超管（AdminUserID）直接放行。
func AuthMiddleware(e *casbin.SyncedEnforcer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 从上下文获取用户信息（假设通过 JWT 或其他方式设置）
		v, exists := ctx.Get("claims")
		if !exists {
			v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
			ctx.Abort()
			return
		}
		uid := v.(*jwt.MyCustomClaims).UserId
		if convertor.ToString(uid) == model.AdminUserID {
			// 防呆设计，超管跳过API权限检查
			ctx.Next()
			return
		}

		// 获取请求的资源和操作
		sub := convertor.ToString(uid)
		obj := model.ApiResourcePrefix + ctx.Request.URL.Path
		act := ctx.Request.Method

		// 检查权限
		allowed, err := e.Enforce(sub, obj, act)
		if err != nil {
			v1.HandleError(ctx, http.StatusForbidden, v1.ErrForbidden, nil)
			ctx.Abort()
			return
		}
		if !allowed {
			v1.HandleError(ctx, http.StatusForbidden, v1.ErrForbidden, nil)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
