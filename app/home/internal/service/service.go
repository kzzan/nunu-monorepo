// Package service 是 home 应用的用例层：实现站点元信息等接口的业务语义。
package service

import (
	"nunu-monorepo/pkg/log"

	"github.com/samber/do/v2"
	"github.com/spf13/viper"
)

// Service 是 home 服务的公共依赖：配置与日志。
type Service struct {
	config *viper.Viper
	logger *log.Logger
}

// Package registers all service-layer providers into the injector.
var Package = do.Package(
	do.Lazy(New),
	do.Lazy(NewSiteService),
)

// New 构造公共依赖，由注入容器调用。
func New(i do.Injector) (*Service, error) {
	return &Service{
		config: do.MustInvoke[*viper.Viper](i),
		logger: do.MustInvoke[*log.Logger](i),
	}, nil
}
