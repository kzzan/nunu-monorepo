package service

import (
	"context"
	v1 "nunu-monorepo/app/admin/api/v1"
	"nunu-monorepo/app/admin/internal/repository"

	"github.com/samber/do/v2"
)

// AdminService 汇集管理后台的用例入口：认证、管理员账号、角色、
// 菜单、API 资源与权限管理。实现分布在同包各文件中。
type AdminService interface {
	Login(ctx context.Context, req *v1.LoginRequest) (string, error)
	GetAdminUsers(ctx context.Context, req *v1.GetAdminUsersRequest) (*v1.GetAdminUsersResponseData, error)
	GetAdminUser(ctx context.Context, uid uint) (*v1.GetAdminUserResponseData, error)
	AdminUserUpdate(ctx context.Context, req *v1.AdminUserUpdateRequest) error
	AdminUserCreate(ctx context.Context, req *v1.AdminUserCreateRequest) error
	AdminUserDelete(ctx context.Context, id uint) error

	GetUserPermissions(ctx context.Context, uid uint) (*v1.GetUserPermissionsData, error)
	GetRolePermissions(ctx context.Context, role string) (*v1.GetRolePermissionsData, error)
	UpdateRolePermission(ctx context.Context, req *v1.UpdateRolePermissionRequest) error

	GetAdminMenus(ctx context.Context) (*v1.GetMenuResponseData, error)
	GetMenus(ctx context.Context, uid uint) (*v1.GetMenuResponseData, error)
	MenuUpdate(ctx context.Context, req *v1.MenuUpdateRequest) error
	MenuCreate(ctx context.Context, req *v1.MenuCreateRequest) error
	MenuDelete(ctx context.Context, id uint) error

	GetRoles(ctx context.Context, req *v1.GetRoleListRequest) (*v1.GetRolesResponseData, error)
	RoleUpdate(ctx context.Context, req *v1.RoleUpdateRequest) error
	RoleCreate(ctx context.Context, req *v1.RoleCreateRequest) error
	RoleDelete(ctx context.Context, id uint) error

	GetApis(ctx context.Context, req *v1.GetApisRequest) (*v1.GetApisResponseData, error)
	ApiUpdate(ctx context.Context, req *v1.ApiUpdateRequest) error
	ApiCreate(ctx context.Context, req *v1.ApiCreateRequest) error
	ApiDelete(ctx context.Context, id uint) error
}

// NewAdminService 构造聚合服务，由注入容器调用。
func NewAdminService(i do.Injector) (AdminService, error) {
	return &adminService{
		Service:         do.MustInvoke[*Service](i),
		adminRepository: do.MustInvoke[repository.AdminRepository](i),
	}, nil
}

// adminService 实现 AdminService，聚合公共依赖与管理后台仓储。
type adminService struct {
	*Service
	adminRepository repository.AdminRepository
}
