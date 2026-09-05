import van from 'vanjs-core'
import { Modal } from 'vanjs-ui'
import { C } from '../ui/kit'

const { div, h3 } = van.tags

export interface ModalOptions {
  title: string
  body: HTMLElement
  footer?: HTMLElement
  onClose?: () => void
}

export interface ModalHandle {
  close: () => void
}

/** Open a vanjs-ui Modal with title / body / footer sections; returns a handle to close it. */
export function openModal(opts: ModalOptions): ModalHandle {
  const closed = van.state(false)

  const header = div(
    { style: `padding: 14px 20px 10px; border-bottom: 1px solid ${C.border};` },
    h3({ style: 'margin: 0; font-size: 1.05rem;' }, opts.title)
  )
  const body = div({ style: 'padding: 16px 20px; max-height: 62vh; overflow-y: auto;' }, opts.body)

  const onKey = (e: KeyboardEvent): void => {
    if (e.key === 'Escape') closed.val = true
  }
  document.addEventListener('keydown', onKey)

  let notified = false
  van.derive(() => {
    if (closed.val) {
      document.removeEventListener('keydown', onKey)
      if (!notified) {
        notified = true
        opts.onClose?.()
      }
    }
  })

  van.add(
    document.body,
    Modal(
      {
        closed,
        blurBackground: true,
        clickBackgroundToClose: true,
        modalStyleOverrides: {
          padding: 0,
          'max-width': 'min(560px, 94vw)',
          'max-height': '88vh',
          'overflow-y': 'auto'
        }
      },
      header,
      body,
      opts.footer ? div({ style: `padding: 12px 20px 16px; border-top: 1px solid ${C.border};` }, opts.footer) : null
    )
  )

  return { close: (): void => { closed.val = true } }
}
