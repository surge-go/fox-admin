<template>
  <div class="page">
    <div class="page__header">
      <div>
        <h1>角色权限</h1>
        <p>管理角色和权限分配</p>
      </div>
      <n-button type="primary">新增角色</n-button>
    </div>

    <n-card size="small">
      <n-table :bordered="false" :single-line="false">
        <thead>
          <tr>
            <th>角色</th>
            <th>权限范围</th>
            <th>成员数</th>
            <th>状态</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>超级管理员</td>
            <td>全部权限</td>
            <td>1</td>
            <td><n-tag type="success" size="small">启用</n-tag></td>
          </tr>
          <tr>
            <td>运营</td>
            <td>基础列表、用户查看</td>
            <td>3</td>
            <td><n-tag type="success" size="small">启用</n-tag></td>
          </tr>
        </tbody>
      </n-table>
    </n-card>
  </div>
</template>
