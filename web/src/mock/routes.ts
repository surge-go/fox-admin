import type { ApiResult } from '../api/http'
import { RouterType, type Router } from '../types/router'

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
  headers: unknown
}

export type MockRoute<T = unknown> = {
  /** mock 地址 */
  path: string

  /** HTTP 方法 */
  method: MockMethod

  /** mock 响应状态码 */
  status?: number

  /** mock 延迟时间，单位毫秒 */
  delay?: number

  /** mock 响应函数 */
  response: (request: MockRequest) => ApiResult<T> | Promise<ApiResult<T>>
}

type LoginResult = {
  /** 访问令牌 */
  token: string

  /** 用户名 */
  username: string
}

type UserProfile = {
  /** 用户 ID */
  id: number

  /** 用户名 */
  username: string

  /** 昵称 */
  nickname: string

  /** 角色标识 */
  roles: string[]

  /** 权限标识 */
  permissions: string[]
}

const menuList: Router[] = [
  {
    id: 1,
    path: '/dashboard',
    name: 'Dashboard',
    type: RouterType.Menu,
    component: 'dashboard/index',
    mate: {
      title: '工作台',
      icon: 'dashboard',
      fixedTab: true,
      keepAlive: true,
    },
  },
  {
    id: 2,
    path: '/system',
    name: 'System',
    type: RouterType.Catalog,
    component: 'layout',
    mate: {
      title: '系统管理',
      icon: 'settings',
    },
  },
]

export const mockRoutes: MockRoute[] = [
  {
    path: '/api/auth/login',
    method: 'POST',
    response: () => ({
      code: 200,
      message: '登录成功',
      data: {
        token: 'mock-token',
        username: 'admin',
      } satisfies LoginResult,
    }),
  },
  {
    path: '/api/user/profile',
    method: 'GET',
    response: () => ({
      code: 200,
      message: '获取用户信息成功',
      data: {
        id: 1,
        nickname: '管理员',
        permissions: ['system:user:list', 'system:user:create'],
        roles: ['admin'],
        username: 'admin',
      } satisfies UserProfile,
    }),
  },
  {
    path: '/api/menu/list',
    method: 'GET',
    response: () => ({
      code: 200,
      message: '获取菜单成功',
      data: menuList,
    }),
  },
]
