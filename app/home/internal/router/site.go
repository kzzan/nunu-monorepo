package router

import "github.com/gin-gonic/gin"

// InitSiteRouter 注册站点路由：首页、健康检查与 /api/v1 元信息接口。
func InitSiteRouter(deps RouterDeps, engine *gin.Engine) {
	engine.GET("/", deps.SiteHandler.Index)
	engine.GET("/healthz", deps.SiteHandler.Health)

	v1 := engine.Group("/api/v1")
	{
		v1.GET("/meta", deps.SiteHandler.Meta)
		v1.GET("/manifest", deps.SiteHandler.Manifest)
	}
}
