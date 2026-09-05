// Package job 承载长驻后台任务（随 server 进程或独立进程运行）。
package job

import (
	"nunu-monorepo/app/admin/internal/repository"
	"nunu-monorepo/pkg/log"
	"nunu-monorepo/pkg/sid"

	"github.com/samber/do/v2"
)

// Job 是长驻任务的公共依赖。
type Job struct {
	logger *log.Logger
	sid    *sid.Sid
	tm     repository.Transaction
}

// Package registers all job-layer providers into the injector.
var Package = do.Package(
	do.Lazy(New),
	do.Lazy(NewUserJob),
)

// New 构造任务公共依赖，由注入容器调用。
func New(i do.Injector) (*Job, error) {
	return &Job{
		logger: do.MustInvoke[*log.Logger](i),
		sid:    do.MustInvoke[*sid.Sid](i),
		tm:     do.MustInvoke[repository.Transaction](i),
	}, nil
}
