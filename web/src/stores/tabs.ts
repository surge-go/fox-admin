import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import type { RouteLocationNormalizedLoaded } from 'vue-router'

export type AdminTab = {
  path: string
  title: string
  closable: boolean
  cacheName?: string
  keepAlive: boolean
}

const HOME_TAB: AdminTab = {
  path: '/dashboard',
  title: 'Dashboard',
  closable: false,
  cacheName: 'DashboardView',
  keepAlive: true,
}

export const useTabsStore = defineStore('tabs', () => {
  const tabs = ref<AdminTab[]>([HOME_TAB])

  const cachedNames = computed(() => {
    return Array.from(
      new Set(
        tabs.value
          .filter((tab) => tab.keepAlive && tab.cacheName)
          .map((tab) => tab.cacheName!),
      ),
    )
  })

  const paths = computed(() => tabs.value.map((tab) => tab.path))

  function addRouteTab(route: RouteLocationNormalizedLoaded) {
    if (!route.meta.requiresAuth || typeof route.meta.title !== 'string') {
      return
    }

    if (paths.value.includes(route.path)) {
      return
    }

    tabs.value.push({
      path: route.path,
      title: route.meta.title,
      closable: route.path !== HOME_TAB.path,
      cacheName: typeof route.meta.cacheName === 'string' ? route.meta.cacheName : String(route.name ?? ''),
      keepAlive: route.meta.keepAlive !== false,
    })
  }

  function removeTab(path: string) {
    const current = tabs.value.find((tab) => tab.path === path)

    if (!current?.closable) {
      return
    }

    tabs.value = tabs.value.filter((tab) => tab.path !== path)
  }

  function resetTabs() {
    tabs.value = [HOME_TAB]
  }

  return {
    tabs,
    cachedNames,
    addRouteTab,
    removeTab,
    resetTabs,
  }
})
