<script setup lang="ts">
import {
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
  IconSettings,
  IconShieldCheck,
  IconSun,
  IconUsers,
} from '@tabler/icons-vue'
import type { MenuOption } from 'naive-ui'
import { darkTheme, dateZhCN, lightTheme, NIcon, zhCN, type GlobalThemeOverrides } from 'naive-ui'

import type { LayoutMenuItem, LayoutTab } from '../types/menu'

const collapsed = ref(false)
const isDark = ref(false)
const activeMenu = ref('/dashboard')
const activeTab = ref('/dashboard')

const theme = computed(() => (isDark.value ? darkTheme : lightTheme))

const themeOverrides: GlobalThemeOverrides = {
  common: {
    borderRadius: '6px',
    borderRadiusSmall: '4px',
    primaryColor: '#2563EB',
    primaryColorHover: '#1D4ED8',
    primaryColorPressed: '#1E40AF',
  },
  Button: {
    borderRadiusMedium: '6px',
  },
  Menu: {
    itemBorderRadius: '6px',
  },
}

const menuItems: LayoutMenuItem[] = [
  {
    key: '/dashboard',
    label: '工作台',
    icon: IconLayoutDashboard,
  },
  {
    key: '/system',
    label: '系统管理',
    icon: IconSettings,
    children: [
      {
        key: '/system/user',
        label: '用户管理',
        icon: IconUsers,
      },
      {
        key: '/system/role',
        label: '角色权限',
        icon: IconShieldCheck,
      },
    ],
  },
  {
    key: '/basic',
    label: '基础数据',
    icon: IconDatabase,
    badge: '12',
  },
  {
    key: '/document',
    label: '文档中心',
    icon: IconFileText,
  },
]

const tabs = ref<LayoutTab[]>([
  {
    key: '/dashboard',
    label: '工作台',
    fixed: true,
  },
  {
    key: '/system/user',
    label: '用户管理',
  },
])

const quickStats = [
  {
    label: '在线用户',
    value: '1,248',
    trend: '+12.4%',
  },
  {
    label: '今日请求',
    value: '86,420',
    trend: '+8.1%',
  },
  {
    label: '异常告警',
    value: '7',
    trend: '-3',
  },
]

const tasks = [
  {
    title: '角色权限矩阵调整',
    desc: '系统管理 / 角色权限',
    status: '进行中',
  },
  {
    title: '菜单路由协议联调',
    desc: '动态路由 / Mock API',
    status: '待确认',
  },
  {
    title: '登录过期处理',
    desc: 'HTTP 拦截器 / 401',
    status: '待开发',
  },
]

function renderIcon(icon?: Component) {
  if (!icon) {
    return undefined
  }

  return () => h(NIcon, null, { default: () => h(icon) })
}

function toMenuOptions(items: LayoutMenuItem[]): MenuOption[] {
  return items.map((item) => ({
    key: item.key,
    label: item.label,
    icon: renderIcon(item.icon),
    children: item.children ? toMenuOptions(item.children) : undefined,
  }))
}

const menuOptions = computed(() => toMenuOptions(menuItems))

function handleMenuUpdate(key: string) {
  activeMenu.value = key
  activeTab.value = key

  const target = findMenuItem(menuItems, key)

  if (target && !tabs.value.some((tab) => tab.key === key)) {
    tabs.value.push({
      key,
      label: target.label,
    })
  }
}

function findMenuItem(items: LayoutMenuItem[], key: string): LayoutMenuItem | undefined {
  for (const item of items) {
    if (item.key === key) {
      return item
    }

    const child = item.children ? findMenuItem(item.children, key) : undefined

    if (child) {
      return child
    }
  }

  return undefined
}

function handleCloseTab(key: string) {
  const tab = tabs.value.find((item) => item.key === key)

  if (tab?.fixed) {
    return
  }

  tabs.value = tabs.value.filter((item) => item.key !== key)

  if (activeTab.value === key) {
    const nextTab = tabs.value[tabs.value.length - 1]

    if (nextTab) {
      activeTab.value = nextTab.key
      activeMenu.value = nextTab.key
    }
  }
}

function toggleCollapsed() {
  collapsed.value = !collapsed.value
}
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
          <n-layout
            class="admin-shell"
            :class="{
              'admin-shell--dark': isDark,
              'admin-shell--collapsed': collapsed,
            }"
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
                  F
                </div>
                <div class="brand__text" :class="{ 'brand__text--hidden': collapsed }">
                  <strong>Fox Admin</strong>
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
                :options="menuOptions"
                :value="activeMenu"
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

                  <n-button quaternary circle size="small" @click="isDark = !isDark">
                    <template #icon>
                      <n-icon>
                        <IconSun v-if="isDark" />
                        <IconMoon v-else />
                      </n-icon>
                    </template>
                  </n-button>

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
                  v-model:value="activeTab"
                  animated
                  size="small"
                  type="card"
                  @update:value="activeMenu = $event"
                >
                  <n-tab-pane
                    v-for="tab in tabs"
                    :key="tab.key"
                    :closable="!tab.fixed"
                    :name="tab.key"
                    :tab="tab.label"
                    @close="handleCloseTab(tab.key)"
                  />
                </n-tabs>
              </div>

              <n-layout-content class="admin-content" :native-scrollbar="false">
                <div class="workspace">
                  <section class="page-heading">
                    <div>
                      <p>Overview</p>
                      <h1>工作台</h1>
                    </div>
                    <n-space>
                      <n-button secondary>导出</n-button>
                      <n-button type="primary">新建任务</n-button>
                    </n-space>
                  </section>

                  <section class="stats-grid">
                    <n-card v-for="item in quickStats" :key="item.label" class="metric-card">
                      <div class="metric-card__label">{{ item.label }}</div>
                      <div class="metric-card__body">
                        <strong>{{ item.value }}</strong>
                        <n-tag :type="item.trend.startsWith('-') ? 'warning' : 'success'" size="small">
                          {{ item.trend }}
                        </n-tag>
                      </div>
                    </n-card>
                  </section>

                  <section class="content-grid">
                    <n-card title="待办事项">
                      <n-list hoverable>
                        <n-list-item v-for="item in tasks" :key="item.title">
                          <n-thing :description="item.desc" :title="item.title">
                            <template #header-extra>
                              <n-tag size="small">{{ item.status }}</n-tag>
                            </template>
                          </n-thing>
                        </n-list-item>
                      </n-list>
                    </n-card>

                    <n-card title="系统状态">
                      <n-space vertical :size="18">
                        <div class="health-row">
                          <span>API 网关</span>
                          <n-progress :percentage="98" :show-indicator="false" type="line" />
                        </div>
                        <div class="health-row">
                          <span>任务队列</span>
                          <n-progress :percentage="76" :show-indicator="false" type="line" />
                        </div>
                        <div class="health-row">
                          <span>缓存命中</span>
                          <n-progress :percentage="91" :show-indicator="false" type="line" />
                        </div>
                      </n-space>
                    </n-card>
                  </section>
                </div>
              </n-layout-content>
            </n-layout>
          </n-layout>
        </n-notification-provider>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<style scoped>
.admin-shell {
  --shell-content-bg: #f6f8fb;
  --shell-heading-color: #0f172a;
  --shell-motion-duration: 0.3s;
  --shell-motion-duration-fast: 0.2s;
  --shell-motion-ease: cubic-bezier(0.22, 1, 0.36, 1);
  --shell-motion-ease-soft: cubic-bezier(0.4, 0, 0.2, 1);
  --shell-muted-color: #64748b;
  --shell-subtle-color: #475569;
  min-height: 100vh;
}

.admin-shell--dark {
  --shell-content-bg: #111827;
  --shell-heading-color: #f8fafc;
  --shell-muted-color: #94a3b8;
  --shell-subtle-color: #cbd5e1;
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
  background: linear-gradient(135deg, #7c5cff 0%, #5b3df5 100%);
  border-radius: 8px;
  box-shadow: 0 8px 18px rgb(91 61 245 / 28%);
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
  background: #eef4ff;
  box-shadow: inset 0 0 0 1px rgb(37 99 235 / 8%);
  color: #2563eb;
}

.admin-menu--collapsed :deep(.n-menu-item-content:hover) {
  background: #f3f6fb;
  transform: translateY(-1px);
}

.admin-menu--collapsed :deep(.n-menu-item-content--selected:hover) {
  background: #e8f1ff;
}

.admin-shell--dark .admin-menu--collapsed :deep(.n-menu-item-content--selected) {
  background: rgb(37 99 235 / 18%);
  color: #93c5fd;
}

.admin-shell--dark .admin-menu--collapsed :deep(.n-menu-item-content:hover) {
  background: rgb(148 163 184 / 12%);
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
  color: #334155;
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

.admin-shell--dark .icon-button {
  color: #e2e8f0;
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
