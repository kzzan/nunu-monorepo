// Package model 定义管理后台（admin 应用）私有的领域实体与 RBAC 语言。
//
// 本包是纯领域模型：不含任何持久化细节（无 db 标签、无 database/sql 依赖、
// 无行扫描类型）——表行结构与 JSON 列编码由 repository 层私有并负责映射。
// 实体只属于 admin 应用的限界上下文：其他应用需要这些数据时应调用 admin
// 暴露的 API 或消费其事件，而不是 import 本包（参见 README 的包边界规则）。
package model

import "time"

// Base 是所有实体的公共审计字段（领域视图）。
// 软删除标记（deleted_at）属于存储细节，由 repository 层的行结构持有。
type Base struct {
	ID        uint
	CreatedAt time.Time
	UpdatedAt time.Time
}
