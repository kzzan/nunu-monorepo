import { App } from './App'
import { registerRoutes, startRouter } from './router/router'
import { LoginView } from './views/LoginView'
import { DashboardView } from './views/DashboardView'
import { UserListView } from './views/users/UserListView'
import { RoleListView } from './views/roles/RoleListView'
import { MenuListView } from './views/menus/MenuListView'
import { ApiListView } from './views/apis/ApiListView'
import { NotFoundView } from './views/NotFoundView'

// Base document reset previously provided by Pico CSS, now applied inline (no CSS files).
const bodyStyle = document.body.style
bodyStyle.margin = '0'
bodyStyle.minHeight = '100vh'
bodyStyle.lineHeight = '1.5'
bodyStyle.fontFamily =
  'system-ui, -apple-system, "Segoe UI", Roboto, "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", sans-serif'

registerRoutes({
  '/login': LoginView,
  '/dashboard/console': DashboardView,
  '/admin/user': UserListView,
  '/admin/role': RoleListView,
  '/admin/menu': MenuListView,
  '/admin/api': ApiListView,
  '*': NotFoundView
})

const mount = document.querySelector<HTMLDivElement>('#app')
if (mount) {
  mount.appendChild(App())
  startRouter()
} else {
  document.body.appendChild(App())
  startRouter()
}
