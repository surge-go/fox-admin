import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

export type ThemeName = 'blue' | 'navy' | 'cyan' | 'violet'
export type NavigationStyle = 'dark' | 'light'

export type ThemePalette = {
  name: ThemeName
  label: string
  primary: string
  hover: string
  strong: string
  supplement: string
  soft: string
  softHover: string
  surfaceHover: string
  borderActive: string
  focusShadow: string
}

export const themePalettes: ThemePalette[] = [
  {
    name: 'blue',
    label: '科技蓝',
    primary: '#2563eb',
    hover: '#3b82f6',
    strong: '#1d4ed8',
    supplement: '#60a5fa',
    soft: '#eef4ff',
    softHover: '#e8f1ff',
    surfaceHover: '#f6faff',
    borderActive: '#93c5fd',
    focusShadow: 'rgba(37, 99, 235, 0.14)',
  },
  {
    name: 'navy',
    label: '深海蓝',
    primary: '#1e40af',
    hover: '#2563eb',
    strong: '#1e3a8a',
    supplement: '#3b82f6',
    soft: '#edf3ff',
    softHover: '#e5eeff',
    surfaceHover: '#f5f8ff',
    borderActive: '#8fb8ff',
    focusShadow: 'rgba(30, 64, 175, 0.14)',
  },
  {
    name: 'cyan',
    label: '晴空蓝',
    primary: '#0284c7',
    hover: '#0ea5e9',
    strong: '#0369a1',
    supplement: '#38bdf8',
    soft: '#eef9ff',
    softHover: '#e5f5ff',
    surfaceHover: '#f6fbff',
    borderActive: '#7dd3fc',
    focusShadow: 'rgba(2, 132, 199, 0.14)',
  },
  {
    name: 'violet',
    label: '靛紫蓝',
    primary: '#4f46e5',
    hover: '#6366f1',
    strong: '#4338ca',
    supplement: '#818cf8',
    soft: '#f1f0ff',
    softHover: '#ebe9ff',
    surfaceHover: '#f8f7ff',
    borderActive: '#a5b4fc',
    focusShadow: 'rgba(79, 70, 229, 0.14)',
  },
]

const fallbackTheme = themePalettes[0]

export const useThemeStore = defineStore('theme', () => {
  const savedTheme = localStorage.getItem('fox-admin-theme') as ThemeName | null
  const activeName = ref<ThemeName>(
    themePalettes.some((theme) => theme.name === savedTheme) ? savedTheme! : fallbackTheme.name,
  )
  const compactMode = ref(localStorage.getItem('fox-admin-compact') === 'true')
  const showBreadcrumb = ref(localStorage.getItem('fox-admin-show-breadcrumb') !== 'false')
  const showTabs = ref(localStorage.getItem('fox-admin-show-tabs') !== 'false')
  const savedNavigationStyle = localStorage.getItem('fox-admin-navigation-style') as NavigationStyle | null
  const navigationStyle = ref<NavigationStyle>(
    savedNavigationStyle === 'dark' || savedNavigationStyle === 'light' ? savedNavigationStyle : 'light',
  )

  const activePalette = computed(
    () => themePalettes.find((theme) => theme.name === activeName.value) ?? fallbackTheme,
  )

  function applyTheme() {
    const root = document.documentElement
    const palette = activePalette.value

    root.style.setProperty('--fox-primary', palette.primary)
    root.style.setProperty('--fox-primary-hover', palette.hover)
    root.style.setProperty('--fox-primary-strong', palette.strong)
    root.style.setProperty('--fox-primary-soft', palette.soft)
    root.style.setProperty('--fox-primary-soft-hover', palette.softHover)
    root.style.setProperty('--fox-primary-border', palette.borderActive)
    root.style.setProperty('--fox-primary-focus', palette.focusShadow)
    root.style.setProperty('--fox-content-padding', compactMode.value ? '18px' : '24px')

    localStorage.setItem('fox-admin-theme', activeName.value)
    localStorage.setItem('fox-admin-compact', String(compactMode.value))
    localStorage.setItem('fox-admin-show-breadcrumb', String(showBreadcrumb.value))
    localStorage.setItem('fox-admin-show-tabs', String(showTabs.value))
    localStorage.setItem('fox-admin-navigation-style', navigationStyle.value)
  }

  function setTheme(name: ThemeName) {
    activeName.value = name
    applyTheme()
  }

  function setCompactMode(value: boolean) {
    compactMode.value = value
    applyTheme()
  }

  function setNavigationStyle(value: NavigationStyle) {
    navigationStyle.value = value
    applyTheme()
  }

  function setShowBreadcrumb(value: boolean) {
    showBreadcrumb.value = value
    applyTheme()
  }

  function setShowTabs(value: boolean) {
    showTabs.value = value
    applyTheme()
  }

  return {
    activeName,
    activePalette,
    compactMode,
    navigationStyle,
    showBreadcrumb,
    showTabs,
    themePalettes,
    applyTheme,
    setTheme,
    setCompactMode,
    setNavigationStyle,
    setShowBreadcrumb,
    setShowTabs,
  }
})
