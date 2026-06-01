<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const iframeUrl = computed(() => {
  return typeof route.meta.iframeUrl === 'string' ? route.meta.iframeUrl : ''
})

const title = computed(() => {
  return typeof route.meta.title === 'string' ? route.meta.title : '内嵌页面'
})
</script>

<template>
  <div class="iframe-page">
    <div v-if="iframeUrl" class="iframe-page__frame">
      <iframe :src="iframeUrl" :title="title" />
    </div>

    <n-empty v-else description="未配置内嵌链接" />
  </div>
</template>
