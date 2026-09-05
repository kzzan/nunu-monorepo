package middleware

import (
	"net/http"

	v1 "nunu-monorepo/app/admin/api/v1"
	"nunu-monorepo/pkg/jwt"
	"nunu-monorepo/pkg/log"

	"github.com/gin-gonic/gin"
)

// StrictAuth 是强制认证中间件：Authorization 头缺失或非法时直接 401，
// 校验通过后把声明写入上下文（键 "claims"）。
func StrictAuth(j *jwt.JWT, logger *log.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := ctx.Request.Header.Get("Authorization")
		if tokenString == "" {
			logger.WithContext(ctx).Warn().Interface("data", map[string]interface{}{
				"url":    ctx.Request.URL,
				"params": ctx.Params,
			}).Msg("No token")
			v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
			ctx.Abort()
			return
		}

		claims, err := j.ParseToken(tokenString)
		if err != nil {
			logger.WithContext(ctx).Error().Interface("data", map[string]interface{}{
				"url":    ctx.Request.URL,
				"params": ctx.Params,
			}).Err(err).Msg("token error")
			v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, nil)
			ctx.Abort()
			return
		}

		ctx.Set("claims", claims)
		recoveryLoggerFunc(ctx, logger)
		ctx.Next()
	}
}

// NoStrictAuth 是宽松认证中间件：依次尝试 Authorization 头、Cookie 与
// Query 中的令牌，取到且合法则注入声明；否则匿名放行。
func NoStrictAuth(j *jwt.JWT, logger *log.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := ctx.Request.Header.Get("Authorization")
		if tokenString == "" {
			tokenString, _ = ctx.Cookie("accessToken")
		}
		if tokenString == "" {
			tokenString = ctx.Query("accessToken")
		}
		if tokenString == "" {
			ctx.Next()
			return
		}

		claims, err := j.ParseToken(tokenString)
		if err != nil {
			ctx.Next()
			return
		}

		ctx.Set("claims", claims)
		recoveryLoggerFunc(ctx, logger)
		ctx.Next()
	}
}

// recoveryLoggerFunc 把当前用户 ID 附加到请求级日志器。
func recoveryLoggerFunc(ctx *gin.Context, logger *log.Logger) {
	if userInfo, ok := ctx.MustGet("claims").(*jwt.MyCustomClaims); ok {
		logger.WithValue(ctx, "UserId", userInfo.UserId)
	}
}
