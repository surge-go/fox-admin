/** 路由类型枚举 */
export enum RouterType {
  /** 目录，仅作为菜单分组，通常不对应具体页面 */
  Catalog = 'catalog',

  /** 菜单，对应一个可访问的业务页面 */
  Menu = 'menu',
}

/** 页面操作权限 */
export type RouterAction = {
  /** 操作名称，用于页面展示，例如：新增用户 */
  title: string

  /** 操作权限编码，用于权限判断，例如：system:user:create */
  code: string
}

/** 路由元信息，用于描述菜单展示、标签页、缓存、权限和外链等配置 */
export type RouterMate = {
  /** 路由标题，通常用于菜单、面包屑和标签页展示 */
  title: string

  /** 路由图标，通常对应图标名称或图标组件 key */
  icon?: string

  /** 是否在菜单中隐藏；隐藏后仍可通过路由访问 */
  isHide?: boolean

  /** 是否在标签页中隐藏；隐藏后访问该路由不会生成 Tab */
  isHideTab?: boolean

  /** 路由访问权限标识列表，满足权限后才允许访问该路由 */
  permissions?: string[]

  /** 页面操作列表，用于控制新增、编辑、删除、导出等按钮 */
  actions?: RouterAction[]

  /** 是否缓存页面，通常配合 keep-alive 使用 */
  keepAlive?: boolean

  /** 是否固定标签页，常用于首页、工作台等固定 Tab */
  fixedTab?: boolean

  /** 外部链接地址，用于跳转站外页面或 iframe 页面 */
  link?: string

  /** 是否外链跳转；为 true 时直接跳转，默认为 iframe 内嵌展示 */
  isExternal?: boolean
}

/** 路由菜单类型 */
export type Router = {
  /** 路由 ID */
  id: number

  /** 路由路径，例如：/system/user */
  path: string

  /** 路由名称，通常需要保持唯一 */
  name: string

  /** 路由类型 */
  type: RouterType

  /** 组件路径，用于动态加载页面组件 */
  component?: string

  /** 路由元信息，用于描述菜单展示相关配置 */
  mate: RouterMate

  /** 子路由列表，通常用于目录下挂载多个菜单页面 */
  children?: Router[]
}
