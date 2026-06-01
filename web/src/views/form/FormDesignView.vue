<script setup lang="ts">
import FoxForm from '../../components/fox-form/FoxForm.vue'
import type { FoxFormSchema } from '../../components/fox-form/types'

const defaultModel = {
  code: '',
  name: '',
  owner: '',
  priority: 'medium',
  status: true,
  type: 'system',
  onlineAt: null,
  remark: '',
}

const formModel = ref<Record<string, unknown>>({ ...defaultModel })

const schemas: FoxFormSchema[] = [
  {
    field: 'name',
    label: '表单名称',
    component: 'input',
    required: true,
  },
  {
    field: 'code',
    label: '业务编码',
    component: 'input',
    required: true,
  },
  {
    field: 'type',
    label: '业务类型',
    component: 'select',
    required: true,
    options: [
      { label: '系统配置', value: 'system' },
      { label: '权限管理', value: 'permission' },
      { label: '基础数据', value: 'basic' },
    ],
  },
  {
    field: 'owner',
    label: '负责人',
    component: 'input',
  },
  {
    field: 'priority',
    label: '优先级',
    component: 'radio',
    options: [
      { label: '低', value: 'low' },
      { label: '中', value: 'medium' },
      { label: '高', value: 'high' },
    ],
  },
  {
    field: 'onlineAt',
    label: '上线日期',
    component: 'date',
  },
  {
    field: 'status',
    label: '启用状态',
    component: 'switch',
  },
  {
    field: 'remark',
    label: '备注',
    component: 'textarea',
    span: 3,
  },
]

function handleSubmit(value: Record<string, unknown>) {
  window.console.log('submit form', value)
}

function handleReset() {
  formModel.value = { ...defaultModel }
}
</script>

<template>
  <div class="page">
    <div class="page__header">
      <div>
        <h1>表单设计</h1>
        <p>面向后台业务的 schema 表单组件</p>
      </div>
    </div>

    <section class="form-design-shell">
      <div class="form-design-shell__header">
        <div>
          <h2>基础信息</h2>
          <p>统一标签、校验、栅格和操作按钮</p>
        </div>
        <n-tag type="info" size="small">FoxForm</n-tag>
      </div>

      <FoxForm
        v-model="formModel"
        :columns="3"
        :schemas="schemas"
        @reset="handleReset"
        @submit="handleSubmit"
      />
    </section>
  </div>
</template>
