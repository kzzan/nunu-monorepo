package service

import (
	"context"
	"fmt"
	"strings"

	v1 "nunu-monorepo/app/admin/api/v1"
	"nunu-monorepo/app/admin/internal/model"
	"nunu-monorepo/app/admin/internal/repository"

	"github.com/duke-git/lancet/v2/convertor"
)

// menuPermission 构造菜单类权限（资源 = menu: + 完整路径，动作 = read）。
func menuPermission(path string) model.Permission {
	return model.Permission{Resource: model.MenuResourcePrefix + path, Action: "read"}
}

// intersectsMenuIDs 判断 API 关联的菜单与勾选菜单是否有交集。
func intersectsMenuIDs(menuIDs []uint, selected map[uint]struct{}) bool {
	for _, menuID := range menuIDs {
		if _, ok := selected[menuID]; ok {
			return true
		}
	}
	return false
}

// includeMenuDescendants 把勾选菜单的全部子孙菜单并入选中集合
// （授予父菜单权限必须同时授予其子菜单），迭代直到不再新增。
func includeMenuDescendants(menus []model.Menu, selected map[uint]struct{}) map[uint]struct{} {
	result := make(map[uint]struct{}, len(selected))
	for id := range selected {
		result[id] = struct{}{}
	}
	changed := true
	for changed {
		changed = false
		for _, menu := range menus {
			if _, parentSelected := result[menu.ParentID]; !parentSelected {
				continue
			}
			if _, exists := result[menu.ID]; exists {
				continue
			}
			result[menu.ID] = struct{}{}
			changed = true
		}
	}
	return result
}

// replaceMenu 在菜单集合中按 ID 替换一项；未找到返回 found=false。
func replaceMenu(menus []model.Menu, replacement model.Menu) ([]model.Menu, bool) {
	result := append([]model.Menu(nil), menus...)
	for index := range result {
		if result[index].ID == replacement.ID {
			result[index] = replacement
			return result, true
		}
	}
	return result, false
}

// validateMenuCollection 校验整棵菜单树的合法性：必填字段与长度、
// name 唯一、完整路径唯一；返回每个菜单的完整路径索引。
func validateMenuCollection(menus []model.Menu) (map[uint]string, error) {
	names := make(map[string]uint, len(menus))
	for _, menu := range menus {
		name := strings.TrimSpace(menu.Name)
		path := strings.TrimSpace(menu.Path)
		title := strings.TrimSpace(menu.Title)
		if name == "" || len(name) > 100 || path == "" || len(path) > 255 || title == "" || len(title) > 100 {
			return nil, v1.ErrBadRequest
		}
		if existingID, exists := names[name]; exists && existingID != menu.ID {
			return nil, v1.ErrBadRequest
		}
		names[name] = menu.ID
	}
	paths, err := buildMenuPathIndex(menus)
	if err != nil {
		return nil, v1.ErrBadRequest
	}
	pathOwners := make(map[string]uint, len(paths))
	for id, path := range paths {
		if existingID, exists := pathOwners[path]; exists && existingID != id {
			return nil, v1.ErrBadRequest
		}
		pathOwners[path] = id
	}
	return paths, nil
}

// buildMenuPathIndex 递归计算每个菜单的完整路由路径（父链拼接），
// 检测重复 ID、缺失父级与环路；绝对路径原样保留。
func buildMenuPathIndex(menus []model.Menu) (map[uint]string, error) {
	menuMap := make(map[uint]model.Menu, len(menus))
	for _, menu := range menus {
		if _, exists := menuMap[menu.ID]; exists {
			return nil, fmt.Errorf("duplicate menu id: %d", menu.ID)
		}
		menuMap[menu.ID] = menu
	}
	paths := make(map[uint]string, len(menus))
	states := make(map[uint]uint8, len(menus))
	var resolve func(uint) (string, error)
	resolve = func(id uint) (string, error) {
		switch states[id] {
		case 1:
			return "", fmt.Errorf("menu hierarchy cycle at id %d", id)
		case 2:
			return paths[id], nil
		}
		menu, exists := menuMap[id]
		if !exists {
			return "", fmt.Errorf("menu %d not found", id)
		}
		states[id] = 1
		parentPath := ""
		if menu.ParentID != 0 {
			if _, exists := menuMap[menu.ParentID]; !exists {
				return "", fmt.Errorf("parent menu %d not found", menu.ParentID)
			}
			var err error
			parentPath, err = resolve(menu.ParentID)
			if err != nil {
				return "", err
			}
		}
		path := strings.TrimSpace(menu.Path)
		switch {
		case strings.HasPrefix(path, "http://"), strings.HasPrefix(path, "https://"), strings.HasPrefix(path, "/"):
			paths[id] = path
		case menu.ParentID == 0:
			paths[id] = "/" + strings.TrimLeft(path, "/")
		default:
			paths[id] = strings.TrimRight(parentPath, "/") + "/" + strings.TrimLeft(path, "/")
		}
		states[id] = 2
		return paths[id], nil
	}
	for id := range menuMap {
		if _, err := resolve(id); err != nil {
			return nil, err
		}
	}
	return paths, nil
}

// MenuUpdate 更新菜单：整树校验后落库；路径变化的菜单会同步把
// Casbin 中的旧权限引用替换为新权限引用（事务内完成）。
func (s *adminService) MenuUpdate(ctx context.Context, req *v1.MenuUpdateRequest) error {
	menu := menuFromPayload(menuPayload{
		ParentID:      req.ParentID,
		Weight:        req.Weight,
		Path:          req.Path,
		Title:         req.Title,
		Name:          req.Name,
		Component:     req.Component,
		Locale:        req.Locale,
		Icon:          req.Icon,
		Redirect:      req.Redirect,
		KeepAlive:     req.KeepAlive,
		HideInMenu:    req.HideInMenu,
		IsEnable:      req.IsEnable,
		IsMenu:        req.IsMenu,
		IsHide:        req.IsHide,
		IsHideTab:     req.IsHideTab,
		Link:          req.Link,
		IsIframe:      req.IsIframe,
		ShowBadge:     req.ShowBadge,
		ShowTextBadge: req.ShowTextBadge,
		FixedTab:      req.FixedTab,
		ActivePath:    req.ActivePath,
		Roles:         req.Roles,
		IsFullPage:    req.IsFullPage,
		AuthList:      req.AuthList,
		Target:        req.Target,
		URL:           req.URL,
	})
	menu.ID = req.ID
	menus, err := s.adminRepository.GetMenuList(ctx)
	if err != nil {
		return err
	}
	oldPaths, err := buildMenuPathIndex(menus)
	if err != nil {
		return err
	}
	updatedMenus, found := replaceMenu(menus, *menu)
	if !found {
		return repository.ErrNotFound
	}
	newPaths, err := validateMenuCollection(updatedMenus)
	if err != nil {
		return err
	}
	replacements := make(map[model.Permission]model.Permission)
	for id, oldPath := range oldPaths {
		newPath := newPaths[id]
		if oldPath != newPath {
			replacements[menuPermission(oldPath)] = menuPermission(newPath)
		}
	}
	if err := s.tm.Transaction(ctx, func(ctx context.Context) error {
		if err := s.adminRepository.MenuUpdate(ctx, menu); err != nil {
			return err
		}
		return s.adminRepository.ReplacePermissionReferences(ctx, replacements)
	}); err != nil {
		return err
	}
	if len(replacements) > 0 {
		return s.adminRepository.ReloadPolicy()
	}
	return nil
}

// MenuCreate 创建菜单：加入现有菜单树一并校验唯一性与层级合法。
func (s *adminService) MenuCreate(ctx context.Context, req *v1.MenuCreateRequest) error {
	menu := menuFromPayload(menuPayload{
		ParentID:      req.ParentID,
		Weight:        req.Weight,
		Path:          req.Path,
		Title:         req.Title,
		Name:          req.Name,
		Component:     req.Component,
		Locale:        req.Locale,
		Icon:          req.Icon,
		Redirect:      req.Redirect,
		KeepAlive:     req.KeepAlive,
		HideInMenu:    req.HideInMenu,
		IsEnable:      req.IsEnable,
		IsMenu:        req.IsMenu,
		IsHide:        req.IsHide,
		IsHideTab:     req.IsHideTab,
		Link:          req.Link,
		IsIframe:      req.IsIframe,
		ShowBadge:     req.ShowBadge,
		ShowTextBadge: req.ShowTextBadge,
		FixedTab:      req.FixedTab,
		ActivePath:    req.ActivePath,
		Roles:         req.Roles,
		IsFullPage:    req.IsFullPage,
		AuthList:      req.AuthList,
		Target:        req.Target,
		URL:           req.URL,
	})
	menus, err := s.adminRepository.GetMenuList(ctx)
	if err != nil {
		return err
	}
	if _, err := validateMenuCollection(append(menus, *menu)); err != nil {
		return err
	}
	return s.adminRepository.MenuCreate(ctx, menu)
}

// menuPayload 是菜单创建/更新请求的中间载体，统一两条入参路径。
type menuPayload struct {
	ParentID      uint
	Weight        int
	Path          string
	Title         string
	Name          string
	Component     string
	Locale        string
	Icon          string
	Redirect      string
	KeepAlive     bool
	HideInMenu    bool
	IsEnable      bool
	IsMenu        bool
	IsHide        bool
	IsHideTab     bool
	Link          string
	IsIframe      bool
	ShowBadge     bool
	ShowTextBadge string
	FixedTab      bool
	ActivePath    string
	Roles         []string
	IsFullPage    bool
	AuthList      []v1.MenuAuthDataItem
	Target        string
	URL           string
}

// menuFromPayload 把请求载荷转换为菜单实体：字符串字段去空格，
// 隐藏标记同时写入新旧两个字段以兼容前端。
func menuFromPayload(payload menuPayload) *model.Menu {
	isHide := payload.IsHide || payload.HideInMenu
	return &model.Menu{
		Component:     strings.TrimSpace(payload.Component),
		Icon:          strings.TrimSpace(payload.Icon),
		KeepAlive:     payload.KeepAlive,
		HideInMenu:    isHide,
		IsHide:        isHide,
		Locale:        strings.TrimSpace(payload.Locale),
		Weight:        payload.Weight,
		Name:          strings.TrimSpace(payload.Name),
		ParentID:      payload.ParentID,
		Path:          strings.TrimSpace(payload.Path),
		Redirect:      strings.TrimSpace(payload.Redirect),
		Title:         strings.TrimSpace(payload.Title),
		URL:           strings.TrimSpace(payload.URL),
		Link:          strings.TrimSpace(payload.Link),
		Target:        strings.TrimSpace(payload.Target),
		ActivePath:    strings.TrimSpace(payload.ActivePath),
		ShowTextBadge: strings.TrimSpace(payload.ShowTextBadge),
		IsEnable:      payload.IsEnable,
		IsMenu:        payload.IsMenu,
		IsHideTab:     payload.IsHideTab,
		IsIframe:      payload.IsIframe,
		ShowBadge:     payload.ShowBadge,
		FixedTab:      payload.FixedTab,
		IsFullPage:    payload.IsFullPage,
		Roles:         payload.Roles,
		AuthList:      menuAuthListFromPayload(payload.AuthList),
	}
}

// menuAuthListFromPayload 转换按钮权限列表，过滤全空项。
func menuAuthListFromPayload(list []v1.MenuAuthDataItem) []model.MenuAuth {
	authList := make([]model.MenuAuth, 0, len(list))
	for _, item := range list {
		if item.Title == "" && item.AuthMark == "" {
			continue
		}
		authList = append(authList, model.MenuAuth{
			Title:    item.Title,
			AuthMark: item.AuthMark,
		})
	}
	return authList
}

// menuDataItemFromModel 把菜单实体转换为前端响应 DTO。
func menuDataItemFromModel(menu model.Menu) v1.MenuDataItem {
	isHide := menu.IsHide || menu.HideInMenu
	return v1.MenuDataItem{
		ID:            menu.ID,
		Name:          menu.Name,
		Title:         menu.Title,
		Path:          menu.Path,
		Component:     menu.Component,
		Redirect:      menu.Redirect,
		KeepAlive:     menu.KeepAlive,
		HideInMenu:    isHide,
		IsHide:        isHide,
		IsEnable:      menu.IsEnable,
		IsMenu:        menu.IsMenu,
		IsHideTab:     menu.IsHideTab,
		Link:          menu.Link,
		IsIframe:      menu.IsIframe,
		ShowBadge:     menu.ShowBadge,
		ShowTextBadge: menu.ShowTextBadge,
		FixedTab:      menu.FixedTab,
		ActivePath:    menu.ActivePath,
		Roles:         menu.Roles,
		IsFullPage:    menu.IsFullPage,
		AuthList:      menuAuthDataListFromModel(menu.AuthList),
		Target:        menu.Target,
		Locale:        menu.Locale,
		Weight:        menu.Weight,
		Icon:          menu.Icon,
		ParentID:      menu.ParentID,
		UpdatedAt:     menu.UpdatedAt.Format("2006-01-02 15:04:05"),
		URL:           menu.URL,
	}
}

// menuAuthDataListFromModel 把按钮权限实体列表转换为响应 DTO。
func menuAuthDataListFromModel(list []model.MenuAuth) []v1.MenuAuthDataItem {
	authList := make([]v1.MenuAuthDataItem, 0, len(list))
	for _, item := range list {
		authList = append(authList, v1.MenuAuthDataItem{
			Title:    item.Title,
			AuthMark: item.AuthMark,
		})
	}
	return authList
}

// MenuDelete 删除菜单：存在子菜单时拒绝；同一事务中从 API 引用
// 摘除该菜单、软删除菜单并清除对应权限引用。
func (s *adminService) MenuDelete(ctx context.Context, id uint) error {
	menus, err := s.adminRepository.GetMenuList(ctx)
	if err != nil {
		return err
	}
	paths, err := buildMenuPathIndex(menus)
	if err != nil {
		return err
	}
	path, exists := paths[id]
	if !exists {
		return repository.ErrNotFound
	}
	for _, menu := range menus {
		if menu.ParentID == id {
			return v1.ErrBadRequest
		}
	}
	permission := menuPermission(path)
	if err := s.tm.Transaction(ctx, func(ctx context.Context) error {
		if err := s.adminRepository.RemoveMenuFromApis(ctx, id); err != nil {
			return err
		}
		if err := s.adminRepository.MenuDelete(ctx, id); err != nil {
			return err
		}
		return s.adminRepository.DeletePermissionReferences(ctx, []model.Permission{permission})
	}); err != nil {
		return err
	}
	return s.adminRepository.ReloadPolicy()
}

// GetMenus 返回当前用户可见的菜单树：超管返回全部，其余用户按
// Casbin 授权的菜单权限过滤，并自动补齐祖先节点保持树结构完整。
func (s *adminService) GetMenus(ctx context.Context, uid uint) (*v1.GetMenuResponseData, error) {
	menuList, err := s.adminRepository.GetMenuList(ctx)
	if err != nil {
		s.logger.WithContext(ctx).Error().Err(err).Msg("GetMenuList error")
		return nil, err
	}
	data := &v1.GetMenuResponseData{
		List: make([]v1.MenuDataItem, 0),
	}

	if convertor.ToString(uid) == model.AdminUserID {
		for _, menu := range menuList {
			data.List = append(data.List, menuDataItemFromModel(menu))
		}
		return data, nil
	}

	// 获取权限的菜单
	permissions, err := s.adminRepository.GetUserPermissions(ctx, uid)
	if err != nil {
		return nil, err
	}
	menuPermMap := map[string]struct{}{}
	for _, permission := range permissions {
		if len(permission) == 3 && strings.HasPrefix(permission[1], model.MenuResourcePrefix) {
			menuPermMap[strings.TrimPrefix(permission[1], model.MenuResourcePrefix)] = struct{}{}
		}
	}

	menuMap := make(map[uint]model.Menu, len(menuList))
	for _, menu := range menuList {
		menuMap[menu.ID] = menu
	}
	menuPaths, err := buildMenuPathIndex(menuList)
	if err != nil {
		return nil, err
	}

	allowedMenuIDs := map[uint]struct{}{}
	for _, menu := range menuList {
		fullPath := menuPaths[menu.ID]
		_, hasFullPath := menuPermMap[fullPath]
		_, hasRawPath := menuPermMap[menu.Path]
		if hasFullPath || hasRawPath {
			allowMenuWithAncestors(menu, menuMap, allowedMenuIDs)
		}
	}

	for _, menu := range menuList {
		if _, ok := allowedMenuIDs[menu.ID]; ok {
			data.List = append(data.List, menuDataItemFromModel(menu))
		}
	}
	return data, nil
}

// allowMenuWithAncestors 把菜单及其全部祖先加入可见集合（环路防护）。
func allowMenuWithAncestors(menu model.Menu, menuMap map[uint]model.Menu, allowed map[uint]struct{}) {
	seen := make(map[uint]struct{})
	current := menu
	for {
		if _, exists := seen[current.ID]; exists {
			return
		}
		seen[current.ID] = struct{}{}
		allowed[current.ID] = struct{}{}
		if current.ParentID == 0 {
			return
		}
		parent, ok := menuMap[current.ParentID]
		if !ok {
			return
		}
		current = parent
	}
}

// GetAdminMenus 返回全部菜单（管理端菜单管理页使用）。
func (s *adminService) GetAdminMenus(ctx context.Context) (*v1.GetMenuResponseData, error) {
	menuList, err := s.adminRepository.GetMenuList(ctx)
	if err != nil {
		s.logger.WithContext(ctx).Error().Err(err).Msg("GetMenuList error")
		return nil, err
	}
	data := &v1.GetMenuResponseData{
		List: make([]v1.MenuDataItem, 0),
	}
	for _, menu := range menuList {
		data.List = append(data.List, menuDataItemFromModel(menu))
	}
	return data, nil
}
