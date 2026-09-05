import { http } from './http'
import type {
  AdminUserItem,
  AdminUserListData,
  AdminUserPayload,
  ApiListData,
  ApiPayload,
  LoginRequest,
  LoginResponseData,
  MenuItem,
  MenuListData,
  RoleListData,
  RolePayload,
  StringListData
} from './types'

export function login(req: LoginRequest): Promise<LoginResponseData> {
  return http.post<LoginResponseData>('/v1/login', req)
}

export function getCurrentAdminUser(): Promise<AdminUserItem> {
  return http.get<AdminUserItem>('/v1/admin/user')
}

/* ---------------- menus ---------------- */

export function getUserMenus(): Promise<MenuListData> {
  return http.get<MenuListData>('/v1/menus')
}

export function getAdminMenus(): Promise<MenuListData> {
  return http.get<MenuListData>('/v1/admin/menus')
}

export function createMenu(item: MenuItem): Promise<unknown> {
  return http.post('/v1/admin/menu', item)
}

export function updateMenu(item: MenuItem): Promise<unknown> {
  return http.put('/v1/admin/menu', item)
}

export function deleteMenu(id: number): Promise<unknown> {
  return http.delete('/v1/admin/menu', { id })
}

/* ---------------- admin users ---------------- */

export interface AdminUsersQuery {
  page: number
  pageSize: number
  id?: number
  username?: string
  nickname?: string
  phone?: string
  email?: string
}

export function getAdminUsers(query: AdminUsersQuery): Promise<AdminUserListData> {
  return http.get<AdminUserListData>('/v1/admin/users', {
    page: query.page,
    pageSize: query.pageSize,
    id: query.id,
    username: query.username,
    nickname: query.nickname,
    phone: query.phone,
    email: query.email
  })
}

export function createAdminUser(payload: AdminUserPayload): Promise<unknown> {
  return http.post('/v1/admin/user', payload)
}

export function updateAdminUser(payload: AdminUserPayload): Promise<unknown> {
  return http.put('/v1/admin/user', payload)
}

export function deleteAdminUser(id: number): Promise<unknown> {
  return http.delete('/v1/admin/user', { id })
}

/* ---------------- roles ---------------- */

export interface RolesQuery {
  page: number
  pageSize: number
  sid?: string
  name?: string
}

export function getRoles(query: RolesQuery): Promise<RoleListData> {
  return http.get<RoleListData>('/v1/admin/roles', {
    page: query.page,
    pageSize: query.pageSize,
    sid: query.sid,
    name: query.name
  })
}

export function createRole(payload: RolePayload): Promise<unknown> {
  return http.post('/v1/admin/role', payload)
}

export function updateRole(payload: RolePayload): Promise<unknown> {
  return http.put('/v1/admin/role', payload)
}

export function deleteRole(id: number): Promise<unknown> {
  return http.delete('/v1/admin/role', { id })
}

/* ---------------- permissions ---------------- */

export function getRolePermissions(roleSid: string): Promise<StringListData> {
  return http.get<StringListData>('/v1/admin/role/permissions', { role: roleSid })
}

export function updateRolePermissions(roleSid: string, list: string[]): Promise<unknown> {
  return http.put('/v1/admin/role/permissions', { role: roleSid, list })
}

export function getUserPermissions(): Promise<StringListData> {
  return http.get<StringListData>('/v1/admin/user/permissions')
}

/* ---------------- apis ---------------- */

export interface ApisQuery {
  page: number
  pageSize: number
  group?: string
  name?: string
  path?: string
  method?: string
}

export function getApis(query: ApisQuery): Promise<ApiListData> {
  return http.get<ApiListData>('/v1/admin/apis', {
    page: query.page,
    pageSize: query.pageSize,
    group: query.group,
    name: query.name,
    path: query.path,
    method: query.method
  })
}

export function createApi(payload: ApiPayload): Promise<unknown> {
  return http.post('/v1/admin/api', payload)
}

export function updateApi(payload: ApiPayload): Promise<unknown> {
  return http.put('/v1/admin/api', payload)
}

export function deleteApi(id: number): Promise<unknown> {
  return http.delete('/v1/admin/api', { id })
}
