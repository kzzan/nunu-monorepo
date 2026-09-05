package repository

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"nunu-monorepo/app/admin/internal/model"
)

// 本文件定义 JSON 列的编码类型：把切片以 JSON 文本形式存入单个 TEXT 列。
// 它们实现 sql.Scanner 与 driver.Valuer，直接参与 sqlx 的读写。
// 这是持久化编码细节，对 model 包（纯领域模型）不可见：
// 领域侧只持有普通切片（[]string / []uint / []model.MenuAuth），
// 在仓储方法的参数与返回处完成包装/解包。

// StringSlice 是以 JSON 文本存储的 []string 列（如 menu.roles）。
type StringSlice []string

// Value 实现 driver.Valuer：nil 与空切片统一编码为 "[]"，避免 NULL 与空数组混用。
func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	return json.Marshal(s)
}

// Scan 实现 sql.Scanner：兼容数据库返回的 []byte、string 与 NULL。
func (s *StringSlice) Scan(src any) error {
	return scanJSONColumn(src, s)
}

// UintSlice 是以 JSON 文本存储的 []uint 列（如 api.menu_ids）。
type UintSlice []uint

// Value 实现 driver.Valuer：nil 与空切片统一编码为 "[]"。
func (u UintSlice) Value() (driver.Value, error) {
	if u == nil {
		return "[]", nil
	}
	return json.Marshal(u)
}

// Scan 实现 sql.Scanner：兼容数据库返回的 []byte、string 与 NULL。
func (u *UintSlice) Scan(src any) error {
	return scanJSONColumn(src, u)
}

// MenuAuths 是以 JSON 文本存储的 []model.MenuAuth 列（如 menu.auth_list）。
type MenuAuths []model.MenuAuth

// Value 实现 driver.Valuer：nil 与空切片统一编码为 "[]"。
func (m MenuAuths) Value() (driver.Value, error) {
	if m == nil {
		return "[]", nil
	}
	return json.Marshal(m)
}

// Scan 实现 sql.Scanner：兼容数据库返回的 []byte、string 与 NULL。
func (m *MenuAuths) Scan(src any) error {
	return scanJSONColumn(src, m)
}

// scanJSONColumn 把数据库列值解码进任意切片指针，是三种 JSON 列类型共用的扫描实现。
func scanJSONColumn(src any, dest any) error {
	switch v := src.(type) {
	case nil:
		return json.Unmarshal([]byte("[]"), dest)
	case []byte:
		return json.Unmarshal(v, dest)
	case string:
		return json.Unmarshal([]byte(v), dest)
	default:
		return fmt.Errorf("unsupported json column source: %T", src)
	}
}
