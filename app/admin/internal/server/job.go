package server

import (
	"context"

	"nunu-monorepo/app/admin/internal/job"
	"nunu-monorepo/pkg/log"

	"github.com/samber/do/v2"
)

// JobServer 承载随 HTTP 进程运行的长驻任务（如消息消费）；
// 任务变重时可独立成进程，只需调整组合根。
type JobServer struct {
	log     *log.Logger
	userJob job.UserJob
}

// NewJobServer 构造任务服务，由注入容器调用。
func NewJobServer(i do.Injector) (*JobServer, error) {
	return &JobServer{
		log:     do.MustInvoke[*log.Logger](i),
		userJob: do.MustInvoke[job.UserJob](i),
	}, nil
}

// Start 启动全部长驻任务。
func (j *JobServer) Start(ctx context.Context) error {
	// Tips: If you want job to start as a separate process, just refer to the task implementation and adjust the code accordingly.

	// eg: kafka consumer
	err := j.userJob.KafkaConsumer(ctx)
	return err
}

// Stop 停止任务服务（当前任务随 ctx 生命周期退出）。
func (j *JobServer) Stop(ctx context.Context) error {
	return nil
}
