import van from 'vanjs-core'
import type { State } from 'vanjs-core'
import { Toggle } from 'vanjs-ui'
import type { MenuItem } from '../../api/types'
import { getAdminMenus, createMenu, updateMenu, deleteMenu } from '../../api/admin'
import { openModal } from '../../components/Modal'
import { confirmDialog } from '../../components/Confirm'
import { toast } from '../../components/Toast'
import { buildMenuTree, flattenTree, resolveTitle, fullPathOf, menuIndexById } from '../../store/menu'
import {
  C, btn, fieldInput, fieldLabel, fieldSelect, dash, monoText,
  tableWrap, dataTable, tdStyle, tdOpsStyle, tag, TAG_OK, TAG_OFF, emptyState, pageSection, toolbar, btnRow
} from '../../ui/kit'

const { div, option, tbody, tr, td, form, button, span } = van.tags

function boolTag(v: boolean | undefined, yes = '是', no = '否'): HTMLElement {
  return v ? tag(yes, TAG_OK.bg, TAG_OK.fg) : tag(no, TAG_OFF.bg, TAG_OFF.fg)
}

/** vanjs-ui Toggle with label text, bound to a boolean state. */
function toggleRow(labelText: string, on: State<boolean>): HTMLElement {
  return div(
    { style: 'display: flex; align-items: center; gap: 8px; font-weight: 400; font-size: 0.9rem;' },
    labelText,
    Toggle({ on, size: 0.9, onColor: C.brand })
  )
}

function menuForm(existing: MenuItem | null, parentOptions: Array<{ id: number; label: string }>, onDone: () => void): void {
  const submitting = van.state(false)
  const isMenu = van.state(existing?.isMenu ?? true)
  const isEnable = van.state(existing?.isEnable ?? true)
  const isHide = van.state(existing?.isHide ?? false)

  const body = form(
    {},
    fieldLabel(
      '父级菜单',
      fieldSelect(
        { name: 'parentId' },
        option({ value: '0' }, '根目录'),
        parentOptions.map((p) => option({ value: String(p.id), selected: existing?.parentId === p.id ? true : null }, p.label))
      )
    ),
    fieldLabel('标题 *', fieldInput({ type: 'text', name: 'title', required: true, value: existing?.title ?? '' })),
    fieldLabel('地址 *', fieldInput({ type: 'text', name: 'path', required: true, value: existing?.path ?? '', placeholder: '如: user 或 /admin' })),
    fieldLabel('路由 Name *', fieldInput({ type: 'text', name: 'name', required: true, value: existing?.name ?? '' })),
    fieldLabel('组件', fieldInput({ type: 'text', name: 'component', value: existing?.component ?? '', placeholder: '如: /admin/user' })),
    fieldLabel('图标', fieldInput({ type: 'text', name: 'icon', value: existing?.icon ?? '', placeholder: '如: ri:user-line' })),
    fieldLabel('重定向', fieldInput({ type: 'text', name: 'redirect', value: existing?.redirect ?? '' })),
    fieldLabel('排序权重 (越大越靠前)', fieldInput({ type: 'number', name: 'weight', value: String(existing?.weight ?? 100) })),
    fieldLabel(
      '选项',
      div(
        { style: 'display: flex; flex-wrap: wrap; gap: 10px 22px; margin-top: 6px;' },
        toggleRow(' 显示为菜单', isMenu),
        toggleRow(' 启用', isEnable),
        toggleRow(' 隐藏', isHide)
      )
    )
  )

  const submit = async (): Promise<void> => {
    if (submitting.val) return
    const formEl = body as HTMLFormElement
    const data = new FormData(formEl)
    const payload: MenuItem = {
      parentId: Number(data.get('parentId') ?? 0),
      title: String(data.get('title') ?? '').trim(),
      path: String(data.get('path') ?? '').trim(),
      name: String(data.get('name') ?? '').trim(),
      component: String(data.get('component') ?? '').trim(),
      icon: String(data.get('icon') ?? '').trim(),
      redirect: String(data.get('redirect') ?? '').trim(),
      weight: Number(data.get('weight') ?? 0) || 0,
      isMenu: isMenu.val,
      isEnable: isEnable.val,
      isHide: isHide.val
    }
    if (!payload.title || !payload.path || !payload.name) {
      toast.error('请填写必填项')
      return
    }
    submitting.val = true
    try {
      if (existing?.id !== undefined) {
        await updateMenu({ ...payload, id: existing.id })
        toast.success('更新成功')
      } else {
        await createMenu(payload)
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
  const handle = openModal({ title: existing ? '编辑菜单' : '新增菜单', body, footer })
}

/** Menu management: tree table + CRUD. */
export function MenuListView(): HTMLElement {
  const rows = van.state<MenuItem[]>([])
  const loading = van.state(false)
  const expanded = van.state<Set<number>>(new Set())

  const load = async (): Promise<void> => {
    loading.val = true
    try {
      const d = await getAdminMenus()
      const list = d.list ?? []
      rows.val = list
      // expand root level by default
      const ids = new Set<number>()
      for (const m of list) {
        if (!m.parentId && m.id !== undefined) ids.add(m.id)
      }
      expanded.val = ids
    } catch (e) {
      toast.error(e instanceof Error ? e.message : '加载失败')
    } finally {
      loading.val = false
    }
  }
  void load()

  const openForm = (existing: MenuItem | null): void => {
    const byId = menuIndexById(rows.val)
    const parentOptions = flattenTree(buildMenuTree(rows.val, { visibleOnly: false })).map(({ node, depth }) => {
      const full = fullPathOf(node.item, byId)
      return { id: node.item.id as number, label: `${'　'.repeat(depth)}${resolveTitle(node.item)} (${full})` }
    })
    if (existing?.id !== undefined) {
      // cannot select itself or its descendants as parent
      const forbidden = new Set<number>([existing.id as number])
      let changed = true
      while (changed) {
        changed = false
        for (const m of rows.val) {
          if (m.parentId && m.id !== undefined && forbidden.has(m.parentId) && !forbidden.has(m.id)) {
            forbidden.add(m.id)
            changed = true
          }
        }
      }
      menuForm(existing, parentOptions.filter((p) => !forbidden.has(p.id)), () => void load())
      return
    }
    menuForm(existing, parentOptions, () => void load())
  }

  const onDelete = async (item: MenuItem): Promise<void> => {
    const title = resolveTitle(item)
    if (!(await confirmDialog(`确定删除菜单「${title}」吗？子菜单将一并失效。`, '删除菜单'))) return
    if (item.id === undefined) return
    try {
      await deleteMenu(item.id)
      toast.success('删除成功')
      void load()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : '删除失败')
    }
  }

  const visibleRows = (): Array<{ item: MenuItem; depth: number; hasChildren: boolean; id: number | undefined }> => {
    const out: Array<{ item: MenuItem; depth: number; hasChildren: boolean; id: number | undefined }> = []
    const walk = (nodes: ReturnType<typeof buildMenuTree>, depth: number): void => {
      for (const n of nodes) {
        const id = n.item.id
        out.push({ item: n.item, depth, hasChildren: n.children.length > 0, id })
        if (id !== undefined && expanded.val.has(id)) walk(n.children, depth + 1)
      }
    }
    walk(buildMenuTree(rows.val, { visibleOnly: false }), 0)
    return out
  }

  const toggle = (id: number): void => {
    const s = new Set(expanded.val)
    if (s.has(id)) s.delete(id)
    else s.add(id)
    expanded.val = s
  }

  const titleCellStyle = `${tdStyle} white-space: nowrap;`

  return pageSection(
    toolbar(
      btn('primary', { onclick: () => openForm(null) }, '＋ 新增菜单')
    ),
    tableWrap(
      { 'aria-busy': (): boolean => loading.val },
      dataTable(
        ['标题', '地址', 'Name', '排序', '启用', '隐藏', '操作'],
        () =>
          tbody(
            visibleRows().map(({ item, depth, hasChildren, id }) =>
              tr(
                td(
                  { style: titleCellStyle },
                  hasChildren
                    ? button(
                        {
                          type: 'button',
                          'aria-label': '展开/折叠',
                          style: `margin: 0 4px 0 0; padding: 2px 7px; width: 1.6rem; font-family: inherit; border-radius: 6px; border: 1px solid ${C.border}; background: #fff; color: ${C.text}; cursor: pointer;`,
                          onclick: () => id !== undefined && toggle(id)
                        },
                        () => (id !== undefined && expanded.val.has(id) ? '▾' : '▸')
                      )
                    : span({ style: `display: inline-block; width: 21px; text-align: center; color: #cbd5e1;` }, '·'),
                  span({ style: `padding-left: ${depth * 14}px;` }),
                  resolveTitle(item)
                ),
                td({ style: tdStyle }, item.path ? monoText(item.path) : dash()),
                td({ style: tdStyle }, item.name || dash()),
                td({ style: tdStyle }, String(item.weight ?? 0)),
                td({ style: tdStyle }, boolTag(item.isEnable ?? true)),
                td({ style: tdStyle }, boolTag(item.isHide ?? false)),
                td(
                  { style: tdOpsStyle },
                  btn('secondary', { small: true, onclick: () => openForm(item) }, '编辑'),
                  ' ',
                  btn('danger', { small: true, onclick: () => void onDelete(item) }, '删除')
                )
              )
            )
          )
      ),
      () => (rows.val.length === 0 && !loading.val ? emptyState('暂无数据') : '')
    )
  )
}
