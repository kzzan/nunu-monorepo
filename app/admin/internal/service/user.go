package service

import (
	"context"

	"nunu-monorepo/app/admin/internal/model"
	"nunu-monorepo/app/admin/internal/repository"

	"github.com/samber/do/v2"
)

// UserService 是脚手架演示用的用户服务，配合 model.User 占位实体
// 展示分层链路；接入真实业务前应替换或删除。
type UserService interface {
	GetUser(ctx context.Context, id int64) (*model.User, error)
}

// NewUserService 构造演示服务，由注入容器调用。
func NewUserService(i do.Injector) (UserService, error) {
	return &userService{
		Service:        do.MustInvoke[*Service](i),
		userRepository: do.MustInvoke[repository.UserRepository](i),
	}, nil
}

// userService 实现 UserService。
type userService struct {
	*Service
	userRepository repository.UserRepository
}

// GetUser 返回占位用户。
func (s *userService) GetUser(ctx context.Context, id int64) (*model.User, error) {
	return s.userRepository.GetUser(ctx, id)
}
