import type { ApiResponse } from './types'
import { clearToken, getToken } from '../store/auth'

export class ApiError extends Error {
  readonly code: number
  readonly status: number

  constructor(code: number, message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.status = status
  }
}

export type QueryParams = Record<string, string | number | undefined>

interface RequestOptions {
  method: string
  url: string
  query?: QueryParams
  body?: unknown
}

function buildQueryString(query: QueryParams): string {
  const params = new URLSearchParams()
  for (const [key, value] of Object.entries(query)) {
    if (value !== undefined && value !== '') {
      params.set(key, String(value))
    }
  }
  const s = params.toString()
  return s ? `?${s}` : ''
}

async function request<T>(opts: RequestOptions): Promise<T> {
  const headers: Record<string, string> = {}
  const token = getToken()
  if (token) {
    headers['Authorization'] = token
  }
  if (opts.body !== undefined) {
    headers['Content-Type'] = 'application/json'
  }

  let res: Response
  try {
    res = await fetch(`${opts.url}${opts.query ? buildQueryString(opts.query) : ''}`, {
      method: opts.method,
      headers,
      body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined
    })
  } catch (e) {
    throw new ApiError(-1, `网络请求失败: ${e instanceof Error ? e.message : String(e)}`, 0)
  }

  if (res.status === 401) {
    if (opts.url === '/v1/login') {
      throw new ApiError(401, '用户名或密码错误', res.status)
    }
    clearToken()
    if (!location.hash.startsWith('#/login')) {
      location.hash = '/login'
    }
    throw new ApiError(401, '登录已过期，请重新登录', res.status)
  }

  let envelope: ApiResponse<T>
  try {
    envelope = (await res.json()) as ApiResponse<T>
  } catch {
    throw new ApiError(-1, `服务响应异常 (HTTP ${res.status})`, res.status)
  }

  if (envelope.code !== 0) {
    throw new ApiError(envelope.code, envelope.message || '请求失败', res.status)
  }
  return envelope.data
}

export const http = {
  get<T>(url: string, query?: QueryParams): Promise<T> {
    return request<T>({ method: 'GET', url, query })
  },
  post<T>(url: string, body?: unknown): Promise<T> {
    return request<T>({ method: 'POST', url, body })
  },
  put<T>(url: string, body?: unknown): Promise<T> {
    return request<T>({ method: 'PUT', url, body })
  },
  delete<T>(url: string, query?: QueryParams): Promise<T> {
    return request<T>({ method: 'DELETE', url, query })
  }
}
