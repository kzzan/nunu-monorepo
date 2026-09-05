// Package repository 是 admin 应用的数据访问层：负责 sqlx 连接与事务管理、
// Casbin 策略适配器，以及各业务实体（管理员/角色/菜单/API/权限）的仓储实现。
//
// 依赖方向：本包只被 service 层与组合根（cmd/*）引用；所有查询通过
// DB(ctx) 获取执行器，事务内的调用自动绑定到当前事务（见 Transaction）。
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/samber/do/v2"
	"github.com/spf13/viper"

	// 数据库驱动：MySQL、PostgreSQL(pgx stdlib) 与纯 Go 实现的 SQLite。
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"

	"github.com/casbin/casbin/v3"
	"github.com/jmoiron/sqlx"
	"nunu-monorepo/pkg/log"
)

// ErrNotFound 是"目标记录不存在"的哨兵错误；service/handler 通过
// errors.Is 判定后映射为 404 类响应。
var ErrNotFound = errors.New("record not found")

// ctxTxKey 是事务对象在 context 中的键类型，避免与其他包的键冲突。
type ctxTxKey struct{}

// Package registers all repository-layer providers into the injector.
var Package = do.Package(
	do.Lazy(NewDB),
	do.Lazy(NewCasbinEnforcer),
	do.Lazy(New),
	do.Lazy(NewTransaction),
	do.Lazy(NewUserRepository),
	do.Lazy(NewAdminRepository),
	//do.Lazy(NewRedis),
)

// Querier 抽象 *sqlx.DB 与 *sqlx.Tx 的共同查询能力，
// 使仓储方法无需关心当前是否处于事务中。
type Querier interface {
	sqlx.ExtContext
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
}

// Repository 是仓储层共享的基础设施：数据库连接、Casbin 执行器与日志。
type Repository struct {
	db     *sqlx.DB
	driver string // mysql / postgres / sqlite，用于方言化 SQL
	e      *casbin.SyncedEnforcer
	logger *log.Logger
}

// New 构造仓储基础对象，由注入容器调用。
func New(i do.Injector) (*Repository, error) {
	return &Repository{
		db:     do.MustInvoke[*sqlx.DB](i),
		driver: DriverFromConf(do.MustInvoke[*viper.Viper](i)),
		e:      do.MustInvoke[*casbin.SyncedEnforcer](i),
		logger: do.MustInvoke[*log.Logger](i),
	}, nil
}

// Transaction 定义事务边界：fn 内所有通过 DB(ctx) 发出的查询
// 都在同一事务中执行；fn 返回错误时整体回滚。
type Transaction interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// NewTransaction 把基础仓储适配为 Transaction，由注入容器调用。
func NewTransaction(i do.Injector) (Transaction, error) {
	return do.MustInvoke[*Repository](i), nil
}

// DB 返回当前上下文使用的查询执行器：
// 若 ctx 由 Transaction 注入则返回事务，否则返回连接池
// （ctx 由各查询方法显式携带，无需包装）。
func (r *Repository) DB(ctx context.Context) Querier {
	if tx, ok := ctx.Value(ctxTxKey{}).(*sqlx.Tx); ok {
		return tx
	}
	return r.db
}

// Transaction 开启事务并执行 fn：fn 通过 ctx 感知事务（DB(ctx) 自动绑定），
// 嵌套调用时直接复用外层事务。fn 出错回滚，成功提交。
func (r *Repository) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(ctxTxKey{}).(*sqlx.Tx); ok {
		return fn(ctx)
	}
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	if err := fn(context.WithValue(ctx, ctxTxKey{}, tx)); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			r.logger.Error().Err(rbErr).Msg("rollback transaction")
		}
		return err
	}
	return tx.Commit()
}

// NewDB 按配置的驱动建立 sqlx 连接并完成连通性检查与连接池参数设置。
// 支持的 driver 取值：mysql、postgres、sqlite。
func NewDB(i do.Injector) (*sqlx.DB, error) {
	conf := do.MustInvoke[*viper.Viper](i)
	driver := DriverFromConf(conf)
	dsn := conf.GetString("data.db.user.dsn")

	var (
		db  *sqlx.DB
		err error
	)
	switch driver {
	case "mysql":
		db, err = sqlx.Connect("mysql", dsn)
	case "postgres":
		db, err = sqlx.Connect("pgx", dsn)
	case "sqlite":
		db, err = sqlx.Connect("sqlite", dsn)
	default:
		return nil, fmt.Errorf("unknown db driver: %s", driver)
	}
	if err != nil {
		return nil, fmt.Errorf("connect %s database: %w", driver, err)
	}

	// Connection Pool config（SQLite 单文件库没有连接池收益，跳过）
	if driver != "sqlite" {
		db.SetMaxIdleConns(10)
		db.SetMaxOpenConns(100)
		db.SetConnMaxLifetime(time.Hour)
	}
	return db, nil
}

// NewCasbinEnforcer 构造 Casbin 同步执行器：
// 策略存储在业务库的 casbin_rule 表（见 casbin_adapter.go），
// RBAC 模型为"角色拥有权限、用户属于角色"的经典三元组。
func NewCasbinEnforcer(i do.Injector) (*casbin.SyncedEnforcer, error) {
	db := do.MustInvoke[*sqlx.DB](i)
	adapter, err := newCasbinAdapter(db, DriverFromConf(do.MustInvoke[*viper.Viper](i)))
	if err != nil {
		return nil, err
	}
	m, err := casbinModelFromString(rbacModel)
	if err != nil {
		return nil, err
	}
	e, err := casbin.NewSyncedEnforcer(m, adapter)
	if err != nil {
		return nil, err
	}

	// 每10秒自动加载策略，防止启动多服务进程策略不一致
	// 如果不想用轮询DB的方式，你也可以使用Casbin Watchers来同步策略，该方式需要基于Redis、Etcd等存储中间件
	// Watchers相关文档：https://casbin.org/zh/docs/watchers
	e.StartAutoLoadPolicy(10 * time.Second)

	// 策略变更（AddPolicy 等）自动写回存储
	e.EnableAutoSave(true)

	return e, nil
}

// rbacModel 是 Casbin 的 RBAC 模型定义：
// 请求 (sub=用户, obj=资源, act=动作)，g 为用户-角色关系，
// 匹配规则要求资源与动作完全相等。
const rbacModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

// NewRedis 构造 Redis 客户端并做连通性检查；当前默认未启用（见注入注册处）。
func NewRedis(i do.Injector) (*redis.Client, error) {
	conf := do.MustInvoke[*viper.Viper](i)
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.GetString("data.redis.addr"),
		Password: conf.GetString("data.redis.password"),
		DB:       conf.GetInt("data.redis.db"),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("redis error: %w", err)
	}

	return rdb, nil
}

// DriverFromConf 从配置读取驱动名，缺省为 sqlite。
func DriverFromConf(conf *viper.Viper) string {
	driver := conf.GetString("data.db.user.driver")
	if driver == "" {
		return "sqlite"
	}
	return driver
}
