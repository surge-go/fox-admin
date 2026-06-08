import axios, {
  type AxiosError,
  type AxiosInstance,
  type AxiosRequestConfig,
  type AxiosResponse,
  type InternalAxiosRequestConfig,
} from 'axios'

import { createMockAdapter } from '../mock'

/** 后端通用响应结构 */
export type ApiResult<T = unknown> = {
  /** 业务状态码，200 表示成功 */
  code: number

  /** 响应消息 */
  message?: string

  /** 响应数据 */
  data: T
}

/** 请求配置 */
export type HttpRequestConfig<D = unknown> = AxiosRequestConfig<D> & {
  /** 是否跳过 token 注入 */
  skipAuth?: boolean

  /** 是否跳过业务响应解包，直接返回 response.data */
  rawData?: boolean
}

/** 标准化后的请求错误 */
export class HttpError<T = unknown> extends Error {
  /** HTTP 状态码 */
  status?: number

  /** 业务状态码 */
  code?: number

  /** 原始响应数据 */
  data?: T

  constructor(message: string, options?: { status?: number; code?: number; data?: T }) {
    super(message)
    this.name = 'HttpError'
    this.status = options?.status
    this.code = options?.code
    this.data = options?.data
  }
}

type InternalRequestConfig = InternalAxiosRequestConfig & {
  skipAuth?: boolean
}

const TOKEN_KEY = 'fox-admin-token'

function getStoredToken() {
  if (typeof window === 'undefined') {
    return ''
  }

  try {
    return window.localStorage.getItem(TOKEN_KEY) || ''
  }
  catch {
    return ''
  }
}

/** Axios 底层实例 */
export const httpInstance: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

if (import.meta.env.DEV && import.meta.env.VITE_USE_MOCK === 'true') {
  httpInstance.defaults.adapter = createMockAdapter(
    axios.getAdapter(httpInstance.defaults.adapter),
  )
}

httpInstance.interceptors.request.use((config: InternalRequestConfig) => {
  if (!config.skipAuth) {
    const token = getStoredToken()

    if (token) {
      config.headers.set('Authorization', `Bearer ${token}`)
    }
  }

  return config
})

httpInstance.interceptors.response.use(
  (response) => response,
  (error: AxiosError<ApiResult>) => {
    const status = error.response?.status
    const code = error.response?.data?.code
    const message = error.response?.data?.message || error.message || '请求失败'

    if (status === 401) {
      window.dispatchEvent(new CustomEvent('auth:unauthorized'))
    }

    return Promise.reject(
      new HttpError(message, {
        status,
        code,
        data: error.response?.data,
      }),
    )
  },
)

function isApiResult<T>(value: unknown): value is ApiResult<T> {
  return typeof value === 'object' && value !== null && 'code' in value && 'data' in value
}

function isSuccessResult(result: ApiResult): boolean {
  return result.code === 200
}

function isSuccessStatus(status: number): boolean {
  return status >= 200 && status < 300
}

function getResponseMessage(value: unknown, fallback = '请求失败'): string {
  if (isApiResult(value) && value.message) {
    return value.message
  }

  return fallback
}

function unwrapResponse<T>(response: AxiosResponse<ApiResult<T> | T>, rawData?: boolean): T {
  const body = response.data

  if (!isSuccessStatus(response.status)) {
    throw new HttpError(getResponseMessage(body, response.statusText || '请求失败'), {
      status: response.status,
      code: isApiResult(body) ? body.code : undefined,
      data: body,
    })
  }

  if (rawData || !isApiResult<T>(body)) {
    return body as T
  }

  if (!isSuccessResult(body)) {
    throw new HttpError(getResponseMessage(body), {
      status: response.status,
      code: body.code,
      data: body,
    })
  }

  return body.data
}

/** 发起请求并返回解包后的数据 */
export async function request<T = unknown, D = unknown>(
  config: HttpRequestConfig<D>,
): Promise<T> {
  const response = await httpInstance.request<
    ApiResult<T> | T,
    AxiosResponse<ApiResult<T> | T>,
    D
  >(config)

  return unwrapResponse<T>(response, config.rawData)
}

/** 发起 GET 请求 */
export function get<T = unknown>(url: string, config?: HttpRequestConfig): Promise<T> {
  return request<T>({ ...config, method: 'GET', url })
}

/** 发起 POST 请求 */
export function post<T = unknown, D = unknown>(
  url: string,
  data?: D,
  config?: HttpRequestConfig<D>,
): Promise<T> {
  return request<T, D>({ ...config, method: 'POST', url, data })
}

/** 发起 PUT 请求 */
export function put<T = unknown, D = unknown>(
  url: string,
  data?: D,
  config?: HttpRequestConfig<D>,
): Promise<T> {
  return request<T, D>({ ...config, method: 'PUT', url, data })
}

/** 发起 PATCH 请求 */
export function patch<T = unknown, D = unknown>(
  url: string,
  data?: D,
  config?: HttpRequestConfig<D>,
): Promise<T> {
  return request<T, D>({ ...config, method: 'PATCH', url, data })
}

/** 发起 DELETE 请求 */
export function deleteRequest<T = unknown>(
  url: string,
  config?: HttpRequestConfig,
): Promise<T> {
  return request<T>({ ...config, method: 'DELETE', url })
}

/** 常用请求方法 */
export const http = {
  request,
  get,
  post,
  put,
  patch,
  delete: deleteRequest,
}

export default http
