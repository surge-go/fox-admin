import type { CSSProperties } from 'vue'

export type FoxTableAlign = 'left' | 'center' | 'right'
export type FoxTableFixed = 'left' | 'right'

export type FoxTableRow = Record<string, any>
export type FoxTableCellClass = string | ((row: FoxTableRow, index: number) => string)
export type FoxTableCellStyle = CSSProperties | ((row: FoxTableRow, index: number) => CSSProperties)

export type FoxTableColumn = {
  key: string
  title: string
  align?: FoxTableAlign
  fixed?: FoxTableFixed
  headerAlign?: FoxTableAlign
  hidden?: boolean
  ellipsis?: boolean | { tooltip?: boolean }
  maxWidth?: number | string
  minWidth?: number | string
  width?: number | string
  className?: string
  cellClassName?: FoxTableCellClass
  cellStyle?: FoxTableCellStyle
  headerClassName?: string
  headerStyle?: CSSProperties
}
