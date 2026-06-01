<script setup lang="ts">
const users = [
  { id: 1, name: 'admin', role: '超级管理员', status: '启用' },
  { id: 2, name: 'operator', role: '运营', status: '启用' },
  { id: 3, name: 'auditor', role: '审计', status: '停用' },
]
</script>

<template>
  <div class="page">
    <div class="page__header">
      <div>
        <h1>用户管理</h1>
        <p>账号、角色和状态</p>
      </div>
      <n-button type="primary">新增用户</n-button>
    </div>

    <n-card size="small">
      <n-space justify="space-between" class="table-toolbar">
        <n-input placeholder="搜索用户名" clearable />
        <n-button>查询</n-button>
      </n-space>

      <n-table :bordered="false" :single-line="false">
        <thead>
          <tr>
            <th>ID</th>
            <th>用户名</th>
            <th>角色</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in users" :key="user.id">
            <td>{{ user.id }}</td>
            <td>{{ user.name }}</td>
            <td>{{ user.role }}</td>
            <td>
              <n-tag :type="user.status === '启用' ? 'success' : 'default'" size="small">
                {{ user.status }}
              </n-tag>
            </td>
            <td>
              <n-button quaternary size="small">编辑</n-button>
            </td>
          </tr>
        </tbody>
      </n-table>
    </n-card>
  </div>
</template>
