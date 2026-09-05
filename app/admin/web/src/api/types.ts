/** Backend API contract types (mirrors app/admin/api/v1). */

export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponseData {
  accessToken: string
}

export interface MenuAuthItem {
  title: string
  authMark: string
}

export interface MenuItem {
  id?: number
  parentId: number
  weight: number
  path: string
  title: string
  name?: string
  component?: string
  locale?: string
  icon?: string
  redirect?: string
  keepAlive?: boolean
  hideInMenu?: boolean
  isEnable?: boolean
  isMenu?: boolean
  isHide?: boolean
  isHideTab?: boolean
  link?: string
  isIframe?: boolean
  showBadge?: boolean
  showTextBadge?: string
  fixedTab?: boolean
  activePath?: string
  roles?: string[]
  isFullPage?: boolean
  authList?: MenuAuthItem[]
  target?: string
  url?: string
  updatedAt?: string
}

export interface MenuListData {
  list: MenuItem[]
}

export interface AdminUserItem {
  id: number
  username: string
  nickname: string
  password?: string
  email: string
  phone: string
  roles?: string[]
  updatedAt: string
  createdAt: string
}

export interface AdminUserListData {
  list: AdminUserItem[]
  total: number
}

export interface AdminUserPayload {
  id?: number
  username: string
  nickname: string
  password?: string
  email: string
  phone: string
  roles?: string[]
}

export interface RoleItem {
  id: number
  name: string
  sid: string
  updatedAt: string
  createdAt: string
}

export interface RoleListData {
  list: RoleItem[]
  total: number
}

export interface RolePayload {
  id?: number
  name: string
  sid: string
}

export interface ApiItem {
  id: number
  name: string
  path: string
  method: string
  group: string
  menuIds?: number[]
  updatedAt: string
  createdAt: string
}

export interface ApiListData {
  list: ApiItem[]
  total: number
  groups?: string[]
}

export interface ApiPayload {
  id?: number
  group: string
  name: string
  path: string
  method: string
  menuIds?: number[]
}

export interface StringListData {
  list: string[]
}

export interface PageQuery {
  page: number
  pageSize: number
}
