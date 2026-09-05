package service

import (
	apiV1 "nunu-monorepo/app/home/api/v1"

	"github.com/samber/do/v2"
)

// SiteService 提供站点健康、元信息与清单三类只读用例。
type SiteService interface {
	Health() apiV1.HealthData
	Meta() apiV1.MetaData
	Manifest() apiV1.ManifestData
}

// siteService 实现 SiteService。
type siteService struct {
	*Service
}

// NewSiteService 构造站点服务，由注入容器调用。
func NewSiteService(i do.Injector) (SiteService, error) {
	return &siteService{Service: do.MustInvoke[*Service](i)}, nil
}

// Health 返回固定健康态。
func (s *siteService) Health() apiV1.HealthData {
	return apiV1.HealthData{
		App:    "home",
		Status: "ok",
	}
}

// Meta 返回运行环境与应用标识等元信息。
func (s *siteService) Meta() apiV1.MetaData {
	return apiV1.MetaData{
		App:   "home",
		Stage: s.config.GetString("env"),
		Entry: "app/home/cmd/server",
		Title: siteTitle(s.config.GetString("site.title")),
	}
}

// Manifest 返回前端引导所需的站点清单（特性入口列表）。
func (s *siteService) Manifest() apiV1.ManifestData {
	return apiV1.ManifestData{
		App:      "home",
		Stage:    s.config.GetString("env"),
		Entry:    "app/home/cmd/server",
		Title:    siteTitle(s.config.GetString("site.title")),
		Headline: siteHeadline(s.config.GetString("site.headline")),
		Features: []apiV1.ManifestFeature{
			{
				Name:        "Public Shell",
				Description: "Serve public pages and lightweight shell APIs from app/home.",
				Route:       "/",
			},
			{
				Name:        "Runtime Health",
				Description: "Expose a small health endpoint for local development and deploy checks.",
				Route:       "/healthz",
			},
			{
				Name:        "Bootstrap Manifest",
				Description: "Return route and product metadata for future front-end bootstrapping.",
				Route:       "/api/v1/manifest",
			},
		},
	}
}

// siteTitle 返回站点标题，空值回退默认标题。
func siteTitle(title string) string {
	if title == "" {
		return "Home App"
	}

	return title
}

// siteHeadline 返回站点标语，空值回退默认标语。
func siteHeadline(headline string) string {
	if headline == "" {
		return "Public-facing shell, ready for real product work."
	}

	return headline
}
