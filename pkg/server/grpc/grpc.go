// Package grpc 提供 gRPC 服务器实现与生命周期管理（预留骨架）。
package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"nunu-monorepo/pkg/log"

	"github.com/samber/do/v2"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// Server 是 gRPC 服务器实现，实现 server.Server 接口。
type Server struct {
	*grpc.Server
	host   string
	port   int
	logger *log.Logger
}

// Package registers the gRPC server provider into the injector.
var Package = do.Package(do.Lazy(NewServer))

// NewServer 按配置构造 gRPC 服务器，由注入容器调用。
func NewServer(i do.Injector) (*Server, error) {
	logger := do.MustInvoke[*log.Logger](i)
	conf := do.MustInvoke[*viper.Viper](i)
	return &Server{
		Server: grpc.NewServer(),
		logger: logger,
		host:   conf.GetString("grpc.host"),
		port:   conf.GetInt("grpc.port"),
	}, nil
}

// Start 监听并阻塞服务 gRPC 请求（被 app 容器以 goroutine 调起）。
func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		s.logger.Fatal().Msgf("Failed to listen: %v", err)
	}
	if err = s.Server.Serve(lis); err != nil {
		s.logger.Fatal().Msgf("Failed to serve: %v", err)
	}
	return nil

}

// Stop 优雅关闭 gRPC 服务器。
func (s *Server) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	s.Server.GracefulStop()

	s.logger.Info().Msg("Server exiting")

	return nil
}
