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

// api 表的接口资源查询与写入。
// 注意：group 是 SQL 保留字，所有引用处按方言加引号。

// GetApiGroups 返回去重后的 API 分组列表。
func (r *adminRepository) GetApiGroups(ctx context.Context) ([]string, error) {
	groupColumn := QuoteIdentifier(r.driver, "group")
	res := make([]string, 0)
	q := r.DB(ctx)
	query := Rebind(r.driver, "SELECT DISTINCT "+groupColumn+" FROM api WHERE deleted_at IS NULL")
	if err := q.SelectContext(ctx, &res, query); err != nil {
		return nil, fmt.Errorf("get api groups: %w", err)
	}
	return res, nil
}

// GetApis 按名称/分组/路径模糊、方法精确的条件分页查询 API，
// 排序为分组升序、id 升序。
func (r *adminRepository) GetApis(ctx context.Context, req *v1.GetApisRequest) ([]model.Api, int64, error) {
	groupColumn := QuoteIdentifier(r.driver, "group")
	conditions := []string{"deleted_at IS NULL"}
	args := []any{}
	if req.Name != "" {
		conditions = append(conditions, "name LIKE ?")
		args = append(args, "%"+req.Name+"%")
	}
	if req.Group != "" {
		conditions = append(conditions, groupColumn+" LIKE ?")
		args = append(args, "%"+req.Group+"%")
	}
	if req.Path != "" {
		conditions = append(conditions, "path LIKE ?")
		args = append(args, "%"+req.Path+"%")
	}
	if req.Method != "" {
		conditions = append(conditions, "method = ?")
		args = append(args, req.Method)
	}
	where := strings.Join(conditions, " AND ")

	var total int64
	q := r.DB(ctx)
	if err := q.GetContext(ctx, &total, Rebind(r.driver, "SELECT COUNT(*) FROM api WHERE "+where), args...); err != nil {
		return nil, total, fmt.Errorf("count apis: %w", err)
	}

	rows := make([]apiRow, 0)
	query := Rebind(r.driver, "SELECT * FROM api WHERE "+where+" ORDER BY "+groupColumn+" ASC, id ASC LIMIT ? OFFSET ?")
	args = append(args, req.PageSize, (req.Page-1)*req.PageSize)
	if err := q.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, total, fmt.Errorf("list apis: %w", err)
	}
	list := make([]model.Api, 0, len(rows))
	for _, row := range rows {
		list = append(list, apiFromRow(row))
	}
	return list, total, nil
}

// GetApi 按 id 查找 API，不存在时返回 ErrNotFound。
func (r *adminRepository) GetApi(ctx context.Context, id uint) (model.Api, error) {
	var row apiRow
	q := r.DB(ctx)
	query := Rebind(r.driver, "SELECT * FROM api WHERE id = ?")
	if err := q.GetContext(ctx, &row, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Api{}, ErrNotFound
		}
		return model.Api{}, fmt.Errorf("get api: %w", err)
	}
	return apiFromRow(row), nil
}

// GetApiList 返回全部 API（按 id 升序），供权限种子与校验使用。
func (r *adminRepository) GetApiList(ctx context.Context) ([]model.Api, error) {
	rows := make([]apiRow, 0)
	q := r.DB(ctx)
	query := Rebind(r.driver, "SELECT * FROM api ORDER BY id ASC")
	if err := q.SelectContext(ctx, &rows, query); err != nil {
		return nil, fmt.Errorf("list apis: %w", err)
	}
	list := make([]model.Api, 0, len(rows))
	for _, row := range rows {
		list = append(list, apiFromRow(row))
	}
	return list, nil
}

// ApiPermissionExists 判断 path+method 是否已被其他 API 占用；
// excludeID 非零时排除自身（更新场景）。
func (r *adminRepository) ApiPermissionExists(ctx context.Context, path string, method string, excludeID uint) (bool, error) {
	conditions := []string{"path = ?", "method = ?"}
	args := []any{path, method}
	if excludeID != 0 {
		conditions = append(conditions, "id <> ?")
		args = append(args, excludeID)
	}
	var count int64
	q := r.DB(ctx)
	query := Rebind(r.driver, "SELECT COUNT(*) FROM api WHERE "+strings.Join(conditions, " AND "))
	if err := q.GetContext(ctx, &count, query, args...); err != nil {
		return false, fmt.Errorf("count api permission: %w", err)
	}
	return count > 0, nil
}

// ApiUpdate 更新 API 的全部业务字段（menu_ids 为 JSON 列）。
func (r *adminRepository) ApiUpdate(ctx context.Context, m *model.Api) error {
	groupColumn := QuoteIdentifier(r.driver, "group")
	q := r.DB(ctx)
	query := Rebind(r.driver, "UPDATE api SET "+groupColumn+" = ?, name = ?, path = ?, method = ?, menu_ids = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?")
	result, err := q.ExecContext(ctx, query, m.Group, m.Name, m.Path, m.Method, UintSlice(m.MenuIDs), m.ID)
	if err != nil {
		return fmt.Errorf("update api: %w", err)
	}
	return rowsAffectedOrNotFound(ctx, r, result, "api", m.ID)
}

// ApiCreate 新增 API；menu_ids 由 UintSlice 自动编码为 JSON 文本，
// id 由数据库自增生成并回填到实体。
func (r *adminRepository) ApiCreate(ctx context.Context, m *model.Api) error {
	groupColumn := QuoteIdentifier(r.driver, "group")
	q := r.DB(ctx)
	query := Rebind(r.driver, "INSERT INTO api ("+groupColumn+", name, path, method, menu_ids) VALUES (?, ?, ?, ?, ?)")
	return r.insertReturningID(ctx, q, query, &m.ID, m.Group, m.Name, m.Path, m.Method, UintSlice(m.MenuIDs))
}

// ApiDelete 物理删除 API 记录（与其他实体的软删除不同：
// path+method 唯一约束要求删除后立即可复用）。
func (r *adminRepository) ApiDelete(ctx context.Context, id uint) error {
	q := r.DB(ctx)
	query := Rebind(r.driver, "DELETE FROM api WHERE id = ?")
	result, err := q.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete api: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}
