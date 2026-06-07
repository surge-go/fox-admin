import { computed, ref } from 'vue'

import { getMenuList } from '../../api/menu'
import type { LayoutTab } from '../../types/menu'
import { RouterType, type Router } from '../../types/router'

const routeList = ref<Router[]>([])
const activeMenu = ref('/dashboard')
const activeTab = ref('/dashboard')
const isLoading = ref(false)
const isLoaded = ref(false)
const tabs = ref<LayoutTab[]>([
  {
    key: '/dashboard',
    label: '工作台',
    fixed: true,
  },
])

function findRouteByPath(routes: Router[], path: string): Router | undefined {
  for (const route of routes) {
    if (route.path === path) {
      return route
    }

    const child = route.children?.length ? findRouteByPath(route.children, path) : undefined

    if (child) {
      return child
    }
  }

  return undefined
}

function collectFixedTabs(routes: Router[], result: LayoutTab[] = []) {
  for (const route of routes) {
    if (route.type === RouterType.Menu && route.mate.fixedTab && !route.mate.isHideTab) {
      result.push({
        key: route.path,
        label: route.mate.title,
        fixed: true,
      })
    }

    if (route.children?.length) {
      collectFixedTabs(route.children, result)
    }
  }

  return result
}

const fixedTabs = computed(() => collectFixedTabs(routeList.value))
const activeRoute = computed(() => findRouteByPath(routeList.value, activeTab.value))

function syncTabsWithFixedTabs() {
  const baseTabs = fixedTabs.value.length
    ? [...fixedTabs.value]
    : [
        {
          key: '/dashboard',
          label: '工作台',
          fixed: true,
        },
      ]

  const openedTabs = tabs.value.filter((tab) => !tab.fixed)

  for (const tab of openedTabs) {
    if (!baseTabs.some((item) => item.key === tab.key)) {
      baseTabs.push(tab)
    }
  }

  tabs.value = baseTabs
}

function setActiveTab(key: string) {
  activeTab.value = key
  activeMenu.value = key
}

function openTab(key: string) {
  const target = findRouteByPath(routeList.value, key)

  if (!target) {
    return
  }

  if (target.mate.link && target.mate.isExternal) {
    window.open(target.mate.link, '_blank', 'noopener,noreferrer')
    return
  }

  setActiveTab(key)

  if (target.mate.isHideTab) {
    return
  }

  if (!tabs.value.some((tab) => tab.key === key)) {
    tabs.value.push({
      key,
      label: target.mate.title,
      fixed: Boolean(target.mate.fixedTab),
    })
  }
}

function closeTab(key: string) {
  const tab = tabs.value.find((item) => item.key === key)

  if (!tab || tab.fixed) {
    return
  }

  tabs.value = tabs.value.filter((item) => item.key !== key)

  if (activeTab.value === key) {
    const nextTab = tabs.value[tabs.value.length - 1]

    if (nextTab) {
      setActiveTab(nextTab.key)
    }
  }
}

async function loadRoutes(force = false) {
  if (isLoading.value) {
    return routeList.value
  }

  if (isLoaded.value && !force) {
    return routeList.value
  }

  isLoading.value = true

  try {
    const routes = await getMenuList()
    routeList.value = routes
    syncTabsWithFixedTabs()
    isLoaded.value = true
    return routes
  }
  finally {
    isLoading.value = false
  }
}

export function useRouterStore() {
  return {
    activeMenu,
    activeTab,
    activeRoute,
    closeTab,
    routeList,
    fixedTabs,
    isLoaded,
    isLoading,
    loadRoutes,
    openTab,
    setActiveTab,
    tabs,
  }
}
