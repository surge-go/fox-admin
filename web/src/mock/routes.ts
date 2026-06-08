import type { ApiResult } from '../api/http'
import type { AppSettings } from '../types/app'
import { RouterCacheBy, RouterType, type Router } from '../types/router'

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

const appSettings: AppSettings = {
  title: 'A Fox Admin',
  titleSuffix: 'Fox Admin',
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
    children: [
      {
        id: 21,
        path: '/system/user',
        name: 'SystemUser',
        type: RouterType.Menu,
        component: 'system/user/index',
        mate: {
          title: '用户管理',
          icon: 'users',
        },
      },
      {
        id: 23,
        path: '/system/user/detail/:id',
        name: 'SystemUserDetail',
        type: RouterType.Menu,
        component: 'system/user/detail/index',
        mate: {
          title: '用户详情',
          isHide: true,
          activeMenu: '/system/user',
          keepAlive: true,
          cacheBy: RouterCacheBy.Path,
          singleTab: true,
        },
      },
      {
        id: 22,
        path: '/system/role',
        name: 'SystemRole',
        type: RouterType.Menu,
        component: 'system/role/index',
        mate: {
          title: '角色权限',
          icon: 'shield-check',
        },
      },
    ],
  },
  {
    id: 3,
    path: '/basic',
    name: 'BasicData',
    type: RouterType.Menu,
    component: 'basic/index',
    mate: {
      title: '基础数据',
      icon: 'database',
    },
  },
  {
    id: 4,
    path: '/document',
    name: 'DocumentCenter',
    type: RouterType.Menu,
    component: 'document/index',
    mate: {
      title: '文档中心',
      icon: 'file-text',
      isExternal: true,
      link: 'https://www.naiveui.com/zh-CN/os-theme/components',
    },
  },
  {
    id: 5,
    path: '/status',
    name: 'StatusDirectory',
    type: RouterType.Catalog,
    component: 'layout',
    mate: {
      title: '状态目录',
      icon: 'activity-heartbeat',
    },
    children: [
      {
        id: 51,
        path: '/status/system',
        name: 'SystemStatus',
        type: RouterType.Menu,
        component: 'status/system/index',
        mate: {
          title: '系统状态',
          icon: 'server',
          keepAlive: true,
        },
      },
      {
        id: 52,
        path: '/status/board',
        name: 'StatusBoard',
        type: RouterType.Menu,
        component: 'status/board/index',
        mate: {
          title: '监控看板',
          icon: 'activity-heartbeat',
          link: '/embedded/status-board.html',
        },
      },
    ],
  },
]

export const mockRoutes: MockRoute[] = [
  {
    path: '/api/app/settings',
    method: 'GET',
    response: () => ({
      code: 200,
      message: '获取应用配置成功',
      data: appSettings,
    }),
  },
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
