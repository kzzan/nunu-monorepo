package server

import (
	nethttp "net/http"

	"nunu-monorepo/app/admin/docs"
	"nunu-monorepo/app/admin/internal/handler"
	"nunu-monorepo/app/admin/internal/middleware"
	"nunu-monorepo/app/admin/web"
	"nunu-monorepo/pkg/jwt"
	"nunu-monorepo/pkg/log"
	"nunu-monorepo/pkg/server/http"

	"github.com/casbin/casbin/v3"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewHTTPServer 组装 admin 的 HTTP 服务：静态资源与 Swagger、
// 公共中间件，以及按认证强度划分的路由组（登录接口免鉴权，
// 管理接口走 JWT + Casbin 双重校验）。
func NewHTTPServer(i do.Injector) (*http.Server, error) {
	logger := do.MustInvoke[*log.Logger](i)
	j := do.MustInvoke[*jwt.JWT](i)
	e := do.MustInvoke[*casbin.SyncedEnforcer](i)
	adminHandler := do.MustInvoke[*handler.AdminHandler](i)
	userHandler := do.MustInvoke[*handler.UserHandler](i)

	gin.SetMode(gin.DebugMode)
	s, err := http.NewServer(i)
	if err != nil {
		return nil, err
	}
	// 设置前端静态资源
	embedFolder, err := static.EmbedFolder(web.Assets(), "dist")
	if err != nil {
		return nil, err
	}
	s.Use(static.Serve("/", embedFolder))
	s.NoRoute(func(c *gin.Context) {
		indexPageData, err := web.Assets().ReadFile("dist/index.html")
		if err != nil {
			c.String(nethttp.StatusNotFound, "404 page not found")
			return
		}
		c.Data(nethttp.StatusOK, "text/html; charset=utf-8", indexPageData)
	})
	// swagger doc
	docs.SwaggerInfo.BasePath = "/"
	s.GET("/swagger/*any", newSwaggerHandler())

	s.Use(
		middleware.CORSMiddleware(),
		middleware.ResponseLogMiddleware(logger),
		middleware.RequestLogMiddleware(logger),
		//middleware.SignMiddleware(log),
	)

	v1 := s.Group("/v1")
	{
		// No route group has permission
		noAuthRouter := v1.Group("/")
		{
			noAuthRouter.POST("/login", adminHandler.Login)
		}

		// Strict permission routing group
		strictAuthRouter := v1.Group("/").Use(middleware.StrictAuth(j, logger), middleware.AuthMiddleware(e))
		{
			strictAuthRouter.GET("/users", userHandler.GetUsers)

			strictAuthRouter.GET("/menus", adminHandler.GetMenus)
			strictAuthRouter.GET("/admin/menus", adminHandler.GetAdminMenus)
			strictAuthRouter.POST("/admin/menu", adminHandler.MenuCreate)
			strictAuthRouter.PUT("/admin/menu", adminHandler.MenuUpdate)
			strictAuthRouter.DELETE("/admin/menu", adminHandler.MenuDelete)

			strictAuthRouter.GET("/admin/users", adminHandler.GetAdminUsers)
			strictAuthRouter.GET("/admin/user", adminHandler.GetAdminUser)
			strictAuthRouter.PUT("/admin/user", adminHandler.AdminUserUpdate)
			strictAuthRouter.POST("/admin/user", adminHandler.AdminUserCreate)
			strictAuthRouter.DELETE("/admin/user", adminHandler.AdminUserDelete)
			strictAuthRouter.GET("/admin/user/permissions", adminHandler.GetUserPermissions)
			strictAuthRouter.GET("/admin/role/permissions", adminHandler.GetRolePermissions)
			strictAuthRouter.PUT("/admin/role/permissions", adminHandler.UpdateRolePermission)
			strictAuthRouter.GET("/admin/roles", adminHandler.GetRoles)
			strictAuthRouter.POST("/admin/role", adminHandler.RoleCreate)
			strictAuthRouter.PUT("/admin/role", adminHandler.RoleUpdate)
			strictAuthRouter.DELETE("/admin/role", adminHandler.RoleDelete)

			strictAuthRouter.GET("/admin/apis", adminHandler.GetApis)
			strictAuthRouter.POST("/admin/api", adminHandler.ApiCreate)
			strictAuthRouter.PUT("/admin/api", adminHandler.ApiUpdate)
			strictAuthRouter.DELETE("/admin/api", adminHandler.ApiDelete)

		}
	}
	return s, nil
}

// newSwaggerHandler 返回 Swagger UI 处理器，doc.json 直接取自
// 内嵌的生成文档，避免文件系统依赖。
func newSwaggerHandler() gin.HandlerFunc {
	swaggerHandler := ginSwagger.WrapHandler(
		swaggerfiles.Handler,
		ginSwagger.URL("/swagger/doc.json"),
		ginSwagger.DefaultModelsExpandDepth(-1),
		ginSwagger.PersistAuthorization(true),
	)
	return func(ctx *gin.Context) {
		if ctx.Param("any") == "/doc.json" {
			ctx.Data(nethttp.StatusOK, "application/json; charset=utf-8", docs.SwaggerJSON)
			return
		}
		swaggerHandler(ctx)
	}
}
