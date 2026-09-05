import van from 'vanjs-core'
import { currentPath } from './router/router'
import { LoginView } from './views/LoginView'
import { AppLayout } from './layout/AppLayout'

const { div } = van.tags

export function App(): HTMLElement {
  const root = div({ style: 'min-height: 100vh;' })
  van.add(
    root,
    () => (currentPath.val === '/login' ? LoginView() : AppLayout())
  )
  return root
}
