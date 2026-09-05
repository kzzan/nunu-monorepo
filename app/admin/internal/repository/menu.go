package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"nunu-monorepo/app/admin/internal/model"
)

// menu 表的管理端菜单查询与写入。

// GetMenuList 返回全部未删除菜单，按权重倒序（与前端展示顺序一致）。
func (r *adminRepository) GetMenuList(ctx context.Context) ([]model.Menu, error) {
	rows := make([]menuRow, 0)
	q := r.DB(ctx)
	query := Rebind(r.driver, "SELECT * FROM menu WHERE deleted_at IS NULL ORDER BY weight DESC")
	if err := q.SelectContext(ctx, &rows, query); err != nil {
		return nil, fmt.Errorf("list menus: %w", err)
	}
	list := make([]model.Menu, 0, len(rows))
	for _, row := range rows {
		list = append(list, menuFromRow(row))
	}
	return list, nil
}

// MenuCreate 新增菜单；roles/auth_list 由 JSON 列类型自动编码。
func (r *adminRepository) MenuCreate(ctx context.Context, m *model.Menu) error {
	q := r.DB(ctx)
	query := Rebind(r.driver, `INSERT INTO menu (
parent_id, path, title, name, component, locale, icon, redirect, url, link, target,
active_path, show_text_badge, weight, is_enable, is_menu, keep_alive, hide_in_menu, is_hide,
is_hide_tab, is_iframe, show_badge, fixed_tab, is_full_page, roles, auth_list
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	args := []any{
		m.ParentID, m.Path, m.Title, m.Name, m.Component, m.Locale, m.Icon, m.Redirect,
		m.URL, m.Link, m.Target, m.ActivePath, m.ShowTextBadge, m.Weight, m.IsEnable,
		m.IsMenu, m.KeepAlive, m.HideInMenu, m.IsHide, m.IsHideTab, m.IsIframe,
		m.ShowBadge, m.FixedTab, m.IsFullPage, StringSlice(m.Roles), MenuAuths(m.AuthList),
	}
	return r.insertReturningID(ctx, q, query, &m.ID, args...)
}

// MenuUpdate 更新菜单的全部业务字段；0 行影响时复核存在性并返回 ErrNotFound。
func (r *adminRepository) MenuUpdate(ctx context.Context, m *model.Menu) error {
	q := r.DB(ctx)
	query := Rebind(r.driver, `UPDATE menu SET
parent_id = ?, path = ?, title = ?, name = ?, component = ?, locale = ?, icon = ?, redirect = ?,
url = ?, link = ?, target = ?, active_path = ?, show_text_badge = ?, weight = ?, is_enable = ?,
is_menu = ?, keep_alive = ?, hide_in_menu = ?, is_hide = ?, is_hide_tab = ?, is_iframe = ?,
show_badge = ?, fixed_tab = ?, is_full_page = ?, roles = ?, auth_list = ?,
updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND deleted_at IS NULL`)
	args := []any{
		m.ParentID, m.Path, m.Title, m.Name, m.Component, m.Locale, m.Icon, m.Redirect,
		m.URL, m.Link, m.Target, m.ActivePath, m.ShowTextBadge, m.Weight, m.IsEnable,
		m.IsMenu, m.KeepAlive, m.HideInMenu, m.IsHide, m.IsHideTab, m.IsIframe,
		m.ShowBadge, m.FixedTab, m.IsFullPage, StringSlice(m.Roles), MenuAuths(m.AuthList), m.ID,
	}
	result, err := q.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update menu: %w", err)
	}
	return rowsAffectedOrNotFound(ctx, r, result, "menu", m.ID)
}

// MenuDelete 软删除菜单（置 deleted_at）。
func (r *adminRepository) MenuDelete(ctx context.Context, id uint) error {
	q := r.DB(ctx)
	query := Rebind(r.driver,
		"UPDATE menu SET deleted_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at IS NULL")
	result, err := q.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete menu: %w", err)
	}
	return rowsAffectedOrNotFound(ctx, r, result, "menu", id)
}

// RemoveMenuFromApis 从所有 API 的关联菜单列表中摘除指定菜单：
// 逐个检查并重写 menu_ids JSON 列。菜单删除事务中调用。
func (r *adminRepository) RemoveMenuFromApis(ctx context.Context, menuID uint) error {
	q := r.DB(ctx)
	query := Rebind(r.driver, "SELECT * FROM api ORDER BY id ASC")
	rows := make([]apiRow, 0)
	if err := q.SelectContext(ctx, &rows, query); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("list apis for menu removal: %w", err)
	}
	updateQuery := Rebind(r.driver, "UPDATE api SET menu_ids = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?")
	for _, row := range rows {
		menuIDs := make(UintSlice, 0, len(row.MenuIDs))
		changed := false
		for _, id := range row.MenuIDs {
			if id == menuID {
				changed = true
				continue
			}
			menuIDs = append(menuIDs, id)
		}
		if !changed {
			continue
		}
		if _, err := q.ExecContext(ctx, updateQuery, menuIDs, row.ID); err != nil {
			return fmt.Errorf("remove menu %d from api %d: %w", menuID, row.ID, err)
		}
	}
	return nil
}
