<script setup lang="ts">
import { useRouterStore } from '../../../store'
import type { RouteParams, Router } from '../../../types/router'
import RoutePage from '../../shared/RoutePage.vue'

defineProps<{
  route: Router
  fullPath: string
  params: RouteParams
}>()

const routerStore = useRouterStore()

const metrics = [
  { label: '用户总量', value: '8,421', trend: '+132' },
  { label: '启用账号', value: '7,985', trend: '+48' },
  { label: '待审核', value: '16', trend: '-2' },
]

function openUserDetail(id: number) {
  routerStore.openTab(`/system/user/detail/${id}`)
}
</script>

<template>
  <RoutePage
    :full-path="fullPath"
    :params="params"
    :route="route"
    accent="#2563eb"
    description="用户管理页面通过 component 字段独立加载，后续可以在这里挂真实列表、筛选和抽屉。"
    :metrics="metrics"
  >
    <template #actions>
      <n-button secondary size="small" @click="openUserDetail(1001)">
        查看用户 1001
      </n-button>
      <n-button secondary size="small" @click="openUserDetail(1002)">
        查看用户 1002
      </n-button>
    </template>
  </RoutePage>
</template>
