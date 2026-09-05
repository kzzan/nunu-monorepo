// Package middleware 提供 admin 应用的 HTTP 中间件：
// 认证（JWT）、鉴权（Casbin RBAC）、跨域、签名与访问日志。
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware 返回跨域中间件：放行全部来源与方法，
// 并响应预检请求。开发态友好，生产环境应按需收紧来源。
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", c.GetHeader("Origin"))
		c.Header("Access-Control-Allow-Credentials", "true")

		if method == "OPTIONS" {
			c.Header("Access-Control-Allow-Methods", c.GetHeader("Access-Control-Request-Method"))
			c.Header("Access-Control-Allow-Headers", c.GetHeader("Access-Control-Request-Headers"))
			c.Header("Access-Control-Max-Age", "7200")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
