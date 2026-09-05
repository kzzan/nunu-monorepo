package model

import "strings"

// RBAC 领域常量：角色标识、种子账号与 Casbin 资源前缀。
const (
	// AdminRole 是超级管理员角色的 Sid，拥有全部权限。
	AdminRole = "admin"
	// AdminUserID 是内置超级管理员账号的用户 ID（种子数据固定为 1）。
	AdminUserID = "1"
	// MenuResourcePrefix 是菜单类权限资源的前缀，形如 menu:/dashboard。
	MenuResourcePrefix = "menu:"
	// ApiResourcePrefix 是接口类权限资源的前缀，形如 api:/v1/menus。
	ApiResourcePrefix = "api:"
	// PermSep 是权限键（Resource+Action）的分隔符。
	PermSep = ","
)

// AdminUser 是后台管理员账号。
type AdminUser struct {
	Base
	Username string
	Nickname string
	Password string // bcrypt 哈希，永不明文存储
	Email    string
	Phone    string
}

// Role 是后台角色；Sid 同时作为 Casbin 中的角色名。
type Role struct {
	Base
	Name string
	Sid  string
}

// Api 是受 RBAC 管控的接口资源注册项；Path+Method 全局唯一。
type Api struct {
	Base
	Group   string // 分组路径，如 "权限管理/菜单"
	Name    string // 展示名
	Path    string // 接口路径，必须以 / 开头
	Method  string // HTTP 方法，统一大写
	MenuIDs []uint // 关联菜单 ID 列表
}

// Permission 表示一条 Casbin 权限：资源（menu:/x 或 api:/v1/y）+ 动作（read/GET…）。
type Permission struct {
	Resource string
	Action   string
}

// Key 返回权限的规范键表示 "Resource,Action"，用于接口传输与去重。
func (p Permission) Key() string {
	return p.Resource + PermSep + p.Action
}

// ParsePermissionKey 解析 "Resource,Action" 形式的权限键。
// 资源本身可能包含逗号（如菜单路径 "api:/v1/items,a"），因此以最后一个分隔符为准；
// 格式非法时返回 ok=false。
func ParsePermissionKey(key string) (Permission, bool) {
	separator := strings.LastIndex(key, PermSep)
	if separator <= 0 || separator == len(key)-1 {
		return Permission{}, false
	}
	permission := Permission{
		Resource: strings.TrimSpace(key[:separator]),
		Action:   strings.TrimSpace(key[separator+len(PermSep):]),
	}
	if permission.Resource == "" || permission.Action == "" {
		return Permission{}, false
	}
	return permission, true
}
