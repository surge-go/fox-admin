<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { createFoxThemeOverrides } from './theme/fox-theme'
import { useThemeStore } from './stores/theme'

const themeStore = useThemeStore()
const themeOverrides = computed(() => createFoxThemeOverrides(themeStore.activePalette))

onMounted(() => {
  themeStore.applyTheme()
})
</script>

<template>
  <n-config-provider :theme-overrides="themeOverrides">
    <n-message-provider>
      <n-dialog-provider>
        <n-notification-provider>
          <router-view />
        </n-notification-provider>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>
