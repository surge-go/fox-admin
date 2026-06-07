<script setup lang="ts">
import type { Router } from '../../types/router'

const props = defineProps<{
  route: Router
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
const routeComponent = computed(() => props.route.component || '-')
const routeLink = computed(() => props.route.mate.link || '-')
const routeCacheText = computed(() => (props.route.mate.keepAlive ? '开启缓存' : '不缓存'))

function increaseVisits() {
  visits.value += 1
}
</script>

<template>
  <div class="route-page">
    <section class="route-page__hero">
      <div>
        <p>{{ routePath }}</p>
        <h1>{{ routeTitle }}</h1>
        <span>{{ description }}</span>
      </div>
      <n-space>
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
          <n-descriptions-item label="链接地址">
            {{ routeLink }}
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
  color: #64748b;
  font-size: 12px;
  margin: 0 0 4px;
  text-transform: uppercase;
}

.route-page__hero h1 {
  color: #0f172a;
  font-size: 24px;
  line-height: 1.2;
  margin: 0;
}

.route-page__hero span {
  color: #475569;
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
  color: #64748b;
  font-size: 13px;
}

.route-page__metric strong {
  color: #0f172a;
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
