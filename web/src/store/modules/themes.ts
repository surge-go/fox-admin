import { computed, ref, watch } from 'vue'
import { darkTheme, lightTheme, type GlobalThemeOverrides } from 'naive-ui'

const THEME_STORAGE_KEY = 'fox-admin-theme'

type ThemePalette = {
  primary: string
  primaryHover: string
  primaryPressed: string
  brandFrom: string
  brandTo: string
  brandShadow: string
  textPrimary: string
  textSecondary: string
  textMuted: string
  contentBg: string
  surfaceBg: string
  surfaceBorder: string
  surfaceShadow: string
  loadingGlow: string
  menuSelectedBg: string
  menuSelectedHoverBg: string
  menuSelectedText: string
  menuChildActiveText: string
  menuSelectedBorder: string
  menuSelectedShadow: string
  menuSelectedHoverShadow: string
  menuHoverBg: string
  iframeBg: string
  iconColor: string
}

const lightPalette: ThemePalette = {
  primary: '#2563EB',
  primaryHover: '#1D4ED8',
  primaryPressed: '#1E40AF',
  brandFrom: '#7C5CFF',
  brandTo: '#5B3DF5',
  brandShadow: 'rgba(91, 61, 245, 0.24)',
  textPrimary: '#0F172A',
  textSecondary: '#475569',
  textMuted: '#64748B',
  contentBg: '#F5F7FB',
  surfaceBg: 'rgba(255, 255, 255, 0.88)',
  surfaceBorder: 'rgba(148, 163, 184, 0.18)',
  surfaceShadow: '0 18px 46px rgba(15, 23, 42, 0.10)',
  loadingGlow: 'rgba(124, 92, 255, 0.10)',
  menuSelectedBg: '#6D4CFF',
  menuSelectedHoverBg: '#5F3EF2',
  menuSelectedText: '#FFFFFF',
  menuChildActiveText: '#6D4CFF',
  menuSelectedBorder: 'rgba(109, 76, 255, 0.38)',
  menuSelectedShadow: 'rgba(15, 23, 42, 0.16)',
  menuSelectedHoverShadow: 'rgba(15, 23, 42, 0.24)',
  menuHoverBg: '#F3F6FB',
  iframeBg: '#FFFFFF',
  iconColor: '#334155',
}

const darkPalette: ThemePalette = {
  primary: '#60A5FA',
  primaryHover: '#3B82F6',
  primaryPressed: '#2563EB',
  brandFrom: '#8B6CFF',
  brandTo: '#6D4CFF',
  brandShadow: 'rgba(109, 76, 255, 0.28)',
  textPrimary: '#F8FAFC',
  textSecondary: '#CBD5E1',
  textMuted: '#94A3B8',
  contentBg: '#0F172A',
  surfaceBg: 'rgba(15, 23, 42, 0.82)',
  surfaceBorder: 'rgba(148, 163, 184, 0.16)',
  surfaceShadow: '0 22px 52px rgba(2, 6, 23, 0.32)',
  loadingGlow: 'rgba(124, 92, 255, 0.16)',
  menuSelectedBg: '#7C5CFF',
  menuSelectedHoverBg: '#6D4CFF',
  menuSelectedText: '#FFFFFF',
  menuChildActiveText: '#C4B5FD',
  menuSelectedBorder: 'rgba(124, 92, 255, 0.46)',
  menuSelectedShadow: 'rgba(2, 6, 23, 0.34)',
  menuSelectedHoverShadow: 'rgba(2, 6, 23, 0.48)',
  menuHoverBg: 'rgba(148, 163, 184, 0.12)',
  iframeBg: '#0B1220',
  iconColor: '#E2E8F0',
}

function readStoredTheme() {
  if (typeof window === 'undefined') {
    return null
  }

  try {
    return window.localStorage.getItem(THEME_STORAGE_KEY)
  }
  catch {
    return null
  }
}

function writeStoredTheme(value: 'dark' | 'light') {
  if (typeof window === 'undefined') {
    return
  }

  try {
    window.localStorage.setItem(THEME_STORAGE_KEY, value)
  }
  catch {
    // 某些内嵌浏览器会禁用 localStorage，这里忽略即可
  }
}

function getInitialDarkMode() {
  if (typeof window === 'undefined') {
    return false
  }

  const savedTheme = readStoredTheme()

  if (savedTheme === 'dark') {
    return true
  }

  if (savedTheme === 'light') {
    return false
  }

  if (typeof window.matchMedia !== 'function') {
    return false
  }

  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

const isDark = ref(getInitialDarkMode())

const naiveTheme = computed(() => (isDark.value ? darkTheme : lightTheme))
const palette = computed(() => (isDark.value ? darkPalette : lightPalette))

const themeOverrides = computed<GlobalThemeOverrides>(() => ({
  common: {
    borderRadius: '6px',
    borderRadiusSmall: '4px',
    cardColor: palette.value.surfaceBg,
    modalColor: palette.value.surfaceBg,
    popoverColor: palette.value.surfaceBg,
    primaryColor: palette.value.primary,
    primaryColorHover: palette.value.primaryHover,
    primaryColorPressed: palette.value.primaryPressed,
    textColorBase: palette.value.textPrimary,
    textColor1: palette.value.textPrimary,
    textColor2: palette.value.textSecondary,
    textColor3: palette.value.textMuted,
  },
  Button: {
    borderRadiusMedium: '6px',
  },
  Card: {
    borderRadius: '8px',
  },
  Input: {
    color: palette.value.surfaceBg,
  },
  Menu: {
    itemBorderRadius: '6px',
    itemColorActive: palette.value.menuSelectedBg,
    itemColorActiveCollapsed: palette.value.menuSelectedBg,
    itemColorActiveHover: palette.value.menuSelectedHoverBg,
    itemColorHover: palette.value.menuHoverBg,
    itemIconColorActive: palette.value.menuSelectedText,
    itemIconColorChildActive: palette.value.menuChildActiveText,
    itemIconColorActiveHover: palette.value.menuSelectedText,
    itemIconColorChildActiveHover: palette.value.menuChildActiveText,
    itemTextColorActive: palette.value.menuSelectedText,
    itemTextColorChildActive: palette.value.menuChildActiveText,
    itemTextColorActiveHover: palette.value.menuSelectedText,
    itemTextColorChildActiveHover: palette.value.menuChildActiveText,
    arrowColorChildActive: palette.value.menuChildActiveText,
    arrowColorChildActiveHover: palette.value.menuChildActiveText,
  },
  Tabs: {
    tabBorderRadiusCard: '6px',
  },
}))

const shellVars = computed<Record<string, string>>(() => ({
  '--shell-brand-from': palette.value.brandFrom,
  '--shell-brand-shadow': palette.value.brandShadow,
  '--shell-brand-to': palette.value.brandTo,
  '--shell-content-bg': palette.value.contentBg,
  '--shell-heading-color': palette.value.textPrimary,
  '--shell-icon-color': palette.value.iconColor,
  '--shell-iframe-bg': palette.value.iframeBg,
  '--shell-loading-glow': palette.value.loadingGlow,
  '--shell-muted-color': palette.value.textMuted,
  '--shell-selected-bg': palette.value.menuSelectedBg,
  '--shell-selected-border': palette.value.menuSelectedBorder,
  '--shell-selected-hover-bg': palette.value.menuSelectedHoverBg,
  '--shell-selected-hover-shadow': palette.value.menuSelectedHoverShadow,
  '--shell-selected-shadow': palette.value.menuSelectedShadow,
  '--shell-selected-text': palette.value.menuSelectedText,
  '--shell-subtle-color': palette.value.textSecondary,
  '--shell-surface-bg': palette.value.surfaceBg,
  '--shell-surface-border': palette.value.surfaceBorder,
  '--shell-surface-shadow': palette.value.surfaceShadow,
  '--shell-hover-bg': palette.value.menuHoverBg,
}))

function setDarkMode(value: boolean) {
  isDark.value = value
}

function toggleTheme() {
  isDark.value = !isDark.value
}

watch(
  isDark,
  (value) => {
    writeStoredTheme(value ? 'dark' : 'light')
  },
  { immediate: true },
)

export function useThemeStore() {
  return {
    isDark,
    naiveTheme,
    palette,
    setDarkMode,
    shellVars,
    themeOverrides,
    toggleTheme,
  }
}
