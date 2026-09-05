import van from 'vanjs-core'
import type { State } from 'vanjs-core'
import type { MenuItem } from '../api/types'
import { getUserMenus } from '../api/admin'
import { t } from '../i18n/zh'

export interface MenuNode {
  item: MenuItem
  children: MenuNode[]
}

const menus = van.state<MenuItem[]>([])

export function menusView(): State<MenuItem[]> {
  return menus
}

export async function fetchMenus(): Promise<void> {
  const data = await getUserMenus()
  menus.val = data.list ?? []
}

export function resetMenus(): void {
  menus.val = []
}

/** Resolve display title: i18n key -> zh-CN text, otherwise raw. */
export function resolveTitle(item: MenuItem): string {
  const raw = item.title ?? ''
  return raw.startsWith('menus.') ? t(raw) : raw
}

/** Full absolute path of a menu item, mirroring backend fullSeedMenuPath logic. */
export function fullPathOf(item: MenuItem, byId: Map<number, MenuItem>): string {
  if (!item.path) return ''
  if (item.path.startsWith('http://') || item.path.startsWith('https://') || item.path.startsWith('/')) {
    return item.path
  }
  if (!item.parentId) {
    return `/${item.path.replace(/^\//, '')}`
  }
  const parent = byId.get(item.parentId)
  if (!parent) {
    return `/${item.path.replace(/^\//, '')}`
  }
  const parentPath = fullPathOf(parent, byId).replace(/\/$/, '')
  const childPath = item.path.replace(/^\/+/, '')
  return parentPath ? `${parentPath}/${childPath}` : `/${childPath}`
}

export function menuIndexById(list: MenuItem[]): Map<number, MenuItem> {
  const map = new Map<number, MenuItem>()
  for (const item of list) {
    if (item.id !== undefined) {
      map.set(item.id, item)
    }
  }
  return map
}

function isVisible(item: MenuItem): boolean {
  return (
    (item.isMenu ?? true) &&
    (item.isEnable ?? true) &&
    !item.isHide &&
    !item.hideInMenu &&
    !item.isIframe
  )
}

export interface TreeOptions {
  /** Filter to sidebar-visible items (default true). Admin views pass false. */
  visibleOnly?: boolean
}

/** Build a menu tree from a flat list. */
export function buildMenuTree(list: MenuItem[], opts: TreeOptions = {}): MenuNode[] {
  const visibleOnly = opts.visibleOnly ?? true
  const nodes = new Map<number, MenuNode>()
  const roots: MenuNode[] = []
  for (const item of list) {
    if (item.id === undefined) continue
    nodes.set(item.id, { item, children: [] })
  }
  for (const item of list) {
    if (item.id === undefined) continue
    const node = nodes.get(item.id)
    if (!node) continue
    const parent = item.parentId ? nodes.get(item.parentId) : undefined
    if (parent) {
      parent.children.push(node)
    } else {
      roots.push(node)
    }
  }
  const sortRec = (treeList: MenuNode[]): MenuNode[] => {
    const filtered = visibleOnly ? treeList.filter((n) => isVisible(n.item)) : treeList
    filtered.sort((a, b) => (b.item.weight ?? 0) - (a.item.weight ?? 0))
    for (const n of filtered) {
      n.children = sortRec(n.children)
    }
    return filtered
  }
  return sortRec(roots)
}

/** Flatten tree depth-first for tree-table rendering. */
export function flattenTree(roots: MenuNode[]): Array<{ node: MenuNode; depth: number }> {
  const out: Array<{ node: MenuNode; depth: number }> = []
  const walk = (list: MenuNode[], depth: number): void => {
    for (const n of list) {
      out.push({ node: n, depth })
      walk(n.children, depth + 1)
    }
  }
  walk(roots, 0)
  return out
}
