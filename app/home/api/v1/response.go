// Package v1 定义 home 应用的 API 契约：请求/响应 DTO 与响应辅助函数。
package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 是 home 接口的统一响应包裹。
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// MetaData 是运行时元信息（应用名、阶段、入口、标题）。
type MetaData struct {
	App   string `json:"app"`
	Stage string `json:"stage"`
	Entry string `json:"entry"`
	Title string `json:"title"`
}

// HealthData 是健康检查结果。
type HealthData struct {
	App    string `json:"app"`
	Status string `json:"status"`
}

// ManifestFeature 描述站点的一个特性入口。
type ManifestFeature struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Route       string `json:"route"`
}

// ManifestData 是前端引导用的站点清单。
type ManifestData struct {
	App      string            `json:"app"`
	Stage    string            `json:"stage"`
	Entry    string            `json:"entry"`
	Title    string            `json:"title"`
	Headline string            `json:"headline"`
	Features []ManifestFeature `json:"features"`
}

// HandleSuccess 以 code=0 写出成功响应。
func HandleSuccess(ctx *gin.Context, data interface{}) {
	if data == nil {
		data = map[string]interface{}{}
	}

	ctx.JSON(http.StatusOK, Response{Code: 0, Message: "ok", Data: data})
}

// HandleError 以指定 HTTP 状态码写出错误响应，业务码与状态码一致。
func HandleError(ctx *gin.Context, httpCode int, message string, data interface{}) {
	if data == nil {
		data = map[string]interface{}{}
	}

	ctx.JSON(httpCode, Response{Code: httpCode, Message: message, Data: data})
}
