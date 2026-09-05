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

// admin_users 表的管理员账号查询与写入。

// GetAdminUserByUsername 按用户名查找未删除的管理员，不存在时返回 ErrNotFound。
func (r *adminRepository) GetAdminUserByUsername(ctx context.Context, username string) (model.AdminUser, error) {
	var row adminUserRow
	q := r.DB(ctx)
	query := Rebind(r.driver, "SELECT * FROM admin_users WHERE username = ? AND deleted_at IS NULL")
	if err := q.GetContext(ctx, &row, query, username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AdminUser{}, ErrNotFound
		}
		return model.AdminUser{}, fmt.Errorf("get admin user by username: %w", err)
	}
	return adminUserFromRow(row), nil
}

// GetAdminUsers 按请求条件分页查询管理员列表，返回列表与总数；排序为 id 倒序。
func (r *adminRepository) GetAdminUsers(ctx context.Context, req *v1.GetAdminUsersRequest) ([]model.AdminUser, int64, error) {
	conditions := []string{"deleted_at IS NULL"}
	args := []any{}
	if req.ID != 0 {
		conditions = append(conditions, "id = ?")
		args = append(args, req.ID)
	}
	if req.Username != "" {
		conditions = append(conditions, "username LIKE ?")
		args = append(args, "%"+req.Username+"%")
	}
	if req.Nickname != "" {
		conditions = append(conditions, "nickname LIKE ?")
		args = append(args, "%"+req.Nickname+"%")
	}
	if req.Email != "" {
		conditions = append(conditions, "email LIKE ?")
		args = append(args, "%"+req.Email+"%")
	}
	if req.Phone != "" {
		conditions = append(conditions, "phone LIKE ?")
		args = append(args, "%"+req.Phone+"%")
	}
	where := strings.Join(conditions, " AND ")

	var total int64
	q := r.DB(ctx)
	if err := q.GetContext(ctx, &total, Rebind(r.driver, "SELECT COUNT(*) FROM admin_users WHERE "+where), args...); err != nil {
		return nil, total, fmt.Errorf("count admin users: %w", err)
	}

	rows := make([]adminUserRow, 0)
	query := Rebind(r.driver, "SELECT * FROM admin_users WHERE "+where+" ORDER BY id DESC LIMIT ? OFFSET ?")
	args = append(args, req.PageSize, (req.Page-1)*req.PageSize)
	if err := q.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, total, fmt.Errorf("list admin users: %w", err)
	}
	list := make([]model.AdminUser, 0, len(rows))
	for _, row := range rows {
		list = append(list, adminUserFromRow(row))
	}
	return list, total, nil
}

// GetAdminUser 按 id 查找未删除的管理员，不存在时返回 ErrNotFound。
func (r *adminRepository) GetAdminUser(ctx context.Context, uid uint) (model.AdminUser, error) {
	var row adminUserRow
	q := r.DB(ctx)
	query := Rebind(r.driver, "SELECT * FROM admin_users WHERE id = ? AND deleted_at IS NULL")
	if err := q.GetContext(ctx, &row, query, uid); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AdminUser{}, ErrNotFound
		}
		return model.AdminUser{}, fmt.Errorf("get admin user: %w", err)
	}
	return adminUserFromRow(row), nil
}

// AdminUserUpdate 更新管理员的资料字段（密码为 bcrypt 哈希后的值）。
func (r *adminRepository) AdminUserUpdate(ctx context.Context, m *model.AdminUser) error {
	q := r.DB(ctx)
	query := Rebind(r.driver, `UPDATE admin_users
SET username = ?, nickname = ?, password = ?, email = ?, phone = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND deleted_at IS NULL`)
	result, err := q.ExecContext(ctx, query,
		m.Username, m.Nickname, m.Password, m.Email, m.Phone, m.ID)
	if err != nil {
		return fmt.Errorf("update admin user: %w", err)
	}
	return rowsAffectedOrNotFound(ctx, r, result, "admin_users", m.ID)
}

// AdminUserCreate 新增管理员，id 由数据库自增生成并回填到实体。
func (r *adminRepository) AdminUserCreate(ctx context.Context, m *model.AdminUser) error {
	q := r.DB(ctx)
	query := Rebind(r.driver,
		"INSERT INTO admin_users (username, nickname, password, email, phone) VALUES (?, ?, ?, ?, ?)")
	return r.insertReturningID(ctx, q, query, &m.ID, m.Username, m.Nickname, m.Password, m.Email, m.Phone)
}

// AdminUserDelete 软删除管理员（置 deleted_at）。
func (r *adminRepository) AdminUserDelete(ctx context.Context, id uint) error {
	q := r.DB(ctx)
	query := Rebind(r.driver,
		"UPDATE admin_users SET deleted_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at IS NULL")
	result, err := q.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete admin user: %w", err)
	}
	return rowsAffectedOrNotFound(ctx, r, result, "admin_users", id)
}
