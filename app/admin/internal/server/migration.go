package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	v1 "nunu-monorepo/app/admin/api/v1"
	"nunu-monorepo/app/admin/internal/model"
	"nunu-monorepo/app/admin/internal/repository"
	"nunu-monorepo/pkg/log"
	"nunu-monorepo/pkg/sid"

	"github.com/casbin/casbin/v3"
	"github.com/jmoiron/sqlx"
	"github.com/samber/do/v2"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

// MigrateServer 是一次性迁移进程：破坏性重建业务表并灌入种子数据
// （管理员、角色、菜单、API 资源、RBAC 策略），完成后退出。
// 仅限本地重置/初始化流程使用，禁止指向有真实数据的库。
type MigrateServer struct {
	db     *sqlx.DB
	driver string
	log    *log.Logger
	sid    *sid.Sid
	e      *casbin.SyncedEnforcer
}

// NewMigrateServer 构造迁移服务，由注入容器调用。
func NewMigrateServer(i do.Injector) (*MigrateServer, error) {
	return &MigrateServer{
		db:     do.MustInvoke[*sqlx.DB](i),
		driver: repository.DriverFromConf(do.MustInvoke[*viper.Viper](i)),
		log:    do.MustInvoke[*log.Logger](i),
		sid:    do.MustInvoke[*sid.Sid](i),
		e:      do.MustInvoke[*casbin.SyncedEnforcer](i),
	}, nil
}

// Start 执行迁移：删表→建表→种子数据→RBAC 策略，成功后进程退出。
func (m *MigrateServer) Start(ctx context.Context) error {
	// 破坏性重建：先删旧表再按方言建表（casbin_rule 由适配器保证存在）
	if err := repository.DropAdminTables(ctx, m.db, m.driver); err != nil {
		m.log.Error().Err(err).Msg("drop tables error")
		return err
	}
	if err := repository.CreateAdminTables(ctx, m.db, m.driver); err != nil {
		m.log.Error().Err(err).Msg("create tables error")
		return err
	}
	if err := m.initialAdminUser(ctx); err != nil {
		m.log.Error().Err(err).Msg("initialAdminUser error")
		return err
	}
	if err := m.initialMenuData(ctx); err != nil {
		m.log.Error().Err(err).Msg("initialMenuData error")
		return err
	}
	if err := m.initialApisData(ctx); err != nil {
		m.log.Error().Err(err).Msg("initialApisData error")
		return err
	}
	if err := m.initialRBAC(ctx); err != nil {
		m.log.Error().Err(err).Msg("initialRBAC error")
		return err
	}
	m.log.Info().Msg("AutoMigrate success")
	os.Exit(0)
	return nil
}

// Stop 迁移是一次性进程，正常路径不会走到 Stop。
func (m *MigrateServer) Stop(ctx context.Context) error {
	m.log.Info().Msg("AutoMigrate stop")
	return nil
}

// initialAdminUser 写入两个内置账号：admin（超管）与 user（运营），
// 密码统一为 123456 的 bcrypt 哈希。
func (m *MigrateServer) initialAdminUser(ctx context.Context) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash seed password: %w", err)
	}
	query := repository.Rebind(m.driver,
		"INSERT INTO admin_users (id, username, nickname, password, email, phone) VALUES (?, ?, ?, ?, ?, ?)")
	users := []model.AdminUser{
		{Base: model.Base{ID: 1}, Username: "admin", Password: string(hashedPassword), Nickname: "Admin"},
		{Base: model.Base{ID: 2}, Username: "user", Password: string(hashedPassword), Nickname: "运营人员"},
	}
	for _, user := range users {
		if _, err := m.db.ExecContext(ctx, query,
			user.ID, user.Username, user.Nickname, user.Password, user.Email, user.Phone); err != nil {
			return fmt.Errorf("seed admin user %s: %w", user.Username, err)
		}
	}
	return nil
}

// initialRBAC 建立角色、用户-角色绑定与各角色的授权策略：
// admin 角色获得全部菜单与接口权限；运营角色获得仪表盘与基础接口权限。
func (m *MigrateServer) initialRBAC(ctx context.Context) error {
	// 角色种子：Sid 即 Casbin 中的角色名
	roles := []model.Role{
		{Sid: model.AdminRole, Name: "超级管理员"},
		{Sid: "1000", Name: "运营人员"},
		{Sid: "1001", Name: "访客"},
	}
	query := repository.Rebind(m.driver, "INSERT INTO roles (name, sid) VALUES (?, ?)")
	for _, role := range roles {
		if _, err := m.db.ExecContext(ctx, query, role.Name, role.Sid); err != nil {
			return fmt.Errorf("seed role %s: %w", role.Sid, err)
		}
	}

	// 清空策略表后重建（SavePolicy 以内存策略全量覆盖存储）
	m.e.ClearPolicy()
	if err := m.e.SavePolicy(); err != nil {
		m.log.Error().Err(err).Msg("m.e.SavePolicy error")
		return err
	}
	if _, err := m.e.AddRoleForUser(model.AdminUserID, model.AdminRole); err != nil {
		m.log.Error().Err(err).Msg("m.e.AddRoleForUser error")
		return err
	}

	// admin 角色拥有全部种子菜单的 read 权限
	menuList := make([]v1.MenuDataItem, 0)
	if err := json.Unmarshal([]byte(menuData), &menuList); err != nil {
		m.log.Error().Err(err).Msg("json.Unmarshal error")
		return err
	}
	menuList = filterSeedMenus(menuList)
	seedMenuMap := make(map[uint]v1.MenuDataItem, len(menuList))
	for _, item := range menuList {
		seedMenuMap[item.ID] = item
	}
	for _, item := range menuList {
		m.addPermissionForRole(model.AdminRole, model.MenuResourcePrefix+fullSeedMenuPath(item, seedMenuMap), "read")
	}

	// admin 角色拥有全部已注册接口权限
	// （迁移进程只需路径与方法两列，用本地最小行结构扫描，不依赖领域实体的存储形态）
	apiList := make([]apiSeedRow, 0)
	apiQuery := repository.Rebind(m.driver, "SELECT path, method FROM api ORDER BY id ASC")
	if err := m.db.SelectContext(ctx, &apiList, apiQuery); err != nil {
		m.log.Error().Err(err).Msg("m.db.SelectContext(apiList) error")
		return err
	}
	for _, api := range apiList {
		m.addPermissionForRole(model.AdminRole, model.ApiResourcePrefix+api.Path, api.Method)
	}

	// 运营人员：绑定角色并授予仪表盘与基础接口权限
	if _, err := m.e.AddRoleForUser("2", "1000"); err != nil {
		m.log.Error().Err(err).Msg("m.e.AddRoleForUser error")
		return err
	}
	m.addPermissionForRole("1000", model.MenuResourcePrefix+"/dashboard", "read")
	m.addPermissionForRole("1000", model.MenuResourcePrefix+"/dashboard/console", "read")
	m.addPermissionForRole("1000", model.MenuResourcePrefix+"/dashboard/analysis", "read")
	m.addPermissionForRole("1000", model.ApiResourcePrefix+"/v1/menus", http.MethodGet)
	m.addPermissionForRole("1000", model.ApiResourcePrefix+"/v1/admin/user", http.MethodGet)
	return nil
}

// apiSeedRow 是迁移进程读取接口资源的最小行结构（仅路径与方法两列）。
type apiSeedRow struct {
	Path   string `db:"path"`
	Method string `db:"method"`
}

// fullSeedMenuPath 计算种子菜单的完整路由路径：绝对路径原样返回，
// 相对路径逐级拼接父级路径。
func fullSeedMenuPath(item v1.MenuDataItem, menuMap map[uint]v1.MenuDataItem) string {
	if item.Path == "" {
		return ""
	}
	if strings.HasPrefix(item.Path, "http://") ||
		strings.HasPrefix(item.Path, "https://") ||
		strings.HasPrefix(item.Path, "/") {
		return item.Path
	}
	if item.ParentID == 0 {
		return "/" + strings.TrimPrefix(item.Path, "/")
	}
	parent, ok := menuMap[item.ParentID]
	if !ok {
		return "/" + strings.TrimPrefix(item.Path, "/")
	}
	parentPath := strings.TrimRight(fullSeedMenuPath(parent, menuMap), "/")
	childPath := strings.TrimLeft(item.Path, "/")
	if parentPath == "" {
		return "/" + childPath
	}
	return parentPath + "/" + childPath
}

// addPermissionForRole 为角色追加一条权限并打印结果；失败不中断迁移，
// 失败项在控制台日志中可人工核对。
func (m *MigrateServer) addPermissionForRole(role, resource, action string) {
	if _, err := m.e.AddPermissionForUser(role, resource, action); err != nil {
		m.log.Info().Msgf("为角色 %s 添加权限 %s:%s 失败: %v", role, resource, action, err)
		return
	}
	fmt.Printf("为角色 %s 添加权限: %s %s\n", role, resource, action)
}

// initialApisData 写入受 RBAC 管控的接口资源注册项；menu_ids 记录
// 接口与菜单的关联，供"勾选菜单自动带出接口权限"使用。
func (m *MigrateServer) initialApisData(ctx context.Context) error {
	initialApis := []model.Api{
		{Group: "基础API", Name: "获取用户菜单列表", Path: "/v1/menus", Method: http.MethodGet},
		{Group: "基础API", Name: "获取管理员信息", Path: "/v1/admin/user", Method: http.MethodGet, MenuIDs: []uint{60}},
		{Group: "权限管理/菜单", Name: "获取管理菜单", Path: "/v1/admin/menus", Method: http.MethodGet, MenuIDs: []uint{62, 63, 64}},
		{Group: "权限管理/菜单", Name: "创建菜单", Path: "/v1/admin/menu", Method: http.MethodPost, MenuIDs: []uint{63}},
		{Group: "权限管理/菜单", Name: "更新菜单", Path: "/v1/admin/menu", Method: http.MethodPut, MenuIDs: []uint{63}},
		{Group: "权限管理/菜单", Name: "删除菜单", Path: "/v1/admin/menu", Method: http.MethodDelete, MenuIDs: []uint{63}},
		{Group: "权限管理/角色", Name: "获取用户权限", Path: "/v1/admin/user/permissions", Method: http.MethodGet, MenuIDs: []uint{62}},
		{Group: "权限管理/角色", Name: "获取角色权限", Path: "/v1/admin/role/permissions", Method: http.MethodGet, MenuIDs: []uint{62}},
		{Group: "权限管理/角色", Name: "更新角色权限", Path: "/v1/admin/role/permissions", Method: http.MethodPut, MenuIDs: []uint{62}},
		{Group: "权限管理/角色", Name: "获取角色列表", Path: "/v1/admin/roles", Method: http.MethodGet, MenuIDs: []uint{62}},
		{Group: "权限管理/角色", Name: "创建角色", Path: "/v1/admin/role", Method: http.MethodPost, MenuIDs: []uint{62}},
		{Group: "权限管理/角色", Name: "更新角色", Path: "/v1/admin/role", Method: http.MethodPut, MenuIDs: []uint{62}},
		{Group: "权限管理/角色", Name: "删除角色", Path: "/v1/admin/role", Method: http.MethodDelete, MenuIDs: []uint{62}},
		{Group: "权限管理/用户", Name: "获取管理员列表", Path: "/v1/admin/users", Method: http.MethodGet, MenuIDs: []uint{61}},
		{Group: "权限管理/用户", Name: "更新管理员信息", Path: "/v1/admin/user", Method: http.MethodPut, MenuIDs: []uint{61}},
		{Group: "权限管理/用户", Name: "创建管理员账号", Path: "/v1/admin/user", Method: http.MethodPost, MenuIDs: []uint{61}},
		{Group: "权限管理/用户", Name: "删除管理员", Path: "/v1/admin/user", Method: http.MethodDelete, MenuIDs: []uint{61}},
		{Group: "权限管理/接口", Name: "获取API列表", Path: "/v1/admin/apis", Method: http.MethodGet, MenuIDs: []uint{62, 64}},
		{Group: "权限管理/接口", Name: "创建API", Path: "/v1/admin/api", Method: http.MethodPost, MenuIDs: []uint{64}},
		{Group: "权限管理/接口", Name: "更新API", Path: "/v1/admin/api", Method: http.MethodPut, MenuIDs: []uint{64}},
		{Group: "权限管理/接口", Name: "删除API", Path: "/v1/admin/api", Method: http.MethodDelete, MenuIDs: []uint{64}},
	}
	groupColumn := repository.QuoteIdentifier(m.driver, "group")
	query := repository.Rebind(m.driver,
		"INSERT INTO api ("+groupColumn+", name, path, method, menu_ids) VALUES (?, ?, ?, ?, ?)")
	for _, api := range initialApis {
		if _, err := m.db.ExecContext(ctx, query, api.Group, api.Name, api.Path, api.Method, repository.UintSlice(api.MenuIDs)); err != nil {
			return fmt.Errorf("seed api %s %s: %w", api.Method, api.Path, err)
		}
	}
	return nil
}

// initialMenuData 按种子 JSON 写入菜单树（保留原 ID，供接口种子引用）。
func (m *MigrateServer) initialMenuData(ctx context.Context) error {
	menuList := make([]v1.MenuDataItem, 0)
	if err := json.Unmarshal([]byte(menuData), &menuList); err != nil {
		m.log.Error().Err(err).Msg("json.Unmarshal error")
		return err
	}
	menuList = filterSeedMenus(menuList)
	query := repository.Rebind(m.driver, `INSERT INTO menu (
id, parent_id, path, title, name, component, locale, icon, redirect, url, link, target,
active_path, show_text_badge, weight, is_enable, is_menu, keep_alive, hide_in_menu, is_hide,
is_hide_tab, is_iframe, show_badge, fixed_tab, is_full_page, roles, auth_list
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	for _, item := range menuList {
		menu := model.Menu{
			Base:          model.Base{ID: item.ID},
			ParentID:      item.ParentID,
			Path:          item.Path,
			Title:         item.Title,
			Name:          item.Name,
			Component:     item.Component,
			Locale:        item.Locale,
			Weight:        item.Weight,
			Icon:          item.Icon,
			Redirect:      item.Redirect,
			URL:           item.URL,
			Link:          item.Link,
			Target:        item.Target,
			ActivePath:    item.ActivePath,
			ShowTextBadge: item.ShowTextBadge,
			KeepAlive:     item.KeepAlive,
			HideInMenu:    item.HideInMenu || item.IsHide,
			IsHide:        item.HideInMenu || item.IsHide,
			IsHideTab:     item.IsHideTab,
			IsIframe:      item.IsIframe,
			ShowBadge:     item.ShowBadge,
			FixedTab:      item.FixedTab,
			IsFullPage:    item.IsFullPage,
			Roles:         item.Roles,
			AuthList:      menuAuthListFromSeed(item.AuthList),
			IsEnable:      true,
			IsMenu:        true,
		}
		args := []any{
			menu.ID, menu.ParentID, menu.Path, menu.Title, menu.Name, menu.Component,
			menu.Locale, menu.Icon, menu.Redirect, menu.URL, menu.Link, menu.Target,
			menu.ActivePath, menu.ShowTextBadge, menu.Weight, menu.IsEnable, menu.IsMenu,
			menu.KeepAlive, menu.HideInMenu, menu.IsHide, menu.IsHideTab, menu.IsIframe,
			menu.ShowBadge, menu.FixedTab, menu.IsFullPage,
			repository.StringSlice(menu.Roles), repository.MenuAuths(menu.AuthList),
		}
		if _, err := m.db.ExecContext(ctx, query, args...); err != nil {
			return fmt.Errorf("seed menu %d: %w", item.ID, err)
		}
	}
	return nil
}

// menuAuthListFromSeed 把种子 JSON 的按钮权限列表转换为实体类型。
func menuAuthListFromSeed(list []v1.MenuAuthDataItem) []model.MenuAuth {
	authList := make([]model.MenuAuth, 0, len(list))
	for _, item := range list {
		authList = append(authList, model.MenuAuth{
			Title:    item.Title,
			AuthMark: item.AuthMark,
		})
	}
	return authList
}

// filterSeedMenus 是种子菜单的过滤钩子：当前原样返回全部菜单，
// 需要裁剪演示菜单时在此处实现。
func filterSeedMenus(menuList []v1.MenuDataItem) []v1.MenuDataItem {
	return menuList
}

// menuData 是内嵌的菜单种子数据（与前端路由配置对应）。
var menuData = `[
  {"id":1,"parentId":0,"path":"/dashboard","name":"Dashboard","title":"menus.dashboard.title","component":"/index/index","icon":"ri:pie-chart-line","roles":["admin","R_SUPER","R_ADMIN"],"weight":100},
  {"id":2,"parentId":1,"path":"console","name":"Console","title":"menus.dashboard.console","component":"/dashboard/console","icon":"ri:home-smile-2-line","fixedTab":true,"weight":100},
  {"id":3,"parentId":1,"path":"analysis","name":"Analysis","title":"menus.dashboard.analysis","component":"/dashboard/analysis","icon":"ri:align-item-bottom-line","weight":90},
  {"id":4,"parentId":1,"path":"ecommerce","name":"Ecommerce","title":"menus.dashboard.ecommerce","component":"/dashboard/ecommerce","icon":"ri:bar-chart-box-line","weight":80},
  {"id":10,"parentId":0,"path":"/template","name":"Template","title":"menus.template.title","component":"/index/index","icon":"ri:apps-2-line","roles":["admin"],"weight":90},
  {"id":11,"parentId":10,"path":"cards","name":"Cards","title":"menus.template.cards","component":"/template/cards","icon":"ri:wallet-line","weight":100},
  {"id":12,"parentId":10,"path":"banners","name":"Banners","title":"menus.template.banners","component":"/template/banners","icon":"ri:rectangle-line","weight":90},
  {"id":13,"parentId":10,"path":"charts","name":"Charts","title":"menus.template.charts","component":"/template/charts","icon":"ri:bar-chart-box-line","weight":80},
  {"id":14,"parentId":10,"path":"map","name":"Map","title":"menus.template.map","component":"/template/map","icon":"ri:map-pin-line","keepAlive":true,"weight":70},
  {"id":15,"parentId":10,"path":"chat","name":"Chat","title":"menus.template.chat","component":"/template/chat","icon":"ri:message-3-line","keepAlive":true,"weight":60},
  {"id":16,"parentId":10,"path":"calendar","name":"Calendar","title":"menus.template.calendar","component":"/template/calendar","icon":"ri:calendar-2-line","keepAlive":true,"weight":50},
  {"id":17,"parentId":10,"path":"pricing","name":"Pricing","title":"menus.template.pricing","component":"/template/pricing","icon":"ri:money-cny-box-line","keepAlive":true,"isFullPage":true,"weight":40},
  {"id":20,"parentId":0,"path":"/widgets","name":"Widgets","title":"menus.widgets.title","component":"/index/index","icon":"ri:apps-2-add-line","roles":["admin"],"weight":80},
  {"id":21,"parentId":20,"path":"icon","name":"Icon","title":"menus.widgets.icon","component":"/widgets/icon","icon":"ri:palette-line","keepAlive":true,"weight":130},
  {"id":22,"parentId":20,"path":"image-crop","name":"ImageCrop","title":"menus.widgets.imageCrop","component":"/widgets/image-crop","icon":"ri:screenshot-line","keepAlive":true,"weight":120},
  {"id":23,"parentId":20,"path":"excel","name":"Excel","title":"menus.widgets.excel","component":"/widgets/excel","icon":"ri:download-2-line","keepAlive":true,"weight":110},
  {"id":24,"parentId":20,"path":"video","name":"Video","title":"menus.widgets.video","component":"/widgets/video","icon":"ri:vidicon-line","keepAlive":true,"weight":100},
  {"id":25,"parentId":20,"path":"count-to","name":"CountTo","title":"menus.widgets.countTo","component":"/widgets/count-to","icon":"ri:anthropic-line","weight":90},
  {"id":26,"parentId":20,"path":"wang-editor","name":"WangEditor","title":"menus.widgets.wangEditor","component":"/widgets/wang-editor","icon":"ri:t-box-line","keepAlive":true,"weight":80},
  {"id":27,"parentId":20,"path":"watermark","name":"Watermark","title":"menus.widgets.watermark","component":"/widgets/watermark","icon":"ri:water-flash-line","keepAlive":true,"weight":70},
  {"id":28,"parentId":20,"path":"context-menu","name":"ContextMenu","title":"menus.widgets.contextMenu","component":"/widgets/context-menu","icon":"ri:menu-2-line","keepAlive":true,"weight":60},
  {"id":29,"parentId":20,"path":"qrcode","name":"Qrcode","title":"menus.widgets.qrcode","component":"/widgets/qrcode","icon":"ri:qr-code-line","keepAlive":true,"weight":50},
  {"id":30,"parentId":20,"path":"drag","name":"Drag","title":"menus.widgets.drag","component":"/widgets/drag","icon":"ri:drag-move-fill","keepAlive":true,"weight":40},
  {"id":31,"parentId":20,"path":"text-scroll","name":"TextScroll","title":"menus.widgets.textScroll","component":"/widgets/text-scroll","icon":"ri:input-method-line","keepAlive":true,"weight":30},
  {"id":32,"parentId":20,"path":"fireworks","name":"Fireworks","title":"menus.widgets.fireworks","component":"/widgets/fireworks","icon":"ri:magic-line","keepAlive":true,"showTextBadge":"Hot","weight":20},
  {"id":33,"parentId":20,"path":"/outside/iframe/elementui","name":"ElementUI","title":"menus.widgets.elementUI","component":"","icon":"ri:apps-2-line","link":"https://element-plus.org/zh-CN/component/overview.html","isIframe":true,"weight":10},
  {"id":40,"parentId":0,"path":"/examples","name":"Examples","title":"menus.examples.title","component":"/index/index","icon":"ri:sparkling-line","roles":["admin"],"weight":70},
  {"id":41,"parentId":40,"path":"permission","name":"Permission","title":"menus.examples.permission.title","component":"","icon":"ri:fingerprint-line","weight":100},
  {"id":42,"parentId":41,"path":"switch-role","name":"PermissionSwitchRole","title":"menus.examples.permission.switchRole","component":"/examples/permission/switch-role","icon":"ri:contacts-line","keepAlive":true,"weight":100},
  {"id":43,"parentId":41,"path":"button-auth","name":"PermissionButtonAuth","title":"menus.examples.permission.buttonAuth","component":"/examples/permission/button-auth","icon":"ri:mouse-line","keepAlive":true,"authList":[{"title":"新增","authMark":"add"},{"title":"编辑","authMark":"edit"},{"title":"删除","authMark":"delete"},{"title":"导出","authMark":"export"},{"title":"查看","authMark":"view"},{"title":"发布","authMark":"publish"},{"title":"配置","authMark":"config"},{"title":"管理","authMark":"manage"}],"weight":90},
  {"id":44,"parentId":41,"path":"page-visibility","name":"PermissionPageVisibility","title":"menus.examples.permission.pageVisibility","component":"/examples/permission/page-visibility","icon":"ri:user-3-line","keepAlive":true,"roles":["admin","R_SUPER"],"weight":80},
  {"id":45,"parentId":40,"path":"tabs","name":"Tabs","title":"menus.examples.tabs","component":"/examples/tabs","icon":"ri:price-tag-line","weight":90},
  {"id":46,"parentId":40,"path":"tables/basic","name":"TablesBasic","title":"menus.examples.tablesBasic","component":"/examples/tables/basic","icon":"ri:layout-grid-line","keepAlive":true,"weight":80},
  {"id":47,"parentId":40,"path":"tables","name":"Tables","title":"menus.examples.tables","component":"/examples/tables","icon":"ri:table-3","keepAlive":true,"weight":70},
  {"id":48,"parentId":40,"path":"forms","name":"Forms","title":"menus.examples.forms","component":"/examples/forms","icon":"ri:table-view","keepAlive":true,"weight":60},
  {"id":49,"parentId":40,"path":"form/search-bar","name":"SearchBar","title":"menus.examples.searchBar","component":"/examples/forms/search-bar","icon":"ri:table-line","keepAlive":true,"weight":50},
  {"id":50,"parentId":40,"path":"tables/tree","name":"TablesTree","title":"menus.examples.tablesTree","component":"/examples/tables/tree","icon":"ri:layout-2-line","keepAlive":true,"weight":40},
  {"id":51,"parentId":40,"path":"socket-chat","name":"SocketChat","title":"menus.examples.socketChat","component":"/examples/socket-chat","icon":"ri:shake-hands-line","keepAlive":true,"weight":30},
  {"id":60,"parentId":0,"path":"/admin","name":"AdminManage","title":"权限管理","component":"/index/index","icon":"ri:shield-user-line","roles":["admin"],"weight":60},
  {"id":61,"parentId":60,"path":"user","name":"AdminUser","title":"用户管理","component":"/admin/user","icon":"ri:user-line","keepAlive":true,"weight":100},
  {"id":62,"parentId":60,"path":"role","name":"AdminRole","title":"角色管理","component":"/admin/role","icon":"ri:user-settings-line","keepAlive":true,"weight":90},
  {"id":63,"parentId":60,"path":"menu","name":"AdminMenu","title":"菜单管理","component":"/admin/menu","icon":"ri:menu-line","keepAlive":true,"weight":80},
  {"id":64,"parentId":60,"path":"api","name":"AdminApi","title":"接口管理","component":"/admin/api","icon":"ri:terminal-window-line","keepAlive":true,"weight":70},
  {"id":70,"parentId":0,"path":"/system","name":"System","title":"menus.system.title","component":"/index/index","icon":"ri:user-3-line","roles":["admin","R_SUPER","R_ADMIN"],"weight":50},
  {"id":71,"parentId":70,"path":"user","name":"User","title":"menus.system.user","component":"/system/user","icon":"ri:user-line","keepAlive":true,"roles":["admin","R_SUPER","R_ADMIN"],"weight":100},
  {"id":72,"parentId":70,"path":"role","name":"Role","title":"menus.system.role","component":"/system/role","icon":"ri:user-settings-line","keepAlive":true,"roles":["admin","R_SUPER"],"weight":90},
  {"id":73,"parentId":70,"path":"user-center","name":"UserCenter","title":"menus.system.userCenter","component":"/system/user-center","icon":"ri:user-line","keepAlive":true,"isHide":true,"isHideTab":true,"weight":80},
  {"id":74,"parentId":70,"path":"menu","name":"Menus","title":"menus.system.menu","component":"/system/menu","icon":"ri:menu-line","keepAlive":true,"roles":["admin","R_SUPER"],"authList":[{"title":"新增","authMark":"add"},{"title":"编辑","authMark":"edit"},{"title":"删除","authMark":"delete"}],"weight":70},
  {"id":75,"parentId":70,"path":"nested","name":"Nested","title":"menus.system.nested","component":"","icon":"ri:menu-unfold-3-line","keepAlive":true,"weight":60},
  {"id":76,"parentId":75,"path":"menu1","name":"NestedMenu1","title":"menus.system.menu1","component":"/system/nested/menu1","icon":"ri:align-justify","keepAlive":true,"weight":100},
  {"id":77,"parentId":75,"path":"menu2","name":"NestedMenu2","title":"menus.system.menu2","component":"","icon":"ri:align-justify","keepAlive":true,"weight":90},
  {"id":78,"parentId":77,"path":"menu2-1","name":"NestedMenu2-1","title":"menus.system.menu21","component":"/system/nested/menu2","icon":"ri:align-justify","keepAlive":true,"weight":100},
  {"id":79,"parentId":75,"path":"menu3","name":"NestedMenu3","title":"menus.system.menu3","component":"","icon":"ri:align-justify","keepAlive":true,"weight":80},
  {"id":80,"parentId":79,"path":"menu3-1","name":"NestedMenu3-1","title":"menus.system.menu31","component":"/system/nested/menu3","keepAlive":true,"weight":100},
  {"id":81,"parentId":79,"path":"menu3-2","name":"NestedMenu3-2","title":"menus.system.menu32","component":"","keepAlive":true,"weight":90},
  {"id":82,"parentId":81,"path":"menu3-2-1","name":"NestedMenu3-2-1","title":"menus.system.menu321","component":"/system/nested/menu3/menu3-2","keepAlive":true,"weight":100},
  {"id":90,"parentId":0,"path":"/article","name":"Article","title":"menus.article.title","component":"/index/index","icon":"ri:book-2-line","roles":["admin","R_SUPER","R_ADMIN"],"weight":40},
  {"id":91,"parentId":90,"path":"article-list","name":"ArticleList","title":"menus.article.articleList","component":"/article/list","icon":"ri:article-line","keepAlive":true,"authList":[{"title":"新增","authMark":"add"},{"title":"编辑","authMark":"edit"}],"weight":100},
  {"id":92,"parentId":90,"path":"detail/:id","name":"ArticleDetail","title":"menus.article.articleDetail","component":"/article/detail","keepAlive":true,"isHide":true,"activePath":"/article/article-list","weight":90},
  {"id":93,"parentId":90,"path":"comment","name":"ArticleComment","title":"menus.article.comment","component":"/article/comment","icon":"ri:mail-line","keepAlive":true,"weight":80},
  {"id":94,"parentId":90,"path":"publish","name":"ArticlePublish","title":"menus.article.articlePublish","component":"/article/publish","icon":"ri:telegram-2-line","keepAlive":true,"authList":[{"title":"发布","authMark":"add"}],"weight":70},
  {"id":100,"parentId":0,"path":"/result","name":"Result","title":"menus.result.title","component":"/index/index","icon":"ri:checkbox-circle-line","roles":["admin"],"weight":30},
  {"id":101,"parentId":100,"path":"success","name":"ResultSuccess","title":"menus.result.success","component":"/result/success","icon":"ri:checkbox-circle-line","keepAlive":true,"weight":100},
  {"id":102,"parentId":100,"path":"fail","name":"ResultFail","title":"menus.result.fail","component":"/result/fail","icon":"ri:close-circle-line","keepAlive":true,"weight":90},
  {"id":110,"parentId":0,"path":"/exception","name":"Exception","title":"menus.exception.title","component":"/index/index","icon":"ri:error-warning-line","roles":["admin"],"weight":20},
  {"id":111,"parentId":110,"path":"403","name":"Exception403","title":"menus.exception.forbidden","component":"/exception/403","keepAlive":true,"isHideTab":true,"isFullPage":true,"weight":100},
  {"id":112,"parentId":110,"path":"404","name":"Exception404","title":"menus.exception.notFound","component":"/exception/404","keepAlive":true,"isHideTab":true,"isFullPage":true,"weight":90},
  {"id":113,"parentId":110,"path":"500","name":"Exception500","title":"menus.exception.serverError","component":"/exception/500","keepAlive":true,"isHideTab":true,"isFullPage":true,"weight":80},
  {"id":120,"parentId":0,"path":"/safeguard","name":"Safeguard","title":"menus.safeguard.title","component":"/index/index","icon":"ri:shield-check-line","roles":["admin"],"weight":10},
  {"id":121,"parentId":120,"path":"server","name":"SafeguardServer","title":"menus.safeguard.server","component":"/safeguard/server","icon":"ri:hard-drive-3-line","keepAlive":true,"weight":100},
  {"id":130,"parentId":0,"path":"/change/log","name":"ChangeLog","title":"menus.plan.log","component":"/change/log","icon":"ri:gamepad-line","showTextBadge":"v3.0.2","weight":1}
]`
