package repository

import (
	"context"
	"fmt"

	"github.com/samber/do/v2"
	v1 "nunu-monorepo/app/admin/api/v1"
	"nunu-monorepo/app/admin/internal/model"
)

// AdminRepository 汇集管理后台全部业务实体的数据访问方法；
// 按实体拆分在 admin_user/role/api/menu/permission 各文件中实现。
type AdminRepository interface {
	GetAdminUsers(ctx context.Context, req *v1.GetAdminUsersRequest) ([]model.AdminUser, int64, error)
	GetAdminUser(ctx context.Context, uid uint) (model.AdminUser, error)
	GetAdminUserByUsername(ctx context.Context, username string) (model.AdminUser, error)
	AdminUserUpdate(ctx context.Context, m *model.AdminUser) error
	AdminUserCreate(ctx context.Context, m *model.AdminUser) error
	AdminUserDelete(ctx context.Context, id uint) error

	GetUserPermissions(ctx context.Context, uid uint) ([][]string, error)
	GetUserRoles(ctx context.Context, uid uint) ([]string, error)
	GetRolePermissions(ctx context.Context, role string) ([][]string, error)
	UpdateRolePermission(ctx context.Context, role string, permissions []model.Permission) error
	UpdateUserRoles(ctx context.Context, uid uint, roles []string) error
	DeleteUserRoles(ctx context.Context, uid uint) error
	ReplacePermissionReferences(ctx context.Context, replacements map[model.Permission]model.Permission) error
	DeletePermissionReferences(ctx context.Context, permissions []model.Permission) error
	ReloadPolicy() error

	GetMenuList(ctx context.Context) ([]model.Menu, error)
	MenuUpdate(ctx context.Context, m *model.Menu) error
	MenuCreate(ctx context.Context, m *model.Menu) error
	MenuDelete(ctx context.Context, id uint) error
	RemoveMenuFromApis(ctx context.Context, menuID uint) error

	GetRoles(ctx context.Context, req *v1.GetRoleListRequest) ([]model.Role, int64, error)
	RoleUpdate(ctx context.Context, m *model.Role) error
	RoleCreate(ctx context.Context, m *model.Role) error
	RoleDelete(ctx context.Context, id uint) error
	CasbinRoleDelete(ctx context.Context, role string) error
	GetRole(ctx context.Context, id uint) (model.Role, error)
	GetRoleBySid(ctx context.Context, sid string) (model.Role, error)

	GetApis(ctx context.Context, req *v1.GetApisRequest) ([]model.Api, int64, error)
	GetApiGroups(ctx context.Context) ([]string, error)
	GetApi(ctx context.Context, id uint) (model.Api, error)
	GetApiList(ctx context.Context) ([]model.Api, error)
	ApiPermissionExists(ctx context.Context, path, method string, excludeID uint) (bool, error)
	ApiUpdate(ctx context.Context, m *model.Api) error
	ApiCreate(ctx context.Context, m *model.Api) error
	ApiDelete(ctx context.Context, id uint) error
}

// NewAdminRepository 构造管理后台聚合仓储，由注入容器调用。
func NewAdminRepository(i do.Injector) (AdminRepository, error) {
	return &adminRepository{
		Repository: do.MustInvoke[*Repository](i),
	}, nil
}

// adminRepository 实现 AdminRepository；通过嵌入 Repository 复用
// 连接、事务、查询执行器与 Casbin 执行器。
type adminRepository struct {
	*Repository
}

// ensureRecordExists 校验指定表中 id 对应的未删除行是否存在，
// 用于把 UPDATE 的 0 行影响转换成明确的 ErrNotFound。
func (r *adminRepository) ensureRecordExists(ctx context.Context, table string, id uint) error {
	var count int
	q := r.DB(ctx)
	query := Rebind(r.driver, "SELECT COUNT(*) FROM "+table+" WHERE id = ? AND deleted_at IS NULL")
	if err := q.GetContext(ctx, &count, query, id); err != nil {
		return fmt.Errorf("count %s id %d: %w", table, id, err)
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}
