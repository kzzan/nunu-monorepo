// admin 定时任务进程入口：装配依赖并启动 gocron 调度。
package main

import (
	"context"
	"flag"

	"nunu-monorepo/app/admin/internal/repository"
	"nunu-monorepo/app/admin/internal/server"
	"nunu-monorepo/app/admin/internal/task"
	"nunu-monorepo/pkg/app"
	"nunu-monorepo/pkg/config"
	"nunu-monorepo/pkg/log"
	"nunu-monorepo/pkg/sid"

	"github.com/samber/do/v2"
)

// main 是定时任务进程的组合根：装配依赖后运行 TaskServer，
// 以 gocron 阻塞调度直至退出。
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
		sid.Package,
		repository.Package,
		task.Package,
		server.Package,
		func(i do.Injector) {
			do.Provide(i, func(i do.Injector) (*app.App, error) {
				return app.New(
					app.WithServer(do.MustInvoke[*server.TaskServer](i)),
					app.WithLogger(do.MustInvoke[*log.Logger](i)),
					app.WithName("demo-task"),
				), nil
			})
		},
	)
	defer func() {
		_ = injector.Shutdown()
	}()

	do.MustInvoke[*log.Logger](injector).Info().Msg("start task")
	if err := do.MustInvoke[*app.App](injector).Run(context.Background()); err != nil {
		panic(err)
	}
}
