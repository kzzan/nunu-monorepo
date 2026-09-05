package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/duke-git/lancet/v2/convertor"
	"nunu-monorepo/app/admin/internal/model"
)

// 权限（Casbin 策略）相关：用户-角色绑定、角色授权、
// 权限引用替换与删除。策略行存于 casbin_rule 表，与
// Casbin 执行器双写协同：批量变更后需 ReloadPolicy 生效。

// DeleteUserRoles 清空用户在 Casbin 中的全部角色绑定。
func (r *adminRepository) DeleteUserRoles(ctx context.Context, uid uint) error {
	if _, err := r.e.DeleteRolesForUser(convertor.ToString(uid)); err != nil {
		return fmt.Errorf("delete roles for user %d: %w", uid, err)
	}
	return nil
}

// UpdateUserRoles 把用户的角色绑定增量同步为目标集合：
// 只删除移除的角色、只新增添加的角色，避免全删全建造成的抖动。
func (r *adminRepository) UpdateUserRoles(ctx context.Context, uid uint, roles []string) error {
	user := convertor.ToString(uid)
	if len(roles) == 0 {
		if _, err := r.e.DeleteRolesForUser(user); err != nil {
			return fmt.Errorf("clear roles for user %d: %w", uid, err)
		}
		return nil
	}
	old, err := r.e.GetRolesForUser(user)
	if err != nil {
		return fmt.Errorf("get roles for user %d: %w", uid, err)
	}

	// 计算需要增删的角色差集
	oldMap := make(map[string]struct{}, len(old))
	newMap := make(map[string]struct{}, len(roles))
	for _, v := range old {
		oldMap[v] = struct{}{}
	}
	for _, v := range roles {
		newMap[v] = struct{}{}
	}
	addRoles := make([]string, 0)
	delRoles := make([]string, 0)
	for key := range oldMap {
		if _, exists := newMap[key]; !exists {
			delRoles = append(delRoles, key)
		}
	}
	for key := range newMap {
		if _, exists := oldMap[key]; !exists {
			addRoles = append(addRoles, key)
		}
	}
	if len(addRoles) == 0 && len(delRoles) == 0 {
		return nil
	}
	for _, role := range delRoles {
		if _, err := r.e.DeleteRoleForUser(user, role); err != nil {
			r.logger.WithContext(ctx).Error().Err(err).Msg("DeleteRoleForUser error")
			return err
		}
	}
	if len(addRoles) > 0 {
		if _, err := r.e.AddRolesForUser(user, addRoles); err != nil {
			return fmt.Errorf("add roles for user %d: %w", uid, err)
		}
	}
	return nil
}

// UpdateRolePermission 用目标权限集合整体替换角色的全部授权：
// 事务内先删后插 casbin_rule 中该角色的 p 规则，随后重载策略。
func (r *adminRepository) UpdateRolePermission(ctx context.Context, role string, permissions []model.Permission) error {
	err := r.Transaction(ctx, func(ctx context.Context) error {
		q := r.DB(ctx)
		query := Rebind(r.driver, "DELETE FROM casbin_rule WHERE ptype = ? AND v0 = ?")
		if _, err := q.ExecContext(ctx, query, "p", role); err != nil {
			return fmt.Errorf("clear role permissions: %w", err)
		}
		if len(permissions) == 0 {
			return nil
		}
		return insertPermissionRules(ctx, q, r.driver, role, permissions)
	})
	if err != nil {
		return err
	}
	return r.ReloadPolicy()
}

// ReplacePermissionReferences 把引用旧权限的规则批量改写为新权限
// （菜单/API 改名后同步 Casbin 策略）。按"删除旧行 + 插入新行"实现。
func (r *adminRepository) ReplacePermissionReferences(ctx context.Context, replacements map[model.Permission]model.Permission) error {
	for oldPermission, newPermission := range replacements {
		if oldPermission == newPermission {
			continue
		}
		q := r.DB(ctx)
		var rules []casbinRule
		query := Rebind(r.driver, "SELECT * FROM casbin_rule WHERE ptype = ? AND v1 = ? AND v2 = ?")
		if err := q.SelectContext(ctx, &rules, query, "p", oldPermission.Resource, oldPermission.Action); err != nil {
			return fmt.Errorf("load permission references: %w", err)
		}
		if len(rules) == 0 {
			continue
		}

		// 删除引用旧权限的行
		ids := make([]any, 0, len(rules))
		placeholders := make([]string, 0, len(rules))
		for _, rule := range rules {
			ids = append(ids, rule.ID)
			placeholders = append(placeholders, "?")
		}
		deleteQuery := "DELETE FROM casbin_rule WHERE id IN (" + strings.Join(placeholders, ", ") + ")"
		if _, err := q.ExecContext(ctx, Rebind(r.driver, deleteQuery), ids...); err != nil {
			return fmt.Errorf("delete permission references: %w", err)
		}

		// 以新权限值重新插入（ID 置零由数据库重新分配，冲突行忽略）
		replacements := make([]model.Permission, 0, len(rules))
		for range rules {
			replacements = append(replacements, newPermission)
		}
		roles := make([]string, 0, len(rules))
		for _, rule := range rules {
			roles = append(roles, rule.V0.String)
		}
		if err := insertPermissionRulesPerRole(ctx, q, r.driver, roles, replacements); err != nil {
			return err
		}
	}
	return nil
}

// DeletePermissionReferences 删除引用指定权限的全部规则（菜单/API 删除时调用）。
func (r *adminRepository) DeletePermissionReferences(ctx context.Context, permissions []model.Permission) error {
	q := r.DB(ctx)
	query := Rebind(r.driver, "DELETE FROM casbin_rule WHERE ptype = ? AND v1 = ? AND v2 = ?")
	for _, permission := range permissions {
		if _, err := q.ExecContext(ctx, query, "p", permission.Resource, permission.Action); err != nil {
			return fmt.Errorf("delete permission references: %w", err)
		}
	}
	return nil
}

// ReloadPolicy 让 Casbin 执行器重新加载库内策略，使直写 casbin_rule
// 的批量变更立即生效。
func (r *adminRepository) ReloadPolicy() error {
	return r.e.LoadPolicy()
}

// GetUserPermissions 返回用户的隐式权限集合（含其所有角色的授权）。
func (r *adminRepository) GetUserPermissions(ctx context.Context, uid uint) ([][]string, error) {
	return r.e.GetImplicitPermissionsForUser(convertor.ToString(uid))
}

// GetRolePermissions 返回角色自身的显式授权集合。
func (r *adminRepository) GetRolePermissions(ctx context.Context, role string) ([][]string, error) {
	return r.e.GetPermissionsForUser(role)
}

// GetUserRoles 返回用户绑定的角色 Sid 列表。
func (r *adminRepository) GetUserRoles(ctx context.Context, uid uint) ([]string, error) {
	return r.e.GetRolesForUser(convertor.ToString(uid))
}

// insertPermissionRules 为单个角色批量插入 p 规则（冲突忽略）。
func insertPermissionRules(ctx context.Context, q Querier, driver, role string, permissions []model.Permission) error {
	roles := make([]string, len(permissions))
	for i := range permissions {
		roles[i] = role
	}
	return insertPermissionRulesPerRole(ctx, q, driver, roles, permissions)
}

// insertPermissionRulesPerRole 按行携带各自角色批量插入 p 规则（冲突忽略）。
func insertPermissionRulesPerRole(ctx context.Context, q Querier, driver string, roles []string, permissions []model.Permission) error {
	d, err := dialectFor(driver)
	if err != nil {
		return err
	}
	var sb strings.Builder
	args := make([]any, 0, len(roles)*3)
	sb.WriteString("INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES ")
	for i := range permissions {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("(?, ?, ?, ?)")
		args = append(args, "p", roles[i], permissions[i].Resource, permissions[i].Action)
	}
	if _, err := q.ExecContext(ctx, Rebind(driver, d.ignore(sb.String())), args...); err != nil {
		return fmt.Errorf("insert casbin rules: %w", err)
	}
	return nil
}
