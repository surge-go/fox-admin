import type {
  DataTableColumnKey,
  DataTableColumns,
  DataTableCreateRowClassName,
  DataTableCreateRowKey,
  DataTableCreateRowProps,
  DataTableFilterState,
  DataTableRowKey,
  DataTableSize,
  DataTableSortState,
  PaginationProps,
} from 'naive-ui'

export type FoxTableRow = object

export type FoxTableColumns<Row extends FoxTableRow = FoxTableRow> = DataTableColumns<Row>

export type FoxTableSize = NonNullable<DataTableSize>

export type FoxTablePagination = false | PaginationProps

export type FoxTableSorter = DataTableSortState | DataTableSortState[] | null

export interface FoxTableColumnOption {
  key: DataTableColumnKey
  title: string
}

export interface FoxTableProps<Row extends FoxTableRow = FoxTableRow> {
  title?: string
  description?: string
  columns?: FoxTableColumns<Row>
  data?: Row[]
  rowKey?: DataTableCreateRowKey<Row>
  rowClassName?: string | DataTableCreateRowClassName<Row>
  rowProps?: DataTableCreateRowProps<Row>
  checkedRowKeys?: DataTableRowKey[]
  defaultCheckedRowKeys?: DataTableRowKey[]
  pagination?: FoxTablePagination
  loading?: boolean
  remote?: boolean
  bordered?: boolean
  striped?: boolean
  singleLine?: boolean
  singleColumn?: boolean
  scrollX?: string | number
  minHeight?: string | number
  maxHeight?: string | number
  flexHeight?: boolean
  tableLayout?: 'auto' | 'fixed'
  virtualScroll?: boolean
  virtualScrollX?: boolean
  virtualScrollHeader?: boolean
  size?: FoxTableSize
  showToolbar?: boolean
  showRefresh?: boolean
  showDensity?: boolean
  showFullscreen?: boolean
  showColumnSetting?: boolean
  showTableSetting?: boolean
  showFooter?: boolean
  emptyText?: string
}

export interface FoxTableEmits<Row extends FoxTableRow = FoxTableRow> {
  refresh: []
  'update:size': [size: FoxTableSize]
  'update:page': [page: number]
  'update:pageSize': [pageSize: number]
  'update:sorter': [sorter: FoxTableSorter]
  'update:filters': [filters: DataTableFilterState]
  'update:checkedRowKeys': [keys: DataTableRowKey[], rows: Row[]]
}
