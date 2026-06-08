<script setup lang="ts">
import { RouterCacheBy, type RouteParams, type Router } from '../../types/router'

const props = defineProps<{
  route: Router
  fullPath: string
  params: RouteParams
  description: string
  accent: string
  metrics: Array<{
    label: string
    value: string
    trend?: string
  }>
}>()

const note = ref('')
const visits = ref(0)

const routeTitle = computed(() => props.route.mate.title)
const routePath = computed(() => props.route.path)
const routeFullPath = computed(() => props.fullPath)
const routeComponent = computed(() => props.route.component || '-')
const routeLink = computed(() => props.route.mate.link || '-')
const routeParamsText = computed(() => {
  return Object.keys(props.params).length ? JSON.stringify(props.params, null, 2) : '无'
})
const routeCacheText = computed(() => {
  if (!props.route.mate.keepAlive) {
    return '不缓存'
  }

  return props.route.mate.cacheBy === RouterCacheBy.FullPath ? '按完整路径缓存' : '按模板路径缓存'
})

function increaseVisits() {
  visits.value += 1
}
</script>

<template>
  <div class="route-page">
    <section class="route-page__hero">
      <div>
        <p>{{ routeFullPath }}</p>
        <h1>{{ routeTitle }}</h1>
        <span>{{ description }}</span>
      </div>
      <n-space>
        <slot name="actions" />
        <n-tag size="small" type="primary">
          {{ routeCacheText }}
        </n-tag>
        <n-tag size="small">
          {{ routeComponent }}
        </n-tag>
      </n-space>
    </section>

    <section class="route-page__metrics">
      <n-card
        v-for="metric in metrics"
        :key="metric.label"
        class="route-page__metric"
      >
        <span>{{ metric.label }}</span>
        <strong>{{ metric.value }}</strong>
        <n-tag v-if="metric.trend" size="small" :style="{ color: accent }">
          {{ metric.trend }}
        </n-tag>
      </n-card>
    </section>

    <section class="route-page__grid">
      <n-card class="route-page__card" title="页面状态">
        <n-descriptions :column="1" label-placement="left" size="small">
          <n-descriptions-item label="路由名称">
            {{ route.name }}
          </n-descriptions-item>
          <n-descriptions-item label="组件路径">
            {{ routeComponent }}
          </n-descriptions-item>
          <n-descriptions-item label="模板路径">
            {{ routePath }}
          </n-descriptions-item>
          <n-descriptions-item label="当前路径">
            {{ routeFullPath }}
          </n-descriptions-item>
          <n-descriptions-item label="链接地址">
            {{ routeLink }}
          </n-descriptions-item>
          <n-descriptions-item label="路由参数">
            <pre class="route-page__params">{{ routeParamsText }}</pre>
          </n-descriptions-item>
        </n-descriptions>
      </n-card>

      <n-card class="route-page__card" title="组件状态">
        <n-space vertical :size="12">
          <n-input
            v-model:value="note"
            placeholder="输入内容后切换标签页，再回来验证组件状态"
          />
          <n-button secondary @click="increaseVisits">
            本地计数 {{ visits }}
          </n-button>
        </n-space>
      </n-card>
    </section>
  </div>
</template>

<style scoped>
.route-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.route-page__hero {
  align-items: flex-start;
  display: flex;
  gap: 16px;
  justify-content: space-between;
}

.route-page__hero p {
  color: var(--shell-muted-color);
  font-size: 12px;
  margin: 0 0 4px;
  text-transform: uppercase;
}

.route-page__hero h1 {
  color: var(--shell-heading-color);
  font-size: 24px;
  line-height: 1.2;
  margin: 0;
}

.route-page__hero span {
  color: var(--shell-subtle-color);
  display: block;
  font-size: 14px;
  margin-top: 8px;
}

.route-page__metrics {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.route-page__metric :deep(.n-card__content) {
  align-items: flex-start;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 16px;
}

.route-page__metric span {
  color: var(--shell-muted-color);
  font-size: 13px;
}

.route-page__metric strong {
  color: var(--shell-heading-color);
  font-size: 28px;
  line-height: 1;
}

.route-page__grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.route-page__card :deep(.n-card__content) {
  padding: 16px;
}

.route-page__params {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}

@media (max-width: 900px) {
  .route-page__hero {
    flex-direction: column;
  }

  .route-page__metrics,
  .route-page__grid {
    grid-template-columns: 1fr;
  }
}
</style>
