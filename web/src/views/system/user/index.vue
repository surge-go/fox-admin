<script setup lang="ts">
import type { Router } from '../../../types/router'
import RoutePage from '../../shared/RoutePage.vue'

defineProps<{
  route: Router
}>()

const metrics = [
  { label: '用户总量', value: '8,421', trend: '+132' },
  { label: '启用账号', value: '7,985', trend: '+48' },
  { label: '待审核', value: '16', trend: '-2' },
]
</script>

<template>
  <RoutePage
    :route="route"
    accent="#2563eb"
    description="用户管理页面通过 component 字段独立加载，后续可以在这里挂真实列表、筛选和抽屉。"
    :metrics="metrics"
  />
</template>
