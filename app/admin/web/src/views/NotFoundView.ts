import van from 'vanjs-core'
import { navigate, HOME_PATH } from '../router/router'
import { C, btn } from '../ui/kit'

const { section, h2, p } = van.tags

/** 404 / not-implemented page for unknown routes (incl. dropped demo pages). */
export function NotFoundView(): HTMLElement {
  return section(
    { style: 'text-align: center; padding: 64px 0;' },
    h2({ style: `font-size: 3rem; margin: 0 0 6px; color: ${C.textMuted};` }, '404'),
    p({ style: `color: ${C.textSecondary}; margin: 0 0 22px;` }, '页面不存在或未在新版前端中实现'),
    btn('secondary', { onclick: () => navigate(HOME_PATH) }, '返回仪表盘')
  )
}
