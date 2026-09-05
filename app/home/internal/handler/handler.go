// Package handler 是 home 应用的传输层：解析请求、调用 service、
// 写出响应；本层不包含业务规则。
package handler

import (
	"nunu-monorepo/pkg/log"

	"github.com/samber/do/v2"
)

// Handler 是 handler 层公共依赖。
type Handler struct {
	logger *log.Logger
}

// Package registers all handler-layer providers into the injector.
var Package = do.Package(
	do.Lazy(New),
	do.Lazy(NewSiteHandler),
)

// New 构造 handler 公共依赖，由注入容器调用。
func New(i do.Injector) (*Handler, error) {
	return &Handler{logger: do.MustInvoke[*log.Logger](i)}, nil
}
