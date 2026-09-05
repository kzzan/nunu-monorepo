package service_test

import (
	"path/filepath"
	"testing"

	"nunu-monorepo/app/home/internal/service"
	"nunu-monorepo/pkg/log"

	"github.com/samber/do/v2"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// TestSiteServiceReturnsConfiguredManifest 验证站点服务按测试配置
// 返回健康态、元信息与清单。
func TestSiteServiceReturnsConfiguredManifest(t *testing.T) {
	injector := newTestInjector(t)
	siteService := do.MustInvoke[service.SiteService](injector)

	health := siteService.Health()
	require.Equal(t, "home", health.App)
	require.Equal(t, "ok", health.Status)

	meta := siteService.Meta()
	require.Equal(t, "test", meta.Stage)
	require.Equal(t, "Home Test", meta.Title)
	require.Equal(t, "app/home/cmd/server", meta.Entry)

	manifest := siteService.Manifest()
	require.Equal(t, "A focused home test shell.", manifest.Headline)
	require.Len(t, manifest.Features, 3)
	require.Equal(t, "/api/v1/manifest", manifest.Features[2].Route)
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
		service.Package,
	)

	return injector
}
