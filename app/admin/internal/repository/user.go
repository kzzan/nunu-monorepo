package repository

import (
	"context"

	"github.com/samber/do/v2"
	"nunu-monorepo/app/admin/internal/model"
)

// UserRepository 是脚手架演示用的用户仓储，配合 model.User 占位实体
// 展示 repository→service→handler 的完整链路。
type UserRepository interface {
	GetUser(ctx context.Context, id int64) (*model.User, error)
}

// NewUserRepository 构造演示仓储，由注入容器调用。
func NewUserRepository(i do.Injector) (UserRepository, error) {
	return &userRepository{
		Repository: do.MustInvoke[*Repository](i),
	}, nil
}

// userRepository 实现 UserRepository。
type userRepository struct {
	*Repository
}

// GetUser 返回占位用户；演示实现不访问数据库。
func (r *userRepository) GetUser(ctx context.Context, id int64) (*model.User, error) {
	return &model.User{}, nil
}
