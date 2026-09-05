package web

import "embed"

//go:embed all:dist
var assets embed.FS

// Assets 返回内嵌的管理端前端构建产物（app/admin/web/dist）。
func Assets() embed.FS {
	return assets
}
