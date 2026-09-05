// Package v1 定义 admin 应用的 API 契约：请求/响应 DTO、统一响应包裹
// 与类型化业务错误。本包是将来对外契约（proto/OpenAPI）的雏形，
// 只允许依赖标准库与传输框架，禁止引入持久化细节。
package v1

// LoginRequest 是登录请求体。
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"1234@gmail.com"`
	Password string `json:"password" binding:"required" example:"123456"`
}

// LoginResponseData 携带签发的访问令牌。
type LoginResponseData struct {
	AccessToken string `json:"accessToken"`
}

// LoginResponse 是登录接口响应。
type LoginResponse struct {
	Response
	Data LoginResponseData
}

// AdminUserDataItem 是管理员列表项 DTO。
type AdminUserDataItem struct {
	ID        uint     `json:"id"`
	Username  string   `json:"username" binding:"required" example:"张三"`
	Nickname  string   `json:"nickname" binding:"required" example:"小Baby"`
	Password  string   `json:"password" binding:"required" example:"123456"`
	Email     string   `json:"email" binding:"required,email" example:"1234@gmail.com"`
	Phone     string   `json:"phone" binding:"" example:"1858888888"`
	Roles     []string `json:"roles" example:""`
	UpdatedAt string   `json:"updatedAt"`
	CreatedAt string   `json:"createdAt"`
}

// GetAdminUsersRequest 是管理员分页查询请求（支持多维过滤）。
type GetAdminUsersRequest struct {
	Page     int    `form:"page" binding:"required" example:"1"`
	PageSize int    `form:"pageSize" binding:"required" example:"10"`
	ID       uint   `form:"id" binding:"" example:"1"`
	Username string `form:"username" binding:"" example:"张三"`
	Nickname string `form:"nickname" binding:"" example:"小Baby"`
	Phone    string `form:"phone" binding:"" example:"1858888888"`
	Email    string `form:"email" binding:"" example:"1234@gmail.com"`
}

// GetAdminUserResponseData 是管理员详情 DTO。
type GetAdminUserResponseData struct {
	ID        uint     `json:"id"`
	Username  string   `json:"username" example:"张三"`
	Nickname  string   `json:"nickname" example:"小Baby"`
	Password  string   `json:"password" example:"123456"`
	Email     string   `json:"email" example:"1234@gmail.com"`
	Phone     string   `json:"phone" example:"1858888888"`
	Roles     []string `json:"roles" example:""`
	UpdatedAt string   `json:"updatedAt"`
	CreatedAt string   `json:"createdAt"`
}

// GetAdminUserResponse 是管理员详情接口响应。
type GetAdminUserResponse struct {
	Response
	Data GetAdminUserResponseData
}

// GetAdminUsersResponseData 是管理员分页列表数据。
type GetAdminUsersResponseData struct {
	List  []AdminUserDataItem `json:"list"`
	Total int64               `json:"total"`
}

// GetAdminUsersResponse 是管理员列表接口响应。
type GetAdminUsersResponse struct {
	Response
	Data GetAdminUsersResponseData
}

// AdminUserCreateRequest 是创建管理员请求体。
type AdminUserCreateRequest struct {
	Username string   `json:"username" binding:"required" example:"张三"`
	Nickname string   `json:"nickname" binding:"" example:"小Baby"`
	Password string   `json:"password" binding:"required" example:"123456"`
	Email    string   `json:"email" binding:"" example:"1234@gmail.com"`
	Phone    string   `json:"phone" binding:"" example:"1858888888"`
	Roles    []string `json:"roles" example:""`
}

// AdminUserUpdateRequest 是更新管理员请求体（密码留空表示不改）。
type AdminUserUpdateRequest struct {
	ID       uint     `json:"id"`
	Username string   `json:"username" binding:"required" example:"张三"`
	Nickname string   `json:"nickname" binding:"" example:"小Baby"`
	Password string   `json:"password" binding:"" example:"123456"`
	Email    string   `json:"email" binding:"" example:"1234@gmail.com"`
	Phone    string   `json:"phone" binding:"" example:"1858888888"`
	Roles    []string `json:"roles" example:""`
}

// AdminUserDeleteRequest 是删除管理员请求体。
type AdminUserDeleteRequest struct {
	ID uint `form:"id" binding:"required" example:"1"`
}

// MenuAuthDataItem 是按钮级权限点 DTO。
type MenuAuthDataItem struct {
	Title    string `json:"title"`    // 权限名称
	AuthMark string `json:"authMark"` // 权限标识
}

// MenuDataItem 是菜单 DTO：字段与前端路由配置一一对应。
type MenuDataItem struct {
	ID            uint               `json:"id,omitempty"`            // 唯一id，使用整数表示
	ParentID      uint               `json:"parentId,omitempty"`      // 父级菜单的id，使用整数表示
	Weight        int                `json:"weight"`                  // 排序权重
	Path          string             `json:"path"`                    // 地址
	Title         string             `json:"title"`                   // 展示名称
	Name          string             `json:"name,omitempty"`          // 同路由中的name，唯一标识
	Component     string             `json:"component,omitempty"`     // 绑定的组件
	Locale        string             `json:"locale,omitempty"`        // 本地化标识
	Icon          string             `json:"icon,omitempty"`          // 图标，使用字符串表示
	Redirect      string             `json:"redirect,omitempty"`      // 重定向地址
	KeepAlive     bool               `json:"keepAlive,omitempty"`     // 是否保活
	HideInMenu    bool               `json:"hideInMenu,omitempty"`    // 兼容旧版隐藏菜单字段
	IsEnable      bool               `json:"isEnable"`                // 是否启用
	IsMenu        bool               `json:"isMenu"`                  // 是否菜单
	IsHide        bool               `json:"isHide,omitempty"`        // 是否隐藏菜单
	IsHideTab     bool               `json:"isHideTab,omitempty"`     // 是否隐藏标签页
	Link          string             `json:"link,omitempty"`          // 外部链接
	IsIframe      bool               `json:"isIframe,omitempty"`      // 是否 iframe
	ShowBadge     bool               `json:"showBadge,omitempty"`     // 是否显示徽章
	ShowTextBadge string             `json:"showTextBadge,omitempty"` // 文本徽章
	FixedTab      bool               `json:"fixedTab,omitempty"`      // 是否固定标签页
	ActivePath    string             `json:"activePath,omitempty"`    // 激活菜单路径
	Roles         []string           `json:"roles,omitempty"`         // 角色权限
	IsFullPage    bool               `json:"isFullPage,omitempty"`    // 是否全屏页面
	AuthList      []MenuAuthDataItem `json:"authList,omitempty"`      // 按钮权限列表
	Target        string             `json:"target,omitempty"`        // 外链打开方式
	URL           string             `json:"url,omitempty"`           // iframe模式下的跳转url，不能与path重复
	UpdatedAt     string             `json:"updatedAt,omitempty"`     // 更新时间
}

// GetMenuResponseData 是菜单列表数据。
type GetMenuResponseData struct {
	List []MenuDataItem `json:"list"`
}

// GetMenuResponse 是菜单查询接口响应。
type GetMenuResponse struct {
	Response
	Data GetMenuResponseData
}

// MenuCreateRequest 是创建菜单请求体。
type MenuCreateRequest struct {
	ParentID      uint               `json:"parentId,omitempty"`       // 父级菜单的id，使用整数表示
	Weight        int                `json:"weight"`                   // 排序权重
	Path          string             `json:"path" binding:"required"`  // 地址
	Title         string             `json:"title" binding:"required"` // 展示名称
	Name          string             `json:"name" binding:"required"`  // 同路由中的name，唯一标识
	Component     string             `json:"component,omitempty"`      // 绑定的组件
	Locale        string             `json:"locale,omitempty"`         // 本地化标识
	Icon          string             `json:"icon,omitempty"`           // 图标，使用字符串表示
	Redirect      string             `json:"redirect,omitempty"`       // 重定向地址
	KeepAlive     bool               `json:"keepAlive,omitempty"`      // 是否保活
	HideInMenu    bool               `json:"hideInMenu,omitempty"`     // 兼容旧版隐藏菜单字段
	IsEnable      bool               `json:"isEnable"`                 // 是否启用
	IsMenu        bool               `json:"isMenu"`                   // 是否菜单
	IsHide        bool               `json:"isHide,omitempty"`         // 是否隐藏菜单
	IsHideTab     bool               `json:"isHideTab,omitempty"`      // 是否隐藏标签页
	Link          string             `json:"link,omitempty"`           // 外部链接
	IsIframe      bool               `json:"isIframe,omitempty"`       // 是否 iframe
	ShowBadge     bool               `json:"showBadge,omitempty"`      // 是否显示徽章
	ShowTextBadge string             `json:"showTextBadge,omitempty"`  // 文本徽章
	FixedTab      bool               `json:"fixedTab,omitempty"`       // 是否固定标签页
	ActivePath    string             `json:"activePath,omitempty"`     // 激活菜单路径
	Roles         []string           `json:"roles,omitempty"`          // 角色权限
	IsFullPage    bool               `json:"isFullPage,omitempty"`     // 是否全屏页面
	AuthList      []MenuAuthDataItem `json:"authList,omitempty"`       // 按钮权限列表
	Target        string             `json:"target,omitempty"`         // 外链打开方式
	URL           string             `json:"url,omitempty"`            // iframe模式下的跳转url，不能与path重复

}

// MenuUpdateRequest 是更新菜单请求体。
type MenuUpdateRequest struct {
	ID            uint               `json:"id" binding:"required"`    // 唯一id，使用整数表示
	ParentID      uint               `json:"parentId,omitempty"`       // 父级菜单的id，使用整数表示
	Weight        int                `json:"weight"`                   // 排序权重
	Path          string             `json:"path" binding:"required"`  // 地址
	Title         string             `json:"title" binding:"required"` // 展示名称
	Name          string             `json:"name" binding:"required"`  // 同路由中的name，唯一标识
	Component     string             `json:"component,omitempty"`      // 绑定的组件
	Locale        string             `json:"locale,omitempty"`         // 本地化标识
	Icon          string             `json:"icon,omitempty"`           // 图标，使用字符串表示
	Redirect      string             `json:"redirect,omitempty"`       // 重定向地址
	KeepAlive     bool               `json:"keepAlive,omitempty"`      // 是否保活
	HideInMenu    bool               `json:"hideInMenu,omitempty"`     // 兼容旧版隐藏菜单字段
	IsEnable      bool               `json:"isEnable"`                 // 是否启用
	IsMenu        bool               `json:"isMenu"`                   // 是否菜单
	IsHide        bool               `json:"isHide,omitempty"`         // 是否隐藏菜单
	IsHideTab     bool               `json:"isHideTab,omitempty"`      // 是否隐藏标签页
	Link          string             `json:"link,omitempty"`           // 外部链接
	IsIframe      bool               `json:"isIframe,omitempty"`       // 是否 iframe
	ShowBadge     bool               `json:"showBadge,omitempty"`      // 是否显示徽章
	ShowTextBadge string             `json:"showTextBadge,omitempty"`  // 文本徽章
	FixedTab      bool               `json:"fixedTab,omitempty"`       // 是否固定标签页
	ActivePath    string             `json:"activePath,omitempty"`     // 激活菜单路径
	Roles         []string           `json:"roles,omitempty"`          // 角色权限
	IsFullPage    bool               `json:"isFullPage,omitempty"`     // 是否全屏页面
	AuthList      []MenuAuthDataItem `json:"authList,omitempty"`       // 按钮权限列表
	Target        string             `json:"target,omitempty"`         // 外链打开方式
	URL           string             `json:"url,omitempty"`            // iframe模式下的跳转url，不能与path重复
	UpdatedAt     string             `json:"updatedAt"`
}

// MenuDeleteRequest 是删除菜单请求体。
type MenuDeleteRequest struct {
	ID uint `form:"id" binding:"required"` // 唯一id，使用整数表示
}

// GetRoleListRequest 是角色分页查询请求。
type GetRoleListRequest struct {
	Page     int    `form:"page" binding:"required" example:"1"`
	PageSize int    `form:"pageSize" binding:"required" example:"10"`
	Sid      string `form:"sid" binding:"" example:"1"`
	Name     string `form:"name" binding:"" example:"Admin"`
}

// RoleDataItem 是角色 DTO。
type RoleDataItem struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Sid       string `json:"sid"`
	UpdatedAt string `json:"updatedAt"`
	CreatedAt string `json:"createdAt"`
}

// GetRolesResponseData 是角色分页列表数据。
type GetRolesResponseData struct {
	List  []RoleDataItem `json:"list"`
	Total int64          `json:"total"`
}

// GetRolesResponse 是角色列表接口响应。
type GetRolesResponse struct {
	Response
	Data GetRolesResponseData
}

// RoleCreateRequest 是创建角色请求体。
type RoleCreateRequest struct {
	Sid  string `json:"sid" form:"sid" binding:"required" example:"1"`
	Name string `json:"name" form:"name" binding:"required" example:"Admin"`
}

// RoleUpdateRequest 是更新角色请求体。
type RoleUpdateRequest struct {
	ID   uint   `json:"id" form:"id" binding:"required" example:"1"`
	Sid  string `json:"sid" form:"sid" binding:"required" example:"1"`
	Name string `json:"name" form:"name" binding:"required" example:"Admin"`
}

// RoleDeleteRequest 是删除角色请求体。
type RoleDeleteRequest struct {
	ID uint `form:"id" binding:"required" example:"1"`
}

// PermissionCreateRequest 是（预留的）权限创建请求体。
type PermissionCreateRequest struct {
	Sid  string `form:"sid" binding:"required" example:"1"`
	Name string `form:"name" binding:"required" example:"Admin"`
}

// GetApisRequest 是 API 分页查询请求（支持分组/名称/路径/方法过滤）。
type GetApisRequest struct {
	Page     int    `form:"page" binding:"required,min=1" example:"1"`
	PageSize int    `form:"pageSize" binding:"required,min=1,max=1000" example:"10"`
	Group    string `form:"group" binding:"" example:"权限管理"`
	Name     string `form:"name" binding:"" example:"菜单列表"`
	Path     string `form:"path" binding:"" example:"/v1/test"`
	Method   string `form:"method" binding:"" example:"GET"`
}

// ApiDataItem 是 API 资源 DTO。
type ApiDataItem struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Method    string `json:"method"`
	Group     string `json:"group"`
	MenuIDs   []uint `json:"menuIds"`
	UpdatedAt string `json:"updatedAt"`
	CreatedAt string `json:"createdAt"`
}

// GetApisResponseData 是 API 分页列表数据（附带全部分组）。
type GetApisResponseData struct {
	List   []ApiDataItem `json:"list"`
	Total  int64         `json:"total"`
	Groups []string      `json:"groups"`
}

// GetApisResponse 是 API 列表接口响应。
type GetApisResponse struct {
	Response
	Data GetApisResponseData
}

// ApiCreateRequest 是创建 API 请求体。
type ApiCreateRequest struct {
	Group   string `json:"group" form:"group" binding:"required" example:"权限管理/菜单"`
	Name    string `json:"name" form:"name" binding:"required" example:"菜单列表"`
	Path    string `json:"path" form:"path" binding:"required" example:"/v1/test"`
	Method  string `json:"method" form:"method" binding:"required" example:"GET"`
	MenuIDs []uint `json:"menuIds" form:"menuIds"`
}

// ApiUpdateRequest 是更新 API 请求体。
type ApiUpdateRequest struct {
	ID      uint   `json:"id" form:"id" binding:"required" example:"1"`
	Group   string `json:"group" form:"group" binding:"required" example:"权限管理/菜单"`
	Name    string `json:"name" form:"name" binding:"required" example:"菜单列表"`
	Path    string `json:"path" form:"path" binding:"required" example:"/v1/test"`
	Method  string `json:"method" form:"method" binding:"required" example:"GET"`
	MenuIDs []uint `json:"menuIds" form:"menuIds"`
}

// ApiDeleteRequest 是删除 API 请求体。
type ApiDeleteRequest struct {
	ID uint `form:"id" binding:"required" example:"1"`
}

// GetUserPermissionsData 携带用户权限键列表。
type GetUserPermissionsData struct {
	List []string `json:"list"`
}

// GetRolePermissionsRequest 是角色权限查询请求体。
type GetRolePermissionsRequest struct {
	Role string `form:"role" binding:"required" example:"admin"`
}

// GetRolePermissionsData 携带角色权限键列表。
type GetRolePermissionsData struct {
	List []string `json:"list"`
}

// UpdateRolePermissionRequest 是整体替换角色权限的请求体：
// List 为权限键（Resource,Action）集合。
type UpdateRolePermissionRequest struct {
	Role string   `json:"role" form:"role" binding:"required" example:"admin"`
	List []string `json:"list" form:"list" example:""`
}
