package job

import (
	"context"

	"nunu-monorepo/app/admin/internal/repository"

	"github.com/samber/do/v2"
)

// UserJob 是用户相关的长驻任务接口（示例：Kafka 消费）。
type UserJob interface {
	KafkaConsumer(ctx context.Context) error
}

// NewUserJob 构造用户任务，由注入容器调用。
func NewUserJob(i do.Injector) (UserJob, error) {
	return &userJob{
		userRepo: do.MustInvoke[repository.UserRepository](i),
		Job:      do.MustInvoke[*Job](i),
	}, nil
}

// userJob 实现用户长驻任务。
type userJob struct {
	userRepo repository.UserRepository
	*Job
}

// KafkaConsumer 是消费任务的占位实现。
func (t userJob) KafkaConsumer(ctx context.Context) error {
	// do something
	return nil
}
