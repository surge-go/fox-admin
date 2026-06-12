<script setup lang="ts" generic="Row extends FoxTableRow = FoxTableRow">
import {
  IconArrowsSort,
  IconArrowsUpDown,
  IconColumns3,
  IconMaximize,
  IconMaximizeOff,
  IconRefresh,
  IconSettings2,
} from '@tabler/icons-vue'
import {
  NButton,
  NCheckbox,
  NCheckboxGroup,
  NDataTable,
  NDropdown,
  NIcon,
  NPopover,
  NSpace,
} from 'naive-ui'
import type {
  DataTableColumnKey,
  DataTableFilterState,
  DataTableRowKey,
  PaginationProps,
} from 'naive-ui'
import type {
  FoxTableColumnOption,
  FoxTableColumns,
  FoxTableEmits,
  FoxTableProps,
  FoxTableRow,
  FoxTableSize,
  FoxTableSorter,
} from './types'

defineOptions({
  inheritAttrs: false,
})

const props = withDefaults(defineProps<FoxTableProps<Row>>(), {
  columns: () => [],
  data: () => [],
  defaultCheckedRowKeys: () => [],
  pagination: false,
  bordered: false,
  flexHeight: true,
  striped: false,
  singleLine: true,
  tableLayout: 'auto',
  size: 'medium',
  showToolbar: true,
  showRefresh: true,
  showDensity: true,
  showFullscreen: true,
  showColumnSetting: true,
  showTableSetting: true,
  showFooter: true,
  emptyText: '暂无数据',
})

const emit = defineEmits<FoxTableEmits<Row>>()

const slots = defineSlots<{
  heading?: () => unknown
  toolbar?: () => unknown
  actions?: () => unknown
  filters?: () => unknown
  footer?: () => unknown
  empty?: () => unknown
}>()

const internalSize = ref<FoxTableSize>(props.size)
const internalCheckedRowKeys = ref<DataTableRowKey[]>([...props.defaultCheckedRowKeys])
const visibleColumnKeys = ref<DataTableColumnKey[]>([])
const fullscreen = ref(false)
const tableBordered = ref(props.bordered)
const tableStriped = ref(props.striped)
const tableHeaderBackground = ref(true)

const densityOptions = [
  {
    label: '宽松',
    key: 'large',
    icon: renderTableIcon,
  },
  {
    label: '默认',
    key: 'medium',
    icon: renderTableIcon,
  },
  {
    label: '紧凑',
    key: 'small',
    icon: renderTableIcon,
  },
]

const columnOptions = computed(() => collectColumnOptions(props.columns))

const displayColumns = computed(() => {
  const columns = withSorterIcons(props.columns)

  if (!visibleColumnKeys.value.length) {
    return columns
  }

  return filterVisibleColumns(columns, visibleColumnKeys.value)
})

const mergedSize = computed(() => props.size ?? internalSize.value)

const mergedCheckedRowKeys = computed(() => {
  return props.checkedRowKeys ?? internalCheckedRowKeys.value
})

const mergedSingleLine = computed(() => {
  return tableBordered.value ? false : props.singleLine
})

const mergedSingleColumn = computed(() => {
  return tableBordered.value ? false : props.singleColumn
})

const hasToolbarContent = computed(() => {
  return (
    props.showToolbar &&
    Boolean(
      props.title ||
        props.description ||
        slots.heading ||
        slots.toolbar ||
        slots.actions ||
        props.showRefresh ||
        props.showDensity ||
        props.showFullscreen ||
        props.showColumnSetting ||
        props.showTableSetting,
    )
  )
})

const hasFooterContent = computed(() => {
  return Boolean(slots.footer)
})

function handleRefresh() {
  emit('refresh')
}

const mergedPagination = computed(() => {
  if (!props.pagination) {
    return false
  }

  return {
    showSizePicker: true,
    pageSizes: [10, 20, 50, 100],
    ...props.pagination,
    'onUpdate:page': handlePageUpdate,
    'onUpdate:pageSize': handlePageSizeUpdate,
  } satisfies PaginationProps
})

watch(
  () => props.size,
  (size) => {
    if (size) {
      internalSize.value = size
    }
  },
)

watch(
  () => props.bordered,
  (bordered) => {
    tableBordered.value = bordered
  },
)

watch(
  () => props.striped,
  (striped) => {
    tableStriped.value = striped
  },
)

watch(
  columnOptions,
  (options) => {
    const optionKeys = options.map((option) => option.key)

    if (!visibleColumnKeys.value.length) {
      visibleColumnKeys.value = optionKeys
      return
    }

    visibleColumnKeys.value = visibleColumnKeys.value.filter((key) => {
      return optionKeys.includes(key)
    })
  },
  { immediate: true },
)

function defaultRowKey(row: Row) {
  const record = row as Record<string, unknown>

  return (record.id ?? record.key) as DataTableRowKey
}

function toggleFullscreen() {
  fullscreen.value = !fullscreen.value
}

function handleSizeSelect(size: FoxTableSize) {
  internalSize.value = size
  emit('update:size', size)
}

function handlePageUpdate(page: number) {
  emit('update:page', page)
}

function handlePageSizeUpdate(pageSize: number) {
  emit('update:pageSize', pageSize)
}

function handleSorterUpdate(sorter: FoxTableSorter) {
  emit('update:sorter', sorter)
}

function handleFiltersUpdate(filters: DataTableFilterState) {
  emit('update:filters', filters)
}

function handleCheckedRowKeysUpdate(keys: DataTableRowKey[], rows: object[]) {
  internalCheckedRowKeys.value = keys
  emit('update:checkedRowKeys', keys, rows as Row[])
}

function handleColumnKeysUpdate(keys: Array<string | number>) {
  visibleColumnKeys.value = keys as DataTableColumnKey[]
}

function getColumnKey(column: unknown) {
  const key = (column as { key?: DataTableColumnKey }).key
  return typeof key === 'string' || typeof key === 'number' ? key : undefined
}

function getColumnTitle(column: unknown, fallback: string) {
  const title = (column as { title?: unknown }).title

  if (typeof title === 'string' && title.trim()) {
    return title
  }

  return fallback
}

function collectColumnOptions(columns: FoxTableColumns<Row>): FoxTableColumnOption[] {
  const options: FoxTableColumnOption[] = []

  columns.forEach((column) => {
    const key = getColumnKey(column)
    const children = (column as { children?: FoxTableColumns<Row> }).children

    if (key !== undefined) {
      options.push({
        key,
        title: getColumnTitle(column, String(key)),
      })
    }

    if (Array.isArray(children)) {
      options.push(...collectColumnOptions(children))
    }
  })

  return options
}

function filterVisibleColumns(
  columns: FoxTableColumns<Row>,
  keys: DataTableColumnKey[],
): FoxTableColumns<Row> {
  return columns
    .map((column) => {
      const key = getColumnKey(column)
      const children = (column as { children?: FoxTableColumns<Row> }).children

      if (Array.isArray(children)) {
        const visibleChildren = filterVisibleColumns(children, keys)

        if (!visibleChildren.length) {
          return null
        }

        return {
          ...column,
          children: visibleChildren,
        }
      }

      if (key !== undefined && !keys.includes(key)) {
        return null
      }

      return column
    })
    .filter(Boolean) as FoxTableColumns<Row>
}

function withSorterIcons(columns: FoxTableColumns<Row>): FoxTableColumns<Row> {
  return columns.map((column) => {
    const children = (column as { children?: FoxTableColumns<Row> }).children

    if (Array.isArray(children)) {
      return {
        ...column,
        children: withSorterIcons(children),
      }
    }

    if ('sorter' in column && column.sorter) {
      return {
        ...column,
        renderSorterIcon: renderFoxSorterIcon,
      }
    }

    return column
  }) as FoxTableColumns<Row>
}

function renderFoxSorterIcon({ order }: { order: false | 'ascend' | 'descend' }) {
  return h(
    'span',
    {
      'aria-hidden': 'true',
      class: [
        'fox-table__sorter',
        order === 'ascend' && 'fox-table__sorter--asc',
        order === 'descend' && 'fox-table__sorter--desc',
      ],
    },
    [
      h('span', {
        class: 'fox-table__sorter-arrow fox-table__sorter-arrow--asc',
      }),
      h('span', {
        class: 'fox-table__sorter-arrow fox-table__sorter-arrow--desc',
      }),
    ],
  )
}

function renderTableIcon() {
  return h(IconArrowsSort)
}
</script>

<template>
  <section
    class="fox-table"
    :class="[
      `fox-table--${mergedSize}`,
      fullscreen && 'fox-table--fullscreen',
      tableBordered ? 'fox-table--bordered' : 'fox-table--borderless',
      tableStriped && 'fox-table--striped',
      tableHeaderBackground && 'fox-table--header-background',
      !tableHeaderBackground && 'fox-table--plain-header',
    ]"
  >
    <header v-if="hasToolbarContent" class="fox-table__toolbar">
      <div class="fox-table__heading">
        <slot name="heading">
        <strong v-if="title">{{ title }}</strong>
        <span v-if="description">{{ description }}</span>
        </slot>
      </div>

      <div class="fox-table__toolbar-main">
        <slot name="toolbar" />
      </div>

      <NSpace align="center" :size="10" class="fox-table__actions">
        <slot name="actions" />

        <NButton
          v-if="showRefresh"
          class="fox-table__tool-button"
          quaternary
          size="small"
          :disabled="loading"
          @click="handleRefresh"
        >
          <template #icon>
            <NIcon>
              <IconRefresh />
            </NIcon>
          </template>
        </NButton>

        <NDropdown
          v-if="showDensity"
          trigger="click"
          :options="densityOptions"
          :value="mergedSize"
          @select="handleSizeSelect"
        >
          <NButton class="fox-table__tool-button" quaternary size="small">
            <template #icon>
              <NIcon>
                <IconArrowsUpDown />
              </NIcon>
            </template>
          </NButton>
        </NDropdown>

        <NButton
          v-if="showFullscreen"
          class="fox-table__tool-button"
          quaternary
          size="small"
          @click="toggleFullscreen"
        >
          <template #icon>
            <NIcon>
              <IconMaximizeOff v-if="fullscreen" />
              <IconMaximize v-else />
            </NIcon>
          </template>
        </NButton>

        <NPopover
          v-if="showColumnSetting && columnOptions.length"
          placement="bottom-end"
          trigger="click"
          :width="180"
        >
          <template #trigger>
            <NButton class="fox-table__tool-button" quaternary size="small">
              <template #icon>
                <NIcon>
                  <IconColumns3 />
                </NIcon>
              </template>
            </NButton>
          </template>

          <div class="fox-table__column-panel">
            <div class="fox-table__column-panel-title">显示列</div>
            <NCheckboxGroup
              :value="visibleColumnKeys"
              @update:value="handleColumnKeysUpdate"
            >
              <NSpace vertical :size="8">
                <NCheckbox
                  v-for="column in columnOptions"
                  :key="column.key"
                  :label="column.title"
                  :value="column.key"
                  :disabled="
                    visibleColumnKeys.length === 1 &&
                    visibleColumnKeys.includes(column.key)
                  "
                />
              </NSpace>
            </NCheckboxGroup>
          </div>
        </NPopover>

        <NPopover
          v-if="showTableSetting"
          placement="bottom-end"
          trigger="click"
          :width="136"
        >
          <template #trigger>
            <NButton class="fox-table__tool-button" quaternary size="small">
              <template #icon>
                <NIcon>
                  <IconSettings2 stroke="2" />
                </NIcon>
              </template>
            </NButton>
          </template>

          <div class="fox-table__style-panel">
            <NCheckbox v-model:checked="tableStriped">
              斑马纹
            </NCheckbox>
            <NCheckbox v-model:checked="tableBordered">
              边框
            </NCheckbox>
            <NCheckbox v-model:checked="tableHeaderBackground">
              表头背景
            </NCheckbox>
          </div>
        </NPopover>
      </NSpace>
    </header>

    <div v-if="$slots.filters" class="fox-table__filters">
      <slot name="filters" />
    </div>

    <NDataTable
      v-bind="$attrs"
      class="fox-table__content"
      :columns="displayColumns"
      :data="data"
      :row-key="rowKey || defaultRowKey"
      :row-class-name="rowClassName"
      :row-props="rowProps"
      :checked-row-keys="mergedCheckedRowKeys"
      :pagination="mergedPagination"
      :loading="loading"
      :remote="remote"
      :bordered="true"
      :striped="tableStriped"
      :single-line="mergedSingleLine"
      :single-column="mergedSingleColumn"
      :scroll-x="scrollX"
      :min-height="minHeight"
      :max-height="maxHeight"
      :flex-height="flexHeight"
      :table-layout="tableLayout"
      :virtual-scroll="virtualScroll"
      :virtual-scroll-x="virtualScrollX"
      :virtual-scroll-header="virtualScrollHeader"
      :size="mergedSize"
      @update:checked-row-keys="handleCheckedRowKeysUpdate"
      @update:sorter="handleSorterUpdate"
      @update:filters="handleFiltersUpdate"
    >
      <template v-if="$slots.empty" #empty>
        <slot name="empty" />
      </template>
      <template v-else #empty>
        <div class="fox-table__empty">{{ emptyText }}</div>
      </template>
    </NDataTable>

    <footer v-if="showFooter && hasFooterContent" class="fox-table__footer">
      <slot name="footer" />
    </footer>
  </section>
</template>

<style scoped>
.fox-table {
  --fox-table-border: var(--shell-surface-border, rgba(148, 163, 184, 0.18));
  --fox-table-control-bg: color-mix(in srgb, var(--shell-surface-bg, #fff) 84%, transparent);
  --fox-table-header-bg: color-mix(
    in srgb,
    var(--fox-table-row-bg) 64%,
    var(--shell-content-bg, #f5f7fb)
  );
  --fox-table-fixed-header-bg: color-mix(
    in srgb,
    var(--fox-table-row-bg) 72%,
    var(--shell-content-bg, #f5f7fb)
  );
  --fox-table-fixed-row-bg: color-mix(
    in srgb,
    var(--fox-table-row-bg) 94%,
    var(--shell-content-bg, #f5f7fb)
  );
  --fox-table-fixed-row-striped-bg: color-mix(
    in srgb,
    var(--shell-content-bg, #f5f7fb) 58%,
    var(--fox-table-row-bg)
  );
  --fox-table-row-bg: var(--shell-iframe-bg, #fff);
  --fox-table-row-hover-bg: var(--shell-hover-bg, #f3f6fb);
  --fox-table-row-striped-bg: color-mix(
    in srgb,
    var(--shell-content-bg, #f5f7fb) 54%,
    var(--fox-table-row-bg)
  );
  --fox-table-sorter-muted: color-mix(
    in srgb,
    var(--shell-muted-color, #94a3b8) 82%,
    transparent
  );
  --fox-table-surface-bg: var(--shell-surface-bg, rgba(255, 255, 255, 0.88));

  backdrop-filter: blur(18px);
  background: var(--fox-table-surface-bg);
  border: 1px solid var(--fox-table-border);
  border-radius: 8px;
  box-shadow: var(--shell-surface-shadow, 0 18px 46px rgba(15, 23, 42, 0.10));
  color: var(--shell-subtle-color, #475569);
  display: flex;
  flex: 1;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
}

.fox-table--fullscreen {
  border-radius: 10px;
  bottom: 16px;
  box-shadow: 0 24px 80px rgba(2, 6, 23, 0.24);
  left: 16px;
  position: fixed;
  right: 16px;
  top: 16px;
  z-index: 2000;
}

.fox-table--fullscreen .fox-table__content {
  flex: 1;
  min-height: 0;
}

.fox-table--plain-header {
  --fox-table-header-bg: var(--fox-table-row-bg);
}

.fox-table--bordered {
  --fox-table-border: color-mix(
    in srgb,
    var(--shell-surface-border, rgba(148, 163, 184, 0.18)) 78%,
    var(--shell-muted-color, #64748b) 22%
  );
}

.fox-table--borderless {
  --fox-table-cell-border: transparent;
}

.fox-table--bordered {
  --fox-table-cell-border: var(--fox-table-border);
}

.fox-table__toolbar {
  align-items: center;
  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--fox-table-surface-bg) 96%, var(--shell-selected-bg, #6d4cff) 4%),
      var(--fox-table-surface-bg)
    );
  border-bottom: 1px solid var(--fox-table-border);
  display: grid;
  gap: 12px;
  grid-template-columns: minmax(0, max-content) 1fr auto;
  min-height: 58px;
  padding: 9px 14px;
}

.fox-table__heading {
  align-items: flex-start;
  display: flex;
  flex-direction: column;
  gap: 4px;
  justify-content: center;
  min-width: 0;
}

.fox-table__heading strong {
  color: var(--shell-heading-color, #111827);
  font-size: 15px;
  font-weight: 700;
  line-height: 1.25;
}

.fox-table__heading span {
  color: var(--shell-muted-color, #6b7280);
  font-size: 12px;
  line-height: 1.4;
}

.fox-table__toolbar-main {
  min-width: 0;
}

.fox-table__actions {
  justify-content: flex-end;
}

.fox-table__tool-button {
  background: color-mix(
    in srgb,
    var(--shell-content-bg, #f5f7fb) 86%,
    var(--shell-muted-color, #64748b) 14%
  );
  border-radius: 8px;
  box-shadow: none;
  color: color-mix(
    in srgb,
    var(--shell-heading-color, #0f172a) 72%,
    var(--shell-muted-color, #64748b)
  );
  height: 38px;
  transition:
    background-color 0.18s ease,
    color 0.18s ease,
    transform 0.18s ease;
  width: 38px;
}

.fox-table__tool-button:hover {
  background: color-mix(
    in srgb,
    var(--shell-content-bg, #f5f7fb) 78%,
    var(--shell-selected-bg, #6d4cff) 22%
  );
  color: var(--shell-selected-bg, #6d4cff);
  transform: translateY(-1px);
}

.fox-table__tool-button :deep(.n-button__border),
.fox-table__tool-button :deep(.n-button__state-border) {
  box-shadow: none;
}

.fox-table__tool-button :deep(.n-button__icon) {
  font-size: 20px;
}

.fox-table__style-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 2px 0;
}

.fox-table__filters {
  background: var(--fox-table-control-bg);
  border-bottom: 1px solid var(--fox-table-border);
  padding: 12px 14px;
}

.fox-table__content {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
  min-width: 0;
}

.fox-table__content :deep(.n-data-table) {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
}

.fox-table__content :deep(.n-data-table-wrapper),
.fox-table__content :deep(.n-data-table-base-table),
.fox-table__content :deep(.n-data-table-base-table-body) {
  background: transparent;
}

.fox-table__content :deep(.n-data-table-wrapper) {
  flex: 1;
  min-height: 0;
}

.fox-table--bordered .fox-table__content :deep(.n-data-table-wrapper) {
  border-color: var(--fox-table-border);
  border-top-color: transparent;
}

.fox-table--borderless .fox-table__content :deep(.n-data-table-wrapper) {
  border-color: transparent;
}

.fox-table__content :deep(.n-data-table-base-table) {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
}

.fox-table__content :deep(.n-data-table-base-table-body) {
  flex: 1;
}

.fox-table__content :deep(.n-data-table-base-table-body .n-scrollbar-container) {
  height: 100%;
}

.fox-table__content :deep(.n-data-table-table) {
  background: var(--fox-table-row-bg);
}

.fox-table__content :deep(.n-data-table-th) {
  background-color: var(--fox-table-header-bg);
  border-color: var(--fox-table-cell-border, var(--fox-table-border));
  color: var(--shell-heading-color, #111827);
  font-size: 12px;
  font-weight: 700;
  height: 40px;
}

.fox-table__content :deep(.n-data-table-td) {
  background-color: var(--fox-table-row-bg);
  border-color: var(--fox-table-cell-border, var(--fox-table-border));
  color: var(--shell-subtle-color, #475569);
  height: 48px;
}

.fox-table--bordered .fox-table__content :deep(.n-data-table-th),
.fox-table--bordered .fox-table__content :deep(.n-data-table-td) {
  border-color: var(--fox-table-cell-border) !important;
}

.fox-table--borderless .fox-table__content :deep(.n-data-table-th),
.fox-table--borderless .fox-table__content :deep(.n-data-table-td) {
  border-color: transparent !important;
}

.fox-table--borderless .fox-table__content :deep(.n-data-table-th:not(:last-child)),
.fox-table--borderless .fox-table__content :deep(.n-data-table-td:not(:last-child)) {
  border-right-color: transparent;
}

.fox-table__content :deep(.n-data-table-td),
.fox-table__content :deep(.n-data-table-th) {
  padding-bottom: 10px;
  padding-top: 10px;
}

.fox-table__content :deep(.n-data-table-tr--striped .n-data-table-td) {
  background-color: var(--fox-table-row-striped-bg);
}

.fox-table__content :deep(.n-data-table-tr:hover .n-data-table-td) {
  background-color: var(--fox-table-row-hover-bg);
}

.fox-table__content :deep(.n-data-table-tr) {
  transition: background-color 0.18s ease;
}

.fox-table__content :deep(.n-data-table__pagination) {
  background: color-mix(in srgb, var(--fox-table-surface-bg) 92%, transparent);
  border-top: 1px solid var(--fox-table-border);
  justify-content: center;
  margin: 0;
  margin-top: auto;
  min-height: 48px;
  padding: 9px 14px;
  position: sticky;
  bottom: 0;
  z-index: 3;
}

.fox-table__content
  :deep(.n-data-table-th--sortable .n-data-table-th__title-wrapper) {
  display: inline-flex;
  width: auto;
}

.fox-table__content
  :deep(.n-data-table-th--sortable .n-data-table-th__title) {
  flex: 0 0 auto;
}

.fox-table__content :deep(.n-data-table-sorter) {
  align-items: center;
  color: var(--fox-table-sorter-muted);
  display: inline-flex;
  height: 14px;
  justify-content: center;
  margin-left: 2px;
  width: 10px;
}

.fox-table__content :deep(.fox-table__sorter) {
  align-items: center;
  display: inline-flex;
  flex-direction: column;
  gap: 2px;
  height: 14px;
  justify-content: center;
  width: 10px;
}

.fox-table__content :deep(.fox-table__sorter-arrow) {
  border-left: 4px solid transparent;
  border-right: 4px solid transparent;
  height: 0;
  opacity: 0.72;
  transition: border-color 0.18s ease, opacity 0.18s ease;
  width: 0;
}

.fox-table__content :deep(.fox-table__sorter-arrow--asc) {
  border-bottom: 5px solid var(--fox-table-sorter-muted);
}

.fox-table__content :deep(.fox-table__sorter-arrow--desc) {
  border-top: 5px solid var(--fox-table-sorter-muted);
}

.fox-table__content
  :deep(.fox-table__sorter--asc .fox-table__sorter-arrow--asc) {
  border-bottom-color: var(--shell-selected-bg, #6d4cff);
  opacity: 1;
}

.fox-table__content
  :deep(.n-data-table-sorter--asc .fox-table__sorter-arrow--asc) {
  border-bottom-color: var(--shell-selected-bg, #6d4cff);
  opacity: 1;
}

.fox-table__content
  :deep(.fox-table__sorter--desc .fox-table__sorter-arrow--desc) {
  border-top-color: var(--shell-selected-bg, #6d4cff);
  opacity: 1;
}

.fox-table__content
  :deep(.n-data-table-sorter--desc .fox-table__sorter-arrow--desc) {
  border-top-color: var(--shell-selected-bg, #6d4cff);
  opacity: 1;
}

.fox-table__content :deep(.n-data-table-th--fixed-left),
.fox-table__content :deep(.n-data-table-th--fixed-right) {
  background-color: var(--fox-table-fixed-header-bg) !important;
}

.fox-table__content :deep(.n-data-table-td--fixed-left),
.fox-table__content :deep(.n-data-table-td--fixed-right) {
  background-color: var(--fox-table-fixed-row-bg) !important;
}

.fox-table__content
  :deep(.n-data-table-tr--striped .n-data-table-td--fixed-left),
.fox-table__content
  :deep(.n-data-table-tr--striped .n-data-table-td--fixed-right) {
  background-color: var(--fox-table-fixed-row-striped-bg) !important;
}

.fox-table__content
  :deep(.n-data-table-tr:hover .n-data-table-td--fixed-left),
.fox-table__content
  :deep(.n-data-table-tr:hover .n-data-table-td--fixed-right) {
  background-color: var(--fox-table-row-hover-bg) !important;
}

.fox-table__content :deep(.n-data-table-th--fixed-right),
.fox-table__content :deep(.n-data-table-td--fixed-right) {
  box-shadow: -1px 0 0 var(--fox-table-border);
}

.fox-table__content :deep(.n-data-table-th--fixed-left),
.fox-table__content :deep(.n-data-table-td--fixed-left) {
  box-shadow: 1px 0 0 var(--fox-table-border);
}

.fox-table__empty {
  color: var(--shell-muted-color, #6b7280);
  font-size: 13px;
  padding: 20px 0;
}

.fox-table__footer {
  align-items: center;
  background: var(--fox-table-control-bg);
  border-top: 1px solid var(--fox-table-border);
  color: var(--shell-muted-color, #6b7280);
  display: flex;
  flex-wrap: wrap;
  font-size: 12px;
  gap: 12px;
  justify-content: space-between;
  min-height: 42px;
  padding: 10px 14px;
}

.fox-table__column-panel {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.fox-table__column-panel-title {
  color: var(--shell-heading-color, #111827);
  font-size: 13px;
  font-weight: 600;
}

.fox-table--small .fox-table__toolbar {
  min-height: 50px;
}

.fox-table--large .fox-table__toolbar {
  min-height: 62px;
}

@media (max-width: 980px) {
  .fox-table__toolbar {
    align-items: stretch;
    grid-template-columns: 1fr;
  }

  .fox-table__actions {
    justify-content: flex-end;
  }

  .fox-table__heading {
    align-items: stretch;
  }

  .fox-table__content :deep(.n-data-table__pagination) {
    justify-content: center;
  }
}

@media (max-width: 760px) {
  .fox-table__actions {
    justify-content: flex-start;
  }
}
</style>
