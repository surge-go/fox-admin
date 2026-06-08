import { computed, ref } from 'vue'

import { getMenuList } from '../../api/menu'
import type { LayoutTab } from '../../types/menu'
import { RouterCacheBy, RouterType, type RouteParams, type Router } from '../../types/router'

type RouteMatch = {
  /** 命中的路由配置 */
  route: Router

  /** 当前访问的完整路径，包含 query 和 hash */
  fullPath: string

  /** 当前访问的标准 pathname */
  pathname: string

  /** 从动态路径中提取出的参数 */
  params: RouteParams
}

type SetActiveTabOptions = {
  /** 是否同步浏览器地址 */
  syncHistory?: 'push' | 'replace' | 'none'

  /** 是否为当前路由创建标签页 */
  ensureTab?: boolean
}

const DEFAULT_TAB: LayoutTab = {
  key: '/dashboard',
  path: '/dashboard',
  fullPath: '/dashboard',
  label: '工作台',
  fixed: true,
}

const routeList = ref<Router[]>([])
const activeMenu = ref(DEFAULT_TAB.path)
const activeTab = ref(getCurrentBrowserFullPath())
const isLoading = ref(false)
const isLoaded = ref(false)
const tabs = ref<LayoutTab[]>([DEFAULT_TAB])

let hasBoundBrowserHistory = false

function getCurrentBrowserFullPath() {
  if (typeof window === 'undefined') {
    return DEFAULT_TAB.fullPath
  }

  return normalizeFullPath(`${window.location.pathname}${window.location.search}${window.location.hash}`)
}

function normalizePathname(pathname: string) {
  const normalized = pathname.trim() || '/'

  if (normalized === '/') {
    return normalized
  }

  return normalized.replace(/\/+$/, '')
}

function parseFullPath(fullPath: string) {
  const [pathAndSearch = '/', hash = ''] = fullPath.split('#', 2)
  const [pathname = '/', search = ''] = pathAndSearch.split('?', 2)

  return {
    pathname: normalizePathname(pathname),
    search: search ? `?${search}` : '',
    hash: hash ? `#${hash}` : '',
  }
}

function normalizeFullPath(fullPath: string) {
  const { pathname, search, hash } = parseFullPath(fullPath)
  return `${pathname}${search}${hash}`
}

function splitPath(path: string) {
  const normalized = normalizePathname(path)

  if (normalized === '/') {
    return []
  }

  return normalized.split('/').filter(Boolean)
}

function matchRoutePath(templatePath: string, pathname: string): RouteParams | undefined {
  const templateSegments = splitPath(templatePath)
  const pathSegments = splitPath(pathname)

  if (templateSegments.length !== pathSegments.length) {
    return undefined
  }

  const params: RouteParams = {}

  for (const [index, segment] of templateSegments.entries()) {
    const currentSegment = pathSegments[index]

    if (segment.startsWith(':')) {
      params[segment.slice(1)] = decodeURIComponent(currentSegment)
      continue
    }

    if (segment !== currentSegment) {
      return undefined
    }
  }

  return params
}

function getRouteScore(path: string) {
  return splitPath(path).reduce((score, segment) => {
    return score + (segment.startsWith(':') ? 1 : 10)
  }, 0)
}

function collectRoutes(routes: Router[], result: Router[] = []) {
  for (const route of routes) {
    result.push(route)

    if (route.children?.length) {
      collectRoutes(route.children, result)
    }
  }

  return result
}

function findRouteMatch(routes: Router[], fullPath: string): RouteMatch | undefined {
  const normalizedFullPath = normalizeFullPath(fullPath)
  const { pathname } = parseFullPath(normalizedFullPath)

  let bestMatch: RouteMatch | undefined
  let bestScore = -1

  for (const route of collectRoutes(routes)) {
    const params = matchRoutePath(route.path, pathname)

    if (!params) {
      continue
    }

    const score = getRouteScore(route.path)

    if (score > bestScore) {
      bestScore = score
      bestMatch = {
        route,
        fullPath: normalizedFullPath,
        pathname,
        params,
      }
    }
  }

  return bestMatch
}

function createTab(route: Router, fullPath = route.path): LayoutTab {
  const normalizedFullPath = normalizeFullPath(fullPath)

  return {
    key: normalizedFullPath,
    path: route.path,
    fullPath: normalizedFullPath,
    label: route.mate.title,
    fixed: Boolean(route.mate.fixedTab),
  }
}

function resolveActiveMenu(route: Router) {
  return route.mate.activeMenu || route.path
}

function resolveCacheKey(match?: RouteMatch) {
  if (!match) {
    return DEFAULT_TAB.fullPath
  }

  if (!match.route.mate.keepAlive) {
    return match.fullPath
  }

  return match.route.mate.cacheBy === RouterCacheBy.FullPath ? match.fullPath : match.route.path
}

function resolveRouteLink(route: Router, params: RouteParams) {
  if (!route.mate.link) {
    return undefined
  }

  return route.mate.link.replace(/:([A-Za-z0-9_]+)/g, (_, key: string) => params[key] || `:${key}`)
}

function syncBrowserLocation(fullPath: string, mode: 'push' | 'replace' | 'none' = 'push') {
  if (typeof window === 'undefined' || mode === 'none') {
    return
  }

  const nextFullPath = normalizeFullPath(fullPath)
  const currentFullPath = getCurrentBrowserFullPath()

  if (nextFullPath === currentFullPath) {
    return
  }

  const nextUrl = nextFullPath || '/'

  if (mode === 'replace') {
    window.history.replaceState(null, '', nextUrl)
    return
  }

  window.history.pushState(null, '', nextUrl)
}

function ensureTabExists(route: Router, fullPath: string) {
  if (route.mate.isHideTab) {
    return
  }

  const normalizedFullPath = normalizeFullPath(fullPath)

  if (route.mate.singleTab) {
    const currentIndex = tabs.value.findIndex((tab) => tab.path === route.path)

    if (currentIndex >= 0) {
      tabs.value[currentIndex] = createTab(route, normalizedFullPath)
      return
    }
  }

  if (!tabs.value.some((tab) => tab.key === normalizedFullPath)) {
    tabs.value.push(createTab(route, normalizedFullPath))
  }
}

function collectFixedTabs(routes: Router[], result: LayoutTab[] = []) {
  for (const route of routes) {
    if (route.type === RouterType.Menu && route.mate.fixedTab && !route.mate.isHideTab) {
      result.push(createTab(route))
    }

    if (route.children?.length) {
      collectFixedTabs(route.children, result)
    }
  }

  return result
}

function bindBrowserHistory() {
  if (typeof window === 'undefined' || hasBoundBrowserHistory) {
    return
  }

  window.addEventListener('popstate', () => {
    const nextFullPath = getCurrentBrowserFullPath()

    if (!isLoaded.value) {
      activeTab.value = nextFullPath
      return
    }

    const target = findRouteMatch(routeList.value, nextFullPath)

    if (target) {
      setActiveTab(target.fullPath, {
        syncHistory: 'none',
        ensureTab: true,
      })
      return
    }

    setActiveTab(DEFAULT_TAB.fullPath, {
      syncHistory: 'replace',
      ensureTab: true,
    })
  })

  hasBoundBrowserHistory = true
}

const fixedTabs = computed(() => collectFixedTabs(routeList.value))
const activeRouteMatch = computed(() => findRouteMatch(routeList.value, activeTab.value))
const activeRoute = computed(() => activeRouteMatch.value?.route)
const activeRouteParams = computed(() => activeRouteMatch.value?.params || {})
const activeRouteCacheKey = computed(() => resolveCacheKey(activeRouteMatch.value))

function syncTabsWithFixedTabs() {
  const baseTabs = fixedTabs.value.length ? [...fixedTabs.value] : [{ ...DEFAULT_TAB }]
  const openedTabs = tabs.value.filter((tab) => !tab.fixed)

  for (const tab of openedTabs) {
    if (!baseTabs.some((item) => item.key === tab.key)) {
      baseTabs.push(tab)
    }
  }

  tabs.value = baseTabs
}

function setActiveTab(fullPath: string, options: SetActiveTabOptions = {}) {
  const {
    syncHistory = 'push',
    ensureTab = false,
  } = options

  const target = findRouteMatch(routeList.value, fullPath)

  if (!target) {
    return
  }

  activeTab.value = target.fullPath
  activeMenu.value = resolveActiveMenu(target.route)
  syncBrowserLocation(target.fullPath, syncHistory)

  if (ensureTab) {
    ensureTabExists(target.route, target.fullPath)
  }
}

function openTab(fullPath: string) {
  const target = findRouteMatch(routeList.value, fullPath)

  if (!target) {
    return
  }

  const routeLink = resolveRouteLink(target.route, target.params)

  if (routeLink && target.route.mate.isExternal) {
    window.open(routeLink, '_blank', 'noopener,noreferrer')
    return
  }

  setActiveTab(target.fullPath, {
    syncHistory: 'push',
    ensureTab: true,
  })
}

function closeTab(key: string) {
  const currentIndex = tabs.value.findIndex((item) => item.key === key)

  if (currentIndex < 0 || tabs.value[currentIndex].fixed) {
    return
  }

  tabs.value = tabs.value.filter((item) => item.key !== key)

  if (activeTab.value === key) {
    const nextTab = tabs.value[currentIndex] || tabs.value[currentIndex - 1] || tabs.value[0]

    if (nextTab) {
      setActiveTab(nextTab.fullPath, {
        syncHistory: 'replace',
        ensureTab: true,
      })
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
    bindBrowserHistory()
    isLoaded.value = true

    const initialFullPath = getCurrentBrowserFullPath()
    const initialRoute = findRouteMatch(routes, initialFullPath)

    if (initialRoute) {
      setActiveTab(initialRoute.fullPath, {
        syncHistory: 'replace',
        ensureTab: true,
      })
    }
    else {
      setActiveTab(DEFAULT_TAB.fullPath, {
        syncHistory: 'replace',
        ensureTab: true,
      })
    }

    return routes
  }
  finally {
    isLoading.value = false
  }
}

export function useRouterStore() {
  return {
    activeMenu,
    activeRoute,
    activeRouteCacheKey,
    activeRouteMatch,
    activeRouteParams,
    activeTab,
    closeTab,
    fixedTabs,
    isLoaded,
    isLoading,
    loadRoutes,
    openTab,
    routeList,
    setActiveTab,
    tabs,
  }
}
