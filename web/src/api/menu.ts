import http from './http'

import type { Router } from '../types/router'

/** 获取菜单路由列表 */
export function getMenuList() {
  return http.get<Router[]>('/api/menu/list')
}
