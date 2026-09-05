package service

import (
	"context"
	"net/http"
	v1 "nunu-monorepo/app/admin/api/v1"
	"nunu-monorepo/app/admin/internal/model"
	"strings"
)

// GetApis 分页查询 API 列表，并附带去重后的分组列表供前端筛选。
func (s *adminService) GetApis(ctx context.Context, req *v1.GetApisRequest) (*v1.GetApisResponseData, error) {
	list, total, err := s.adminRepository.GetApis(ctx, req)
	if err != nil {
		return nil, err
	}
	groups, err := s.adminRepository.GetApiGroups(ctx)
	if err != nil {
		return nil, err
	}
	data := &v1.GetApisResponseData{
		List:   make([]v1.ApiDataItem, 0),
		Total:  total,
		Groups: groups,
	}
	for _, api := range list {
		data.List = append(data.List, v1.ApiDataItem{
			CreatedAt: api.CreatedAt.Format("2006-01-02 15:04:05"),
			Group:     api.Group,
			ID:        api.ID,
			MenuIDs:   append([]uint{}, api.MenuIDs...),
			Method:    api.Method,
			Name:      api.Name,
			Path:      api.Path,
			UpdatedAt: api.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return data, nil
}

// ApiUpdate 更新 API 资源：校验输入与菜单引用、检查 path+method 唯一，
// 路径或方法变化时同步替换 Casbin 中的权限引用并重载策略。
func (s *adminService) ApiUpdate(ctx context.Context, req *v1.ApiUpdateRequest) error {
	api, err := normalizeApiInput(req.Group, req.Name, req.Path, req.Method)
	if err != nil {
		return err
	}
	api.ID = req.ID
	menuIDs := uniqueUintIDs(req.MenuIDs)
	if err := s.validateApiMenuIDs(ctx, menuIDs); err != nil {
		return err
	}
	api.MenuIDs = menuIDs
	oldApi, err := s.adminRepository.GetApi(ctx, req.ID)
	if err != nil {
		return err
	}
	exists, err := s.adminRepository.ApiPermissionExists(ctx, api.Path, api.Method, req.ID)
	if err != nil {
		return err
	}
	if exists {
		return v1.ErrBadRequest
	}

	oldPermission := apiPermission(oldApi.Path, oldApi.Method)
	newPermission := apiPermission(api.Path, api.Method)
	err = s.tm.Transaction(ctx, func(ctx context.Context) error {
		if err := s.adminRepository.ApiUpdate(ctx, api); err != nil {
			return err
		}
		return s.adminRepository.ReplacePermissionReferences(ctx, map[model.Permission]model.Permission{
			oldPermission: newPermission,
		})
	})
	if err != nil {
		return err
	}
	if oldPermission != newPermission {
		return s.adminRepository.ReloadPolicy()
	}
	return nil
}

// ApiCreate 创建 API 资源：校验输入与菜单引用，path+method 重复时拒绝。
func (s *adminService) ApiCreate(ctx context.Context, req *v1.ApiCreateRequest) error {
	api, err := normalizeApiInput(req.Group, req.Name, req.Path, req.Method)
	if err != nil {
		return err
	}
	menuIDs := uniqueUintIDs(req.MenuIDs)
	if err := s.validateApiMenuIDs(ctx, menuIDs); err != nil {
		return err
	}
	exists, err := s.adminRepository.ApiPermissionExists(ctx, api.Path, api.Method, 0)
	if err != nil {
		return err
	}
	if exists {
		return v1.ErrBadRequest
	}
	api.MenuIDs = menuIDs
	return s.adminRepository.ApiCreate(ctx, api)
}

// validateApiMenuIDs 校验 API 关联的菜单 ID 都真实存在。
func (s *adminService) validateApiMenuIDs(ctx context.Context, menuIDs []uint) error {
	if len(menuIDs) == 0 {
		return nil
	}
	menus, err := s.adminRepository.GetMenuList(ctx)
	if err != nil {
		return err
	}
	validIDs := make(map[uint]struct{}, len(menus))
	for _, menu := range menus {
		validIDs[menu.ID] = struct{}{}
	}
	for _, id := range menuIDs {
		if _, exists := validIDs[id]; !exists {
			return v1.ErrBadRequest
		}
	}
	return nil
}

// normalizeApiGroup 规范化分组路径：按 / 分段去空格后重新拼接，
// 空段与超长（>255）视为非法。
func normalizeApiGroup(group string) (string, error) {
	parts := strings.Split(strings.TrimSpace(group), "/")
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return "", v1.ErrBadRequest
		}
		normalized = append(normalized, part)
	}
	if len(normalized) == 0 {
		return "", v1.ErrBadRequest
	}
	result := strings.Join(normalized, "/")
	if len(result) > 255 {
		return "", v1.ErrBadRequest
	}
	return result, nil
}

// normalizeApiInput 校验并规范化 API 创建/更新入参：
// 名称/路径长度、路径前缀、HTTP 方法白名单。
func normalizeApiInput(group, name, path, method string) (*model.Api, error) {
	normalizedGroup, err := normalizeApiGroup(group)
	if err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	path = strings.TrimSpace(path)
	method = strings.ToUpper(strings.TrimSpace(method))
	if name == "" || len(name) > 100 || path == "" || len(path) > 255 || !strings.HasPrefix(path, "/") {
		return nil, v1.ErrBadRequest
	}
	allowedMethods := map[string]struct{}{
		http.MethodGet: {}, http.MethodPost: {}, http.MethodPut: {}, http.MethodPatch: {},
		http.MethodDelete: {}, http.MethodHead: {}, http.MethodOptions: {},
	}
	if _, ok := allowedMethods[method]; !ok {
		return nil, v1.ErrBadRequest
	}
	return &model.Api{Group: normalizedGroup, Name: name, Path: path, Method: method}, nil
}

// uniqueUintIDs 去除 0 与重复项，保持首次出现顺序。
func uniqueUintIDs(ids []uint) []uint {
	result := make([]uint, 0, len(ids))
	seen := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

// apiPermission 构造 API 类权限（资源 = api: + 路径，动作 = 大写方法）。
func apiPermission(path, method string) model.Permission {
	return model.Permission{
		Resource: model.ApiResourcePrefix + strings.TrimSpace(path),
		Action:   strings.ToUpper(strings.TrimSpace(method)),
	}
}

// ApiDelete 删除 API 并在同一事务中清除其 Casbin 权限引用。
func (s *adminService) ApiDelete(ctx context.Context, id uint) error {
	api, err := s.adminRepository.GetApi(ctx, id)
	if err != nil {
		return err
	}
	permission := apiPermission(api.Path, api.Method)
	if err := s.tm.Transaction(ctx, func(ctx context.Context) error {
		if err := s.adminRepository.ApiDelete(ctx, id); err != nil {
			return err
		}
		return s.adminRepository.DeletePermissionReferences(ctx, []model.Permission{permission})
	}); err != nil {
		return err
	}
	return s.adminRepository.ReloadPolicy()
}
