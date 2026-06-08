<script setup lang="ts">
import type { RouteParams, Router } from '../../types/router'
import RoutePage from '../shared/RoutePage.vue'

defineProps<{
  route: Router
  fullPath: string
  params: RouteParams
}>()

const metrics = [
  { label: '字典项', value: '1,204', trend: '+18' },
  { label: '组织单元', value: '96', trend: '+2' },
  { label: '待同步', value: '5', trend: '待处理' },
]
</script>

<template>
  <RoutePage
    :full-path="fullPath"
    :params="params"
    :route="route"
    accent="#ea580c"
    description="基础数据页已按 component 动态分发，适合继续拆成字典、组织、区域等业务模块。"
    :metrics="metrics"
  />
</template>
