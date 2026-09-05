package task

import (
	"context"

	"nunu-monorepo/app/admin/internal/repository"

	"github.com/samber/do/v2"
)

// UserTask 是用户相关的定时任务接口（示例：用户状态巡检）。
type UserTask interface {
	CheckUser(ctx context.Context) error
}

// NewUserTask 构造用户定时任务，由注入容器调用。
func NewUserTask(i do.Injector) (UserTask, error) {
	return &userTask{
		userRepo: do.MustInvoke[repository.UserRepository](i),
		Task:     do.MustInvoke[*Task](i),
	}, nil
}

// userTask 实现用户定时任务。
type userTask struct {
	userRepo repository.UserRepository
	*Task
}

// CheckUser 是巡检任务的占位实现。
func (t userTask) CheckUser(ctx context.Context) error {
	// do something
	t.logger.Info().Msg("CheckUser")
	return nil
}
