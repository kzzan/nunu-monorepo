package repository

import (
	"fmt"
	"strings"
)

// dialect 描述某一种数据库驱动在 DDL/DML 上的差异，
// 使同一套建表与查询语句可以适配 sqlite/mysql/postgres。
type dialect struct {
	idCol     string              // 自增主键列定义
	tsCol     string              // 普通时间戳列类型
	tsDefault string              // 带默认值的审计时间戳列（created_at/updated_at 用）
	boolCol   string              // 布尔列类型
	quote     func(string) string // 标识符引用函数
	ignore    func(string) string // 把 INSERT 语句包装为"冲突忽略"形式
}

// dialectFor 返回驱动对应的方言；未知驱动直接报错，避免静默生成错误 SQL。
func dialectFor(driver string) (dialect, error) {
	switch driver {
	case "sqlite":
		return dialect{
			idCol:     "INTEGER PRIMARY KEY AUTOINCREMENT",
			tsCol:     "DATETIME",
			tsDefault: "DATETIME DEFAULT CURRENT_TIMESTAMP",
			boolCol:   "INTEGER",
			quote:     func(name string) string { return "`" + name + "`" },
			ignore:    insertKeywordWrapper("INSERT OR IGNORE INTO"),
		}, nil
	case "mysql":
		return dialect{
			idCol:     "BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY",
			tsCol:     "DATETIME(3)",
			tsDefault: "DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3)",
			boolCol:   "TINYINT(1)",
			quote:     func(name string) string { return "`" + name + "`" },
			ignore:    insertKeywordWrapper("INSERT IGNORE INTO"),
		}, nil
	case "postgres":
		return dialect{
			idCol:     "BIGSERIAL PRIMARY KEY",
			tsCol:     "TIMESTAMPTZ",
			tsDefault: "TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP",
			boolCol:   "BOOLEAN",
			quote:     func(name string) string { return `"` + name + `"` },
			ignore:    func(sql string) string { return sql + " ON CONFLICT DO NOTHING" },
		}, nil
	default:
		return dialect{}, fmt.Errorf("unknown db driver: %s", driver)
	}
}

// insertKeywordWrapper 返回把语句开头 "INSERT INTO" 替换为方言等价
// "冲突忽略"形式的包装函数（sqlite 的 INSERT OR IGNORE、mysql 的 INSERT IGNORE）。
func insertKeywordWrapper(keyword string) func(string) string {
	return func(sql string) string {
		return strings.Replace(sql, "INSERT INTO", keyword, 1)
	}
}

// QuoteIdentifier 按驱动引用标识符（group 等保留字列名必须引用）。
func QuoteIdentifier(driver, name string) string {
	d, err := dialectFor(driver)
	if err != nil {
		return name
	}
	return d.quote(name)
}
