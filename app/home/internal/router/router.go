// Package router 注册 home 应用的 HTTP 路由。
package router

import (
	"nunu-monorepo/app/home/internal/handler"
	"nunu-monorepo/pkg/log"

	"github.com/samber/do/v2"
	"github.com/spf13/viper"
)

// RouterDeps 聚合站点路由所需的依赖，解耦 router 与注入容器。
type RouterDeps struct {
	Logger      *log.Logger
	Config      *viper.Viper
	SiteHandler *handler.SiteHandler
}

// Package registers the RouterDeps provider into the injector.
var Package = do.Package(
	do.Lazy(func(i do.Injector) (RouterDeps, error) {
		return RouterDeps{
			Logger:      do.MustInvoke[*log.Logger](i),
			Config:      do.MustInvoke[*viper.Viper](i),
			SiteHandler: do.MustInvoke[*handler.SiteHandler](i),
		}, nil
	}),
)
