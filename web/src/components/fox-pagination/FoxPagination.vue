<script setup lang="ts">
withDefaults(
  defineProps<{
    page: number
    pageSize: number
    pageSlot?: number
    pageSizes?: number[]
    showQuickJumper?: boolean
    showSizePicker?: boolean
    total: number
  }>(),
  {
    pageSlot: 7,
    pageSizes: () => [10, 20, 30, 50],
    showQuickJumper: true,
    showSizePicker: true,
  },
)

const emit = defineEmits<{
  'update:page': [value: number]
  'update:pageSize': [value: number]
}>()
</script>

<template>
  <div class="fox-pagination">
    <div class="fox-pagination__summary">共 {{ total }} 条</div>

    <n-pagination
      class="fox-pagination__pager"
      :item-count="total"
      :page="page"
      :page-size="pageSize"
      :page-slot="pageSlot"
      :page-sizes="pageSizes"
      :show-quick-jumper="showQuickJumper"
      :show-size-picker="showSizePicker"
      @update:page="emit('update:page', $event)"
      @update:page-size="emit('update:pageSize', $event)"
    />
  </div>
</template>
