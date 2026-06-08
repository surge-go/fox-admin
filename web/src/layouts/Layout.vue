<script setup lang="ts">
import { defineAsyncComponent, markRaw, type Component } from 'vue'
import {
  IconActivityHeartbeat,
  IconBell,
  IconChevronDown,
  IconChevronsLeft,
  IconChevronsRight,
  IconDatabase,
  IconFileText,
  IconHelpCircle,
  IconHome,
  IconLayoutDashboard,
  IconMoon,
  IconSearch,
  IconServer,
  IconSettings,
  IconShieldCheck,
  IconSun,
  IconUsers,
} from '@tabler/icons-vue'
import type { MenuOption } from 'naive-ui'
import { dateZhCN, NIcon, zhCN } from 'naive-ui'

import { useAppStore, useRouterStore, useThemeStore } from '../store'
import type { LayoutMenuItem } from '../types/menu'
import { RouterType, type Router } from '../types/router'

const collapsed = ref(false)
const expandedMenuKeys = ref<string[]>([])
const appStore = useAppStore()
const routerStore = useRouterStore()
const themeStore = useThemeStore()
const logo = appStore.logo
const markText = appStore.markText
const appTitle = appStore.title
const appIsLoaded = appStore.isLoaded
const appIsLoading = appStore.isLoading
const activeMenu = routerStore.activeMenu
const activeTab = routerStore.activeTab
const activeRoute = routerStore.activeRoute
const activeRouteParams = routerStore.activeRouteParams
const activeRouteCacheKey = routerStore.activeRouteCacheKey
const isDark = themeStore.isDark
const shellVars = themeStore.shellVars
const themeToggleAnchor = ref<HTMLElement | null>(null)
const routerIsLoaded = routerStore.isLoaded
const routerIsLoading = routerStore.isLoading
const routeList = routerStore.routeList
const theme = themeStore.naiveTheme
const themeOverrides = themeStore.themeOverrides
const tabs = routerStore.tabs
const bootstrapError = ref('')
const isBootstrapping = computed(() => {
  return appIsLoading.value || routerIsLoading.value || !appIsLoaded.value || !routerIsLoaded.value
})

const iconMap: Record<string, Component> = {
  'activity-heartbeat': IconActivityHeartbeat,
  dashboard: IconLayoutDashboard,
  server: IconServer,
  settings: IconSettings,
  users: IconUsers,
  'shield-check': IconShieldCheck,
  database: IconDatabase,
  'file-text': IconFileText,
}

const viewModules = import.meta.glob('../views/**/*.vue') as Record<
  string,
  () => Promise<{ default: Component }>
>
const viewComponentMap = new Map<string, Component>()

function renderIcon(icon?: Component) {
  if (!icon) {
    return undefined
  }

  return () => h(NIcon, null, { default: () => h(icon) })
}

function resolveIcon(icon?: string): Component | undefined {
  if (!icon) {
    return undefined
  }

  return iconMap[icon]
}

function normalizeMenuItem(route: Router): LayoutMenuItem | undefined {
  if (route.mate.isHide) {
    return undefined
  }

  const children = route.children
    ?.map(normalizeMenuItem)
    .filter((item): item is LayoutMenuItem => Boolean(item))

  if (route.type === RouterType.Catalog) {
    return {
      key: route.path,
      label: route.mate.title,
      icon: resolveIcon(route.mate.icon),
      children,
    }
  }

  return {
    key: route.path,
    label: route.mate.title,
    icon: resolveIcon(route.mate.icon),
    children,
  }
}

function toLayoutMenuItems(routes: Router[]) {
  return routes
    .map(normalizeMenuItem)
    .filter((item): item is LayoutMenuItem => Boolean(item))
}

function toMenuOptions(items: LayoutMenuItem[]): MenuOption[] {
  return items.map((item) => ({
    key: item.key,
    label: item.label,
    icon: renderIcon(item.icon),
    children: item.children ? toMenuOptions(item.children) : undefined,
  }))
}

function findAncestorMenuKeys(routes: Router[], targetKey: string, parents: string[] = []): string[] {
  for (const route of routes) {
    if (route.path === targetKey) {
      return parents
    }

    if (route.children?.length) {
      const nextParents = route.type === RouterType.Catalog ? [...parents, route.path] : parents
      const matchedParents = findAncestorMenuKeys(route.children, targetKey, nextParents)

      if (matchedParents.length) {
        return matchedParents
      }
    }
  }

  return []
}

const menuItems = computed(() => toLayoutMenuItems(routeList.value))
const menuOptions = computed(() => toMenuOptions(menuItems.value))
const autoExpandedMenuKeys = computed(() => findAncestorMenuKeys(routeList.value, activeMenu.value))
const mergedExpandedMenuKeys = computed(() => {
  return Array.from(new Set([...expandedMenuKeys.value, ...autoExpandedMenuKeys.value]))
})
const activeRouteFullPath = computed(() => routerStore.activeRouteMatch.value?.fullPath || activeTab.value)
const activeRouteLink = computed(() => {
  const link = activeRoute.value?.mate.link

  if (!link) {
    return undefined
  }

  return link.replace(
    /:([A-Za-z0-9_]+)/g,
    (_, key: string) => activeRouteParams.value[key] || `:${key}`,
  )
})
const activeViewComponent = computed(() => {
  const componentPath = activeRoute.value?.component

  if (!componentPath || isIframeRoute.value) {
    return null
  }

  const cachedComponent = viewComponentMap.get(componentPath)

  if (cachedComponent) {
    return cachedComponent
  }

  const modulePath = `../views/${componentPath}.vue`
  const loader = viewModules[modulePath]

  if (!loader) {
    return null
  }

  const component = markRaw(defineAsyncComponent(loader))
  viewComponentMap.set(componentPath, component)
  return component
})
let fallbackThemeOverlay: HTMLDivElement | null = null

function handleMenuUpdate(key: string) {
  routerStore.openTab(key)
}

function handleExpandedKeysUpdate(keys: string[]) {
  expandedMenuKeys.value = keys
}

function handleCloseTab(key: string) {
  routerStore.closeTab(key)
}

const isIframeRoute = computed(() => Boolean(activeRoute.value?.mate.link && !activeRoute.value?.mate.isExternal))
const isMissingLocalView = computed(() => {
  return Boolean(activeRoute.value && !isIframeRoute.value && !activeViewComponent.value)
})

function toggleCollapsed() {
  collapsed.value = !collapsed.value
}

type DocumentWithViewTransition = Document & {
  startViewTransition?: (callback: () => Promise<void> | void) => {
    ready: Promise<void>
    finished: Promise<void>
  }
}

async function toggleTheme(event?: MouseEvent) {
  const anchor = themeToggleAnchor.value
  const prefersReducedMotion = typeof window.matchMedia === 'function'
    && window.matchMedia('(prefers-reduced-motion: reduce)').matches
  const transitionApi = (document as DocumentWithViewTransition).startViewTransition?.bind(document)
  const root = document.documentElement

  if (!anchor || prefersReducedMotion) {
    themeStore.toggleTheme()
    return
  }

  const { left, top, width, height } = anchor.getBoundingClientRect()
  const x = left + width / 2
  const y = top + height / 2
  const endRadius = Math.hypot(
    Math.max(x, window.innerWidth - x),
    Math.max(y, window.innerHeight - y),
  ) + 8

  root.style.setProperty('--theme-switch-x', `${x}px`)
  root.style.setProperty('--theme-switch-y', `${y}px`)
  root.style.setProperty('--theme-switch-radius', `${endRadius}px`)

  if (typeof transitionApi !== 'function') {
    runFallbackThemeSwitch()
    return
  }

  root.classList.add('theme-switching')

  try {
    const transition = transitionApi(async () => {
      themeStore.toggleTheme()
      await nextTick()
    })

    void transition.finished.finally(() => {
      clearThemeTransitionState()
    })
  }
  catch {
    clearThemeTransitionState()
    themeStore.toggleTheme()
  }
}

function clearThemeTransitionState() {
  const root = document.documentElement

  root.classList.remove('theme-switching')
  root.style.removeProperty('--theme-switch-x')
  root.style.removeProperty('--theme-switch-y')
  root.style.removeProperty('--theme-switch-radius')
  fallbackThemeOverlay?.remove()
  fallbackThemeOverlay = null
}

function runFallbackThemeSwitch() {
  const overlay = document.createElement('div')

  fallbackThemeOverlay?.remove()
  fallbackThemeOverlay = overlay
  overlay.className = 'theme-switch-fallback'
  overlay.style.setProperty('--theme-switch-fallback-bg', isDark.value ? '#F5F7FB' : '#0F172A')
  document.body.appendChild(overlay)

  overlay.addEventListener('animationend', () => {
    themeStore.toggleTheme()

    requestAnimationFrame(() => {
      overlay.classList.add('theme-switch-fallback--done')
    })
  }, { once: true })

  overlay.addEventListener('transitionend', () => {
    if (fallbackThemeOverlay === overlay) {
      fallbackThemeOverlay = null
    }

    overlay.remove()
    clearThemeTransitionState()
  }, { once: true })
}

onBeforeUnmount(() => {
  clearThemeTransitionState()
})

onMounted(() => {
  void Promise.all([
    appStore.loadSettings(),
    routerStore.loadRoutes(),
  ]).catch((error: unknown) => {
    bootstrapError.value = error instanceof Error ? error.message : '系统初始化失败'
  })
})

watch(
  [appTitle, activeRoute],
  ([currentAppTitle, currentRoute]) => {
    appStore.applyDocumentTitle(currentRoute?.mate.title || currentAppTitle)
  },
  { immediate: true },
)
</script>

<template>
  <n-config-provider
    :date-locale="dateZhCN"
    :locale="zhCN"
    :theme="theme"
    :theme-overrides="themeOverrides"
  >
    <n-message-provider>
      <n-dialog-provider>
        <n-notification-provider>
          <div
            v-if="isBootstrapping"
            class="app-loading-screen"
            :class="{ 'app-loading-screen--dark': isDark }"
            :style="shellVars"
          >
            <div class="app-loading-panel">
              <div class="app-loading-panel__mark">
                <img
                  v-if="logo"
                  :src="logo"
                  :alt="appTitle"
                  class="app-loading-panel__logo"
                >
                <template v-else>
                  {{ markText }}
                </template>
              </div>

              <div class="app-loading-panel__body">
                <strong>加载中</strong>
                <span>正在准备系统资源</span>
              </div>

              <n-spin size="small" />
            </div>
          </div>

          <div v-else-if="bootstrapError" class="app-loading-screen" :style="shellVars">
            <n-card class="startup-error-card" title="页面初始化失败">
              <p class="startup-error-card__text">{{ bootstrapError }}</p>
            </n-card>
          </div>

          <template v-else>
            <n-layout
              class="admin-shell"
              :class="{
                'admin-shell--dark': isDark,
                'admin-shell--collapsed': collapsed,
              }"
              :style="shellVars"
              has-sider
            >
              <n-layout-sider
                bordered
                class="admin-sider"
                :class="{ 'admin-sider--collapsed': collapsed }"
                collapse-mode="width"
                :collapsed="collapsed"
                :collapsed-width="56"
                :native-scrollbar="false"
                :width="248"
              >
                <div class="brand" :class="{ 'brand--collapsed': collapsed }">
                  <div class="brand__mark">
                    <img
                      v-if="logo"
                      :src="logo"
                      :alt="appTitle"
                      class="brand__logo"
                    >
                    <template v-else>
                      {{ markText }}
                    </template>
                  </div>
                  <div class="brand__text" :class="{ 'brand__text--hidden': collapsed }">
                    <strong>{{ appTitle }}</strong>
                  </div>
                </div>

                <n-menu
                  class="admin-menu"
                  :class="{
                    'admin-menu--collapsed': collapsed,
                    'admin-menu--labels-hidden': collapsed,
                  }"
                  :collapsed="false"
                  :collapsed-icon-size="20"
                  :collapsed-width="56"
                  :expanded-keys="mergedExpandedMenuKeys"
                  :options="menuOptions"
                  :value="activeMenu"
                  @update:expanded-keys="handleExpandedKeysUpdate"
                  @update:value="handleMenuUpdate"
                />
              </n-layout-sider>

              <n-layout class="admin-main">
                <n-layout-header bordered class="admin-header">
                  <div class="header-left">
                    <button
                      class="icon-button"
                      type="button"
                      @click="toggleCollapsed"
                    >
                      <span
                        class="collapse-toggle-icon"
                        :class="{ 'collapse-toggle-icon--collapsed': collapsed }"
                      >
                        <span class="collapse-toggle-icon__item collapse-toggle-icon__item--close">
                          <IconChevronsLeft class="collapse-toggle-icon__svg" />
                        </span>
                        <span class="collapse-toggle-icon__item collapse-toggle-icon__item--open">
                          <IconChevronsRight class="collapse-toggle-icon__svg" />
                        </span>
                      </span>
                    </button>

                    <n-breadcrumb class="header-breadcrumb">
                      <n-breadcrumb-item :clickable="false">
                        <n-icon :component="IconHome" />
                        工作台
                      </n-breadcrumb-item>
                      <n-breadcrumb-item class="breadcrumb-extra">总览</n-breadcrumb-item>
                    </n-breadcrumb>
                  </div>

                  <div class="header-actions">
                    <n-input class="header-search" placeholder="搜索菜单或页面" round size="small">
                      <template #prefix>
                        <n-icon :component="IconSearch" />
                      </template>
                    </n-input>

                    <span ref="themeToggleAnchor" class="theme-toggle-anchor">
                      <n-button quaternary circle size="small" @click="toggleTheme($event)">
                        <template #icon>
                          <n-icon>
                            <IconSun v-if="isDark" />
                            <IconMoon v-else />
                          </n-icon>
                        </template>
                      </n-button>
                    </span>

                    <n-button class="help-action" quaternary circle size="small">
                      <template #icon>
                        <n-icon :component="IconHelpCircle" />
                      </template>
                    </n-button>

                    <n-badge :value="3" processing>
                      <n-button quaternary circle size="small">
                        <template #icon>
                          <n-icon :component="IconBell" />
                        </template>
                      </n-button>
                    </n-badge>

                    <n-dropdown
                      :options="[
                        { label: '个人中心', key: 'profile' },
                        { label: '系统设置', key: 'settings' },
                        { type: 'divider', key: 'divider' },
                        { label: '退出登录', key: 'logout' },
                      ]"
                    >
                      <button class="user-entry" type="button">
                        <n-avatar round size="small">A</n-avatar>
                        <span class="user-entry__name">Admin</span>
                        <n-icon :component="IconChevronDown" />
                      </button>
                    </n-dropdown>
                  </div>
                </n-layout-header>

                <div class="tab-strip">
                  <n-tabs
                    :value="activeTab"
                    animated
                    @close="handleCloseTab"
                    size="small"
                    type="card"
                    @update:value="routerStore.setActiveTab($event)"
                  >
                    <n-tab-pane
                      v-for="tab in tabs"
                      :key="tab.key"
                      :closable="!tab.fixed"
                      :name="tab.fullPath"
                      :tab="tab.label"
                    />
                  </n-tabs>
                </div>

                <n-layout-content class="admin-content" :native-scrollbar="false">
                  <div class="workspace">
                    <template v-if="activeRoute">
                      <section v-if="isIframeRoute" class="iframe-page">
                        <div class="iframe-page__header">
                          <div>
                            <p>{{ activeRouteFullPath }}</p>
                            <h1>{{ activeRoute.mate.title }}</h1>
                          </div>
                          <n-tag size="small" type="info">
                            iframe
                          </n-tag>
                        </div>
                        <iframe
                          class="iframe-page__frame"
                          :src="activeRouteLink"
                          :title="activeRoute.mate.title"
                        />
                      </section>

                      <template v-else>
                        <KeepAlive>
                          <component
                            :is="activeViewComponent"
                            v-if="activeRoute.mate.keepAlive && activeViewComponent"
                            :key="activeRouteCacheKey"
                            :full-path="activeRouteFullPath"
                            :params="activeRouteParams"
                            :route="activeRoute"
                          />
                        </KeepAlive>

                        <component
                          :is="activeViewComponent"
                          v-if="!activeRoute.mate.keepAlive && activeViewComponent"
                          :key="activeRouteFullPath"
                          :full-path="activeRouteFullPath"
                          :params="activeRouteParams"
                          :route="activeRoute"
                        />

                        <n-card v-if="isMissingLocalView" class="missing-view-card" title="页面未找到">
                          <p class="missing-view-card__text">
                            未找到与 <code>{{ activeRoute.component }}</code> 对应的视图文件。
                          </p>
                        </n-card>
                      </template>
                    </template>
                  </div>
                </n-layout-content>
              </n-layout>
            </n-layout>
          </template>
        </n-notification-provider>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<style scoped>
.admin-shell {
  --shell-motion-duration: 0.3s;
  --shell-motion-duration-fast: 0.2s;
  --shell-motion-ease: cubic-bezier(0.22, 1, 0.36, 1);
  --shell-motion-ease-soft: cubic-bezier(0.4, 0, 0.2, 1);
  min-height: 100vh;
}

.app-loading-screen {
  align-items: center;
  background:
    radial-gradient(circle at top, var(--shell-loading-glow), transparent 36%),
    var(--shell-content-bg);
  display: flex;
  justify-content: center;
  min-height: 100vh;
  overflow: hidden;
  padding: 24px;
}

.app-loading-screen--dark {
  background:
    radial-gradient(circle at top, var(--shell-loading-glow), transparent 34%),
    var(--shell-content-bg);
}

.app-loading-panel {
  align-items: center;
  animation: loading-panel-in 0.28s var(--shell-motion-ease-soft);
  backdrop-filter: blur(8px);
  background: var(--shell-surface-bg);
  border: 1px solid var(--shell-surface-border);
  border-radius: 16px;
  box-shadow: var(--shell-surface-shadow);
  display: grid;
  gap: 16px;
  justify-items: center;
  min-width: 240px;
  padding: 28px 24px;
}

.app-loading-panel__mark {
  align-items: center;
  background: linear-gradient(135deg, var(--shell-brand-from) 0%, var(--shell-brand-to) 100%);
  border-radius: 14px;
  box-shadow: 0 14px 30px var(--shell-brand-shadow);
  color: #fff;
  display: inline-flex;
  font-size: 26px;
  font-weight: 800;
  height: 58px;
  justify-content: center;
  letter-spacing: 0;
  line-height: 1;
  position: relative;
  width: 58px;
}

.app-loading-panel__mark::after {
  border: 1px solid rgba(255, 255, 255, 0.22);
  border-radius: 10px;
  content: '';
  inset: 6px;
  position: absolute;
}

.app-loading-panel__logo {
  display: block;
  height: 28px;
  object-fit: contain;
  width: 28px;
}

.app-loading-panel__body {
  display: grid;
  gap: 6px;
  justify-items: center;
  text-align: center;
}

.app-loading-panel__body strong {
  color: var(--shell-heading-color);
  font-size: 16px;
  font-weight: 700;
}

.app-loading-panel__body span {
  color: var(--shell-muted-color);
  font-size: 13px;
}

.app-loading-screen--dark .app-loading-panel {
  background: var(--shell-surface-bg);
  border-color: var(--shell-surface-border);
  box-shadow: var(--shell-surface-shadow);
}

.app-loading-screen--dark .app-loading-panel__body strong {
  color: var(--shell-heading-color);
}

.app-loading-screen--dark .app-loading-panel__body span {
  color: var(--shell-muted-color);
}

@keyframes loading-panel-in {
  from {
    opacity: 0;
    transform: translateY(10px) scale(0.98);
  }

  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.admin-main {
  min-width: 0;
  transition: background-color var(--shell-motion-duration-fast) ease;
}

.admin-sider {
  background: var(--n-color);
  transform: translateZ(0);
  transition:
    width var(--shell-motion-duration) 0.08s var(--shell-motion-ease),
    box-shadow var(--shell-motion-duration-fast) ease;
  will-change: width;
}

.admin-sider--collapsed {
  overflow: hidden;
}

.brand {
  align-items: center;
  display: flex;
  height: 56px;
  justify-content: flex-start;
  overflow: hidden;
  padding: 0 11px;
  position: relative;
  transform-origin: left center;
  white-space: nowrap;
}

.brand--collapsed {
  height: 56px;
  padding: 0 11px;
}

.brand__mark {
  align-items: center;
  background: linear-gradient(135deg, var(--shell-brand-from) 0%, var(--shell-brand-to) 100%);
  border-radius: 8px;
  box-shadow: 0 8px 18px var(--shell-brand-shadow);
  color: #fff;
  display: inline-flex;
  flex: none;
  font-size: 16px;
  font-weight: 800;
  height: 34px;
  justify-content: center;
  line-height: 1;
  width: 34px;
}

.brand__logo {
  display: block;
  height: 18px;
  object-fit: contain;
  width: 18px;
}

.brand__text {
  display: grid;
  left: 58px;
  line-height: 1.15;
  opacity: 1;
  pointer-events: none;
  position: absolute;
  overflow: hidden;
  top: 50%;
  transform: translateY(-50%);
  transform-origin: left center;
  width: 118px;
  will-change: opacity, transform;
  transition:
    opacity 0.1s ease,
    transform 0.16s var(--shell-motion-ease);
}

.brand__text--hidden {
  opacity: 0;
  transform: translateY(-50%) translateX(-6px);
}

.brand__text strong {
  display: block;
  font-size: 16px;
  font-weight: 800;
  letter-spacing: 0;
}

.admin-menu {
  padding: 8px;
  transition: padding var(--shell-motion-duration) var(--shell-motion-ease);
}

.admin-menu :deep(.n-menu-item-content) {
  transition:
    background-color var(--shell-motion-duration-fast) ease,
    border-radius var(--shell-motion-duration) var(--shell-motion-ease),
    box-shadow var(--shell-motion-duration-fast) ease,
    color var(--shell-motion-duration-fast) ease,
    margin var(--shell-motion-duration) var(--shell-motion-ease),
    padding var(--shell-motion-duration) var(--shell-motion-ease),
    transform var(--shell-motion-duration-fast) var(--shell-motion-ease-soft),
    width var(--shell-motion-duration) var(--shell-motion-ease);
}

.admin-menu :deep(.n-menu-item-content::before) {
  transition: opacity var(--shell-motion-duration-fast) ease;
}

.admin-menu :deep(.n-menu-item-content-header) {
  max-width: 100px;
  overflow: hidden;
  transform: translateX(0);
  transform-origin: left center;
  transition:
    max-width 0.14s var(--shell-motion-ease),
    opacity 0.1s ease,
    transform 0.14s var(--shell-motion-ease);
  white-space: nowrap;
}

.admin-menu :deep(.n-menu-item-content__icon) {
  transition:
    color var(--shell-motion-duration-fast) ease,
    left var(--shell-motion-duration) var(--shell-motion-ease),
    top var(--shell-motion-duration) var(--shell-motion-ease),
    transform var(--shell-motion-duration) var(--shell-motion-ease);
}

.admin-menu :deep(.n-menu-item-content__arrow) {
  transition:
    opacity 0.1s ease,
    transform 0.14s var(--shell-motion-ease);
}

.admin-menu--collapsed {
  padding: 4px 6px 0;
}

.admin-menu--collapsed :deep(.n-menu-item-content) {
  align-items: center;
  border-radius: 8px;
  display: flex;
  height: 40px;
  margin: 6px auto;
  padding: 0;
  position: relative;
  width: 40px;
}

.admin-menu--labels-hidden :deep(.n-menu-item-content::before) {
  opacity: 0;
}

.admin-menu--labels-hidden :deep(.n-menu-item-content-header) {
  max-width: 0;
  opacity: 0;
  transform: translateX(-10px);
}

.admin-menu--labels-hidden :deep(.n-menu-item-content__arrow) {
  opacity: 0;
  transform: translateX(-6px);
}

.admin-menu--collapsed :deep(.n-menu-item-content__icon) {
  align-items: center;
  display: flex;
  flex: none;
  font-size: 20px;
  height: 20px;
  justify-content: center;
  left: 50%;
  margin: 0;
  position: absolute;
  top: 50%;
  transform: translate(-50%, -50%);
  width: 20px;
}

.admin-menu--collapsed :deep(.n-menu-item-content__icon .n-icon),
.admin-menu--collapsed :deep(.n-menu-item-content__icon svg) {
  display: block;
  height: 20px;
  line-height: 1;
  margin: 0;
  width: 20px;
}

.admin-menu--collapsed :deep(.n-menu-item-content--selected) {
  background: var(--shell-selected-bg);
  box-shadow:
    inset 0 0 0 1px var(--shell-selected-border),
    0 8px 18px var(--shell-selected-shadow);
  color: var(--shell-selected-text);
}

.admin-menu--collapsed :deep(.n-menu-item-content--selected .n-menu-item-content__icon),
.admin-menu--collapsed :deep(.n-menu-item-content--selected .n-icon),
.admin-menu--collapsed :deep(.n-menu-item-content--selected svg) {
  color: var(--shell-selected-text);
}

.admin-menu--collapsed :deep(.n-menu-item-content:hover) {
  background: var(--shell-hover-bg);
  transform: translateY(-1px);
}

.admin-menu :deep(.n-menu-item-content--selected:hover) {
  box-shadow: 0 10px 24px var(--shell-selected-hover-shadow);
}

.admin-menu :deep(.n-menu-item-content--selected:hover::before) {
  box-shadow: 0 10px 24px var(--shell-selected-hover-shadow);
}

.admin-menu--collapsed :deep(.n-menu-item-content--selected:hover) {
  background: var(--shell-selected-hover-bg);
  box-shadow:
    inset 0 0 0 1px var(--shell-selected-border),
    0 12px 24px var(--shell-selected-hover-shadow);
}

.collapse-toggle-icon {
  align-items: center;
  display: inline-flex;
  height: 18px;
  justify-content: center;
  line-height: 0;
  position: relative;
  width: 18px;
}

.collapse-toggle-icon__item {
  align-items: center;
  display: flex;
  justify-content: center;
  inset: 0;
  position: absolute;
  transition:
    opacity 0.16s ease,
    transform 0.22s var(--shell-motion-ease);
}

.collapse-toggle-icon__svg {
  display: block;
  height: 18px;
  width: 18px;
}

.collapse-toggle-icon__item--close {
  opacity: 1;
  transform: scale(1) rotate(0deg);
}

.collapse-toggle-icon__item--open {
  opacity: 0;
  transform: scale(0.82) rotate(-12deg);
}

.collapse-toggle-icon--collapsed .collapse-toggle-icon__item--close {
  opacity: 0;
  transform: scale(0.82) rotate(12deg);
}

.collapse-toggle-icon--collapsed .collapse-toggle-icon__item--open {
  opacity: 1;
  transform: scale(1) rotate(0deg);
}

.icon-button {
  align-self: center;
  align-items: center;
  appearance: none;
  background: transparent;
  border: 0;
  border-radius: 8px;
  color: var(--shell-icon-color);
  cursor: pointer;
  display: inline-flex;
  height: 36px;
  justify-content: center;
  line-height: 0;
  padding: 0;
  transition: transform 0.18s ease;
  width: 36px;
}

.icon-button:hover {
  transform: none;
}

.icon-button:active {
  transform: translateY(0);
}

.icon-button:focus,
.icon-button:focus-visible {
  outline: none;
}

.admin-header {
  align-items: center;
  display: flex;
  height: 56px;
  justify-content: space-between;
  padding: 0 16px;
  transition:
    background-color var(--shell-motion-duration-fast) ease,
    border-color var(--shell-motion-duration-fast) ease;
}

.header-left,
.header-actions,
.user-entry {
  align-items: center;
  display: flex;
}

.header-left {
  align-items: center;
  height: 36px;
  gap: 12px;
}

.header-breadcrumb {
  align-items: center;
  display: flex;
  min-height: 36px;
}

.header-breadcrumb :deep(ul) {
  align-items: center;
  display: flex;
}

.header-breadcrumb :deep(.n-breadcrumb-item) {
  align-items: center;
}

.header-breadcrumb :deep(.n-breadcrumb-item__link) {
  align-items: center;
  display: inline-flex;
  gap: 6px;
  line-height: 1;
  min-height: 36px;
  padding: 0;
}

.header-breadcrumb :deep(.n-breadcrumb-item__link:hover),
.header-breadcrumb :deep(.n-breadcrumb-item__link:active) {
  background-color: transparent;
  color: inherit;
}

.header-breadcrumb :deep(.n-icon),
.header-breadcrumb :deep(svg) {
  display: block;
  flex: none;
  height: 16px;
  width: 16px;
}

.header-breadcrumb :deep(.n-breadcrumb-item__separator) {
  align-self: center;
}

.header-actions {
  gap: 10px;
}

.theme-toggle-anchor {
  display: inline-flex;
}

.header-search {
  width: 240px;
}

.user-entry {
  background: transparent;
  border: 0;
  border-radius: 6px;
  color: var(--n-text-color);
  cursor: pointer;
  gap: 8px;
  height: 34px;
  padding: 0 8px;
  transition: background-color 0.2s ease;
}

.user-entry:hover {
  background: var(--n-action-color);
}

.tab-strip {
  background: var(--n-color);
  border-bottom: 1px solid var(--n-border-color);
  padding: 6px 12px 0;
  transition:
    background-color var(--shell-motion-duration-fast) ease,
    border-color var(--shell-motion-duration-fast) ease;
}

.admin-content {
  background: var(--shell-content-bg);
  height: calc(100vh - 97px);
  transition: background-color var(--shell-motion-duration-fast) ease;
}

.workspace {
  margin: 0 auto;
  max-width: 1440px;
  padding: 20px;
  transition: padding var(--shell-motion-duration-fast) ease;
}

.iframe-page {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: calc(100vh - 137px);
}

.iframe-page__header {
  align-items: center;
  display: flex;
  justify-content: space-between;
  gap: 16px;
}

.iframe-page__header p {
  color: var(--shell-muted-color);
  font-size: 12px;
  margin: 0 0 4px;
  text-transform: uppercase;
}

.iframe-page__header h1 {
  color: var(--shell-heading-color);
  font-size: 24px;
  line-height: 1.2;
  margin: 0;
}

.iframe-page__frame {
  background: var(--shell-iframe-bg);
  border: 1px solid var(--n-border-color);
  border-radius: 8px;
  flex: 1;
  min-height: 720px;
  width: 100%;
}

.missing-view-card :deep(.n-card__content) {
  padding: 18px;
}

.startup-error-card {
  max-width: 420px;
  width: 100%;
}

.startup-error-card :deep(.n-card__content) {
  padding: 18px;
}

.startup-error-card__text {
  color: var(--shell-subtle-color);
  margin: 0;
}

.missing-view-card__text {
  color: var(--shell-subtle-color);
  margin: 0;
}

.page-heading {
  align-items: center;
  display: flex;
  justify-content: space-between;
  margin-bottom: 16px;
}

.page-heading p {
  color: var(--shell-muted-color);
  font-size: 12px;
  letter-spacing: 0;
  margin: 0 0 4px;
  text-transform: uppercase;
}

.page-heading h1 {
  color: var(--shell-heading-color);
  font-size: 24px;
  line-height: 1.2;
  margin: 0;
}

.stats-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  margin-bottom: 12px;
}

.metric-card :deep(.n-card__content) {
  padding: 16px;
}

.metric-card__label {
  color: var(--shell-muted-color);
  font-size: 13px;
}

.metric-card__body {
  align-items: end;
  display: flex;
  gap: 10px;
  justify-content: space-between;
  margin-top: 10px;
}

.metric-card__body strong {
  color: var(--shell-heading-color);
  font-size: 28px;
  line-height: 1;
}

.content-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: minmax(0, 1.4fr) minmax(320px, 0.6fr);
}

.health-row {
  display: grid;
  gap: 8px;
}

.health-row span {
  color: var(--shell-subtle-color);
  font-size: 13px;
}

@media (max-width: 900px) {
  .header-search {
    display: none;
  }

  .stats-grid,
  .content-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .admin-sider {
    display: none;
  }

  .admin-header {
    gap: 8px;
    padding: 0 10px;
  }

  .header-left,
  .header-actions {
    gap: 8px;
    min-width: 0;
  }

  .header-breadcrumb {
    min-width: 0;
  }

  .breadcrumb-extra,
  .help-action,
  .user-entry__name {
    display: none;
  }

  .user-entry {
    gap: 4px;
    padding: 0 4px;
  }

  .page-heading {
    align-items: flex-start;
    flex-direction: column;
    gap: 12px;
  }

  .workspace {
    padding: 14px;
  }
}
</style>

<style>
:root {
  --theme-switch-x: calc(100vw - 72px);
  --theme-switch-y: 72px;
  --theme-switch-radius: 0px;
}

:root.theme-switching::view-transition-old(root),
:root.theme-switching::view-transition-new(root) {
  animation: none;
  inset: 0;
  mix-blend-mode: normal;
}

:root.theme-switching::view-transition-group(root),
:root.theme-switching::view-transition-image-pair(root) {
  animation: none;
}

:root.theme-switching::view-transition-old(root) {
  opacity: 1;
  z-index: 0;
}

:root.theme-switching::view-transition-new(root) {
  clip-path: circle(0 at var(--theme-switch-x) var(--theme-switch-y));
  z-index: 1;
  animation: theme-reveal 0.92s cubic-bezier(0.22, 1, 0.36, 1) forwards;
}

.theme-switch-fallback {
  animation: theme-reveal 0.76s cubic-bezier(0.22, 1, 0.36, 1) forwards;
  background: var(--theme-switch-fallback-bg);
  clip-path: circle(0 at var(--theme-switch-x) var(--theme-switch-y));
  inset: 0;
  opacity: 1;
  pointer-events: none;
  position: fixed;
  transition: opacity 0.24s ease;
  z-index: 2147483647;
}

.theme-switch-fallback--done {
  opacity: 0;
}

@keyframes theme-reveal {
  from {
    clip-path: circle(0 at var(--theme-switch-x) var(--theme-switch-y));
  }

  to {
    clip-path: circle(var(--theme-switch-radius) at var(--theme-switch-x) var(--theme-switch-y));
  }
}
</style>
