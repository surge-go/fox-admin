import http from './http'

import type { AppSettings } from '../types/app'

/** 获取前端应用基础配置 */
export function getAppSettings() {
  return http.get<AppSettings>('/api/app/settings')
}
