import { computed, ref } from 'vue'

import { getAppSettings } from '../../api/app'
import type { AppSettings } from '../../types/app'

const DEFAULT_APP_SETTINGS: AppSettings = {
  title: 'Fox Admin',
  titleSuffix: 'Fox Admin',
}

const settings = ref<AppSettings>({ ...DEFAULT_APP_SETTINGS })
const isLoading = ref(false)
const isLoaded = ref(false)

function ensureFaviconLink() {
  let faviconLink = document.querySelector<HTMLLinkElement>('link[rel="icon"]')

  if (!faviconLink) {
    faviconLink = document.createElement('link')
    faviconLink.rel = 'icon'
    document.head.appendChild(faviconLink)
  }

  return faviconLink
}

function isValidAssetLink(value?: string) {
  if (!value) {
    return false
  }

  const normalizedValue = value.trim()

  if (!normalizedValue) {
    return false
  }

  return /^(https?:)?\/\//.test(normalizedValue)
    || normalizedValue.startsWith('/')
    || normalizedValue.startsWith('data:image/')
}

function applyFavicon(favicon?: string) {
  if (typeof document === 'undefined' || !isValidAssetLink(favicon)) {
    return
  }

  ensureFaviconLink().href = favicon!.trim()
}

function applyDocumentTitle(pageTitle?: string) {
  if (typeof document === 'undefined') {
    return
  }

  const titleSuffix = settings.value.titleSuffix || settings.value.title

  document.title = pageTitle ? `${pageTitle} - ${titleSuffix}` : titleSuffix
}

async function loadSettings(force = false) {
  if (isLoading.value) {
    return settings.value
  }

  if (isLoaded.value && !force) {
    return settings.value
  }

  isLoading.value = true

  try {
    const nextSettings = await getAppSettings()
    settings.value = {
      ...DEFAULT_APP_SETTINGS,
      ...nextSettings,
    }
    applyFavicon(settings.value.favicon)
    isLoaded.value = true
    return settings.value
  }
  finally {
    isLoading.value = false
  }
}

export function useAppStore() {
  return {
    applyDocumentTitle,
    isLoaded,
    isLoading,
    isValidAssetLink,
    loadSettings,
    logo: computed(() => (isValidAssetLink(settings.value.favicon) ? settings.value.favicon!.trim() : '')),
    settings,
    markText: computed(() => settings.value.title.slice(0, 1).toUpperCase() || 'F'),
    title: computed(() => settings.value.title),
  }
}
