package v1

// Error 是业务错误的类型化载体：被 fmt.Errorf("%w") 包装后仍可通过
// errors.As 取回 Code，供 HandleError 统一映射响应码。
type Error struct {
	Code    int
	Message string
}

// Error 实现 error 接口，消息即用户可读的业务描述。
func (e *Error) Error() string {
	return e.Message
}

// newError 是包内构造哨兵错误的辅助函数。
func newError(code int, msg string) *Error {
	return &Error{Code: code, Message: msg}
}

// 哨兵业务错误：Code 对应响应体中的业务码，跨层透传时用
// errors.Is/errors.As 识别。
var (
	// common errors
	ErrSuccess             = newError(0, "ok")
	ErrBadRequest          = newError(400, "参数错误")
	ErrUnauthorized        = newError(401, "登录失效，请重新登录~")
	ErrNotFound            = newError(404, "数据不存在")
	ErrForbidden           = newError(403, "权限不足，请联系管理员开通权限~")
	ErrInternalServerError = newError(500, "服务器错误~")

	// more biz errors
	ErrUsernameAlreadyUse = newError(1001, "The username is already in use.")
	ErrRoleAlreadyUse     = newError(1002, "The role sid is already in use.")
)
