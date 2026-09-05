// Package middleware 提供 home 应用的 HTTP 中间件：安全头与访问日志。
package middleware

import (
	"time"

	"nunu-monorepo/pkg/log"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders 设置基础安全响应头（防嗅探、防点击劫持等）。
func SecurityHeaders() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		ctx.Header("X-Content-Type-Options", "nosniff")
		ctx.Header("X-Frame-Options", "SAMEORIGIN")
		ctx.Header("X-XSS-Protection", "0")
		ctx.Next()
	}
}

// RequestLog 记录请求方法、路径、状态码与耗时。
func RequestLog(logger *log.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()

		logger.WithContext(ctx).Info().
			Str("method", ctx.Request.Method).
			Str("path", ctx.Request.URL.Path).
			Int("status", ctx.Writer.Status()).
			Dur("latency", time.Since(start)).
			Msg("home request")
	}
}
