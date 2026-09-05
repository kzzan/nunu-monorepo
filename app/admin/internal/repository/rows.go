package repository

import (
	"database/sql"
)

// 本文件定义 repository 私有的表行结构（带 db 标签）。
// 行结构持有全部存储细节（软删除列、可空时间戳、JSON 列编码类型），
// 与 model 包的领域实体一一对应，经 mapping.go 完成双向转换——
// model 因此保持零持久化依赖。

// rowBase 是全部行结构的公共审计列（对应 model.Base + 软删除列）。
type rowBase struct {
	ID        uint         `db:"id"`
	CreatedAt sql.NullTime `db:"created_at"`
	UpdatedAt sql.NullTime `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

// adminUserRow 映射 admin_users 表一行。
type adminUserRow struct {
	rowBase
	Username string `db:"username"`
	Nickname string `db:"nickname"`
	Password string `db:"password"`
	Email    string `db:"email"`
	Phone    string `db:"phone"`
}

// roleRow 映射 roles 表一行。
type roleRow struct {
	rowBase
	Name string `db:"name"`
	Sid  string `db:"sid"`
}

// apiRow 映射 api 表一行；menu_ids 为 JSON 列。
type apiRow struct {
	rowBase
	Group   string    `db:"group"`
	Name    string    `db:"name"`
	Path    string    `db:"path"`
	Method  string    `db:"method"`
	MenuIDs UintSlice `db:"menu_ids"`
}

// menuRow 映射 menu 表一行；roles/auth_list 为 JSON 列。
type menuRow struct {
	rowBase
	ParentID      uint        `db:"parent_id"`
	Path          string      `db:"path"`
	Title         string      `db:"title"`
	Name          string      `db:"name"`
	Component     string      `db:"component"`
	Locale        string      `db:"locale"`
	Icon          string      `db:"icon"`
	Redirect      string      `db:"redirect"`
	URL           string      `db:"url"`
	Link          string      `db:"link"`
	Target        string      `db:"target"`
	ActivePath    string      `db:"active_path"`
	ShowTextBadge string      `db:"show_text_badge"`
	Weight        int         `db:"weight"`
	IsEnable      bool        `db:"is_enable"`
	IsMenu        bool        `db:"is_menu"`
	KeepAlive     bool        `db:"keep_alive"`
	HideInMenu    bool        `db:"hide_in_menu"`
	IsHide        bool        `db:"is_hide"`
	IsHideTab     bool        `db:"is_hide_tab"`
	IsIframe      bool        `db:"is_iframe"`
	ShowBadge     bool        `db:"show_badge"`
	FixedTab      bool        `db:"fixed_tab"`
	IsFullPage    bool        `db:"is_full_page"`
	Roles         StringSlice `db:"roles"`
	AuthList      MenuAuths   `db:"auth_list"`
}
