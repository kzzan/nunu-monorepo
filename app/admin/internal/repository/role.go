package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	v1 "nunu-monorepo/app/admin/api/v1"
	"nunu-monorepo/app/admin/internal/model"
)

// roles 表的后台角色查询与写入，以及 Casbin 侧的角色清理。

// CasbinRoleDelete 删除 Casbin 中该角色的全部策略（角色本身在 roles 表删除之外，
// 还要清掉它关联的授权与用户绑定）。
func (r *adminRepository) CasbinRoleDelete(ctx context.Context, role string) error {
	if _, err := r.e.DeleteRole(role); err != nil {
		return fmt.Errorf("delete casbin role %s: %w", role, err)
	}
	return nil
}

// GetRole 按 id 查找未删除角色，不存在时返回 ErrNotFound。
func (r *adminRepository) GetRole(ctx context.Context, id uint) (model.Role, error) {
	var row roleRow
	q := r.DB(ctx)
	query := Rebind(r.driver, "SELECT * FROM roles WHERE id = ? AND deleted_at IS NULL")
	if err := q.GetContext(ctx, &row, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Role{}, ErrNotFound
		}
		return model.Role{}, fmt.Errorf("get role: %w", err)
	}
	return roleFromRow(row), nil
}

// GetRoleBySid 按角色 Sid 查找未删除角色（Sid 即 Casbin 中的角色名）。
func (r *adminRepository) GetRoleBySid(ctx context.Context, sid string) (model.Role, error) {
	var row roleRow
	q := r.DB(ctx)
	query := Rebind(r.driver, "SELECT * FROM roles WHERE sid = ? AND deleted_at IS NULL")
	if err := q.GetContext(ctx, &row, query, sid); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Role{}, ErrNotFound
		}
		return model.Role{}, fmt.Errorf("get role by sid: %w", err)
	}
	return roleFromRow(row), nil
}

// RoleUpdate 仅更新角色展示名；Sid 创建后不可变更。
func (r *adminRepository) RoleUpdate(ctx context.Context, m *model.Role) error {
	q := r.DB(ctx)
	query := Rebind(r.driver,
		"UPDATE roles SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at IS NULL")
	result, err := q.ExecContext(ctx, query, m.Name, m.ID)
	if err != nil {
		return fmt.Errorf("update role: %w", err)
	}
	return rowsAffectedOrNotFound(ctx, r, result, "roles", m.ID)
}

// RoleCreate 新增角色，id 由数据库自增生成并回填到实体。
func (r *adminRepository) RoleCreate(ctx context.Context, m *model.Role) error {
	q := r.DB(ctx)
	query := Rebind(r.driver, "INSERT INTO roles (name, sid) VALUES (?, ?)")
	return r.insertReturningID(ctx, q, query, &m.ID, m.Name, m.Sid)
}

// RoleDelete 软删除角色（置 deleted_at）。
func (r *adminRepository) RoleDelete(ctx context.Context, id uint) error {
	q := r.DB(ctx)
	query := Rebind(r.driver,
		"UPDATE roles SET deleted_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at IS NULL")
	result, err := q.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete role: %w", err)
	}
	return rowsAffectedOrNotFound(ctx, r, result, "roles", id)
}

// GetRoles 按名称模糊、Sid 精确的条件分页查询角色列表。
func (r *adminRepository) GetRoles(ctx context.Context, req *v1.GetRoleListRequest) ([]model.Role, int64, error) {
	conditions := []string{"deleted_at IS NULL"}
	args := []any{}
	if req.Name != "" {
		conditions = append(conditions, "name LIKE ?")
		args = append(args, "%"+req.Name+"%")
	}
	if req.Sid != "" {
		conditions = append(conditions, "sid = ?")
		args = append(args, req.Sid)
	}
	where := strings.Join(conditions, " AND ")

	var total int64
	q := r.DB(ctx)
	if err := q.GetContext(ctx, &total, Rebind(r.driver, "SELECT COUNT(*) FROM roles WHERE "+where), args...); err != nil {
		return nil, total, fmt.Errorf("count roles: %w", err)
	}

	rows := make([]roleRow, 0)
	query := Rebind(r.driver, "SELECT * FROM roles WHERE "+where+" ORDER BY id DESC LIMIT ? OFFSET ?")
	args = append(args, req.PageSize, (req.Page-1)*req.PageSize)
	if err := q.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, total, fmt.Errorf("list roles: %w", err)
	}
	list := make([]model.Role, 0, len(rows))
	for _, row := range rows {
		list = append(list, roleFromRow(row))
	}
	return list, total, nil
}
