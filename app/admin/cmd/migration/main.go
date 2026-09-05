// admin 数据库迁移命令：重建表结构并灌入种子数据（破坏性）。
package main

import (
	"context"
	"flag"

	"nunu-monorepo/app/admin/internal/repository"
	"nunu-monorepo/app/admin/internal/server"
	"nunu-monorepo/pkg/app"
	"nunu-monorepo/pkg/config"
	"nunu-monorepo/pkg/log"
	"nunu-monorepo/pkg/sid"

	"github.com/samber/do/v2"
)

// main 是迁移命令的组合根：装配依赖后运行 MigrateServer，
// 完成破坏性重建与种子灌入即退出。
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
		server.Package,
		func(i do.Injector) {
			do.Provide(i, func(i do.Injector) (*app.App, error) {
				return app.New(
					app.WithServer(do.MustInvoke[*server.MigrateServer](i)),
					app.WithLogger(do.MustInvoke[*log.Logger](i)),
					app.WithName("demo-migrate"),
				), nil
			})
		},
	)
	defer func() {
		_ = injector.Shutdown()
	}()

	if err := do.MustInvoke[*app.App](injector).Run(context.Background()); err != nil {
		panic(err)
	}
}
