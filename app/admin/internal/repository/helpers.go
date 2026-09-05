package repository

import (
	"context"
	"database/sql"
	"fmt"
)

// 本文件存放跨仓储复用的小工具：影响行数判定、自增主键回填。

// rowsAffectedOrNotFound 把 UPDATE 的 0 行影响转换为 ErrNotFound：
// 先区分"行不存在"与"值未变化"两种情况（后者也会产生 0 行影响），
// 通过 ensureRecordExists 复核后返回明确错误。
func rowsAffectedOrNotFound(ctx context.Context, r *adminRepository, result sql.Result, table string, id uint) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read rows affected for %s: %w", table, err)
	}
	if affected == 0 {
		return r.ensureRecordExists(ctx, table, id)
	}
	return nil
}

// insertReturningID 执行 INSERT 并把数据库生成的自增主键回填到 *dest：
// postgres 需要 RETURNING id 子句，sqlite/mysql 通过 LastInsertId 获取。
func (r *Repository) insertReturningID(ctx context.Context, q Querier, query string, dest *uint, args ...any) error {
	if r.driver == "postgres" {
		if err := q.GetContext(ctx, dest, Rebind(r.driver, query+" RETURNING id"), args...); err != nil {
			return err
		}
		return nil
	}
	result, err := q.ExecContext(ctx, Rebind(r.driver, query), args...)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("read last insert id: %w", err)
	}
	*dest = uint(id)
	return nil
}
