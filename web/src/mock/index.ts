import axios, {
  type AxiosAdapter,
  type AxiosResponse,
  type InternalAxiosRequestConfig,
} from 'axios'

import { mockRoutes, type MockRoute } from './routes'

type MockMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'

type MockRequest = {
  /** 请求地址 */
  url: string

  /** 请求方法 */
  method: MockMethod

  /** 查询参数 */
  params?: unknown

  /** 请求体 */
  data?: unknown

  /** 请求头 */
  headers: InternalAxiosRequestConfig['headers']
}

function isAbsoluteUrl(url: string): boolean {
  return /^[a-z][a-z\d+\-.]*:\/\//i.test(url)
}

function joinUrl(baseURL = '', url = ''): string {
  if (!baseURL || isAbsoluteUrl(url)) {
    return url
  }

  const normalizedBaseURL = baseURL.replace(/\/+$/, '')
  const normalizedUrl = url.replace(/^\/+/, '')

  if (url.startsWith(`${normalizedBaseURL}/`) || url === normalizedBaseURL) {
    return url
  }

  return `${normalizedBaseURL}/${normalizedUrl}`
}

function normalizePath(url?: string): string {
  if (!url) {
    return '/'
  }

  const [path] = url.split(/[?#]/)
  return path || '/'
}

function normalizeMethod(method?: string): MockMethod {
  return (method || 'GET').toUpperCase() as MockMethod
}

function parseData(data: unknown): unknown {
  if (typeof data !== 'string') {
    return data
  }

  try {
    return JSON.parse(data)
  } catch {
    return data
  }
}

function matchRoute(path: string, method: MockMethod): MockRoute | undefined {
  return mockRoutes.find((route) => {
    return route.path === path && route.method === method
  })
}

function createResponse<T>(
  config: InternalAxiosRequestConfig,
  status: number,
  data: T,
): AxiosResponse<T> {
  return {
    config,
    data,
    headers: {},
    status,
    statusText: status >= 200 && status < 300 ? 'OK' : 'Error',
  }
}

/** 创建 Axios mock adapter */
export function createMockAdapter(originalAdapter?: AxiosAdapter): AxiosAdapter {
  const fallbackAdapter = originalAdapter || axios.getAdapter(axios.defaults.adapter)

  return async (config) => {
    const path = normalizePath(joinUrl(config.baseURL, config.url))
    const method = normalizeMethod(config.method)
    const route = matchRoute(path, method)

    if (!route) {
      return fallbackAdapter(config)
    }

    const request: MockRequest = {
      data: parseData(config.data),
      headers: config.headers,
      method,
      params: config.params,
      url: path,
    }

    const result = await route.response(request)
    const delay = route.delay ?? 300

    if (delay > 0) {
      await new Promise((resolve) => window.setTimeout(resolve, delay))
    }

    return createResponse(config, route.status ?? 200, result)
  }
}
