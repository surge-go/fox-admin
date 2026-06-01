<script setup lang="ts">
import { computed, useSlots } from 'vue'
import type { CSSProperties } from 'vue'

const props = withDefaults(
  defineProps<{
    bordered?: boolean
    description?: string
    gap?: number | string
    padding?: number | string
    shadow?: boolean
    showHeader?: boolean
    title?: string
  }>(),
  {
    bordered: true,
    gap: 16,
    padding: '16px 20px',
    shadow: true,
    showHeader: true,
  },
)

const slots = useSlots()

const panelStyle = computed<CSSProperties>(() => ({
  '--fox-panel-gap': toCssSize(props.gap),
  '--fox-panel-padding': toCssSize(props.padding),
}))

const hasHeader = computed(() => {
  return props.showHeader && Boolean(props.title || props.description || slots.header || slots.actions)
})

function toCssSize(value: number | string) {
  return typeof value === 'number' ? `${value}px` : value
}
</script>

<template>
  <section
    class="fox-panel"
    :class="{
      'fox-panel--bordered': bordered,
      'fox-panel--shadow': shadow,
      'fox-panel--with-header': hasHeader,
    }"
    :style="panelStyle"
  >
    <div v-if="hasHeader" class="fox-panel__header">
      <slot name="header">
        <div>
          <h2 v-if="title">{{ title }}</h2>
          <p v-if="description">{{ description }}</p>
        </div>
      </slot>

      <div v-if="$slots.actions" class="fox-panel__actions">
        <slot name="actions" />
      </div>
    </div>

    <div class="fox-panel__body">
      <slot />
    </div>
  </section>
</template>
