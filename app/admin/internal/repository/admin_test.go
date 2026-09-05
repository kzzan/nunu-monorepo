package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/casbin/casbin/v3"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
	v1 "nunu-monorepo/app/admin/api/v1"
	"nunu-monorepo/app/admin/internal/model"
)

// newPermissionTestRepository 在共享缓存的内存 SQLite 上构造带 Casbin
// 执行器的仓储，供权限相关仓储测试使用。
func newPermissionTestRepository(t *testing.T) *adminRepository {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := sqlx.Connect("sqlite", dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := CreateAdminTables(context.Background(), db, "sqlite"); err != nil {
		t.Fatalf("create tables: %v", err)
	}
	casbinModel, err := casbinModelFromString(rbacModel)
	if err != nil {
		t.Fatalf("create casbin model: %v", err)
	}
	adapter, err := newCasbinAdapter(db, "sqlite")
	if err != nil {
		t.Fatalf("create casbin adapter: %v", err)
	}
	enforcer, err := casbin.NewSyncedEnforcer(casbinModel, adapter)
	if err != nil {
		t.Fatalf("create casbin enforcer: %v", err)
	}
	enforcer.EnableAutoSave(true)
	return &adminRepository{Repository: &Repository{db: db, driver: "sqlite", e: enforcer}}
}

// TestUpdateRolePermissionReplacesOnlyTargetRole 验证整体替换角色权限时
// 只影响目标角色，其他角色的策略保持不变。
func TestUpdateRolePermissionReplacesOnlyTargetRole(t *testing.T) {
	repository := newPermissionTestRepository(t)
	if _, err := repository.e.AddPermissionForUser("role-a", "api:/old", "GET"); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.e.AddPermissionForUser("role-b", "api:/keep", "GET"); err != nil {
		t.Fatal(err)
	}
	want := model.Permission{Resource: "api:/items,a", Action: "GET"}
	if err := repository.UpdateRolePermission(context.Background(), "role-a", []model.Permission{want}); err != nil {
		t.Fatalf("UpdateRolePermission() error = %v", err)
	}
	permissions, err := repository.e.GetPermissionsForUser("role-a")
	if err != nil {
		t.Fatal(err)
	}
	if len(permissions) != 1 || permissions[0][1] != want.Resource || permissions[0][2] != want.Action {
		t.Fatalf("role-a permissions = %#v", permissions)
	}
	kept, err := repository.e.HasPermissionForUser("role-b", "api:/keep", "GET")
	if err != nil || !kept {
		t.Fatalf("role-b permission lost, kept=%v err=%v", kept, err)
	}
}

// TestPermissionReferenceLifecycle 验证权限引用的替换与删除：
// 替换后旧引用消失、新引用生效；删除后引用彻底移除。
func TestPermissionReferenceLifecycle(t *testing.T) {
	repository := newPermissionTestRepository(t)
	oldPermission := model.Permission{Resource: "menu:/old", Action: "read"}
	newPermission := model.Permission{Resource: "menu:/new", Action: "read"}
	if _, err := repository.e.AddPermissionForUser("role-a", oldPermission.Resource, oldPermission.Action); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := repository.ReplacePermissionReferences(ctx, map[model.Permission]model.Permission{
		oldPermission: newPermission,
	}); err != nil {
		t.Fatalf("ReplacePermissionReferences() error = %v", err)
	}
	if err := repository.ReloadPolicy(); err != nil {
		t.Fatal(err)
	}
	hasNew, _ := repository.e.HasPermissionForUser("role-a", newPermission.Resource, newPermission.Action)
	hasOld, _ := repository.e.HasPermissionForUser("role-a", oldPermission.Resource, oldPermission.Action)
	if !hasNew || hasOld {
		t.Fatalf("replacement result old=%v new=%v", hasOld, hasNew)
	}
	if err := repository.DeletePermissionReferences(ctx, []model.Permission{newPermission}); err != nil {
		t.Fatalf("DeletePermissionReferences() error = %v", err)
	}
	if err := repository.ReloadPolicy(); err != nil {
		t.Fatal(err)
	}
	hasNew, _ = repository.e.HasPermissionForUser("role-a", newPermission.Resource, newPermission.Action)
	if hasNew {
		t.Fatal("permission was not deleted")
	}
}

// TestApiPermissionIsUniqueAndReusableAfterDelete 验证 path+method 唯一约束：
// 重复创建报错，删除后同一路径方法可再次创建。
func TestApiPermissionIsUniqueAndReusableAfterDelete(t *testing.T) {
	repository := newPermissionTestRepository(t)
	first := &model.Api{Group: "test", Name: "first", Path: "/v1/items", Method: "GET"}
	if err := repository.ApiCreate(context.Background(), first); err != nil {
		t.Fatalf("ApiCreate() error = %v", err)
	}
	duplicate := &model.Api{Group: "test", Name: "duplicate", Path: "/v1/items", Method: "GET"}
	if err := repository.ApiCreate(context.Background(), duplicate); err == nil {
		t.Fatal("ApiCreate() accepted duplicate path and method")
	}
	if err := repository.ApiDelete(context.Background(), first.ID); err != nil {
		t.Fatalf("ApiDelete() error = %v", err)
	}
	replacement := &model.Api{Group: "test", Name: "replacement", Path: "/v1/items", Method: "GET"}
	if err := repository.ApiCreate(context.Background(), replacement); err != nil {
		t.Fatalf("ApiCreate() after delete error = %v", err)
	}
}

// TestApiListAndGroupQuery 验证 API 分页列表与分组查询的行扫描
// （回归用例：行结构漏映射会在运行时报 missing destination）。
func TestApiListAndGroupQuery(t *testing.T) {
	repository := newPermissionTestRepository(t)
	ctx := context.Background()
	if err := repository.ApiCreate(ctx, &model.Api{Group: "g1", Name: "n1", Path: "/v1/a", Method: "GET", MenuIDs: []uint{1}}); err != nil {
		t.Fatalf("seed api: %v", err)
	}
	list, total, err := repository.GetApis(ctx, &v1.GetApisRequest{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("GetApis() error = %v", err)
	}
	if total != 1 || len(list) != 1 || list[0].MenuIDs[0] != 1 {
		t.Fatalf("GetApis() = %d items, total %d, %#v", len(list), total, list)
	}
	groups, err := repository.GetApiGroups(ctx)
	if err != nil || len(groups) != 1 || groups[0] != "g1" {
		t.Fatalf("GetApiGroups() = %v, err %v", groups, err)
	}
}
