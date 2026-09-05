// admin HTTP 服务入口：装配依赖并启动 Gin 服务器。
package main

import (
	"context"
	"flag"
	"fmt"

	"nunu-monorepo/app/admin/internal/handler"
	"nunu-monorepo/app/admin/internal/job"
	"nunu-monorepo/app/admin/internal/repository"
	"nunu-monorepo/app/admin/internal/server"
	"nunu-monorepo/app/admin/internal/service"
	"nunu-monorepo/pkg/app"
	"nunu-monorepo/pkg/config"
	"nunu-monorepo/pkg/jwt"
	"nunu-monorepo/pkg/log"
	httpx "nunu-monorepo/pkg/server/http"
	"nunu-monorepo/pkg/sid"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

// @title           Nunu Admin API
// @version         1.0.0
// @description     Admin API for the nunu monorepo.
// @termsOfService  http://swagger.io/terms/
// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io
// @license.name  Apache 2.0
// @license.url    http://www.apache.org/licenses/LICENSE-2.0.html
// @host      localhost:8000
// @BasePath  /
// @securityDefinitions.apiKey Bearer
// @in header
// @name Authorization
// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
// main 是 admin HTTP 服务的组合根：按顺序装配配置、日志、JWT、
// ID 生成、各业务层与服务器，随后启动并阻塞直至收到退出信号。
func main() {
	var envConf = flag.String("conf", "config/admin/local.yml", "config path, eg: -conf ./config/admin/local.yml")
	flag.Parse()
	conf, err := config.New(*envConf)
	if err != nil {
		panic(err)
	}

	injector := do.New(
		func(i do.Injector) { do.ProvideValue(i, conf) },
		log.Package,
		jwt.Package,
		sid.Package,
		func(i do.Injector) {
			do.Provide(i, func(i do.Injector) (*gin.Engine, error) {
				return gin.Default(), nil
			})
		},
		repository.Package,
		service.Package,
		handler.Package,
		job.Package,
		server.Package,
		func(i do.Injector) {
			do.Provide(i, func(i do.Injector) (*app.App, error) {
				return app.New(
					app.WithServer(do.MustInvoke[*httpx.Server](i), do.MustInvoke[*server.JobServer](i)),
					app.WithLogger(do.MustInvoke[*log.Logger](i)),
					app.WithName("demo-server"),
				), nil
			})
		},
	)
	defer func() {
		_ = injector.Shutdown()
	}()

	application := do.MustInvoke[*app.App](injector)

	logger := do.MustInvoke[*log.Logger](injector)
	logger.Info().Str("host", fmt.Sprintf("http://%s:%d", conf.GetString("http.host"), conf.GetInt("http.port"))).Msg("server start")
	logger.Info().Str("addr", fmt.Sprintf("http://%s:%d/swagger/index.html", conf.GetString("http.host"), conf.GetInt("http.port"))).Msg("docs addr")
	if err := application.Run(context.Background()); err != nil {
		panic(err)
	}
}
