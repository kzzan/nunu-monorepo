/**
 * UI kit: design tokens + inline-styled primitives for the admin console.
 *
 * Replaces the old Pico CSS + styles.css stack. All styling is inline
 * (no CSS files, no CSS classes); shared looks live here as helpers.
 * Tokens mirror the previous styles.css values 1:1.
 */
import van from 'vanjs-core'
import type { ChildDom } from 'vanjs-core'

const { button, input, select, label, div, span, table, thead, tr, th } = van.tags

// ---------- design tokens ----------

export const C = {
  brand: '#1095c1',
  brandDark: '#0c7a9e',
  brandLight: '#e8f0f6',
  bgAside: '#f7f9fb',
  bgSoft: '#f8fafc',
  border: '#e2e8f0',
  borderSoft: '#f1f5f9',
  text: '#334155',
  textHead: '#475569',
  textMuted: '#94a3b8',
  textSecondary: '#64748b',
  danger: '#dc2626',
  dangerDark: '#b91c1c',
  dangerBg: '#fef2f2',
  dangerBorder: '#fecaca',
  ok: '#047857',
  okBg: '#ecfdf5',
  offBg: '#f1f5f9'
} as const

// ---------- buttons ----------

export type BtnKind = 'primary' | 'secondary' | 'danger'

interface BtnVisual {
  bg: string
  fg: string
  border: string
  hoverBg: string
}

const btnVisual: Record<BtnKind, BtnVisual> = {
  primary: { bg: C.brand, fg: '#fff', border: C.brand, hoverBg: C.brandDark },
  secondary: { bg: '#fff', fg: C.text, border: C.border, hoverBg: C.brandLight },
  danger: { bg: '#fff', fg: C.danger, border: C.dangerBorder, hoverBg: C.dangerBg }
}

export interface BtnProps {
  type?: 'button' | 'submit'
  small?: boolean
  onclick?: (e: MouseEvent) => void
  disabled?: boolean | (() => boolean)
  ariaBusy?: boolean | (() => boolean)
  title?: string
}

/** Inline-styled button with JS hover effect. Kinds: primary (solid), secondary (outline), danger (destructive outline). */
export function btn(kind: BtnKind, props: BtnProps, ...children: ChildDom[]): HTMLButtonElement {
  const v = btnVisual[kind]
  const pad = props.small ? '4px 11px' : '8px 16px'
  const fontSize = props.small ? '0.84rem' : '0.95rem'
  const b = button(
    {
      type: props.type ?? 'button',
      onclick: props.onclick ?? null,
      disabled: props.disabled ?? null,
      'aria-busy': props.ariaBusy ?? null,
      title: props.title ?? null,
      style: `padding: ${pad}; font-size: ${fontSize}; font-family: inherit; border-radius: 8px; border: 1px solid ${v.border}; background: ${v.bg}; color: ${v.fg}; cursor: pointer; margin: 0; line-height: 1.4;`
    },
    ...children
  )
  b.addEventListener('mouseenter', (): void => {
    if (!b.disabled) b.style.backgroundColor = v.hoverBg
  })
  b.addEventListener('mouseleave', (): void => {
    if (!b.disabled) b.style.backgroundColor = v.bg
  })
  return b
}

// ---------- form fields ----------

const inputStyle = `width: 100%; box-sizing: border-box; padding: 8px 12px; font-size: 0.92rem; font-family: inherit; border: 1px solid ${C.border}; border-radius: 8px; background: #fff; color: ${C.text}; margin: 4px 0 0;`

/** Block form label wrapping its control(s). */
export function fieldLabel(text: string, ...controls: ChildDom[]): HTMLLabelElement {
  return label(
    { style: `display: block; margin: 0 0 12px; font-weight: 600; font-size: 0.88rem; color: ${C.textHead};` },
    text,
    ...controls
  )
}

/** Full-width text input styled like the old Pico fields. */
export function fieldInput(props: Record<string, unknown>): HTMLInputElement {
  return input({ ...props, style: inputStyle })
}

/** Full-width select styled to match fieldInput. */
export function fieldSelect(props: Record<string, unknown>, ...options: ChildDom[]): HTMLSelectElement {
  return select({ ...props, style: `${inputStyle} cursor: pointer;` }, ...options)
}

/** Fixed-width search input for toolbars. */
export function searchInput(props: Record<string, unknown>, width = '180px'): HTMLInputElement {
  return input({
    ...props,
    type: 'search',
    style: `width: ${width}; box-sizing: border-box; padding: 8px 12px; font-size: 0.9rem; font-family: inherit; border: 1px solid ${C.border}; border-radius: 8px; background: #fff; color: ${C.text}; margin: 0;`
  })
}

/** Inline checkbox with label text (for role/menu checkbox groups). */
export function checkItem(checked: boolean, onchange: (e: Event) => void, text: string): HTMLLabelElement {
  return label(
    {
      style: `display: inline-flex; align-items: center; gap: 6px; margin: 0; font-size: 0.9rem; font-weight: 400; cursor: pointer;`
    },
    input({ type: 'checkbox', checked, onchange }),
    text
  )
}

// ---------- table ----------

export const thStyle = `padding: 10px 13px; font-size: 0.86rem; text-align: left; border-bottom: 1px solid ${C.border}; background: ${C.bgSoft}; color: ${C.textHead}; white-space: nowrap;`
export const tdStyle = `padding: 10px 13px; font-size: 0.9rem; border-bottom: 1px solid ${C.borderSoft}; vertical-align: middle;`
export const tdOpsStyle = `${tdStyle} white-space: nowrap;`

/** White card wrapper with horizontal scroll for data tables. */
export function tableWrap(props: Record<string, unknown>, ...children: ChildDom[]): HTMLDivElement {
  return div(
    { ...props, style: `overflow-x: auto; background: #fff; border: 1px solid ${C.border}; border-radius: 10px;` },
    ...children
  )
}

export function dataTable(headCells: ChildDom[], ...body: ChildDom[]): HTMLTableElement {
  return table(
    { style: 'width: 100%; border-collapse: collapse; margin: 0;' },
    thead(tr(...headCells.map((c) => th({ style: thStyle }, c)))),
    ...body
  )
}

// ---------- small bits ----------

/** Muted "—" placeholder for empty values. */
export function dash(): HTMLSpanElement {
  return span({ style: `color: ${C.textMuted};` }, '—')
}

export function mutedText(text: string): HTMLSpanElement {
  return span({ style: `color: ${C.textMuted};` }, text)
}

export function monoText(text: string): HTMLSpanElement {
  return span(
    { style: 'font-family: ui-monospace, SFMono-Regular, Menlo, monospace; font-size: 0.84rem;' },
    text
  )
}

/** Small rounded chip (e.g. role names in the user table). */
export function chip(text: string): HTMLSpanElement {
  return span(
    {
      style: `display: inline-block; padding: 2px 9px; border-radius: 999px; background: ${C.brandLight}; color: ${C.brandDark}; font-size: 0.78rem;`
    },
    text
  )
}

/** Colored tag span (boolean pills, HTTP methods). */
export function tag(text: string, bg: string, fg: string): HTMLSpanElement {
  return span(
    { style: `display: inline-block; padding: 2px 8px; border-radius: 6px; background: ${bg}; color: ${fg}; font-size: 0.78rem;` },
    text
  )
}

export const TAG_OK = { bg: C.okBg, fg: C.ok }
export const TAG_OFF = { bg: C.offBg, fg: C.textSecondary }

export const METHOD_COLORS: Record<string, { bg: string; fg: string }> = {
  GET: { bg: '#e0f2fe', fg: '#0369a1' },
  POST: { bg: '#ecfdf5', fg: '#047857' },
  PUT: { bg: '#fef9c3', fg: '#a16207' },
  DELETE: { bg: '#fee2e2', fg: '#b91c1c' },
  PATCH: { bg: '#f3e8ff', fg: '#7e22ce' }
}

export function methodTag(method: string): HTMLSpanElement {
  const c = METHOD_COLORS[method.toUpperCase()] ?? { bg: C.offBg, fg: C.textSecondary }
  return tag(method.toUpperCase(), c.bg, c.fg)
}

/** Centered muted empty-state block under tables. */
export function emptyState(text: string): HTMLDivElement {
  return div({ style: `padding: 36px; text-align: center; color: ${C.textMuted};` }, text)
}

// ---------- page scaffolding ----------

/** Standard page container: column flex with gap. */
export function pageSection(...children: ChildDom[]): HTMLElement {
  return div({ style: 'display: flex; flex-direction: column; gap: 16px;' }, ...children)
}

/** Toolbar row above tables: search inputs + action buttons. */
export function toolbar(...children: ChildDom[]): HTMLDivElement {
  return div({ style: 'display: flex; flex-wrap: wrap; gap: 10px; align-items: center;' }, ...children)
}

/** Right-aligned button row for modal footers. */
export function btnRow(...children: ChildDom[]): HTMLDivElement {
  return div({ style: 'display: flex; justify-content: flex-end; gap: 10px;' }, ...children)
}
