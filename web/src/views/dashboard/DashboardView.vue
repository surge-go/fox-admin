<script setup lang="ts">
const metrics = [
  { label: '今日访问', value: '12,408', trend: '+12.4%' },
  { label: '活跃用户', value: '2,936', trend: '+8.1%' },
  { label: '接口成功率', value: '99.92%', trend: '+0.3%' },
  { label: '待处理事项', value: '18', trend: '-4' },
]
</script>

<template>
  <div class="page">
    <div class="page__header">
      <div>
        <h1>工作台</h1>
        <p>服务运行状态和关键指标</p>
      </div>
      <n-button type="primary">刷新数据</n-button>
    </div>

    <div class="metric-grid">
      <n-card v-for="item in metrics" :key="item.label" size="small">
        <div class="metric-card">
          <span>{{ item.label }}</span>
          <strong>{{ item.value }}</strong>
          <n-tag size="small" type="success">{{ item.trend }}</n-tag>
        </div>
      </n-card>
    </div>

    <n-card title="最近任务" size="small">
      <n-table :bordered="false" :single-line="false">
        <thead>
          <tr>
            <th>任务</th>
            <th>负责人</th>
            <th>状态</th>
            <th>更新时间</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>用户权限校验</td>
            <td>admin</td>
            <td><n-tag type="success" size="small">正常</n-tag></td>
            <td>09:20</td>
          </tr>
          <tr>
            <td>接口响应监控</td>
            <td>ops</td>
            <td><n-tag type="warning" size="small">观察</n-tag></td>
            <td>09:12</td>
          </tr>
        </tbody>
      </n-table>
    </n-card>
  </div>
</template>
