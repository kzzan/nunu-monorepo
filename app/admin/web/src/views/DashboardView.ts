import van from 'vanjs-core'
import { getAdminUsers, getRoles, getAdminMenus, getApis } from '../api/admin'
import { currentUserView } from '../store/auth'
import { C } from '../ui/kit'

const { div, h2, p, section, span } = van.tags

interface Stat {
  label: string
  value: () => string
}

function statCard(stat: Stat): HTMLElement {
  return div(
    {
      style: `background: #fff; border: 1px solid ${C.border}; border-radius: 10px; padding: 18px 19px; display: flex; flex-direction: column; gap: 5px;`
    },
    span({ style: `font-size: 1.8rem; font-weight: 700; color: ${C.brand};` }, stat.value),
    span({ style: `color: ${C.textSecondary}; font-size: 0.88rem;` }, stat.label)
  )
}

/** Dashboard: welcome + resource stat cards from real API totals. */
export function DashboardView(): HTMLElement {
  const userTotal = van.state('–')
  const roleTotal = van.state('–')
  const menuTotal = van.state('–')
  const apiTotal = van.state('–')

  void getAdminUsers({ page: 1, pageSize: 1 })
    .then((d) => (userTotal.val = String(d.total)))
    .catch(() => (userTotal.val = '—'))
  void getRoles({ page: 1, pageSize: 1 })
    .then((d) => (roleTotal.val = String(d.total)))
    .catch(() => (roleTotal.val = '—'))
  void getAdminMenus()
    .then((d) => (menuTotal.val = String(d.list?.length ?? 0)))
    .catch(() => (menuTotal.val = '—'))
  void getApis({ page: 1, pageSize: 1 })
    .then((d) => (apiTotal.val = String(d.total)))
    .catch(() => (apiTotal.val = '—'))

  const stats: Stat[] = [
    { label: '管理员', value: () => userTotal.val },
    { label: '角色', value: () => roleTotal.val },
    { label: '菜单', value: () => menuTotal.val },
    { label: 'API 接口', value: () => apiTotal.val }
  ]

  return section(
    { style: 'display: flex; flex-direction: column; gap: 20px;' },
    div(
      {},
      h2({ style: 'margin: 0 0 4px;' }, () => `你好，${currentUserView().val?.nickname ?? '管理员'}`),
      p({ style: `color: ${C.textSecondary}; margin: 0;` }, '欢迎使用 Nunu Admin 管理控制台（VanJS + VanUI + Vite）')
    ),
    div(
      { style: 'display: grid; grid-template-columns: repeat(auto-fit, minmax(170px, 1fr)); gap: 16px;' },
      stats.map(statCard)
    ),
    div(
      {
        style: `background: ${C.bgSoft}; border: 1px dashed ${C.border}; border-radius: 10px; padding: 14px 19px; color: ${C.textSecondary}; font-size: 0.88rem;`
      },
      p(
        { style: 'margin: 0;' },
        '提示：左侧菜单由后端 /v1/menus 接口按角色权限动态生成；',
        '用户、角色、菜单、接口管理支持完整的增删改查与权限分配。'
      )
    )
  )
}
