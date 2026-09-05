package handler

import (
	"path/filepath"
	"runtime"

	apiV1 "nunu-monorepo/app/home/api/v1"
	"nunu-monorepo/app/home/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

// SiteHandler 承载 home 站点的页面与元信息接口。
type SiteHandler struct {
	*Handler
	siteService service.SiteService
	indexFile   string
}

// NewSiteHandler 构造站点处理器并定位静态首页文件，由注入容器调用。
func NewSiteHandler(i do.Injector) (*SiteHandler, error) {
	return &SiteHandler{
		Handler:     do.MustInvoke[*Handler](i),
		siteService: do.MustInvoke[service.SiteService](i),
		indexFile:   resolveHomeIndexFile(),
	}, nil
}

// Index 返回静态首页。
func (h *SiteHandler) Index(ctx *gin.Context) {
	ctx.File(h.indexFile)
}

// Health 返回健康检查结果。
func (h *SiteHandler) Health(ctx *gin.Context) {
	apiV1.HandleSuccess(ctx, h.siteService.Health())
}

// Meta 返回运行时元信息。
func (h *SiteHandler) Meta(ctx *gin.Context) {
	apiV1.HandleSuccess(ctx, h.siteService.Meta())
}

// Manifest 返回站点清单。
func (h *SiteHandler) Manifest(ctx *gin.Context) {
	apiV1.HandleSuccess(ctx, h.siteService.Manifest())
}

// resolveHomeIndexFile 定位静态首页 index.html：优先按源码相对路径
// （开发态 go run），失败时退回工作目录相对路径。
func resolveHomeIndexFile() string {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Clean("app/home/web/index.html")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "../../web/index.html"))
}
