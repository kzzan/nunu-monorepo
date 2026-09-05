package service

import (
	"context"
	v1 "nunu-monorepo/app/admin/api/v1"
	"nunu-monorepo/app/admin/internal/model"

	"golang.org/x/crypto/bcrypt"
)

// GetAdminUser 查询单个管理员详情并附带其角色列表；
// 角色查询失败不阻塞详情返回（角色为空）。
func (s *adminService) GetAdminUser(ctx context.Context, uid uint) (*v1.GetAdminUserResponseData, error) {
	user, err := s.adminRepository.GetAdminUser(ctx, uid)
	if err != nil {
		return nil, err
	}
	roles, _ := s.adminRepository.GetUserRoles(ctx, uid)

	return &v1.GetAdminUserResponseData{
		Email:     user.Email,
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Phone:     user.Phone,
		Roles:     roles,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// GetAdminUsers 分页查询管理员列表并逐个附带角色；单个用户的
// 角色查询失败仅记录日志并跳过，不影响整页返回。
func (s *adminService) GetAdminUsers(ctx context.Context, req *v1.GetAdminUsersRequest) (*v1.GetAdminUsersResponseData, error) {
	list, total, err := s.adminRepository.GetAdminUsers(ctx, req)
	if err != nil {
		return nil, err
	}
	data := &v1.GetAdminUsersResponseData{
		List:  make([]v1.AdminUserDataItem, 0),
		Total: total,
	}
	for _, user := range list {
		roles, err := s.adminRepository.GetUserRoles(ctx, user.ID)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetUserRoles error")
			continue
		}
		data.List = append(data.List, v1.AdminUserDataItem{
			Email:     user.Email,
			ID:        user.ID,
			Nickname:  user.Nickname,
			Username:  user.Username,
			Phone:     user.Phone,
			Roles:     roles,
			CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return data, nil
}

// AdminUserUpdate 更新管理员资料并同步角色绑定；密码留空时保留原值，
// 传入新密码则先做 bcrypt 哈希。
func (s *adminService) AdminUserUpdate(ctx context.Context, req *v1.AdminUserUpdateRequest) error {
	old, err := s.adminRepository.GetAdminUser(ctx, req.ID)
	if err != nil {
		return err
	}
	password := old.Password
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		password = string(hash)
	}
	return s.tm.Transaction(ctx, func(ctx context.Context) error {
		if err := s.adminRepository.AdminUserUpdate(ctx, &model.AdminUser{
			Base:     model.Base{ID: req.ID},
			Email:    req.Email,
			Nickname: req.Nickname,
			Password: password,
			Phone:    req.Phone,
			Username: req.Username,
		}); err != nil {
			return err
		}
		return s.adminRepository.UpdateUserRoles(ctx, req.ID, req.Roles)
	})

}

// AdminUserCreate 创建管理员（密码 bcrypt 哈希后落库）并绑定角色，
// 两步在同一事务中完成。
func (s *adminService) AdminUserCreate(ctx context.Context, req *v1.AdminUserCreateRequest) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	req.Password = string(hash)
	return s.tm.Transaction(ctx, func(ctx context.Context) error {
		if err := s.adminRepository.AdminUserCreate(ctx, &model.AdminUser{
			Email:    req.Email,
			Nickname: req.Nickname,
			Phone:    req.Phone,
			Username: req.Username,
			Password: req.Password,
		}); err != nil {
			return err
		}
		user, err := s.adminRepository.GetAdminUserByUsername(ctx, req.Username)
		if err != nil {
			return err
		}
		return s.adminRepository.UpdateUserRoles(ctx, user.ID, req.Roles)
	})

}

// AdminUserDelete 在同一事务中软删除管理员并清空其角色绑定。
func (s *adminService) AdminUserDelete(ctx context.Context, id uint) error {
	return s.tm.Transaction(ctx, func(ctx context.Context) error {
		if err := s.adminRepository.AdminUserDelete(ctx, id); err != nil {
			return err
		}
		return s.adminRepository.DeleteUserRoles(ctx, id)
	})
}
