import van from 'vanjs-core'
import type { AdminUserItem, RoleItem } from '../../api/types'
import { getAdminUsers, createAdminUser, updateAdminUser, deleteAdminUser, getRoles } from '../../api/admin'
import { openModal } from '../../components/Modal'
import { confirmDialog } from '../../components/Confirm'
import { Pagination } from '../../components/Pagination'
import { toast } from '../../components/Toast'
import {
  btn, fieldInput, fieldLabel, searchInput, checkItem,
  tableWrap, dataTable, tdStyle, tdOpsStyle, chip, dash, emptyState, pageSection, toolbar, btnRow
} from '../../ui/kit'

const { div, tbody, tr, td, form } = van.tags

const PAGE_SIZE = 10

function roleChips(roles: string[] | undefined): HTMLElement {
  const list = roles ?? []
  if (list.length === 0) return dash()
  return div({ style: 'display: flex; flex-wrap: wrap; gap: 5px;' }, list.map((r) => chip(r)))
}

function userForm(existing: AdminUserItem | null, roles: RoleItem[], onDone: () => void): void {
  const selected = new Set<string>(existing?.roles ?? [])
  const submitting = van.state(false)

  const body = form(
    {},
    fieldLabel('用户名 *', fieldInput({ type: 'text', name: 'username', required: true, value: existing?.username ?? '' })),
    fieldLabel('昵称', fieldInput({ type: 'text', name: 'nickname', value: existing?.nickname ?? '' })),
    fieldLabel(
      existing ? '密码（留空则不修改）' : '密码 *',
      fieldInput({ type: 'password', name: 'password', required: !existing })
    ),
    fieldLabel('邮箱', fieldInput({ type: 'email', name: 'email', value: existing?.email ?? '' })),
    fieldLabel('手机号', fieldInput({ type: 'text', name: 'phone', value: existing?.phone ?? '' })),
    fieldLabel(
      '角色',
      div(
        { style: 'display: flex; flex-wrap: wrap; gap: 6px 16px; margin-top: 6px;' },
        roles.map((r) =>
          checkItem(selected.has(r.sid), (e: Event) => {
            const el = e.target as HTMLInputElement
            if (el.checked) selected.add(r.sid)
            else selected.delete(r.sid)
          }, `${r.name} (${r.sid})`)
        )
      )
    )
  )

  const submit = async (): Promise<void> => {
    if (submitting.val) return
    const data = new FormData(body as HTMLFormElement)
    const payload = {
      username: String(data.get('username') ?? '').trim(),
      nickname: String(data.get('nickname') ?? '').trim(),
      password: String(data.get('password') ?? ''),
      email: String(data.get('email') ?? '').trim(),
      phone: String(data.get('phone') ?? '').trim(),
      roles: [...selected]
    }
    if (!payload.username || (!existing && !payload.password)) {
      toast.error('请填写必填项')
      return
    }
    submitting.val = true
    try {
      if (existing) {
        await updateAdminUser({ ...payload, id: existing.id })
        toast.success('更新成功')
      } else {
        await createAdminUser(payload)
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
    btn('primary', { disabled: (): boolean => submitting.val, ariaBusy: (): boolean => submitting.val, onclick: () => void submit() },
      existing ? '保存' : '创建')
  )

  const handle = openModal({ title: existing ? '编辑管理员' : '新增管理员', body, footer })
}

/** Admin user management: search + table + pagination + CRUD. */
export function UserListView(): HTMLElement {
  const rows = van.state<AdminUserItem[]>([])
  const total = van.state(0)
  const page = van.state(1)
  const loading = van.state(false)
  const kwUsername = van.state('')
  const kwEmail = van.state('')

  const load = async (): Promise<void> => {
    loading.val = true
    try {
      const d = await getAdminUsers({
        page: page.val,
        pageSize: PAGE_SIZE,
        username: kwUsername.val || undefined,
        email: kwEmail.val || undefined
      })
      rows.val = d.list ?? []
      total.val = d.total
    } catch (e) {
      toast.error(e instanceof Error ? e.message : '加载失败')
    } finally {
      loading.val = false
    }
  }
  void load()

  const openForm = (existing: AdminUserItem | null): void => {
    void getRoles({ page: 1, pageSize: 200 })
      .then((d) => userForm(existing, d.list ?? [], () => void load()))
      .catch(() => {
        toast.error('角色列表加载失败')
        userForm(existing, [], () => void load())
      })
  }

  const onDelete = async (row: AdminUserItem): Promise<void> => {
    if (!(await confirmDialog(`确定删除管理员「${row.username}」吗？`, '删除管理员'))) return
    try {
      await deleteAdminUser(row.id)
      toast.success('删除成功')
      void load()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : '删除失败')
    }
  }

  return pageSection(
    toolbar(
      searchInput({
        placeholder: '用户名',
        oninput: (e: Event) => (kwUsername.val = (e.target as HTMLInputElement).value)
      }),
      searchInput({
        placeholder: '邮箱',
        oninput: (e: Event) => (kwEmail.val = (e.target as HTMLInputElement).value)
      }),
      btn('secondary', { onclick: () => { page.val = 1; void load() } }, '搜索'),
      btn('primary', { onclick: () => openForm(null) }, '＋ 新增管理员')
    ),
    tableWrap(
      { 'aria-busy': (): boolean => loading.val },
      dataTable(
        ['ID', '用户名', '昵称', '邮箱', '手机号', '角色', '创建时间', '操作'],
        () =>
          tbody(
            rows.val.map((row) =>
              tr(
                td({ style: tdStyle }, String(row.id)),
                td({ style: tdStyle }, row.username),
                td({ style: tdStyle }, row.nickname || dash()),
                td({ style: tdStyle }, row.email || dash()),
                td({ style: tdStyle }, row.phone || dash()),
                td({ style: tdStyle }, roleChips(row.roles)),
                td({ style: tdStyle }, row.createdAt?.slice(0, 19).replace('T', ' ') || dash()),
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
