<script setup lang="ts">
import type { Router } from '../../../types/router'
import RoutePage from '../../shared/RoutePage.vue'

defineProps<{
  route: Router
}>()

const metrics = [
  { label: '服务可用率', value: '99.98%', trend: '+0.03%' },
  { label: '队列积压', value: '14', trend: '-6' },
  { label: '链路告警', value: '2', trend: '关注中' },
]
</script>

<template>
  <RoutePage
    :route="route"
    accent="#0891b2"
    description="系统状态页已由 component 字段动态加载，同时开启 keepAlive，适合承载实时监控卡片。"
    :metrics="metrics"
  />
</template>
