import van from 'vanjs-core'
import type { RoleItem, MenuItem, ApiItem } from '../../api/types'
import { getRoles, createRole, updateRole, deleteRole, getRolePermissions, updateRolePermissions, getAdminMenus, getApis } from '../../api/admin'
import { openModal } from '../../components/Modal'
import { confirmDialog } from '../../components/Confirm'
import { Pagination } from '../../components/Pagination'
import { toast } from '../../components/Toast'
import { buildMenuTree, flattenTree, resolveTitle, fullPathOf, menuIndexById } from '../../store/menu'
import {
  C, btn, fieldInput, fieldLabel, searchInput, checkItem,
  tableWrap, dataTable, tdStyle, tdOpsStyle, dash, emptyState, pageSection, toolbar, btnRow
} from '../../ui/kit'

const { div, tbody, tr, td, form, fieldset, legend } = van.tags

const PAGE_SIZE = 10

const permGridStyle = 'display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 4px 16px;'
const fieldsetStyle = `margin: 0 0 16px; padding: 10px 14px 12px; border: 1px solid ${C.border}; border-radius: 8px;`
const legendStyle = `padding: 0 6px; font-size: 0.88rem; font-weight: 600; color: ${C.textHead};`

function roleForm(existing: RoleItem | null, onDone: () => void): void {
  const submitting = van.state(false)
  const body = form(
    {},
    fieldLabel('角色标识 (Sid) *', fieldInput({ type: 'text', name: 'sid', required: true, value: existing?.sid ?? '', placeholder: '如: 1001' })),
    fieldLabel('角色名称 *', fieldInput({ type: 'text', name: 'name', required: true, value: existing?.name ?? '', placeholder: '如: 运营' }))
  )

  const submit = async (): Promise<void> => {
    if (submitting.val) return
    const data = new FormData(body as HTMLFormElement)
    const payload = {
      sid: String(data.get('sid') ?? '').trim(),
      name: String(data.get('name') ?? '').trim()
    }
    if (!payload.sid || !payload.name) {
      toast.error('请填写必填项')
      return
    }
    submitting.val = true
    try {
      if (existing) {
        await updateRole({ ...payload, id: existing.id })
        toast.success('更新成功')
      } else {
        await createRole(payload)
        toast.success('创建成功')
      }
      handle.close()
      onDone()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : '操作失败')
    } finally {
      submitting.val = false
    }
  }

  const footer = btnRow(
    btn('secondary', { onclick: () => handle.close() }, '取消'),
    btn('primary', { disabled: (): boolean => submitting.val, ariaBusy: (): boolean => submitting.val, onclick: () => void submit() }, existing ? '保存' : '创建')
  )
  const handle = openModal({ title: existing ? '编辑角色' : '新增角色', body, footer })
}

interface PermOption {
  key: string
  label: string
}

function menuPermOptions(menus: MenuItem[]): PermOption[] {
  const byId = menuIndexById(menus)
  return flattenTree(buildMenuTree(menus, { visibleOnly: false })).map(({ node, depth }) => {
    const full = fullPathOf(node.item, byId)
    const indent = '　'.repeat(depth)
    return { key: `menu:${full},read`, label: `${indent}${resolveTitle(node.item)} (${full})` }
  })
}

function apiPermOptions(apis: ApiItem[]): PermOption[] {
  return apis.map((a) => ({ key: `api:${a.path},${a.method.toUpperCase()}`, label: `${a.name} — ${a.method.toUpperCase()} ${a.path}` }))
}

function permissionDialog(role: RoleItem, onDone: () => void): void {
  const loading = van.state(true)
  const menuOpts = van.state<PermOption[]>([])
  const apiOpts = van.state<PermOption[]>([])
  const checked = new Set<string>()
  const saving = van.state(false)

  const permItem = (option: PermOption): HTMLElement =>
    checkItem(checked.has(option.key), (e: Event) => {
      const el = e.target as HTMLInputElement
      if (el.checked) checked.add(option.key)
      else checked.delete(option.key)
    }, option.label)

  const body = div(
    {},
    () =>
      loading.val
        ? div({ style: `padding: 24px; text-align: center; color: ${C.textMuted};`, 'aria-busy': 'true' }, '加载中…')
        : div(
            {},
            fieldset(
              { style: fieldsetStyle },
              legend({ style: legendStyle }, '菜单权限'),
              div({ style: permGridStyle }, menuOpts.val.map(permItem))
            ),
            fieldset(
              { style: fieldsetStyle },
              legend({ style: legendStyle }, 'API 权限'),
              div({ style: permGridStyle }, apiOpts.val.map(permItem))
            )
          )
  )

  const save = async (): Promise<void> => {
    if (saving.val) return
    saving.val = true
    try {
      await updateRolePermissions(role.sid, [...checked])
      toast.success('权限已更新')
      handle.close()
      onDone()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : '保存失败')
    } finally {
      saving.val = false
    }
  }

  const footer = btnRow(
    btn('secondary', { onclick: () => handle.close() }, '取消'),
    btn('primary', { disabled: (): boolean => saving.val || loading.val, ariaBusy: (): boolean => saving.val, onclick: () => void save() }, '保存权限')
  )
  const handle = openModal({ title: `分配权限 — ${role.name} (${role.sid})`, body, footer })

  void Promise.all([
    getRolePermissions(role.sid),
    getAdminMenus(),
    getApis({ page: 1, pageSize: 1000 })
  ])
    .then(([permData, menuData, apiData]) => {
      for (const key of permData.list ?? []) checked.add(key)
      menuOpts.val = menuPermOptions(menuData.list ?? [])
      apiOpts.val = apiPermOptions(apiData.list ?? [])
      loading.val = false
    })
    .catch((e: unknown) => {
      toast.error(e instanceof Error ? e.message : '权限数据加载失败')
      handle.close()
    })
}

/** Role management: table + CRUD + permission assignment. */
export function RoleListView(): HTMLElement {
  const rows = van.state<RoleItem[]>([])
  const total = van.state(0)
  const page = van.state(1)
  const loading = van.state(false)
  const kwName = van.state('')

  const load = async (): Promise<void> => {
    loading.val = true
    try {
      const d = await getRoles({ page: page.val, pageSize: PAGE_SIZE, name: kwName.val || undefined })
      rows.val = d.list ?? []
      total.val = d.total
    } catch (e) {
      toast.error(e instanceof Error ? e.message : '加载失败')
    } finally {
      loading.val = false
    }
  }
  void load()

  const onDelete = async (row: RoleItem): Promise<void> => {
    if (!(await confirmDialog(`确定删除角色「${row.name}」吗？`, '删除角色'))) return
    try {
      await deleteRole(row.id)
      toast.success('删除成功')
      void load()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : '删除失败')
    }
  }

  return pageSection(
    toolbar(
      searchInput({
        placeholder: '角色名称',
        oninput: (e: Event) => (kwName.val = (e.target as HTMLInputElement).value)
      }),
      btn('secondary', { onclick: () => { page.val = 1; void load() } }, '搜索'),
      btn('primary', { onclick: () => roleForm(null, () => void load()) }, '＋ 新增角色')
    ),
    tableWrap(
      { 'aria-busy': (): boolean => loading.val },
      dataTable(
        ['ID', '角色标识', '角色名称', '创建时间', '操作'],
        () =>
          tbody(
            rows.val.map((row) =>
              tr(
                td({ style: tdStyle }, String(row.id)),
                td({ style: tdStyle }, row.sid),
                td({ style: tdStyle }, row.name),
                td({ style: tdStyle }, row.createdAt?.slice(0, 19).replace('T', ' ') || dash()),
                td(
                  { style: tdOpsStyle },
                  btn('secondary', { small: true, onclick: () => permissionDialog(row, () => void load()) }, '权限'),
                  ' ',
                  btn('secondary', { small: true, onclick: () => roleForm(row, () => void load()) }, '编辑'),
                  ' ',
                  btn('danger', { small: true, onclick: () => void onDelete(row) }, '删除')
                )
              )
            )
          )
      ),
      () => (rows.val.length === 0 && !loading.val ? emptyState('暂无数据') : '')
    ),
    Pagination({
      page,
      total,
      pageSize: PAGE_SIZE,
      onChange: (p) => {
        page.val = p
        void load()
      }
    })
  )
}
