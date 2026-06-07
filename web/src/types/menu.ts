import type { Component } from 'vue'

/** 布局菜单项 */
export type LayoutMenuItem = {
  /** 菜单唯一值 */
  key: string

  /** 菜单标题 */
  label: string

  /** 菜单图标组件 */
  icon?: Component

  /** 菜单徽章 */
  badge?: string | number

  /** 子菜单 */
  children?: LayoutMenuItem[]
}

/** 标签页 */
export type LayoutTab = {
  /** 标签页唯一值 */
  key: string

  /** 标签页标题 */
  label: string

  /** 是否固定 */
  fixed?: boolean
}

