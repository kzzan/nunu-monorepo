package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 是所有接口的统一响应包裹：业务码 + 消息 + 数据。
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// HandleSuccess 写出 code=0 的成功响应；data 为 nil 时以空对象占位。
func HandleSuccess(ctx *gin.Context, data interface{}) {
	if data == nil {
		data = map[string]interface{}{}
	}
	ctx.JSON(http.StatusOK, Response{Code: ErrSuccess.Code, Message: ErrSuccess.Message, Data: data})
}

// HandleError 写出错误响应：优先从 err 链中提取 *Error 的业务码，
// 非业务错误统一回退 500/unknown error，不泄漏内部细节。
func HandleError(ctx *gin.Context, httpCode int, err error, data interface{}) {
	if data == nil {
		data = map[string]string{}
	}
	var bizErr *Error
	if !errors.As(err, &bizErr) {
		// 非业务错误不向客户端透出内部细节
		bizErr = &Error{Code: http.StatusInternalServerError, Message: "unknown error"}
	}
	ctx.JSON(httpCode, Response{Code: bizErr.Code, Message: bizErr.Message, Data: data})
}
