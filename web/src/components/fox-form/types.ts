import type { FormItemRule } from 'naive-ui'

export type FoxFormComponent =
  | 'date'
  | 'input'
  | 'input-number'
  | 'radio'
  | 'select'
  | 'switch'
  | 'textarea'

export type FoxFormOption = {
  label: string
  value: string | number | boolean
}

export type FoxFormSchema = {
  field: string
  label: string
  component: FoxFormComponent
  placeholder?: string
  options?: FoxFormOption[]
  span?: number
  disabled?: boolean
  required?: boolean
  rules?: FormItemRule | FormItemRule[]
}
