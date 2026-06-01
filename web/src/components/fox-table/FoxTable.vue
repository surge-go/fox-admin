<script setup lang="ts">
import { computed, h, useSlots } from 'vue'
import type { DataTableColumns, DataTableCreateRowKey, DataTableRowKey } from 'naive-ui'
import type { FoxTableColumn, FoxTableRow } from './types'

const props = withDefaults(
  defineProps<{
    bordered?: boolean
    columns: FoxTableColumn[]
    emptyText?: string
    loading?: boolean
    rowKey?: string
    rows: FoxTableRow[]
    scrollX?: number | string
    singleLine?: boolean
    size?: 'small' | 'medium' | 'large'
  }>(),
  {
    bordered: false,
    emptyText: '暂无数据',
    loading: false,
    rowKey: 'id',
    singleLine: false,
    size: 'medium',
  },
)

const slots = useSlots()

const visibleColumns = computed(() => props.columns.filter((column) => !column.hidden))

const rowIndexMap = computed(() => {
  const map = new WeakMap<FoxTableRow, number>()

  props.rows.forEach((row, index) => {
    map.set(row, index)
  })

  return map
})

const mergedScrollX = computed(() => props.scrollX ?? getColumnsWidth(visibleColumns.value))

const dataTableColumns = computed<DataTableColumns<FoxTableRow>>(() => {
  return visibleColumns.value.map((column) => ({
    key: column.key,
    title: () => renderHeader(column),
    align: column.align,
    className: column.className,
    ellipsis: column.ellipsis,
    fixed: column.fixed,
    maxWidth: column.maxWidth,
    minWidth: column.minWidth,
    titleAlign: column.headerAlign ?? column.align,
    width: column.width,
    cellProps: (row, index) => ({
      class: getCellClass(column, row, index),
      style: getCellStyle(column, row, index),
    }),
    render: (row, index) => renderCell(column, row, index),
  }))
})

function getColumnsWidth(columns: FoxTableColumn[]) {
  const width = columns.reduce((total, column) => {
    return total + (getColumnScrollWidth(column) ?? 0)
  }, 0)

  return width > 0 ? width : undefined
}

function getColumnScrollWidth(column: FoxTableColumn) {
  return normalizeColumnWidth(column.width) ?? normalizeColumnWidth(column.minWidth)
}

function normalizeColumnWidth(width: number | string | undefined) {
  if (typeof width === 'number') {
    return Number.isFinite(width) && width > 0 ? width : undefined
  }

  if (typeof width !== 'string') {
    return undefined
  }

  const matched = width.trim().match(/^(\d+(?:\.\d+)?)px?$/)

  return matched ? Number(matched[1]) : undefined
}

const getRowKey: DataTableCreateRowKey<FoxTableRow> = (row) => {
  const rowKey = normalizeRowKey(row[props.rowKey])

  return rowKey ?? rowIndexMap.value.get(row) ?? 0
}

function normalizeRowKey(value: unknown): DataTableRowKey | undefined {
  if (typeof value === 'string' || typeof value === 'number') {
    return value
  }

  return undefined
}

function getCellClass(column: FoxTableColumn, row: FoxTableRow, index: number) {
  if (typeof column.cellClassName === 'function') {
    return column.cellClassName(row, index)
  }

  return column.cellClassName
}

function getCellStyle(column: FoxTableColumn, row: FoxTableRow, index: number) {
  if (typeof column.cellStyle === 'function') {
    return column.cellStyle(row, index)
  }

  return column.cellStyle
}

function renderHeader(column: FoxTableColumn) {
  const slot = slots[`header-${column.key}`]
  const content = slot ? slot({ column }) : column.title

  if (!column.headerClassName && !column.headerStyle) {
    return content
  }

  return h(
    'span',
    {
      class: column.headerClassName,
      style: column.headerStyle,
    },
    content,
  )
}

function renderCell(column: FoxTableColumn, row: FoxTableRow, index: number) {
  const slot = slots[`cell-${column.key}`]

  if (slot) {
    return slot({
      column,
      index,
      row,
      value: row[column.key],
    })
  }

  return row[column.key] ?? ''
}
</script>

<template>
  <n-data-table
    class="fox-table"
    :bordered="bordered"
    :columns="dataTableColumns"
    :data="rows"
    :loading="loading"
    :pagination="false"
    :row-key="getRowKey"
    :scroll-x="mergedScrollX"
    :single-line="singleLine"
    :size="size"
  >
    <template #empty>
      <slot name="empty">
        <n-empty :description="emptyText" />
      </slot>
    </template>
  </n-data-table>
</template>
