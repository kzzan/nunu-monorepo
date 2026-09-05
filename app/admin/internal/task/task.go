// Package task 承载基于 gocron 的定时任务。
package task

import (
	"nunu-monorepo/app/admin/internal/repository"
	"nunu-monorepo/pkg/log"
	"nunu-monorepo/pkg/sid"

	"github.com/samber/do/v2"
)

// Task 是定时任务的公共依赖。
type Task struct {
	logger *log.Logger
	sid    *sid.Sid
	tm     repository.Transaction
}

// Package registers all task-layer providers into the injector.
var Package = do.Package(
	do.Lazy(New),
	do.Lazy(NewUserTask),
)

// New 构造定时任务公共依赖，由注入容器调用。
func New(i do.Injector) (*Task, error) {
	return &Task{
		logger: do.MustInvoke[*log.Logger](i),
		sid:    do.MustInvoke[*sid.Sid](i),
		tm:     do.MustInvoke[repository.Transaction](i),
	}, nil
}
