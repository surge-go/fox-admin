<script setup lang="ts">
import type { Router } from '../../../types/router'
import RoutePage from '../../shared/RoutePage.vue'

defineProps<{
  route: Router
}>()

const metrics = [
  { label: '角色数量', value: '24', trend: '+1' },
  { label: '菜单授权', value: '182', trend: '+12' },
  { label: '待调整角色', value: '3', trend: '进行中' },
]
</script>

<template>
  <RoutePage
    :route="route"
    accent="#7c3aed"
    description="角色权限页面由 component 字段驱动，可在这里继续接按钮权限和菜单授权树。"
    :metrics="metrics"
  />
</template>
