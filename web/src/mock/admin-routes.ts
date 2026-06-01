export type MockRouteIcon = 'GridOutline' | 'HomeOutline' | 'ListOutline' | 'OptionsOutline'

export type MockRouteComponent =
  | 'BasicListView'
  | 'DashboardView'
  | 'FormDesignView'
  | 'FormTableExampleView'
  | 'IFrameView'
  | 'MenuPermissionView'
  | 'RolePermissionView'

export type MockAdminRoute = {
  path: string
  name: string
  title: string
  icon?: MockRouteIcon
  component?: MockRouteComponent
  externalUrl?: string
  iframeUrl?: string
  keepAlive?: boolean
  children?: MockAdminRoute[]
}

export const mockAdminRoutes: MockAdminRoute[] = [
  {
    path: '/dashboard',
    name: 'Dashboard',
    title: 'Dashboard',
    icon: 'HomeOutline',
    component: 'DashboardView',
  },
  {
    path: '/system',
    name: 'System',
    title: '系统设置',
    icon: 'OptionsOutline',
    children: [
      {
        path: '/system/menu-permission',
        name: 'MenuPermission',
        title: '菜单权限',
        component: 'MenuPermissionView',
      },
      {
        path: '/system/role-permission',
        name: 'RolePermission',
        title: '角色权限',
        component: 'RolePermissionView',
      },
    ],
  },
  {
    path: '/basic/list',
    name: 'BasicList',
    title: '基础列表',
    icon: 'GridOutline',
    component: 'BasicListView',
  },
  {
    path: '/form/design',
    name: 'FormDesign',
    title: '表单设计',
    icon: 'ListOutline',
    component: 'FormDesignView',
  },
  {
    path: '/examples/form-table',
    name: 'FormTableExample',
    title: '组件示例',
    icon: 'ListOutline',
    component: 'FormTableExampleView',
  },
  {
    path: '/links',
    name: 'Links',
    title: '链接示例',
    icon: 'ListOutline',
    children: [
      {
        path: '/links/iframe',
        name: 'IFrameExample',
        title: '内嵌链接',
        component: 'IFrameView',
        iframeUrl: 'https://example.com',
      },
      {
        path: '/links/naive-ui',
        name: 'NaiveUiExternal',
        title: 'Naive UI 外链',
        externalUrl: 'https://www.naiveui.com/zh-CN/os-theme',
      },
    ],
  },
]
