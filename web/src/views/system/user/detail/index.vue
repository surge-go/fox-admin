<script setup lang="ts">
import type { RouteParams, Router } from '../../../../types/router'
import RoutePage from '../../../shared/RoutePage.vue'

defineProps<{
  route: Router
  fullPath: string
  params: RouteParams
}>()

const metrics = [
  { label: '登录次数', value: '248', trend: '+12' },
  { label: '关联角色', value: '3', trend: '稳定' },
  { label: '待处理工单', value: '2', trend: '-1' },
]
</script>

<template>
  <RoutePage
    :full-path="fullPath"
    :params="params"
    :route="route"
    accent="#0f766e"
    description="当前详情页使用隐藏菜单路由 + 动态参数渲染，可用于用户详情、订单详情、设备详情等场景。"
    :metrics="metrics"
  />
</template>
