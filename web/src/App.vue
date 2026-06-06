<template>
  <n-config-provider>
    <n-message-provider>
      <main class="app">
        <n-space vertical align="center" :size="16">
          <n-h1>fox-admin-web</n-h1>
          <n-text depth="3">Vue 3 + Vite + TypeScript + Naive UI</n-text>
          <n-button type="primary">Ready</n-button>
        </n-space>
      </main>
    </n-message-provider>
  </n-config-provider>
</template>

<style scoped>
.app {
  display: grid;
  min-height: 100vh;
  place-items: center;
  padding: 24px;
}
</style>
