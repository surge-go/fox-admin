<script setup lang="ts">
import type { Router } from '../../types/router'
import RoutePage from '../shared/RoutePage.vue'

defineProps<{
  route: Router
}>()

const metrics = [
  { label: '在线用户', value: '1,248', trend: '+12.4%' },
  { label: '今日请求', value: '86,420', trend: '+8.1%' },
  { label: '异常告警', value: '7', trend: '-3' },
]
</script>

<template>
  <RoutePage
    :route="route"
    accent="#16a34a"
    description="工作台页面已通过 component 字段动态加载，可作为首页和固定标签页入口。"
    :metrics="metrics"
  />
</template>
