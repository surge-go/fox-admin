import type { Router, RouteRecordRaw } from 'vue-router'
import type { MockAdminRoute } from '../mock/admin-routes'
import { mockAdminRoutes } from '../mock/admin-routes'

const viewModules = {
  BasicListView: () => import('../views/basic/BasicListView.vue'),
  DashboardView: () => import('../views/dashboard/DashboardView.vue'),
  FormDesignView: () => import('../views/form/FormDesignView.vue'),
  FormTableExampleView: () => import('../views/examples/FormTableExampleView.vue'),
  IFrameView: () => import('../views/frame/IFrameView.vue'),
  MenuPermissionView: () => import('../views/system/MenuPermissionView.vue'),
  RolePermissionView: () => import('../views/system/RolePermissionView.vue'),
}

let mockDynamicRoutesReady = false

function toChildPath(path: string) {
  return path.replace(/^\//, '')
}

function toRouteRecord(route: MockAdminRoute): RouteRecordRaw {
  if (!route.component) {
    throw new Error(`Mock route "${route.name}" is missing component or iframeUrl.`)
  }

  return {
    path: toChildPath(route.path),
    name: route.name,
    component: viewModules[route.component],
    meta: {
      cacheName: route.component,
      iframeUrl: route.iframeUrl,
      keepAlive: route.keepAlive !== false,
      title: route.title,
      requiresAuth: true,
    },
  }
}

function collectPageRoutes(routes: MockAdminRoute[]): MockAdminRoute[] {
  return routes.flatMap((route) => {
    if (route.children?.length) {
      return collectPageRoutes(route.children)
    }

    return route.component ? [route] : []
  })
}

export function areMockDynamicRoutesReady() {
  return mockDynamicRoutesReady
}

export function setupMockDynamicRoutes(router: Router) {
  if (mockDynamicRoutesReady) {
    return
  }

  collectPageRoutes(mockAdminRoutes).forEach((route) => {
    router.addRoute('AdminRoot', toRouteRecord(route))
  })

  mockDynamicRoutesReady = true
}
