// home 公开门面服务入口：装配依赖并启动 Gin 服务器。
package main

import (
	"context"
	"flag"
	"fmt"

	"nunu-monorepo/app/home/internal/handler"
	"nunu-monorepo/app/home/internal/router"
	"nunu-monorepo/app/home/internal/server"
	"nunu-monorepo/app/home/internal/service"
	"nunu-monorepo/pkg/app"
	"nunu-monorepo/pkg/config"
	"nunu-monorepo/pkg/log"
	httpx "nunu-monorepo/pkg/server/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
	"github.com/spf13/viper"
)

// main 是 home 服务的组合根：装配配置、日志与各层后启动 HTTP 服务。
func main() {
	var envConf = flag.String("conf", "config/home/local.yml", "config path, eg: -conf ./config/home/local.yml")
	flag.Parse()
	conf, err := config.New(*envConf)
	if err != nil {
		panic(err)
	}

	injector := do.New(
		func(i do.Injector) { do.ProvideValue(i, conf) },
		log.Package,
		func(i do.Injector) {
			do.Provide(i, func(i do.Injector) (*gin.Engine, error) {
				if do.MustInvoke[*viper.Viper](i).GetString("env") == "prod" {
					gin.SetMode(gin.ReleaseMode)
				}
				return gin.New(), nil
			})
		},
		service.Package,
		handler.Package,
		router.Package,
		server.Package,
		func(i do.Injector) {
			do.Provide(i, func(i do.Injector) (*app.App, error) {
				return app.New(
					app.WithServer(do.MustInvoke[*httpx.Server](i)),
					app.WithLogger(do.MustInvoke[*log.Logger](i)),
					app.WithName("home-server"),
				), nil
			})
		},
	)
	defer func() {
		_ = injector.Shutdown()
	}()

	application := do.MustInvoke[*app.App](injector)

	do.MustInvoke[*log.Logger](injector).
		Info().
		Str("host", fmt.Sprintf("http://%s:%d", conf.GetString("http.host"), conf.GetInt("http.port"))).
		Msg("home server start")
	if err := application.Run(context.Background()); err != nil {
		panic(err)
	}
}
