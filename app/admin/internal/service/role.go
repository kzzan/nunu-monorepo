package service

import (
	"context"
	"errors"

	v1 "nunu-monorepo/app/admin/api/v1"
	"nunu-monorepo/app/admin/internal/model"
	"nunu-monorepo/app/admin/internal/repository"
)

// RoleUpdate 更新角色展示名；Sid 创建后不可变更，由仓储层约束。
func (s *adminService) RoleUpdate(ctx context.Context, req *v1.RoleUpdateRequest) error {
	return s.adminRepository.RoleUpdate(ctx, &model.Role{
		Base: model.Base{ID: req.ID},
		Name: req.Name,
		Sid:  req.Sid,
	})
}

// RoleCreate 创建角色：Sid 重复时返回 ErrRoleAlreadyUse。
func (s *adminService) RoleCreate(ctx context.Context, req *v1.RoleCreateRequest) error {
	_, err := s.adminRepository.GetRoleBySid(ctx, req.Sid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return s.adminRepository.RoleCreate(ctx, &model.Role{
				Name: req.Name,
				Sid:  req.Sid,
			})
		} else {
			return err
		}
	}
	return v1.ErrRoleAlreadyUse
}

// RoleDelete 在同一事务中删除角色并清空其在 Casbin 中的全部授权。
func (s *adminService) RoleDelete(ctx context.Context, id uint) error {
	return s.tm.Transaction(ctx, func(ctx context.Context) error {
		old, err := s.adminRepository.GetRole(ctx, id)
		if err != nil {
			return err
		}
		if err := s.adminRepository.RoleDelete(ctx, id); err != nil {
			return err
		}
		return s.adminRepository.CasbinRoleDelete(ctx, old.Sid)
	})
}

// GetRoles 分页查询角色列表并转换为响应 DTO。
func (s *adminService) GetRoles(ctx context.Context, req *v1.GetRoleListRequest) (*v1.GetRolesResponseData, error) {
	list, total, err := s.adminRepository.GetRoles(ctx, req)
	if err != nil {
		return nil, err
	}
	data := &v1.GetRolesResponseData{
		List:  make([]v1.RoleDataItem, 0),
		Total: total,
	}
	for _, role := range list {
		data.List = append(data.List, v1.RoleDataItem{
			ID:        role.ID,
			Name:      role.Name,
			Sid:       role.Sid,
			UpdatedAt: role.UpdatedAt.Format("2006-01-02 15:04:05"),
			CreatedAt: role.CreatedAt.Format("2006-01-02 15:04:05"),
		})

	}
	return data, nil
}
