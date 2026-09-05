package server

import (
	"context"
	"time"

	"nunu-monorepo/app/admin/internal/task"
	"nunu-monorepo/pkg/log"

	"github.com/go-co-op/gocron"
	"github.com/samber/do/v2"
)

// TaskServer 承载基于 gocron 的定时任务进程（cmd/task 入口）。
type TaskServer struct {
	log       *log.Logger
	scheduler *gocron.Scheduler
	userTask  task.UserTask
}

// NewTaskServer 构造定时任务服务，由注入容器调用。
func NewTaskServer(i do.Injector) (*TaskServer, error) {
	return &TaskServer{
		log:      do.MustInvoke[*log.Logger](i),
		userTask: do.MustInvoke[task.UserTask](i),
	}, nil
}

// Start 注册全部定时任务并阻塞运行（StartBlocking 直至 Stop）。
func (t *TaskServer) Start(ctx context.Context) error {
	gocron.SetPanicHandler(func(jobName string, recoverData interface{}) {
		t.log.Error().Str("job", jobName).Interface("recover", recoverData).Msg("TaskServer Panic")
	})

	// eg: crontab task
	t.scheduler = gocron.NewScheduler(time.UTC)
	// if you are in China, you will need to change the time zone as follows
	// t.scheduler = gocron.NewScheduler(time.FixedZone("PRC", 8*60*60))

	//_, err := t.scheduler.Every("3s").Do(func()
	_, err := t.scheduler.CronWithSeconds("0/3 * * * * *").Do(func() {
		err := t.userTask.CheckUser(ctx)
		if err != nil {
			t.log.Error().Err(err).Msg("CheckUser error")
		}
	})
	if err != nil {
		t.log.Error().Err(err).Msg("CheckUser error")
	}

	t.scheduler.StartBlocking()
	return nil
}

// Stop 停止调度器。
func (t *TaskServer) Stop(ctx context.Context) error {
	t.scheduler.Stop()
	t.log.Info().Msg("TaskServer stop...")
	return nil
}
