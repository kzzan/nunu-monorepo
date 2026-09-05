import van from 'vanjs-core'
import { currentPath } from '../router/router'
import { buildMenuTree, menusView, resolveTitle, fullPathOf, menuIndexById, type MenuNode } from '../store/menu'
import { C } from '../ui/kit'

const { nav, a, div, span, ul, li } = van.tags

const LINK_BASE = `display: flex; align-items: center; gap: 7px; padding: 8px 10px; margin: 2px 0; border-radius: 8px; font-size: 0.92rem; text-decoration: none; cursor: pointer; transition: background 0.15s, color 0.15s;`

/** Inline hover effect for sidebar links (skipped when the link is active). */
function hover(el: HTMLElement, isActive: () => boolean): HTMLElement {
  el.addEventListener('mouseenter', () => {
    if (!isActive()) {
      el.style.background = C.brandLight
      el.style.color = C.brandDark
    }
  })
  el.addEventListener('mouseleave', () => {
    if (!isActive()) {
      el.style.background = 'transparent'
      el.style.color = C.text
    }
  })
  return el
}

function leafLink(node: MenuNode): HTMLElement {
  const byId = menuIndexById(menusView().val)
  const full = fullPathOf(node.item, byId)
  const active = () => currentPath.val === full || currentPath.val.startsWith(`${full}/`)
  return hover(
    a(
      {
        href: `#${full}`,
        style: () => (active() ? `${LINK_BASE} background: ${C.brand}; color: #fff;` : `${LINK_BASE} color: ${C.text};`)
      },
      span(resolveTitle(node.item))
    ),
    active
  )
}

function groupItem(node: MenuNode): HTMLElement {
  const byId = menuIndexById(menusView().val)
  const full = fullPathOf(node.item, byId)
  const open = van.state(currentPath.val.startsWith(`${full}/`))
  const active = () => currentPath.val.startsWith(`${full}/`)

  const head = hover(
    div(
      {
        role: 'button',
        tabindex: '0',
        style: () =>
          active()
            ? `${LINK_BASE} font-weight: 600; background: ${C.brand}; color: #fff;`
            : `${LINK_BASE} font-weight: 600; color: ${C.text};`,
        onclick: () => {
          open.val = !open.val
        },
        onkeydown: (e: KeyboardEvent) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            open.val = !open.val
          }
        }
      },
      span(resolveTitle(node.item)),
      span(
        {
          style: () =>
            `margin-left: auto; transition: transform 0.15s; ${open.val ? 'transform: rotate(90deg);' : ''} color: ${active() ? '#fff' : C.textMuted};`
        },
        '›'
      )
    ),
    active
  )

  return li(
    head,
    () =>
      open.val
        ? ul(
            { style: `list-style: none; margin: 0 0 4px; padding: 0 0 0 1rem; border-left: 1px dashed ${C.border}; margin-left: 11px;` },
            node.children.map((c) => (c.children.length > 0 ? groupItem(c) : li(leafLink(c))))
          )
        : ''
  )
}

/** Sidebar rendered from the current user's menu tree. */
export function Sidebar(): HTMLElement {
  return nav(
    { 'aria-label': '主导航', style: `padding: 8px 10px 24px; ${'display: flex; flex-direction: column;'}` },
    () => {
      const tree = buildMenuTree(menusView().val)
      if (tree.length === 0) {
        return div({ style: `padding: 16px 13px; color: ${C.textMuted}; font-size: 0.88rem;` }, '暂无菜单')
      }
      return ul(
        { style: 'list-style: none; margin: 0; padding: 0;' },
        tree.map((n) => (n.children.length > 0 ? groupItem(n) : li(leafLink(n))))
      )
    }
  )
}
