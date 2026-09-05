package server_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"nunu-monorepo/app/home/internal/handler"
	"nunu-monorepo/app/home/internal/router"
	server "nunu-monorepo/app/home/internal/server"
	"nunu-monorepo/app/home/internal/service"
	"nunu-monorepo/pkg/log"
	httpx "nunu-monorepo/pkg/server/http"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// TestHomeHTTPServerRoutes 逐一路由验证 home 服务的页面与 API 响应。
func TestHomeHTTPServerRoutes(t *testing.T) {
	server := newTestServer(t)

	t.Run("index page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		resp := httptest.NewRecorder()

		server.ServeHTTP(resp, req)

		require.Equal(t, http.StatusOK, resp.Code)
		require.Contains(t, resp.Body.String(), "Public-facing shell")
	})

	t.Run("health", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		resp := httptest.NewRecorder()

		server.ServeHTTP(resp, req)

		require.Equal(t, http.StatusOK, resp.Code)
		require.Contains(t, resp.Body.String(), "\"status\":\"ok\"")
		require.Contains(t, resp.Body.String(), "\"app\":\"home\"")
	})

	t.Run("meta", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/meta", nil)
		resp := httptest.NewRecorder()

		server.ServeHTTP(resp, req)

		require.Equal(t, http.StatusOK, resp.Code)
		require.Contains(t, resp.Body.String(), "\"entry\":\"app/home/cmd/server\"")
		require.Contains(t, resp.Body.String(), "\"title\":\"Home Test\"")
	})

	t.Run("manifest", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/manifest", nil)
		resp := httptest.NewRecorder()

		server.ServeHTTP(resp, req)

		require.Equal(t, http.StatusOK, resp.Code)
		require.Contains(t, resp.Body.String(), "\"headline\":\"A focused home test shell.\"")
		require.Contains(t, resp.Body.String(), "Bootstrap Manifest")
	})
}

// newTestServer 构造带完整路由的测试服务器。
func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	return do.MustInvoke[*httpx.Server](newTestInjector(t))
}

// newTestInjector 构造带测试配置的注入容器。
func newTestInjector(t *testing.T) do.Injector {
	t.Helper()

	conf := viper.New()
	conf.Set("env", "test")
	conf.Set("http.host", "127.0.0.1")
	conf.Set("http.port", 8081)
	conf.Set("site.title", "Home Test")
	conf.Set("site.headline", "A focused home test shell.")
	conf.Set("log.log_level", "debug")
	conf.Set("log.mode", "console")
	conf.Set("log.encoding", "console")
	conf.Set("log.log_file_name", filepath.Join(t.TempDir(), "home-test.log"))
	conf.Set("log.max_backups", 1)
	conf.Set("log.max_age", 1)
	conf.Set("log.max_size", 1)
	conf.Set("log.compress", false)

	injector := do.New(
		func(i do.Injector) { do.ProvideValue(i, conf) },
		log.Package,
		func(i do.Injector) {
			do.Provide(i, func(i do.Injector) (*gin.Engine, error) {
				return gin.New(), nil
			})
		},
		service.Package,
		handler.Package,
		router.Package,
		server.Package,
	)

	return injector
}
