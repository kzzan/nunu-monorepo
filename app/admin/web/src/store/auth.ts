import van from 'vanjs-core'
import type { State } from 'vanjs-core'
import type { AdminUserItem } from '../api/types'

const TOKEN_KEY = 'nunu_admin_token'

const token = van.state<string>(localStorage.getItem(TOKEN_KEY) ?? '')
const currentUser = van.state<AdminUserItem | null>(null)

export function getToken(): string {
  return token.val
}

export function setToken(value: string): void {
  token.val = value
  localStorage.setItem(TOKEN_KEY, value)
}

export function clearToken(): void {
  token.val = ''
  currentUser.val = null
  localStorage.removeItem(TOKEN_KEY)
}

export function setCurrentUser(user: AdminUserItem | null): void {
  currentUser.val = user
}

export function currentUserView(): State<AdminUserItem | null> {
  return currentUser
}
