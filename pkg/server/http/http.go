// Package http 提供基于 Gin 的 HTTP 服务器实现与生命周期管理。
package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"nunu-monorepo/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
	"github.com/spf13/viper"
)

// Server 是嵌入 gin.Engine 的 HTTP 服务器，实现 server.Server 接口。
type Server struct {
	*gin.Engine
	httpSrv *http.Server
	host    string
	port    int
	logger  *log.Logger
}

// Package registers the HTTP server provider into the injector.
var Package = do.Package(do.Lazy(NewServer))

// NewServer 按配置构造 HTTP 服务器，由注入容器调用。
func NewServer(i do.Injector) (*Server, error) {
	engine := do.MustInvoke[*gin.Engine](i)
	logger := do.MustInvoke[*log.Logger](i)
	conf := do.MustInvoke[*viper.Viper](i)
	return &Server{
		Engine: engine,
		logger: logger,
		host:   conf.GetString("http.host"),
		port:   conf.GetInt("http.port"),
	}, nil
}

// Start 监听并阻塞服务 HTTP 请求（被 app 容器以 goroutine 调起）。
func (s *Server) Start(ctx context.Context) error {
	s.httpSrv = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.host, s.port),
		Handler: s,
	}

	if err := s.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Fatal().Msgf("listen: %s", err)
	}

	return nil
}

// Stop 优雅关闭 HTTP 服务器，给在途请求 5 秒收尾时间。
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info().Msg("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.httpSrv.Shutdown(ctx); err != nil {
		s.logger.Fatal().Msgf("Server forced to shutdown: %v", err)
	}

	s.logger.Info().Msg("Server exiting")
	return nil
}
