<script setup lang="ts">
import { h } from 'vue'
import { ref } from 'vue'
import type { Component } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import type { RouteLocationNormalizedLoaded } from 'vue-router'
import type { DropdownOption, MenuOption } from 'naive-ui'
import { NIcon, useMessage } from 'naive-ui'
import {
  ChevronDownOutline,
  CheckmarkOutline,
  CloseOutline,
  GridOutline,
  HomeOutline,
  LogOutOutline,
  MenuOutline,
  OptionsOutline,
  PersonOutline,
  SettingsOutline,
} from '@vicons/ionicons5'
import { useAuthStore } from '../stores/auth'
import { useTabsStore } from '../stores/tabs'
import { useThemeStore } from '../stores/theme'
import type { MockAdminRoute, MockRouteIcon } from '../mock/admin-routes'
import { mockAdminRoutes } from '../mock/admin-routes'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const authStore = useAuthStore()
const tabsStore = useTabsStore()
const themeStore = useThemeStore()
const collapsed = ref(false)
const activeCollapsedPopover = ref('')
const showSettingsDrawer = ref(false)
const isDarkNavigation = computed(() => themeStore.navigationStyle === 'dark')
const iconMap: Record<MockRouteIcon, Component> = {
  GridOutline,
  HomeOutline,
  ListOutline: GridOutline,
  OptionsOutline,
}

function renderIcon(icon: Component) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

function toMenuOption(item: MockAdminRoute): MenuOption {
  if (item.children?.length) {
    return {
      label: item.title,
      key: item.path,
      icon: item.icon ? renderIcon(iconMap[item.icon]) : undefined,
      children: item.children.map(toMenuOption),
    }
  }

  if (item.externalUrl) {
    return {
      label: () =>
        h(
          'a',
          {
            class: 'admin-menu-external-link',
            href: item.externalUrl,
            onClick: handleExternalLinkClick,
            onMousedown: handleExternalLinkMouseDown,
            rel: 'noreferrer',
            target: '_blank',
          },
          item.title,
        ),
      key: item.path,
      icon: item.icon ? renderIcon(iconMap[item.icon]) : undefined,
    }
  }

  return {
    label: () => h(RouterLink, { to: item.path }, { default: () => item.title }),
    key: item.path,
    icon: item.icon ? renderIcon(iconMap[item.icon]) : undefined,
    children: item.children?.map(toMenuOption),
  }
}

const menuOptions: MenuOption[] = mockAdminRoutes.map(toMenuOption)
function isCollapsedActive(item: MockAdminRoute) {
  if (route.path === item.path) {
    return true
  }

  return item.children?.some((child) => route.path.startsWith(child.path)) ?? false
}

function openExternalLink(url: string) {
  window.open(url, '_blank', 'noopener,noreferrer')
}

function handleExternalLinkClick(event: MouseEvent) {
  ;(event.currentTarget as HTMLElement).blur()
}

function handleExternalLinkMouseDown(event: MouseEvent) {
  event.preventDefault()
}

function getCollapsedIcon(item: MockAdminRoute) {
  return item.icon ? iconMap[item.icon] : undefined
}

function handleCollapsedMenuClick(item: MockAdminRoute) {
  if (item.externalUrl) {
    openExternalLink(item.externalUrl)
    return
  }

  if (!item.children?.length) {
    router.push(item.path)
  }
}

function handleCollapsedChildClick(item: MockAdminRoute) {
  activeCollapsedPopover.value = ''

  if (item.externalUrl) {
    openExternalLink(item.externalUrl)
    return
  }

  router.push(item.path)
}

const userDropdownOptions: DropdownOption[] = [
  {
    key: 'profile',
    type: 'render',
    render: () =>
      h('div', { class: 'user-dropdown-profile' }, [
        h('div', { class: 'user-dropdown-profile__name' }, authStore.username || 'admin'),
        h('div', { class: 'user-dropdown-profile__role' }, '超级管理员'),
      ]),
  },
  {
    key: 'divider',
    type: 'divider',
  },
  {
    label: '个人设置',
    key: 'settings',
    icon: renderIcon(PersonOutline),
  },
  {
    label: '注销登录',
    key: 'logout',
    icon: renderIcon(LogOutOutline),
  },
]
const defaultExpandedKeys = computed(() => {
  const expanded = mockAdminRoutes
    .filter((item) => item.children?.some((child) => route.path.startsWith(child.path)))
    .map((item) => item.path)

  return expanded.length > 0 ? expanded : ['/system']
})

function handleLogout() {
  authStore.clearSession()
  tabsStore.resetTabs()
  router.replace('/login')
}

function handleUserDropdownSelect(key: string | number) {
  if (key === 'logout') {
    handleLogout()
    return
  }

  if (key === 'settings') {
    message.info('个人设置暂未开放')
  }
}

function handleTabClick(path: string) {
  if (path !== route.path) {
    router.push(path)
  }
}

function handleTabClose(path: string) {
  const currentIndex = tabsStore.tabs.findIndex((tab) => tab.path === path)
  const isActiveTab = route.path === path

  tabsStore.removeTab(path)

  if (!isActiveTab) {
    return
  }

  const nextTab = tabsStore.tabs[currentIndex] ?? tabsStore.tabs[currentIndex - 1] ?? tabsStore.tabs[0]
  router.push(nextTab.path)
}

function getPageCacheKey(currentRoute: RouteLocationNormalizedLoaded) {
  return String(currentRoute.meta.cacheKey ?? currentRoute.fullPath)
}

watch(
  () => route.fullPath,
  () => tabsStore.addRouteTab(route),
  { immediate: true },
)
</script>

<template>
  <n-layout :class="['admin-shell', `admin-shell--nav-${themeStore.navigationStyle}`]" has-sider>
    <n-layout-sider
      :class="[
        'admin-sider',
        {
          'admin-sider--collapsed': collapsed,
          'admin-sider--dark': isDarkNavigation,
        },
      ]"
      bordered
      collapse-mode="width"
      :collapsed="collapsed"
      :collapsed-width="64"
      :inverted="isDarkNavigation"
      :width="220"
      @collapse="collapsed = true"
      @expand="collapsed = false"
    >
      <div class="brand" :class="{ 'brand--collapsed': collapsed }">
        <div class="brand__mark">F</div>
        <span v-if="!collapsed">Fox Admin</span>
      </div>

      <div v-if="collapsed" class="admin-collapse-menu">
        <template v-for="item in mockAdminRoutes" :key="item.path">
          <n-popover
            v-if="item.children?.length"
            placement="right-start"
            trigger="hover"
            :show-arrow="false"
            :show="activeCollapsedPopover === item.path"
            :class="['admin-collapse-popover', { 'admin-collapse-popover--dark': isDarkNavigation }]"
            @update:show="(value) => (activeCollapsedPopover = value ? item.path : '')"
          >
            <template #trigger>
              <button
                class="admin-collapse-menu__item"
                :class="{ 'admin-collapse-menu__item--active': isCollapsedActive(item) }"
                :title="item.title"
                type="button"
              >
                <n-icon v-if="getCollapsedIcon(item)" :component="getCollapsedIcon(item)" />
              </button>
            </template>

            <div class="admin-collapse-submenu" :class="{ 'admin-collapse-submenu--dark': isDarkNavigation }">
              <button
                v-for="child in item.children"
                :key="child.path"
                class="admin-collapse-submenu__item"
                :class="{ 'admin-collapse-submenu__item--active': route.path === child.path }"
                type="button"
                @click="handleCollapsedChildClick(child)"
              >
                {{ child.title }}
              </button>
            </div>
          </n-popover>

          <button
            v-else
            class="admin-collapse-menu__item"
            :class="{ 'admin-collapse-menu__item--active': isCollapsedActive(item) }"
            :title="item.title"
            type="button"
            @click="handleCollapsedMenuClick(item)"
          >
            <n-icon v-if="getCollapsedIcon(item)" :component="getCollapsedIcon(item)" />
          </button>
        </template>
      </div>

      <n-menu
        v-else
        :collapsed="collapsed"
        :default-expanded-keys="defaultExpandedKeys"
        :inverted="isDarkNavigation"
        :options="menuOptions"
        :value="route.path"
      />
    </n-layout-sider>

    <n-layout class="admin-main">
      <n-layout-header bordered class="admin-header">
        <div class="admin-header__left">
          <button
            class="sidebar-toggle"
            title="折叠侧边栏"
            type="button"
            @click="collapsed = !collapsed"
          >
            <n-icon :component="MenuOutline" />
          </button>

          <n-breadcrumb v-if="themeStore.showBreadcrumb">
            <n-breadcrumb-item>后台管理</n-breadcrumb-item>
            <n-breadcrumb-item>{{ route.meta.title }}</n-breadcrumb-item>
          </n-breadcrumb>
        </div>

        <div class="admin-header__user">
          <n-dropdown
            placement="bottom-end"
            trigger="click"
            :options="userDropdownOptions"
            @select="handleUserDropdownSelect"
          >
            <button class="user-trigger" type="button">
              <n-avatar round class="user-avatar">F</n-avatar>
              <n-icon class="user-trigger__arrow" :component="ChevronDownOutline" />
            </button>
          </n-dropdown>

          <n-button circle quaternary size="small" title="系统设置" @click="showSettingsDrawer = true">
            <template #icon>
              <n-icon :component="SettingsOutline" />
            </template>
          </n-button>
        </div>
      </n-layout-header>

      <div v-if="themeStore.showTabs" class="admin-tabs">
        <button
          v-for="tab in tabsStore.tabs"
          :key="tab.path"
          class="admin-tab"
          :class="{ 'admin-tab--active': route.path === tab.path }"
          type="button"
          @click="handleTabClick(tab.path)"
        >
          <span class="admin-tab__dot" />
          <span class="admin-tab__title">{{ tab.title }}</span>
          <span
            v-if="tab.closable"
            class="admin-tab__close"
            role="button"
            tabindex="0"
            @click.stop="handleTabClose(tab.path)"
            @keydown.enter.stop.prevent="handleTabClose(tab.path)"
          >
            <n-icon :component="CloseOutline" />
          </span>
        </button>
      </div>

      <n-layout-content class="admin-content">
        <router-view v-slot="{ Component, route: currentRoute }">
          <keep-alive :include="tabsStore.cachedNames">
            <component :is="Component" :key="getPageCacheKey(currentRoute)" />
          </keep-alive>
        </router-view>
      </n-layout-content>
    </n-layout>

    <n-drawer v-model:show="showSettingsDrawer" :width="360" placement="right">
      <n-drawer-content closable>
        <template #header>
          <span class="theme-drawer-title">项目配置</span>
        </template>

        <div class="theme-drawer">
          <section class="theme-section">
            <div class="theme-section__title">
              <span />
              <strong>主题</strong>
              <span />
            </div>

            <div class="theme-palette-grid">
              <button
                v-for="item in themeStore.themePalettes"
                :key="item.name"
                class="theme-palette"
                :class="{ 'theme-palette--active': themeStore.activeName === item.name }"
                :style="{ '--theme-color': item.primary, '--theme-soft': item.soft }"
                type="button"
                @click="themeStore.setTheme(item.name)"
              >
                <span class="theme-palette__swatch">
                  <n-icon v-if="themeStore.activeName === item.name" :component="CheckmarkOutline" />
                </span>
                <span>{{ item.label }}</span>
              </button>
            </div>
          </section>

          <section class="theme-section">
            <div class="theme-section__title">
              <span />
              <strong>导航风格</strong>
              <span />
            </div>

            <div class="layout-choice-grid">
              <button
                class="layout-choice"
                :class="{ 'layout-choice--active': themeStore.navigationStyle === 'dark' }"
                type="button"
                @click="themeStore.setNavigationStyle('dark')"
              >
                <span class="layout-choice__preview layout-choice__preview--dark">
                  <span class="layout-choice__nav">
                    <span />
                    <span />
                    <span />
                  </span>
                  <span class="layout-choice__content">
                    <span class="layout-choice__bar" />
                    <span class="layout-choice__body">
                      <span />
                      <span />
                    </span>
                  </span>
                  <span v-if="themeStore.navigationStyle === 'dark'" class="layout-choice__check">
                    <n-icon :component="CheckmarkOutline" />
                  </span>
                </span>
                <strong>深色</strong>
                <small>沉稳侧栏</small>
              </button>

              <button
                class="layout-choice"
                :class="{ 'layout-choice--active': themeStore.navigationStyle === 'light' }"
                type="button"
                @click="themeStore.setNavigationStyle('light')"
              >
                <span class="layout-choice__preview layout-choice__preview--light">
                  <span class="layout-choice__nav">
                    <span />
                    <span />
                    <span />
                  </span>
                  <span class="layout-choice__content">
                    <span class="layout-choice__bar" />
                    <span class="layout-choice__body">
                      <span />
                      <span />
                    </span>
                  </span>
                  <span v-if="themeStore.navigationStyle === 'light'" class="layout-choice__check">
                    <n-icon :component="CheckmarkOutline" />
                  </span>
                </span>
                <strong>浅色</strong>
                <small>清爽简洁</small>
              </button>
            </div>
          </section>

          <section class="theme-section">
            <div class="theme-setting-row">
              <div>
                <strong>显示标签页</strong>
                <p>展示已访问页面快捷入口</p>
              </div>
              <n-switch
                :value="themeStore.showTabs"
                @update:value="themeStore.setShowTabs"
              />
            </div>

            <div class="theme-setting-row">
              <div>
                <strong>显示面包屑</strong>
                <p>展示当前页面层级位置</p>
              </div>
              <n-switch
                :value="themeStore.showBreadcrumb"
                @update:value="themeStore.setShowBreadcrumb"
              />
            </div>

            <div class="theme-setting-row">
              <div>
                <strong>紧凑模式</strong>
                <p>减少内容区留白</p>
              </div>
              <n-switch
                :value="themeStore.compactMode"
                @update:value="themeStore.setCompactMode"
              />
            </div>

            <div class="theme-setting-row">
              <div>
                <strong>固定侧边栏</strong>
                <p>保持导航全高显示</p>
              </div>
              <n-switch :value="true" />
            </div>
          </section>
        </div>
      </n-drawer-content>
    </n-drawer>
  </n-layout>
</template>
