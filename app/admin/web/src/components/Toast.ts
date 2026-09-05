import { MessageBoard } from 'vanjs-ui'

type ToastKind = 'success' | 'error' | 'info'

const ICONS: Record<ToastKind, string> = {
  success: '✓',
  error: '✕',
  info: 'ℹ'
}

let board: MessageBoard | null = null

/** Lazily create the single top-center message board (vanjs-ui). */
function getBoard(): MessageBoard {
  if (board === null) {
    board = new MessageBoard({
      top: '14px',
      messageStyleOverrides: { padding: '10px 16px', 'font-size': '0.92rem' }
    })
  }
  return board
}

function push(kind: ToastKind, text: string): void {
  getBoard().show({ message: `${ICONS[kind]}  ${text}`, durationSec: 3 })
}

export const toast = {
  success: (text: string): void => push('success', text),
  error: (text: string): void => push('error', text),
  info: (text: string): void => push('info', text)
}
