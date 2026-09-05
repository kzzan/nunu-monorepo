package middleware

import (
	"bytes"
	"io"
	"time"

	"nunu-monorepo/pkg/log"

	"github.com/duke-git/lancet/v2/cryptor"
	"github.com/duke-git/lancet/v2/random"
	"github.com/gin-gonic/gin"
)

// RequestLogMiddleware 记录入站请求的方法、路径与耗时。
func RequestLogMiddleware(logger *log.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// The configuration is initialized once per request
		uuid, err := random.UUIdV4()
		if err != nil {
			return
		}
		trace := cryptor.Md5String(uuid)
		logger.WithValue(ctx, "trace", trace)
		logger.WithValue(ctx, "request_method", ctx.Request.Method)
		logger.WithValue(ctx, "request_headers", ctx.Request.Header)
		logger.WithValue(ctx, "request_url", ctx.Request.URL.String())
		if ctx.Request.Body != nil {
			bodyBytes, _ := ctx.GetRawData()
			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // 关键点
			logger.WithValue(ctx, "request_params", string(bodyBytes))
		}
		logger.WithContext(ctx).Info().Msg("Request")
		ctx.Next()
	}
}

// ResponseLogMiddleware 记录响应体（开发排查用，生产可按级别关闭）。
func ResponseLogMiddleware(logger *log.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
		ctx.Writer = blw
		startTime := time.Now()
		ctx.Next()
		duration := time.Since(startTime).String()
		logger.WithContext(ctx).Info().Interface("response_body", blw.body.String()).Str("time", duration).Msg("Response")
	}
}

// bodyLogWriter 包装 gin 响应写入器以捕获响应体用于日志。
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 同时写入底层响应与日志缓冲。
func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
