import van from 'vanjs-core'
import { openModal } from './Modal'
import { btn, btnRow } from '../ui/kit'

const { p } = van.tags

export function confirmDialog(message: string, title = '操作确认'): Promise<boolean> {
  return new Promise((resolve) => {
    let settled = false
    const settle = (v: boolean): void => {
      if (!settled) {
        settled = true
        resolve(v)
      }
    }

    const body = p({ style: 'margin: 0 0 4px;' }, message)
    const footer = btnRow(
      btn('secondary', { onclick: () => { settle(false); handle.close() } }, '取消'),
      btn('primary', { onclick: () => { settle(true); handle.close() } }, '确定')
    )

    const handle = openModal({ title, body, footer, onClose: () => settle(false) })
  })
}
