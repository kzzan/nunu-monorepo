package service

import (
	"context"
	"errors"
	v1 "nunu-monorepo/app/admin/api/v1"
	"nunu-monorepo/app/admin/internal/model"
	"nunu-monorepo/app/admin/internal/repository"
	"sort"
	"strings"
)

// UpdateRolePermission 用请求中的权限键列表整体替换角色授权：
// 1) 校验每个键都对应真实存在的菜单/API 权限；
// 2) 由勾选的菜单自动带上全部子孙菜单与关联 API（前端体验语义）；
// 3) 事务内替换 casbin 规则后重载策略使其立即生效。
func (s *adminService) UpdateRolePermission(ctx context.Context, req *v1.UpdateRolePermissionRequest) error {
	role := strings.TrimSpace(req.Role)
	if role == "" {
		return v1.ErrBadRequest
	}
	if _, err := s.adminRepository.GetRoleBySid(ctx, role); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return v1.ErrBadRequest
		}
		return err
	}

	menus, err := s.adminRepository.GetMenuList(ctx)
	if err != nil {
		return err
	}
	menuPaths, err := buildMenuPathIndex(menus)
	if err != nil {
		return err
	}
	apis, err := s.adminRepository.GetApiList(ctx)
	if err != nil {
		return err
	}

	validPermissions := make(map[string]model.Permission, len(menus)+len(apis))
	menuIDByPermission := make(map[string]uint, len(menus))
	for _, menu := range menus {
		permission := menuPermission(menuPaths[menu.ID])
		validPermissions[permission.Key()] = permission
		menuIDByPermission[permission.Key()] = menu.ID
	}
	for _, api := range apis {
		permission := apiPermission(api.Path, api.Method)
		validPermissions[permission.Key()] = permission
	}

	permissionSet := make(map[model.Permission]struct{}, len(req.List))
	selectedMenuIDs := make(map[uint]struct{})
	for _, key := range req.List {
		permission, ok := model.ParsePermissionKey(key)
		if !ok {
			return v1.ErrBadRequest
		}
		valid, exists := validPermissions[permission.Key()]
		if !exists || valid != permission {
			return v1.ErrBadRequest
		}
		permissionSet[permission] = struct{}{}
		if menuID, isMenu := menuIDByPermission[permission.Key()]; isMenu {
			selectedMenuIDs[menuID] = struct{}{}
		}
	}

	selectedMenuIDs = includeMenuDescendants(menus, selectedMenuIDs)
	for menuID := range selectedMenuIDs {
		permissionSet[menuPermission(menuPaths[menuID])] = struct{}{}
	}
	for _, api := range apis {
		if intersectsMenuIDs(api.MenuIDs, selectedMenuIDs) {
			permissionSet[apiPermission(api.Path, api.Method)] = struct{}{}
		}
	}

	permissions := make([]model.Permission, 0, len(permissionSet))
	for permission := range permissionSet {
		permissions = append(permissions, permission)
	}
	sort.Slice(permissions, func(i, j int) bool {
		return permissions[i].Key() < permissions[j].Key()
	})
	return s.adminRepository.UpdateRolePermission(ctx, role, permissions)
}

// GetUserPermissions 返回用户的隐式权限键列表（含角色继承）。
func (s *adminService) GetUserPermissions(ctx context.Context, uid uint) (*v1.GetUserPermissionsData, error) {
	data := &v1.GetUserPermissionsData{
		List: []string{},
	}
	list, err := s.adminRepository.GetUserPermissions(ctx, uid)
	if err != nil {
		return nil, err
	}
	for _, v := range list {
		if len(v) == 3 {
			data.List = append(data.List, strings.Join([]string{v[1], v[2]}, model.PermSep))
		}
	}
	return data, nil
}

// GetRolePermissions 返回角色自身的显式权限键列表。
func (s *adminService) GetRolePermissions(ctx context.Context, role string) (*v1.GetRolePermissionsData, error) {
	data := &v1.GetRolePermissionsData{
		List: []string{},
	}
	list, err := s.adminRepository.GetRolePermissions(ctx, role)
	if err != nil {
		return nil, err
	}
	for _, v := range list {
		if len(v) == 3 {
			data.List = append(data.List, strings.Join([]string{v[1], v[2]}, model.PermSep))
		}
	}
	return data, nil
}
