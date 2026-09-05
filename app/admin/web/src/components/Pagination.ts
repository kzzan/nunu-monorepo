import van from 'vanjs-core'
import type { State } from 'vanjs-core'
import { C } from '../ui/kit'

const { nav, ul, li, button, small } = van.tags

export interface PaginationOptions {
  page: State<number>
  total: State<number>
  pageSize: number
  onChange: (page: number) => void
}

const PAGE_BASE = `min-width: 30px; padding: 3px 7px; font-family: inherit; text-align: center; border-radius: 6px; border: 1px solid ${C.border}; background: #fff; color: ${C.text}; font-size: 0.86rem; cursor: pointer;`
const PAGE_ACTIVE = `min-width: 30px; padding: 3px 7px; font-family: inherit; text-align: center; border-radius: 6px; border: 1px solid ${C.brand}; background: ${C.brand}; color: #fff; font-size: 0.86rem; cursor: pointer;`

function pageBtn(label: string, opts: { active?: boolean; ariaLabel?: string; ariaCurrent?: boolean; onClick: () => void }): HTMLButtonElement {
  return button(
    {
      type: 'button',
      'aria-label': opts.ariaLabel ?? null,
      'aria-current': opts.ariaCurrent ? 'page' : null,
      style: opts.active ? PAGE_ACTIVE : PAGE_BASE,
      onclick: opts.onClick
    },
    label
  )
}

function pageWindow(current: number, last: number): number[] {
  if (last <= 7) {
    return Array.from({ length: Math.max(last, 1) }, (_, i) => i + 1)
  }
  const pages = new Set<number>([1, last, current - 1, current, current + 1])
  const sorted = [...pages].filter((p) => p >= 1 && p <= last).sort((a, b) => a - b)
  const out: number[] = []
  let prev = 0
  for (const p of sorted) {
    if (p - prev > 1) out.push(0)
    out.push(p)
    prev = p
  }
  return out
}

export function Pagination(opts: PaginationOptions): HTMLElement {
  const last = van.derive(() => Math.max(1, Math.ceil(opts.total.val / opts.pageSize)))
  const go = (p: number): void => {
    const target = Math.min(Math.max(1, p), last.val)
    if (target !== opts.page.val) {
      opts.onChange(target)
    }
  }

  return nav(
    { style: 'display: flex; align-items: center; justify-content: space-between; gap: 16px; flex-wrap: wrap;', 'aria-label': '分页' },
    small({ style: `color: ${C.textSecondary};` }, () => `共 ${opts.total.val} 条 · 第 ${opts.page.val}/${last.val} 页`),
    () =>
      ul(
        { style: 'display: flex; gap: 4px; list-style: none; margin: 0; padding: 0; align-items: center;' },
        li(pageBtn('‹', { ariaLabel: '上一页', onClick: () => go(opts.page.val - 1) })),
        pageWindow(opts.page.val, last.val).map((p) =>
          p === 0
            ? li({ style: `padding: 0 6px; color: ${C.textMuted};` }, '…')
            : li(
                pageBtn(String(p), {
                  active: p === opts.page.val,
                  ariaCurrent: p === opts.page.val,
                  onClick: () => go(p)
                })
              )
        ),
        li(pageBtn('›', { ariaLabel: '下一页', onClick: () => go(opts.page.val + 1) }))
      )
  )
}
