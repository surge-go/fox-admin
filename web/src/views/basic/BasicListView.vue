<script setup lang="ts">
import FoxPagination from '../../components/fox-pagination/FoxPagination.vue'

const page = ref(1)
const pageSize = ref(10)
const total = ref(42)
</script>

<template>
  <div class="page">
    <div class="page__header">
      <div>
        <h1>基础列表</h1>
        <p>常规数据列表和操作入口</p>
      </div>
      <n-button type="primary">新增数据</n-button>
    </div>

    <n-card size="small">
      <n-space justify="space-between" class="table-toolbar">
        <n-input placeholder="搜索名称" clearable />
        <n-button>查询</n-button>
      </n-space>

      <n-table :bordered="false" :single-line="false">
        <thead>
          <tr>
            <th>ID</th>
            <th>名称</th>
            <th>分类</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>1001</td>
            <td>示例数据 A</td>
            <td>默认分类</td>
            <td><n-tag type="success" size="small">启用</n-tag></td>
            <td><n-button quaternary size="small">编辑</n-button></td>
          </tr>
          <tr>
            <td>1002</td>
            <td>示例数据 B</td>
            <td>扩展分类</td>
            <td><n-tag size="small">停用</n-tag></td>
            <td><n-button quaternary size="small">编辑</n-button></td>
          </tr>
        </tbody>
      </n-table>

      <FoxPagination
        v-model:page="page"
        v-model:page-size="pageSize"
        :total="total"
      />
    </n-card>
  </div>
</template>
