import van from 'vanjs-core'
import { currentPath, navigate, HOME_PATH, resolveView } from '../router/router'
import { Sidebar } from './Sidebar'
import { currentUserView, clearToken, setCurrentUser, getToken } from '../store/auth'
import { fetchMenus, resetMenus } from '../store/menu'
import { getCurrentAdminUser } from '../api/admin'
import { C, btn } from '../ui/kit'

const { div, aside, header, main, h1, span } = van.tags

let bootstrapped = false

// Responsive breakpoint (mirrors the old CSS @media (max-width: 860px)) via JS, since no CSS is used.
const mql = window.matchMedia('(max-width: 860px)')
const compact = van.state(mql.matches)
mql.addEventListener('change', (e) => {
  compact.val = e.matches
})

/** Fetch menus + current user once per login session. */
function bootstrap(): void {
  if (bootstrapped || getToken() === '') return
  bootstrapped = true
  void fetchMenus().catch(() => {
    bootstrapped = false
  })
  void getCurrentAdminUser()
    .then(setCurrentUser)
    .catch(() => setCurrentUser(null))
}

function routeTitle(path: string): string {
  if (path === HOME_PATH) return '仪表盘'
  if (path.startsWith('/admin/')) {
    const seg = path.slice('/admin/'.length)
    if (seg === 'user') return '用户管理'
    if (seg === 'role') return '角色管理'
    if (seg === 'menu') return '菜单管理'
    if (seg === 'api') return '接口管理'
  }
  return '管理后台'
}

function logout(): void {
  clearToken()
  resetMenus()
  bootstrapped = false
  navigate('/login')
}

/** Application shell: sidebar + header + routed main content. */
export function AppLayout(): HTMLElement {
  bootstrap()

  return div(
    { style: () => (compact.val ? 'display: flex; min-height: 100vh; flex-direction: column;' : 'display: flex; min-height: 100vh;') },
    aside(
      {
        style: () =>
          compact.val
            ? `width: 100%; background: ${C.bgAside}; border-bottom: 1px solid ${C.border};`
            : `width: 232px; flex: 0 0 232px; background: ${C.bgAside}; border-right: 1px solid ${C.border}; display: flex; flex-direction: column; position: sticky; top: 0; height: 100vh; overflow-y: auto;`
      },
      div(
        { style: `display: flex; align-items: center; gap: 10px; padding: 16px 18px; font-weight: 700; border-bottom: 1px solid ${C.border};` },
        span(
          {
            style: `display: inline-flex; align-items: center; justify-content: center; width: 30px; height: 30px; border-radius: 8px; background: ${C.brand}; color: #fff; font-size: 1rem;`
          },
          'N'
        ),
        span({ style: 'font-size: 1.02rem; letter-spacing: 0.02em;' }, 'Nunu Admin')
      ),
      Sidebar()
    ),
    div(
      { style: 'flex: 1; min-width: 0; display: flex; flex-direction: column;' },
      header(
        { style: `display: flex; align-items: center; justify-content: space-between; padding: 10px 22px; background: #fff; border-bottom: 1px solid ${C.border};` },
        h1({ style: 'margin: 0; font-size: 1.05rem;' }, () => routeTitle(currentPath.val)),
        div(
          { style: 'display: flex; align-items: center; gap: 12px;' },
          span(
            { style: `color: ${C.textSecondary}; font-size: 0.92rem;` },
            () => currentUserView().val?.nickname ?? '管理员'
          ),
          btn('secondary', { small: true, onclick: logout }, '退出登录')
        )
      ),
      main(
        { style: 'flex: 1; padding: 22px;', id: 'main' },
        // reactive outlet: re-resolves the view whenever the route changes
        () => resolveView(currentPath.val)()
      )
    )
  )
}
