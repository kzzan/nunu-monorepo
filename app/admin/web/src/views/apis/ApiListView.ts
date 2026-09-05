import van from 'vanjs-core'
import type { ApiItem, MenuItem } from '../../api/types'
import { getApis, createApi, updateApi, deleteApi, getAdminMenus } from '../../api/admin'
import { openModal } from '../../components/Modal'
import { confirmDialog } from '../../components/Confirm'
import { Pagination } from '../../components/Pagination'
import { toast } from '../../components/Toast'
import { buildMenuTree, flattenTree, resolveTitle } from '../../store/menu'
import {
  C, btn, fieldInput, fieldLabel, fieldSelect, searchInput, checkItem, methodTag,
  monoText, dash, tableWrap, dataTable, tdStyle, tdOpsStyle, emptyState, pageSection, toolbar, btnRow
} from '../../ui/kit'

const { div, option, tbody, tr, td, form } = van.tags

const PAGE_SIZE = 10
const METHODS = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH']

function apiForm(existing: ApiItem | null, menus: MenuItem[], onDone: () => void): void {
  const selected = new Set<number>(existing?.menuIds ?? [])
  const submitting = van.state(false)
  const menuRows = flattenTree(buildMenuTree(menus, { visibleOnly: false }))

  const body = form(
    {},
    fieldLabel('分组 *', fieldInput({ type: 'text', name: 'group', required: true, value: existing?.group ?? '', placeholder: '如: 权限管理/菜单' })),
    fieldLabel('名称 *', fieldInput({ type: 'text', name: 'name', required: true, value: existing?.name ?? '' })),
    fieldLabel('路径 *', fieldInput({ type: 'text', name: 'path', required: true, value: existing?.path ?? '', placeholder: '如: /v1/admin/menu' })),
    fieldLabel(
      '方法 *',
      fieldSelect(
        { name: 'method' },
        METHODS.map((m) => option({ value: m, selected: (existing?.method ?? 'GET').toUpperCase() === m ? true : null }, m))
      )
    ),
    fieldLabel(
      '关联菜单',
      div(
        {
          style: `display: flex; flex-wrap: wrap; gap: 6px 16px; margin-top: 6px; max-height: 180px; overflow-y: auto; border: 1px solid ${C.border}; border-radius: 8px; padding: 8px 11px;`
        },
        menuRows.map(({ node }) =>
          checkItem(
            node.item.id !== undefined && selected.has(node.item.id),
            (e: Event) => {
              const el = e.target as HTMLInputElement
              const id = node.item.id
              if (id === undefined) return
              if (el.checked) selected.add(id)
              else selected.delete(id)
            },
            `${resolveTitle(node.item)}#${node.item.id ?? '?'}`
          )
        )
      )
    )
  )

  const submit = async (): Promise<void> => {
    if (submitting.val) return
    const data = new FormData(body as HTMLFormElement)
    const payload = {
      group: String(data.get('group') ?? '').trim(),
      name: String(data.get('name') ?? '').trim(),
      path: String(data.get('path') ?? '').trim(),
      method: String(data.get('method') ?? 'GET'),
      menuIds: [...selected]
    }
    if (!payload.group || !payload.name || !payload.path) {
      toast.error('请填写必填项')
      return
    }
    submitting.val = true
    try {
      if (existing) {
        await updateApi({ ...payload, id: existing.id })
        toast.success('更新成功')
      } else {
        await createApi(payload)
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
  const handle = openModal({ title: existing ? '编辑接口' : '新增接口', body, footer })
}

/** API resource management: table + pagination + CRUD + menu binding. */
export function ApiListView(): HTMLElement {
  const rows = van.state<ApiItem[]>([])
  const total = van.state(0)
  const page = van.state(1)
  const loading = van.state(false)
  const kwPath = van.state('')

  const load = async (): Promise<void> => {
    loading.val = true
    try {
      const d = await getApis({ page: page.val, pageSize: PAGE_SIZE, path: kwPath.val || undefined })
      rows.val = d.list ?? []
      total.val = d.total
    } catch (e) {
      toast.error(e instanceof Error ? e.message : '加载失败')
    } finally {
      loading.val = false
    }
  }
  void load()

  const openForm = (existing: ApiItem | null): void => {
    void getAdminMenus()
      .then((d) => apiForm(existing, d.list ?? [], () => void load()))
      .catch(() => apiForm(existing, [], () => void load()))
  }

  const onDelete = async (row: ApiItem): Promise<void> => {
    if (!(await confirmDialog(`确定删除接口「${row.name}」吗？`, '删除接口'))) return
    try {
      await deleteApi(row.id)
      toast.success('删除成功')
      void load()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : '删除失败')
    }
  }

  return pageSection(
    toolbar(
      searchInput({
        placeholder: '接口路径',
        oninput: (e: Event) => (kwPath.val = (e.target as HTMLInputElement).value)
      }),
      btn('secondary', { onclick: () => { page.val = 1; void load() } }, '搜索'),
      btn('primary', { onclick: () => openForm(null) }, '＋ 新增接口')
    ),
    tableWrap(
      { 'aria-busy': (): boolean => loading.val },
      dataTable(
        ['ID', '分组', '名称', '方法', '路径', '关联菜单', '操作'],
        () =>
          tbody(
            rows.val.map((row) =>
              tr(
                td({ style: tdStyle }, String(row.id)),
                td({ style: tdStyle }, row.group || dash()),
                td({ style: tdStyle }, row.name),
                td({ style: tdStyle }, methodTag(row.method)),
                td({ style: tdStyle }, monoText(row.path)),
                td({ style: tdStyle }, row.menuIds?.length ? `${row.menuIds.length} 项` : dash()),
                td(
                  { style: tdOpsStyle },
                  btn('secondary', { small: true, onclick: () => openForm(row) }, '编辑'),
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
