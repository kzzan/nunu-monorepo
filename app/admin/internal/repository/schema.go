package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

// 本文件集中生成业务表的 DDL，供迁移命令（cmd/migration，破坏性重建）
// 与测试（内存库建表）共用。列定义与 model 包的实体字段一一对应。

// CreateAdminTables 按驱动方言创建全部业务表（IF NOT EXISTS，幂等），
// 包括 admin_users、roles、api、menu 与 casbin_rule。
func CreateAdminTables(ctx context.Context, db *sqlx.DB, driver string) error {
	statements, err := createTableStatements(driver)
	if err != nil {
		return err
	}
	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("create table: %w", err)
		}
	}
	return nil
}

// DropAdminTables 删除全部业务表（IF EXISTS）。迁移命令用它做破坏性重建，
// 仅限本地重置/种子流程使用。
func DropAdminTables(ctx context.Context, db *sqlx.DB, driver string) error {
	tables := []string{"admin_users", "roles", "api", "menu"}
	for _, table := range tables {
		if _, err := db.ExecContext(ctx, "DROP TABLE IF EXISTS "+QuoteIdentifier(driver, table)); err != nil {
			return fmt.Errorf("drop table %s: %w", table, err)
		}
	}
	return nil
}

// createTableStatements 拼装全部建表语句；group 是 SQL 保留字，需要按方言引用。
func createTableStatements(driver string) ([]string, error) {
	d, err := dialectFor(driver)
	if err != nil {
		return nil, err
	}
	group := d.quote("group")
	build := func(table string, columns []string, extras ...string) string {
		return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n  %s\n)",
			d.quote(table), strings.Join(append(columns, extras...), ",\n  "))
	}

	adminUsers := build("admin_users", []string{
		"id " + d.idCol,
		"created_at " + d.tsDefault,
		"updated_at " + d.tsDefault,
		"deleted_at " + d.tsCol + " NULL",
		"username VARCHAR(50) NOT NULL UNIQUE",
		"nickname VARCHAR(50) NOT NULL",
		"password VARCHAR(255) NOT NULL",
		"email VARCHAR(100) NOT NULL",
		"phone VARCHAR(20) NOT NULL",
	})

	roles := build("roles", []string{
		"id " + d.idCol,
		"created_at " + d.tsDefault,
		"updated_at " + d.tsDefault,
		"deleted_at " + d.tsCol + " NULL",
		"name VARCHAR(100) UNIQUE",
		"sid VARCHAR(100) UNIQUE",
	})

	apis := build("api", []string{
		"id " + d.idCol,
		"created_at " + d.tsDefault,
		"updated_at " + d.tsDefault,
		"deleted_at " + d.tsCol + " NULL",
		group + " VARCHAR(255) NOT NULL",
		"name VARCHAR(100) NOT NULL",
		"path VARCHAR(255) NOT NULL",
		"method VARCHAR(20) NOT NULL",
		"menu_ids TEXT",
	}, "UNIQUE ("+d.quote("path")+", "+d.quote("method")+")")

	menus := build("menu", []string{
		"id " + d.idCol,
		"created_at " + d.tsDefault,
		"updated_at " + d.tsDefault,
		"deleted_at " + d.tsCol + " NULL",
		"parent_id " + d.intCol(),
		"path VARCHAR(255)",
		"title VARCHAR(100)",
		"name VARCHAR(100)",
		"component VARCHAR(255)",
		"locale VARCHAR(100)",
		"icon VARCHAR(100)",
		"redirect VARCHAR(255)",
		"url VARCHAR(255)",
		"link VARCHAR(255)",
		"target VARCHAR(20)",
		"active_path VARCHAR(255)",
		"show_text_badge VARCHAR(50)",
		"weight " + d.intCol() + " DEFAULT 0",
		"is_enable " + d.boolCol,
		"is_menu " + d.boolCol,
		"keep_alive " + d.boolCol + " DEFAULT " + d.boolFalse(),
		"hide_in_menu " + d.boolCol + " DEFAULT " + d.boolFalse(),
		"is_hide " + d.boolCol + " DEFAULT " + d.boolFalse(),
		"is_hide_tab " + d.boolCol + " DEFAULT " + d.boolFalse(),
		"is_iframe " + d.boolCol + " DEFAULT " + d.boolFalse(),
		"show_badge " + d.boolCol + " DEFAULT " + d.boolFalse(),
		"fixed_tab " + d.boolCol + " DEFAULT " + d.boolFalse(),
		"is_full_page " + d.boolCol + " DEFAULT " + d.boolFalse(),
		"roles TEXT",
		"auth_list TEXT",
	})

	casbinRule := casbinRuleStatement(d)

	return []string{adminUsers, roles, apis, menus, casbinRule}, nil
}

// casbinRuleStatement 生成 casbin_rule 表的建表语句，供整体建表与
// Casbin 适配器（仅确保自身表存在）共用。
func casbinRuleStatement(d dialect) string {
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n  %s\n)",
		d.quote("casbin_rule"), strings.Join([]string{
			"id " + d.idCol,
			"ptype VARCHAR(100)",
			"v0 VARCHAR(100)",
			"v1 VARCHAR(100)",
			"v2 VARCHAR(100)",
			"v3 VARCHAR(100)",
			"v4 VARCHAR(100)",
			"v5 VARCHAR(100)",
		}, ",\n  "))
}

// intCol 返回普通整型列类型。
func (d dialect) intCol() string {
	if d.boolCol == "BOOLEAN" { // postgres 方言
		return "BIGINT"
	}
	return "INTEGER"
}

// boolFalse 返回布尔假值字面量（postgres 与 sqlite/mysql 字面量不同）。
func (d dialect) boolFalse() string {
	if d.boolCol == "BOOLEAN" {
		return "FALSE"
	}
	return "0"
}
