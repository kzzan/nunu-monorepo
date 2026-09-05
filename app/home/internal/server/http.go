package server

import (
	"nunu-monorepo/app/home/internal/middleware"
	"nunu-monorepo/app/home/internal/router"
	pkghttp "nunu-monorepo/pkg/server/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

// Package registers all server-layer providers into the injector.
var Package = do.Package(
	do.Lazy(NewHTTPServer),
)

// NewHTTPServer 组装 home 的 HTTP 服务：安全头、访问日志中间件
// 与站点路由；prod 环境切换 Gin 到 Release 模式。
func NewHTTPServer(i do.Injector) (*pkghttp.Server, error) {
	deps := do.MustInvoke[router.RouterDeps](i)

	if deps.Config.GetString("env") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	s, err := pkghttp.NewServer(i)
	if err != nil {
		return nil, err
	}

	s.Use(
		gin.Recovery(),
		middleware.SecurityHeaders(),
		middleware.RequestLog(deps.Logger),
	)

	router.InitSiteRouter(deps, s.Engine)

	return s, nil
}
