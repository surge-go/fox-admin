import { createRouter, createWebHistory } from 'vue-router'
import { areMockDynamicRoutesReady, setupMockDynamicRoutes } from './dynamic'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'AdminRoot',
      component: () => import('../layouts/AdminLayout.vue'),
      redirect: '/dashboard',
      meta: { requiresAuth: true },
      children: [],
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/login/LoginView.vue'),
      meta: { title: '登录' },
    },
  ],
})

router.beforeEach((to) => {
  const token = localStorage.getItem('fox-admin-token')

  if (to.path !== '/login' && !token) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }

  if (to.path === '/login' && token) {
    setupMockDynamicRoutes(router)
    return '/dashboard'
  }

  if (token && !areMockDynamicRoutesReady()) {
    setupMockDynamicRoutes(router)
    return to.fullPath
  }

  return true
})

export default router
