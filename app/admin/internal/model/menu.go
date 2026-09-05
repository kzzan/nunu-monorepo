package model

// MenuAuth 描述菜单页面上一个按钮级权限点（如"新增/编辑/删除"）。
type MenuAuth struct {
	Title    string `json:"title"`    // 按钮权限展示名
	AuthMark string `json:"authMark"` // 前端鉴权标识
}

// Menu 是管理端导航菜单；字段语义与前端路由配置一一对应。
// Roles/AuthList 在领域模型中是普通切片，JSON 列编码由 repository 负责。
type Menu struct {
	Base
	ParentID      uint
	Path          string
	Title         string
	Name          string
	Component     string
	Locale        string
	Icon          string
	Redirect      string
	URL           string
	Link          string
	Target        string
	ActivePath    string
	ShowTextBadge string
	Weight        int
	IsEnable      bool
	IsMenu        bool
	KeepAlive     bool
	HideInMenu    bool
	IsHide        bool
	IsHideTab     bool
	IsIframe      bool
	ShowBadge     bool
	FixedTab      bool
	IsFullPage    bool
	Roles         []string   // 可见角色列表
	AuthList      []MenuAuth // 按钮权限列表
}
