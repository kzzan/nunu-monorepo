import van from 'vanjs-core'
import { getToken, clearToken } from '../store/auth'

export const HOME_PATH = '/dashboard/console'
export const LOGIN_PATH = '/login'

/** Reactive current path (always normalized, guard-processed). */
export const currentPath = van.state<string>(LOGIN_PATH)

export type ViewFactory = () => HTMLElement

const routes = new Map<string, ViewFactory>()

export function registerRoutes(table: Record<string, ViewFactory>): void {
  for (const [path, factory] of Object.entries(table)) {
    routes.set(path, factory)
  }
}

export function resolveView(path: string): ViewFactory {
  return routes.get(path) ?? routes.get('*') ?? (() => document.createElement('main'))
}

export function navigate(path: string): void {
  if (`#${path}` === location.hash) return
  location.hash = path
}

function normalizePath(raw: string): string {
  let p = raw.trim()
  if (!p.startsWith('/')) p = `/${p}`
  if (p.length > 1 && p.endsWith('/')) p = p.slice(0, -1)
  return p
}

function handleRoute(): void {
  const raw = normalizePath(location.hash.slice(1) || '/')
  const authed = getToken() !== ''

  if (!authed) {
    clearToken()
    if (raw !== LOGIN_PATH) {
      location.replace(`#${LOGIN_PATH}`)
      return
    }
    currentPath.val = LOGIN_PATH
    return
  }

  let p = raw === '/' || raw === '/dashboard' ? HOME_PATH : raw
  if (p === LOGIN_PATH) {
    location.replace(`#${HOME_PATH}`)
    return
  }
  currentPath.val = p
}

/** Wire hashchange listener and run the initial guard pass. Call once at boot. */
export function startRouter(): void {
  window.addEventListener('hashchange', handleRoute)
  handleRoute()
}
