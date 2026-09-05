// Package service 是 admin 应用的用例编排层：组合各仓储实现业务规则，
// 向 handler 层提供面向用例的接口。业务错误统一使用 api/v1 的哨兵错误。
package service

import (
	"nunu-monorepo/app/admin/internal/repository"
	"nunu-monorepo/pkg/jwt"
	"nunu-monorepo/pkg/log"
	"nunu-monorepo/pkg/sid"

	"github.com/samber/do/v2"
)

// Service 是所有业务服务的公共依赖：日志、ID 生成、JWT 与事务边界。
type Service struct {
	logger *log.Logger
	sid    *sid.Sid
	jwt    *jwt.JWT
	tm     repository.Transaction
}

// Package registers all service-layer providers into the injector.
var Package = do.Package(
	do.Lazy(New),
	do.Lazy(NewUserService),
	do.Lazy(NewAdminService),
)

// New 构造公共依赖，由注入容器调用。
func New(i do.Injector) (*Service, error) {
	return &Service{
		logger: do.MustInvoke[*log.Logger](i),
		sid:    do.MustInvoke[*sid.Sid](i),
		jwt:    do.MustInvoke[*jwt.JWT](i),
		tm:     do.MustInvoke[repository.Transaction](i),
	}, nil
}
