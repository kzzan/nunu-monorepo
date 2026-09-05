package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	casbinmodel "github.com/casbin/casbin/v3/model"
	"github.com/casbin/casbin/v3/persist"
	"github.com/jmoiron/sqlx"
)

// 本文件实现基于 sqlx 的 Casbin 策略适配器，取代 casbin/gorm-adapter。
// 策略行存储在业务库 casbin_rule 表（与 gorm-adapter 默认表结构一致，
// 已有数据库可无缝沿用），列固定为 ptype + v0..v5。

// casbinRule 映射 casbin_rule 表的一行；ptype 为 "p"（授权）或 "g"（用户-角色）。
// 值列使用 sql.NullString：短规则只填 v0..v2，v3..v5 可能为 NULL。
type casbinRule struct {
	ID    uint           `db:"id"`
	Ptype string         `db:"ptype"`
	V0    sql.NullString `db:"v0"`
	V1    sql.NullString `db:"v1"`
	V2    sql.NullString `db:"v2"`
	V3    sql.NullString `db:"v3"`
	V4    sql.NullString `db:"v4"`
	V5    sql.NullString `db:"v5"`
}

// ruleColumns 是策略值列，按 Casbin 惯例最多 6 个。
var ruleColumns = []string{"v0", "v1", "v2", "v3", "v4", "v5"}

// casbinAdapter 把 casbin_rule 表适配为 casbin 的 persist.Adapter。
type casbinAdapter struct {
	db     *sqlx.DB
	driver string
}

// 编译期校验适配器满足 Casbin 存储接口。
var _ persist.Adapter = (*casbinAdapter)(nil)

// newCasbinAdapter 构造适配器并确保 casbin_rule 表已存在（幂等），
// 使全新数据库上服务启动与迁移命令都能直接工作。
func newCasbinAdapter(db *sqlx.DB, driver string) (*casbinAdapter, error) {
	d, err := dialectFor(driver)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(casbinRuleStatement(d)); err != nil {
		return nil, fmt.Errorf("ensure casbin_rule table: %w", err)
	}
	return &casbinAdapter{db: db, driver: driver}, nil
}

// casbinModelFromString 从 Casbin 模型文本构造模型对象。
func casbinModelFromString(text string) (casbinmodel.Model, error) {
	return casbinmodel.NewModelFromString(text)
}

// LoadPolicy 把表内全部策略规则加载进内存模型。
func (a *casbinAdapter) LoadPolicy(m casbinmodel.Model) error {
	var rules []casbinRule
	if err := a.db.Select(&rules, "SELECT * FROM casbin_rule"); err != nil {
		return fmt.Errorf("load casbin policy: %w", err)
	}
	for _, rule := range rules {
		tokens := append([]string{rule.Ptype}, rule.values()...)
		if err := persist.LoadPolicyArray(tokens, m); err != nil {
			return err
		}
	}
	return nil
}

// SavePolicy 用内存中的全量策略覆盖表内数据（先清空再写入，
// 单事务保证不留中间态）。
func (a *casbinAdapter) SavePolicy(m casbinmodel.Model) error {
	ctx := context.Background()
	tx, err := a.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin save policy: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, "DELETE FROM casbin_rule"); err != nil {
		return fmt.Errorf("clear casbin policy: %w", err)
	}
	// p 规则（角色-资源-动作）与 g 规则（用户-角色）依次落库；
	// g 规则经由断言的 Policy 字段读取（v3 无 GetGroupingPolicy 方法）
	groupingRules := [][]string{}
	if g := m["g"]["g"]; g != nil {
		groupingRules = g.Policy
	}
	pRules, err := m.GetPolicy("p", "p")
	if err != nil {
		return fmt.Errorf("read in-memory casbin policy: %w", err)
	}
	for ptype, rules := range map[string][][]string{
		"p": pRules,
		"g": groupingRules,
	} {
		for _, rule := range rules {
			if err := insertRule(ctx, tx, a.driver, ptype, rule); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

// AddPolicy 追加单条策略（Auto-Save 模式下由执行器自动调用；
// 执行器仅在规则不存在时触发，因此无需方言冲突忽略子句之外的防护）。
func (a *casbinAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	return insertRule(context.Background(), a.db, a.driver, ptype, rule)
}

// RemovePolicy 删除与给定规则完全匹配的策略行。
func (a *casbinAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	conditions, args := ruleConditions(ptype, rule)
	query := "DELETE FROM casbin_rule WHERE " + strings.Join(conditions, " AND ")
	if _, err := a.db.Exec(Rebind(a.driver, query), args...); err != nil {
		return fmt.Errorf("remove casbin policy: %w", err)
	}
	return nil
}

// RemoveFilteredPolicy 删除从 fieldIndex 起与 fieldValues 前缀匹配的策略行；
// 空 fieldValue 表示该位置不参与过滤。
func (a *casbinAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	conditions := []string{"ptype = ?"}
	args := []any{ptype}
	for offset, value := range fieldValues {
		if value == "" {
			continue
		}
		column := fieldIndex + offset
		if column < 0 || column >= len(ruleColumns) {
			return fmt.Errorf("casbin filter field index out of range: %d", column)
		}
		conditions = append(conditions, ruleColumns[column]+" = ?")
		args = append(args, value)
	}
	query := "DELETE FROM casbin_rule WHERE " + strings.Join(conditions, " AND ")
	if _, err := a.db.Exec(Rebind(a.driver, query), args...); err != nil {
		return fmt.Errorf("remove filtered casbin policy: %w", err)
	}
	return nil
}

// insertRule 插入一条策略；忽略与现存行冲突的情况（幂等写入）。
// 不足 6 段的规则以空串补齐 v 列，保证列数与占位符始终对齐。
func insertRule(ctx context.Context, db sqlx.ExtContext, driver, ptype string, rule []string) error {
	columns := append([]string{"ptype"}, ruleColumns...)
	placeholders := make([]string, len(columns))
	args := make([]any, 0, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
		if i == 0 {
			args = append(args, ptype)
			continue
		}
		if valueIndex := i - 1; valueIndex < len(rule) {
			args = append(args, rule[valueIndex])
		} else {
			args = append(args, "")
		}
	}
	base := fmt.Sprintf("INSERT INTO casbin_rule (%s) VALUES (%s)",
		strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	d, err := dialectFor(driver)
	if err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, Rebind(driver, d.ignore(base)), args...); err != nil {
		return fmt.Errorf("insert casbin policy: %w", err)
	}
	return nil
}

// ruleConditions 生成精确匹配一条规则的 WHERE 片段与参数。
func ruleConditions(ptype string, rule []string) ([]string, []any) {
	conditions := []string{"ptype = ?"}
	args := []any{ptype}
	for i, value := range rule {
		if i >= len(ruleColumns) {
			break
		}
		conditions = append(conditions, ruleColumns[i]+" = ?")
		args = append(args, value)
	}
	return conditions, args
}

// values 返回规则行的非空值切片（v0..v5 依次截断）。
func (r casbinRule) values() []string {
	values := []string{r.V0.String, r.V1.String, r.V2.String, r.V3.String, r.V4.String, r.V5.String}
	last := 0
	for i, v := range values {
		if v != "" {
			last = i + 1
		}
	}
	return values[:last]
}

// Rebind 按方言把 ? 占位符转换为 $N（仅 postgres 需要，其余原样返回）。
func Rebind(driver, query string) string {
	if driver != "postgres" {
		return query
	}
	return sqlx.Rebind(sqlx.DOLLAR, query)
}
