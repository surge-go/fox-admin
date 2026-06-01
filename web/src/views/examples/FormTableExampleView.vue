<script setup lang="ts">
import FoxForm from '../../components/fox-form/FoxForm.vue'
import type { FoxFormSchema } from '../../components/fox-form/types'
import FoxPagination from '../../components/fox-pagination/FoxPagination.vue'
import FoxPanel from '../../components/fox-panel/FoxPanel.vue'
import FoxTable from '../../components/fox-table/FoxTable.vue'
import type { FoxTableColumn, FoxTableRow } from '../../components/fox-table/types'

type ExampleRow = {
  avatar: string
  id: number
  code: string
  name: string
  owner: string
  priority: 'high' | 'medium' | 'low'
  status: 'enabled' | 'disabled' | 'pending'
  type: 'system' | 'permission' | 'basic'
  typeLabel: string
  successRate: number
  updatedAt: string
}

const queryModel = ref<Record<string, unknown>>({
  keyword: '',
  type: null,
  priority: null,
  status: null,
})

const page = ref(1)
const pageSize = ref(10)
const selectedRowIds = ref<number[]>([])

const querySchemas: FoxFormSchema[] = [
  {
    field: 'keyword',
    label: '关键词',
    component: 'input',
    placeholder: '名称 / 负责人',
  },
  {
    field: 'type',
    label: '业务类型',
    component: 'select',
    options: [
      { label: '系统配置', value: 'system' },
      { label: '权限管理', value: 'permission' },
      { label: '基础数据', value: 'basic' },
    ],
  },
  {
    field: 'priority',
    label: '优先级',
    component: 'select',
    options: [
      { label: '高', value: 'high' },
      { label: '中', value: 'medium' },
      { label: '低', value: 'low' },
    ],
  },
  {
    field: 'status',
    label: '状态',
    component: 'select',
    options: [
      { label: '启用', value: 'enabled' },
      { label: '待确认', value: 'pending' },
      { label: '停用', value: 'disabled' },
    ],
  },
]

const tableColumns: FoxTableColumn[] = [
  { key: 'selection', title: '', width: 42, align: 'center', headerAlign: 'center' },
  { key: 'id', title: 'ID', width: 68, align: 'center', headerAlign: 'center' },
  { key: 'name', title: '配置名称', width: 220, align: 'left', headerAlign: 'left' },
  { key: 'typeLabel', title: '类型', width: 130, align: 'center', headerAlign: 'center' },
  { key: 'priority', title: '优先级', width: 120, align: 'center', headerAlign: 'center' },
  {
    key: 'avatar',
    title: '头像',
    width: 76,
    align: 'center',
    headerAlign: 'center',
    cellClassName: 'example-table__avatar-cell',
  },
  { key: 'owner', title: '负责人', width: 130, align: 'left', headerAlign: 'left' },
  { key: 'status', title: '状态', width: 120, align: 'center', headerAlign: 'center' },
  { key: 'successRate', title: '成功率', width: 130, align: 'left', headerAlign: 'left' },
  { key: 'updatedAt', title: '更新时间', width: 150, align: 'left', headerAlign: 'left' },
  { key: 'actions', title: '操作', width: 132, align: 'center', fixed: 'right', headerAlign: 'center' },
]

const typeOptions = [
  { label: '系统配置', value: 'system' },
  { label: '权限管理', value: 'permission' },
  { label: '基础数据', value: 'basic' },
] as const

const sceneOptions = ['方案', '策略', '规则', '模板', '任务', '流程']
const ownerOptions = ['admin', 'operator', 'auditor', 'manager']
const priorityOptions = ['high', 'medium', 'low'] as const
const statusOptions = ['enabled', 'pending', 'enabled', 'disabled'] as const

function createAvatarDataUrl(owner: string, index: number) {
  const colors = [
    ['#2563eb', '#dbeafe'],
    ['#7c3aed', '#ede9fe'],
    ['#0891b2', '#cffafe'],
    ['#ea580c', '#ffedd5'],
  ]
  const [textColor, bgColor] = colors[index % colors.length]
  const label = owner.slice(0, 2).toUpperCase()
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="40" height="40"><rect width="40" height="40" rx="10" fill="${bgColor}"/><text x="20" y="24" text-anchor="middle" font-family="Arial, sans-serif" font-size="13" font-weight="700" fill="${textColor}">${label}</text></svg>`

  return `data:image/svg+xml;utf8,${encodeURIComponent(svg)}`
}

const rows = ref<ExampleRow[]>(
  Array.from({ length: 128 }, (_, index) => {
    const type = typeOptions[index % typeOptions.length]
    const scene = sceneOptions[index % sceneOptions.length]
    const owner = ownerOptions[index % ownerOptions.length]

    return {
      avatar: createAvatarDataUrl(owner, index),
      id: index + 1,
      code: `FOX-${String(index + 1).padStart(4, '0')}`,
      name: `${type.label}${scene} ${index + 1}`,
      owner,
      priority: priorityOptions[index % priorityOptions.length],
      status: statusOptions[index % statusOptions.length],
      type: type.value,
      typeLabel: type.label,
      successRate: 92 + (index % 8),
      updatedAt: `2026-${String((index % 6) + 1).padStart(2, '0')}-${String((index % 28) + 1).padStart(2, '0')}`,
    }
  }),
)

const filteredRows = computed(() => {
  const keyword = String(queryModel.value.keyword ?? '').trim()
  const type = queryModel.value.type
  const priority = queryModel.value.priority
  const status = queryModel.value.status

  return rows.value.filter((row) => {
    const keywordMatched =
      !keyword || row.name.includes(keyword) || row.owner.includes(keyword) || row.code.includes(keyword)
    const typeMatched = !type || row.type === type
    const priorityMatched = !priority || row.priority === priority
    const statusMatched = !status || row.status === status

    return keywordMatched && typeMatched && priorityMatched && statusMatched
  })
})

const pagedRows = computed(() => {
  const start = (page.value - 1) * pageSize.value
  return filteredRows.value.slice(start, start + pageSize.value)
})

const pageRowIds = computed(() => pagedRows.value.map((row) => row.id))

const allPageSelected = computed(
  () => pageRowIds.value.length > 0 && pageRowIds.value.every((id) => selectedRowIds.value.includes(id)),
)

const partiallyPageSelected = computed(
  () => !allPageSelected.value && pageRowIds.value.some((id) => selectedRowIds.value.includes(id)),
)

const metrics = computed(() => {
  const enabled = rows.value.filter((row) => row.status === 'enabled').length
  const pending = rows.value.filter((row) => row.status === 'pending').length
  const averageRate = Math.round(
    rows.value.reduce((total, row) => total + row.successRate, 0) / rows.value.length,
  )

  return [
    { label: '全部配置', value: rows.value.length, meta: 'Mock 数据', type: 'default' },
    { label: '启用中', value: enabled, meta: '可正常访问', type: 'success' },
    { label: '待确认', value: pending, meta: '需要处理', type: 'warning' },
    { label: '平均成功率', value: `${averageRate}%`, meta: '近 7 天', type: 'info' },
  ] as const
})

watch([filteredRows, pageSize], () => {
  const maxPage = Math.max(1, Math.ceil(filteredRows.value.length / pageSize.value))
  if (page.value > maxPage) {
    page.value = maxPage
  }
})

function handleSearch() {
  page.value = 1
  selectedRowIds.value = []
}

function handleReset() {
  queryModel.value = {
    keyword: '',
    type: null,
    priority: null,
    status: null,
  }
  page.value = 1
  selectedRowIds.value = []
}

function handleTogglePage(checked: boolean) {
  const current = new Set(selectedRowIds.value)

  pageRowIds.value.forEach((id) => {
    if (checked) {
      current.add(id)
    } else {
      current.delete(id)
    }
  })

  selectedRowIds.value = Array.from(current)
}

function handleToggleRow(rowId: number, checked: boolean) {
  selectedRowIds.value = checked
    ? Array.from(new Set([...selectedRowIds.value, rowId]))
    : selectedRowIds.value.filter((id) => id !== rowId)
}

function getStatusTag(row: FoxTableRow) {
  if (row.status === 'enabled') {
    return { label: '启用', type: 'success' as const }
  }

  if (row.status === 'pending') {
    return { label: '待确认', type: 'warning' as const }
  }

  return { label: '停用', type: 'default' as const }
}

function getPriorityTag(row: FoxTableRow) {
  if (row.priority === 'high') {
    return { label: '高', type: 'error' as const }
  }

  if (row.priority === 'medium') {
    return { label: '中', type: 'warning' as const }
  }

  return { label: '低', type: 'info' as const }
}
</script>

<template>
  <div class="page">
    <div class="page__header">
      <div>
        <h1>组件示例</h1>
        <p>查询表单、数据表格、批量选择和分页组合</p>
      </div>
      <n-space>
        <n-button>导出</n-button>
        <n-button type="primary">新增记录</n-button>
      </n-space>
    </div>

    <div class="metric-grid">
      <n-card v-for="item in metrics" :key="item.label" size="small">
        <div class="example-metric">
          <span>{{ item.label }}</span>
          <strong>{{ item.value }}</strong>
          <n-tag :type="item.type" size="small">{{ item.meta }}</n-tag>
        </div>
      </n-card>
    </div>

    <FoxPanel :show-header="false" padding="16px 18px" :gap="0">
      <FoxForm
        v-model="queryModel"
        actions-inline
        auto-fit
        :columns="4"
        :control-max-width="176"
        :item-min-width="228"
        reset-text="清空"
        :row-gap="16"
        :schemas="querySchemas"
        :show-feedback="false"
        submit-text="查询"
        @reset="handleReset"
        @submit="handleSearch"
      />
    </FoxPanel>

    <n-card size="small">
      <div class="example-table-toolbar">
        <div>
          <h2>配置列表</h2>
          <p>
            共 {{ filteredRows.length }} 条结果，已选择 {{ selectedRowIds.length }} 条
          </p>
        </div>
        <n-space>
          <n-button size="small" :disabled="selectedRowIds.length === 0">批量停用</n-button>
          <n-button size="small" @click="selectedRowIds = []">清除选择</n-button>
          <n-button size="small" type="primary">同步数据</n-button>
        </n-space>
      </div>

      <FoxTable :columns="tableColumns" empty-text="暂无匹配数据" :rows="pagedRows">
        <template #header-selection>
          <n-checkbox
            :checked="allPageSelected"
            :indeterminate="partiallyPageSelected"
            @update:checked="handleTogglePage"
          />
        </template>

        <template #cell-selection="{ row }">
          <n-checkbox
            :checked="selectedRowIds.includes(row.id)"
            @update:checked="handleToggleRow(row.id, $event)"
          />
        </template>

        <template #cell-name="{ row }">
          <div class="example-table__main">
            <strong>{{ row.name }}</strong>
            <span>{{ row.code }}</span>
          </div>
        </template>

        <template #cell-priority="{ row }">
          <n-tag :type="getPriorityTag(row).type" size="small">
            {{ getPriorityTag(row).label }}
          </n-tag>
        </template>

        <template #cell-avatar="{ row }">
          <img class="example-owner-avatar" :src="row.avatar" :alt="row.owner" />
        </template>

        <template #cell-status="{ row }">
          <n-tag :type="getStatusTag(row).type" size="small">
            {{ getStatusTag(row).label }}
          </n-tag>
        </template>

        <template #cell-successRate="{ row }">
          <n-progress
            class="example-table__progress"
            :height="6"
            :percentage="row.successRate"
            :show-indicator="false"
            type="line"
          />
        </template>

        <template #cell-actions>
          <div class="example-table__actions">
            <n-button size="small" type="primary">编辑</n-button>
            <n-button size="small" type="error">删除</n-button>
          </div>
        </template>
      </FoxTable>

      <FoxPagination
        v-model:page="page"
        v-model:page-size="pageSize"
        :total="filteredRows.length"
      />
    </n-card>
  </div>
</template>
