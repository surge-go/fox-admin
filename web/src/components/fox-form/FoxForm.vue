<script setup lang="ts">
import { RefreshOutline, SearchOutline } from '@vicons/ionicons5'
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { FormInst, FormRules, SelectOption } from 'naive-ui'
import type { FoxFormSchema } from './types'

const props = withDefaults(
  defineProps<{
    autoFit?: boolean
    actionsInline?: boolean
    columnGap?: number | string
    columns?: number
    controlMaxWidth?: number | string
    itemMinWidth?: number
    labelAlign?: 'left' | 'right'
    labelWidth?: number
    loading?: boolean
    modelValue: Record<string, unknown>
    resetText?: string
    rowGap?: number | string
    schemas: FoxFormSchema[]
    showActions?: boolean
    showFeedback?: boolean
    submitText?: string
  }>(),
  {
    autoFit: false,
    actionsInline: false,
    columnGap: 24,
    columns: 3,
    controlMaxWidth: '100%',
    itemMinWidth: 280,
    labelAlign: 'left',
    labelWidth: 92,
    loading: false,
    resetText: '重置',
    rowGap: 14,
    showActions: true,
    showFeedback: true,
    submitText: '提交',
  },
)

const emit = defineEmits<{
  reset: []
  submit: [value: Record<string, unknown>]
  'update:modelValue': [value: Record<string, unknown>]
}>()

const formRef = ref<FormInst | null>(null)
const gridRef = ref<HTMLElement | null>(null)
const actionsRef = ref<HTMLElement | null>(null)
const actionsWrapped = ref(false)
let resizeObserver: ResizeObserver | undefined

const gridStyle = computed(() => ({
  '--fox-form-column-gap': toCssSize(props.columnGap),
  '--fox-form-control-max-width':
    typeof props.controlMaxWidth === 'number' ? `${props.controlMaxWidth}px` : props.controlMaxWidth,
  '--fox-form-item-min-width': `${props.itemMinWidth}px`,
  '--fox-form-row-gap': toCssSize(props.rowGap),
  gridTemplateColumns: props.autoFit ? undefined : `repeat(${props.columns}, minmax(0, 1fr))`,
}))

const resolvedLabelWidth = computed(() => {
  return props.autoFit ? undefined : props.labelWidth
})

function updateActionsWrapped() {
  if (!props.actionsInline) {
    actionsWrapped.value = false
    return
  }

  const gridEl = gridRef.value
  const actionsEl = actionsRef.value
  const firstItem = gridEl?.querySelector<HTMLElement>('.fox-form__item')

  if (!gridEl || !actionsEl || !firstItem) {
    actionsWrapped.value = false
    return
  }

  actionsWrapped.value = actionsEl.offsetTop > firstItem.offsetTop + 4
}

function scheduleActionsMeasure() {
  void nextTick(updateActionsWrapped)
}

const formRules = computed<FormRules>(() => {
  return props.schemas.reduce<FormRules>((rules, item) => {
    const itemRules = Array.isArray(item.rules) ? item.rules : item.rules ? [item.rules] : []

    if (item.required) {
      itemRules.unshift({
        required: true,
        message: `请输入${item.label}`,
        trigger: ['blur', 'change'],
      })
    }

    if (itemRules.length > 0) {
      rules[item.field] = itemRules
    }

    return rules
  }, {})
})

function setFieldValue(field: string, value: unknown) {
  emit('update:modelValue', {
    ...props.modelValue,
    [field]: value,
  })
}

function getPlaceholder(item: FoxFormSchema) {
  if (item.placeholder) {
    return item.placeholder
  }

  if (item.component === 'select' || item.component === 'date' || item.component === 'radio') {
    return `请选择${item.label}`
  }

  return `请输入${item.label}`
}

function getSelectOptions(item: FoxFormSchema): SelectOption[] {
  return (item.options ?? []).map((option) => ({
    label: option.label,
    value: option.value as string | number,
  }))
}

function toCssSize(value: number | string) {
  return typeof value === 'number' ? `${value}px` : value
}

async function handleSubmit() {
  await formRef.value?.validate()
  emit('submit', props.modelValue)
}

function handleReset() {
  emit('reset')
}

onMounted(() => {
  scheduleActionsMeasure()

  if (typeof ResizeObserver !== 'undefined' && gridRef.value) {
    resizeObserver = new ResizeObserver(scheduleActionsMeasure)
    resizeObserver.observe(gridRef.value)
  }
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
})

watch(
  () => [
    props.actionsInline,
    props.autoFit,
    props.columnGap,
    props.controlMaxWidth,
    props.itemMinWidth,
    props.rowGap,
    props.schemas.length,
  ],
  scheduleActionsMeasure,
)
</script>

<template>
  <n-form
    ref="formRef"
    class="fox-form"
    :class="{
      'fox-form--auto-fit': autoFit,
      'fox-form--compact': !showFeedback,
      'fox-form--actions-inline': actionsInline,
      'fox-form--actions-wrapped': actionsWrapped,
    }"
    :label-align="labelAlign"
    :label-width="resolvedLabelWidth"
    label-placement="left"
    :model="modelValue"
    :rules="formRules"
    :show-feedback="showFeedback"
  >
    <div ref="gridRef" class="fox-form__grid" :style="gridStyle">
      <n-form-item
        v-for="item in schemas"
        :key="item.field"
        class="fox-form__item"
        :class="{ 'fox-form__item--full': item.span === columns }"
        :label="item.label"
        :path="item.field"
      >
        <n-input
          v-if="item.component === 'input'"
          :disabled="item.disabled"
          :placeholder="getPlaceholder(item)"
          :value="modelValue[item.field] as string"
          @update:value="setFieldValue(item.field, $event)"
        />

        <n-input
          v-else-if="item.component === 'textarea'"
          :disabled="item.disabled"
          :placeholder="getPlaceholder(item)"
          type="textarea"
          :value="modelValue[item.field] as string"
          @update:value="setFieldValue(item.field, $event)"
        />

        <n-input-number
          v-else-if="item.component === 'input-number'"
          clearable
          :disabled="item.disabled"
          :placeholder="getPlaceholder(item)"
          :value="modelValue[item.field] as number"
          @update:value="setFieldValue(item.field, $event)"
        />

        <n-select
          v-else-if="item.component === 'select'"
          clearable
          :disabled="item.disabled"
          :options="getSelectOptions(item)"
          :placeholder="getPlaceholder(item)"
          :value="modelValue[item.field] as string | number"
          @update:value="setFieldValue(item.field, $event)"
        />

        <n-date-picker
          v-else-if="item.component === 'date'"
          clearable
          :disabled="item.disabled"
          :placeholder="getPlaceholder(item)"
          type="date"
          :value="modelValue[item.field] as number"
          @update:value="setFieldValue(item.field, $event)"
        />

        <n-radio-group
          v-else-if="item.component === 'radio'"
          :disabled="item.disabled"
          :value="modelValue[item.field] as string | number | boolean"
          @update:value="setFieldValue(item.field, $event)"
        >
          <n-space>
            <n-radio v-for="option in item.options" :key="String(option.value)" :value="option.value">
              {{ option.label }}
            </n-radio>
          </n-space>
        </n-radio-group>

        <n-switch
          v-else-if="item.component === 'switch'"
          :disabled="item.disabled"
          :value="modelValue[item.field] as boolean"
          @update:value="setFieldValue(item.field, $event)"
        />
      </n-form-item>

      <div v-if="showActions" ref="actionsRef" class="fox-form__actions">
        <n-button :loading="loading" type="primary" @click="handleSubmit">
          <template #icon>
            <n-icon :component="SearchOutline" />
          </template>
          {{ submitText }}
        </n-button>
        <n-button @click="handleReset">
          <template #icon>
            <n-icon :component="RefreshOutline" />
          </template>
          {{ resetText }}
        </n-button>
      </div>
    </div>
  </n-form>
</template>
