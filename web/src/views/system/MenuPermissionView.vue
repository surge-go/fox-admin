<script setup lang="ts">
import { AddOutline, CreateOutline, ReloadOutline, SaveOutline, SearchOutline } from '@vicons/ionicons5'
import type { TreeOption } from 'naive-ui'

type MenuType = 'catalog' | 'menu' | 'button' | 'external' | 'iframe'

type PermissionAction = {
  id: number
  name: string
  code: string
  method: string
  status: 'enabled' | 'disabled'
}

type MenuRecord = {
  key: string
  parentKey?: string
  title: string
  type: MenuType
  path: string
  component: string
  permission: string
  icon: string
  sort: number
  redirect: string
  link: string
  visible: boolean
  keepAlive: boolean
  status: 'enabled' | 'disabled'
  actions: PermissionAction[]
}

const selectedKeys = ref<Array<string | number>>(['menu-permission'])
const keyword = ref('')
const showCreateModal = ref(false)
const createModel = ref({
  title: '',
  type: 'menu',
  parentKey: 'system',
  path: '',
  component: '',
  permission: '',
  icon: '',
  sort: 10,
  redirect: '',
  link: '',
  visible: true,
  keepAlive: true,
  status: true,
})

const menuRecords: MenuRecord[] = [
  {
    key: 'system',
    title: '系统设置',
    type: 'catalog',
    path: '/system',
    component: 'LAYOUT',
    permission: 'system',
    icon: 'OptionsOutline',
    sort: 20,
    redirect: '/system/menu-permission',
    link: '',
    visible: true,
    keepAlive: true,
    status: 'enabled',
    actions: [],
  },
  {
    key: 'menu-permission',
    parentKey: 'system',
    title: '菜单权限',
    type: 'menu',
    path: '/system/menu-permission',
    component: '/system/menu-permission',
    permission: 'system:menu:list',
    icon: 'MenuOutline',
    sort: 10,
    redirect: '',
    link: '',
    visible: true,
    keepAlive: true,
    status: 'enabled',
    actions: [
      { id: 1, name: '新增菜单', code: 'system:menu:add', method: 'POST', status: 'enabled' },
      { id: 2, name: '编辑菜单', code: 'system:menu:update', method: 'PUT', status: 'enabled' },
      { id: 3, name: '删除菜单', code: 'system:menu:delete', method: 'DELETE', status: 'enabled' },
    ],
  },
  {
    key: 'role-permission',
    parentKey: 'system',
    title: '角色权限',
    type: 'menu',
    path: '/system/role-permission',
    component: '/system/role-permission',
    permission: 'system:role:list',
    icon: 'PersonOutline',
    sort: 20,
    redirect: '',
    link: '',
    visible: true,
    keepAlive: true,
    status: 'enabled',
    actions: [
      { id: 4, name: '分配权限', code: 'system:role:assign', method: 'POST', status: 'enabled' },
      { id: 5, name: '停用角色', code: 'system:role:disable', method: 'PATCH', status: 'enabled' },
    ],
  },
  {
    key: 'links',
    title: '链接示例',
    type: 'catalog',
    path: '/links',
    component: 'LAYOUT',
    permission: 'links',
    icon: 'LinkOutline',
    sort: 60,
    redirect: '/links/iframe',
    link: '',
    visible: true,
    keepAlive: true,
    status: 'enabled',
    actions: [],
  },
  {
    key: 'iframe-link',
    parentKey: 'links',
    title: '内嵌链接',
    type: 'iframe',
    path: '/links/iframe',
    component: 'IFrameView',
    permission: 'links:iframe:view',
    icon: 'BrowsersOutline',
    sort: 10,
    redirect: '',
    link: 'https://example.com',
    visible: true,
    keepAlive: true,
    status: 'enabled',
    actions: [],
  },
  {
    key: 'external-link',
    parentKey: 'links',
    title: 'Naive UI 外链',
    type: 'external',
    path: '/links/naive-ui',
    component: '',
    permission: 'links:external:view',
    icon: 'OpenOutline',
    sort: 20,
    redirect: '',
    link: 'https://www.naiveui.com/zh-CN/os-theme',
    visible: true,
    keepAlive: false,
    status: 'enabled',
    actions: [],
  },
]

const treeData: TreeOption[] = [
  {
    key: 'system',
    label: '系统设置',
    children: [
      { key: 'menu-permission', label: '菜单权限' },
      { key: 'role-permission', label: '角色权限' },
    ],
  },
  {
    key: 'links',
    label: '链接示例',
    children: [
      { key: 'iframe-link', label: '内嵌链接' },
      { key: 'external-link', label: 'Naive UI 外链' },
    ],
  },
]

const activeMenu = computed(() => {
  return menuRecords.find((item) => item.key === selectedKeys.value[0]) ?? menuRecords[0]
})

const childMenus = computed(() => {
  return menuRecords.filter((item) => item.parentKey === activeMenu.value.key)
})

const filteredTreeData = computed(() => {
  const value = keyword.value.trim()

  if (!value) {
    return treeData
  }

  return treeData
    .map((item) => ({
      ...item,
      children: item.children?.filter((child) => String(child.label).includes(value)),
    }))
    .filter((item) => String(item.label).includes(value) || item.children?.length)
})

const parentOptions = computed(() => {
  return menuRecords
    .filter((item) => item.type === 'catalog')
    .map((item) => ({
      label: item.title,
      value: item.key,
    }))
})

const menuTypeOptions = [
  { label: '目录', value: 'catalog' },
  { label: '菜单', value: 'menu' },
  { label: '按钮', value: 'button' },
  { label: '内嵌链接', value: 'iframe' },
  { label: '外部链接', value: 'external' },
]

function getTypeTag(type: MenuType) {
  const tagMap: Record<MenuType, { label: string; type: 'default' | 'info' | 'success' | 'warning' }> = {
    button: { label: '按钮', type: 'warning' },
    catalog: { label: '目录', type: 'default' },
    external: { label: '外链', type: 'info' },
    iframe: { label: '内嵌', type: 'success' },
    menu: { label: '菜单', type: 'success' },
  }

  return tagMap[type]
}

function openCreateModal() {
  showCreateModal.value = true
}
</script>

<template>
  <div class="page menu-admin-page">
    <div class="page__header">
      <div>
        <h1>菜单权限</h1>
        <p>管理目录、菜单、按钮权限、内嵌页面和外部链接</p>
      </div>
      <n-space>
        <n-button>
          <template #icon>
            <n-icon :component="ReloadOutline" />
          </template>
          刷新
        </n-button>
        <n-button type="primary" @click="openCreateModal">
          <template #icon>
            <n-icon :component="AddOutline" />
          </template>
          新增菜单
        </n-button>
      </n-space>
    </div>

    <div class="menu-admin-layout">
      <section class="menu-tree-panel">
        <div class="menu-panel-header">
          <div>
            <strong>菜单结构</strong>
            <span>{{ menuRecords.length }} 个节点</span>
          </div>
          <n-button quaternary size="small" @click="openCreateModal">
            <template #icon>
              <n-icon :component="AddOutline" />
            </template>
          </n-button>
        </div>

        <n-input v-model:value="keyword" clearable placeholder="搜索菜单名称">
          <template #prefix>
            <n-icon :component="SearchOutline" />
          </template>
        </n-input>

        <n-tree
          v-model:selected-keys="selectedKeys"
          block-line
          class="menu-tree"
          :data="filteredTreeData"
          :default-expanded-keys="['system', 'links']"
          selectable
        />
      </section>

      <section class="menu-detail-panel">
        <div class="menu-detail-header">
          <div>
            <span>当前节点</span>
            <h2>{{ activeMenu.title }}</h2>
          </div>
          <n-space>
            <n-tag :type="getTypeTag(activeMenu.type).type">
              {{ getTypeTag(activeMenu.type).label }}
            </n-tag>
            <n-tag :type="activeMenu.status === 'enabled' ? 'success' : 'default'">
              {{ activeMenu.status === 'enabled' ? '启用' : '停用' }}
            </n-tag>
          </n-space>
        </div>

        <div class="menu-form-grid">
          <label>
            <span>菜单名称</span>
            <n-input :value="activeMenu.title" />
          </label>
          <label>
            <span>菜单类型</span>
            <n-select
              :value="activeMenu.type"
              :options="[
                { label: '目录', value: 'catalog' },
                { label: '菜单', value: 'menu' },
                { label: '按钮', value: 'button' },
                { label: '内嵌链接', value: 'iframe' },
                { label: '外部链接', value: 'external' },
              ]"
            />
          </label>
          <label>
            <span>路由地址</span>
            <n-input :value="activeMenu.path" />
          </label>
          <label>
            <span>组件路径</span>
            <n-input :value="activeMenu.component" placeholder="LAYOUT / IFrameView / 页面组件路径" />
          </label>
          <label>
            <span>权限标识</span>
            <n-input :value="activeMenu.permission" />
          </label>
          <label>
            <span>菜单图标</span>
            <n-input :value="activeMenu.icon" />
          </label>
          <label>
            <span>重定向</span>
            <n-input :value="activeMenu.redirect" placeholder="可选" />
          </label>
          <label>
            <span>排序</span>
            <n-input-number :value="activeMenu.sort" />
          </label>
          <label class="menu-form-grid__full">
            <span>链接地址</span>
            <n-input :value="activeMenu.link" placeholder="内嵌链接或外链地址" />
          </label>
        </div>

        <div class="menu-switch-grid">
          <div>
            <strong>显示菜单</strong>
            <p>关闭后保留路由但不在侧边栏展示</p>
            <n-switch :value="activeMenu.visible" />
          </div>
          <div>
            <strong>页面缓存</strong>
            <p>切换标签页时保留页面状态</p>
            <n-switch :value="activeMenu.keepAlive" />
          </div>
          <div>
            <strong>启用状态</strong>
            <p>停用后角色无法分配该权限</p>
            <n-switch :value="activeMenu.status === 'enabled'" />
          </div>
        </div>

        <div class="menu-relation-grid">
          <div class="menu-sub-panel">
            <div class="menu-sub-panel__title">
              <strong>下级菜单</strong>
              <n-button size="small">
                <template #icon>
                  <n-icon :component="AddOutline" />
                </template>
                新增下级
              </n-button>
            </div>

            <n-empty v-if="childMenus.length === 0" description="暂无下级菜单" />
            <div v-else class="menu-child-list">
              <button v-for="item in childMenus" :key="item.key" type="button" @click="selectedKeys = [item.key]">
                <span>{{ item.title }}</span>
                <n-tag size="small" :type="getTypeTag(item.type).type">
                  {{ getTypeTag(item.type).label }}
                </n-tag>
              </button>
            </div>
          </div>

          <div class="menu-sub-panel">
            <div class="menu-sub-panel__title">
              <strong>按钮权限</strong>
              <n-button size="small">
                <template #icon>
                  <n-icon :component="AddOutline" />
                </template>
                新增按钮
              </n-button>
            </div>

            <n-table v-if="activeMenu.actions.length" :bordered="false" size="small">
              <thead>
                <tr>
                  <th>名称</th>
                  <th>权限标识</th>
                  <th>方法</th>
                  <th>状态</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in activeMenu.actions" :key="item.id">
                  <td>{{ item.name }}</td>
                  <td>{{ item.code }}</td>
                  <td>{{ item.method }}</td>
                  <td>
                    <n-tag :type="item.status === 'enabled' ? 'success' : 'default'" size="small">
                      {{ item.status === 'enabled' ? '启用' : '停用' }}
                    </n-tag>
                  </td>
                  <td>
                    <n-button quaternary size="small">
                      <template #icon>
                        <n-icon :component="CreateOutline" />
                      </template>
                    </n-button>
                  </td>
                </tr>
              </tbody>
            </n-table>
            <n-empty v-else description="暂无按钮权限" />
          </div>
        </div>

        <div class="menu-detail-actions">
          <n-button>取消</n-button>
          <n-button type="primary">
            <template #icon>
              <n-icon :component="SaveOutline" />
            </template>
            保存配置
          </n-button>
        </div>
      </section>
    </div>

    <n-modal
      v-model:show="showCreateModal"
      class="menu-create-modal"
      preset="card"
      title="新增菜单"
      transform-origin="center"
    >
      <div class="menu-create-modal-grid">
        <label>
          <span>菜单名称</span>
          <n-input v-model:value="createModel.title" placeholder="请输入菜单名称" />
        </label>
        <label>
          <span>上级菜单</span>
          <n-select v-model:value="createModel.parentKey" :options="parentOptions" />
        </label>
        <label>
          <span>菜单类型</span>
          <n-select v-model:value="createModel.type" :options="menuTypeOptions" />
        </label>
        <label>
          <span>排序</span>
          <n-input-number v-model:value="createModel.sort" :min="0" />
        </label>
        <label>
          <span>路由地址</span>
          <n-input v-model:value="createModel.path" placeholder="/system/example" />
        </label>
        <label>
          <span>组件路径</span>
          <n-input v-model:value="createModel.component" placeholder="LAYOUT / IFrameView / 页面组件路径" />
        </label>
        <label>
          <span>权限标识</span>
          <n-input v-model:value="createModel.permission" placeholder="system:menu:add" />
        </label>
        <label>
          <span>菜单图标</span>
          <n-input v-model:value="createModel.icon" placeholder="OptionsOutline" />
        </label>
        <label>
          <span>重定向</span>
          <n-input v-model:value="createModel.redirect" placeholder="可选" />
        </label>
        <label>
          <span>链接地址</span>
          <n-input v-model:value="createModel.link" placeholder="外链或内嵌链接地址" />
        </label>
      </div>

      <div class="menu-create-modal-switches">
        <div>
          <strong>显示菜单</strong>
          <n-switch v-model:value="createModel.visible" />
        </div>
        <div>
          <strong>页面缓存</strong>
          <n-switch v-model:value="createModel.keepAlive" />
        </div>
        <div>
          <strong>启用状态</strong>
          <n-switch v-model:value="createModel.status" />
        </div>
      </div>

      <template #footer>
        <div class="menu-create-modal-actions">
          <n-button @click="showCreateModal = false">取消</n-button>
          <n-button type="primary" @click="showCreateModal = false">
            <template #icon>
              <n-icon :component="SaveOutline" />
            </template>
            保存
          </n-button>
        </div>
      </template>
    </n-modal>
  </div>
</template>
