<script setup lang="ts">
import { IconChevronDown, IconRefresh, IconSearch } from '@tabler/icons-vue'
import { NButton, NIcon, NSpace } from 'naive-ui'

type QueryBarCssSize = number | string
const COMPACT_ITEM_MIN_WIDTH = 178
const COMPACT_ITEM_MAX_WIDTH = 246

const props = withDefaults(defineProps<{
  advancedItemCount?: number
  baseItemCount?: number
  collapseText?: string
  controlWidth?: QueryBarCssSize
  expandText?: string
  expanded?: boolean
  itemWidth?: QueryBarCssSize
  minItemWidth?: number
  resetText?: string
  searchText?: string
  showReset?: boolean
  showSearch?: boolean
}>(), {
  advancedItemCount: 0,
  baseItemCount: 0,
  collapseText: '收起',
  expandText: '展开',
  expanded: false,
  minItemWidth: 190,
  resetText: '重置',
  searchText: '查询',
  showReset: true,
  showSearch: true,
})

const emit = defineEmits<{
  reset: []
  search: []
  'update:expanded': [expanded: boolean]
}>()

const slots = defineSlots<{
  actions?: (props: {
    canExpand: boolean
    expanded: boolean
    reset: () => void
    search: () => void
    toggle: () => void
  }) => unknown
  advanced?: () => unknown
  default?: () => unknown
}>()

const fieldsRef = ref<HTMLElement | null>(null)
const filtersRef = ref<HTMLElement | null>(null)
const canExpand = ref(props.advancedItemCount > 0 || Boolean(slots.advanced))
const hiddenOptionalCount = ref(0)
const internalExpanded = ref(props.expanded)
let resizeObserver: ResizeObserver | undefined

const mergedExpanded = computed({
  get: () => internalExpanded.value,
  set: (expanded: boolean) => {
    if (internalExpanded.value === expanded) {
      return
    }

    internalExpanded.value = expanded
    emit('update:expanded', expanded)
  },
})

const hasAdvanced = computed(() => {
  return props.advancedItemCount > 0 || Boolean(slots.advanced)
})

const showAdvanced = computed(() => {
  return hasAdvanced.value && mergedExpanded.value
})

const queryBarStyle = computed(() => {
  const style: Record<string, string> = {
    '--fox-query-bar-item-min-width': formatCssSize(props.minItemWidth),
    '--fox-query-bar-responsive-item-width': formatCssSize(props.minItemWidth),
  }

  if (props.itemWidth !== undefined) {
    style['--fox-query-bar-item-width'] = formatCssSize(props.itemWidth)
  }

  if (props.controlWidth !== undefined) {
    style['--fox-query-bar-control-width'] = formatCssSize(props.controlWidth)
  }

  return style
})

onMounted(() => {
  nextTick(() => {
    updateExpandVisibility()

    if (typeof ResizeObserver === 'function' && fieldsRef.value) {
      resizeObserver = new ResizeObserver(updateExpandVisibility)
      resizeObserver.observe(fieldsRef.value)
    }
  })
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
})

watch(
  () => [
    props.baseItemCount,
    props.advancedItemCount,
    props.controlWidth,
    props.itemWidth,
    props.minItemWidth,
  ],
  () => nextTick(updateExpandVisibility),
)

watch(
  () => props.expanded,
  (expanded) => {
    internalExpanded.value = expanded
    nextTick(updateExpandVisibility)
  },
)

function updateExpandVisibility() {
  const fieldsElement = fieldsRef.value
  const filtersElement = filtersRef.value

  if (!fieldsElement || !filtersElement) {
    hiddenOptionalCount.value = 0
    canExpand.value = hasAdvanced.value
    mergedExpanded.value = false
    return
  }

  const styles = window.getComputedStyle(filtersElement)
  const columnGap = Number.parseFloat(styles.columnGap || styles.gap || '0') || 0
  let itemWidth = getItemWidth(fieldsElement.clientWidth)

  fieldsElement.style.setProperty(
    '--fox-query-bar-responsive-item-width',
    `${itemWidth}px`,
  )

  const baseItems = getBaseItems(filtersElement)
  const optionalItems = getOptionalBaseItems(filtersElement)
  const requiredCount = baseItems.length - optionalItems.length
  const actionsWidth = getActionsWidth(filtersElement, optionalItems.length > 0)

  fieldsElement.style.setProperty('--fox-query-bar-actions-width', `${actionsWidth}px`)

  if (requiredCount > 0 && requiredCount <= 2) {
    itemWidth = getCompressedItemWidth(
      fieldsElement.clientWidth,
      itemWidth,
      actionsWidth,
      columnGap,
      requiredCount,
    )

    fieldsElement.style.setProperty(
      '--fox-query-bar-responsive-item-width',
      `${itemWidth}px`,
    )
  }

  const previousCanExpand = canExpand.value
  syncOptionalBaseItems(
    filtersElement,
    optionalItems,
    fieldsElement.clientWidth,
    actionsWidth,
    columnGap,
  )

  canExpand.value = hasAdvanced.value || hiddenOptionalCount.value > 0

  if (!canExpand.value) {
    mergedExpanded.value = false
  }

  if (canExpand.value !== previousCanExpand) {
    nextTick(updateExpandVisibility)
  }
}

function syncOptionalBaseItems(
  fieldsElement: HTMLElement,
  optionalItems: HTMLElement[],
  containerWidth: number,
  actionsWidth: number,
  columnGap: number,
) {
  const baseItems = getBaseItems(fieldsElement)

  baseItems.forEach((item) => {
    item.classList.remove('fox-query-bar__item--auto-hidden')
  })

  if (mergedExpanded.value || !optionalItems.length) {
    hiddenOptionalCount.value = 0
    return baseItems.length
  }

  const visibleItems = [...baseItems]
  const hiddenItems: HTMLElement[] = []

  while (
    optionalItems.length > hiddenItems.length &&
    !canItemsShareFirstRow(visibleItems, containerWidth, actionsWidth, columnGap)
  ) {
    const item = optionalItems[optionalItems.length - hiddenItems.length - 1]
    hiddenItems.unshift(item)
    visibleItems.splice(visibleItems.indexOf(item), 1)
  }

  optionalItems.forEach((item) => {
    item.classList.toggle('fox-query-bar__item--auto-hidden', hiddenItems.includes(item))
  })

  hiddenOptionalCount.value = hiddenItems.length
  return baseItems.length - hiddenItems.length
}

function toggleExpanded() {
  mergedExpanded.value = !mergedExpanded.value
  nextTick(updateExpandVisibility)
}

function getBaseItems(fieldsElement: HTMLElement) {
  return Array.from(
    fieldsElement.querySelectorAll<HTMLElement>(':scope > .fox-query-bar__item'),
  )
}

function getOptionalBaseItems(fieldsElement: HTMLElement) {
  return getBaseItems(fieldsElement).filter((item) => {
    return !item.querySelector('.fox-query-bar__label--required')
  })
}

function getActionsWidth(fieldsElement: HTMLElement, mayNeedExpandButton: boolean) {
  const actionsElement = fieldsElement.querySelector<HTMLElement>('.fox-query-bar__actions')

  if (!actionsElement) {
    return 0
  }

  const styles = window.getComputedStyle(actionsElement)
  const gap = Number.parseFloat(styles.columnGap || styles.gap || '0') || 12
  const actionItems = Array.from(actionsElement.children).filter((item) => {
    return window.getComputedStyle(item).display !== 'none'
  })
  const actionsWidth = actionItems.reduce((total, item, index) => {
    const itemWidth = item.getBoundingClientRect().width
    return total + itemWidth + (index > 0 ? gap : 0)
  }, 0)
  const hasExpandButton = Boolean(actionsElement?.querySelector('.fox-query-bar__expand-button'))

  if (!hasExpandButton && (hasAdvanced.value || mayNeedExpandButton)) {
    return actionsWidth + (actionItems.length > 0 ? gap : 0) + 62
  }

  return actionsWidth
}

function getCompressedItemWidth(
  containerWidth: number,
  preferredItemWidth: number,
  actionsWidth: number,
  columnGap: number,
  requiredCount: number,
) {
  const sameRowGapWidth = columnGap * requiredCount
  const sameRowAvailableWidth = containerWidth - actionsWidth - sameRowGapWidth

  if (sameRowAvailableWidth <= 0) {
    return preferredItemWidth
  }

  const sameRowItemWidth = Math.floor(sameRowAvailableWidth / requiredCount)
  const maxItemWidth = Math.max(preferredItemWidth, COMPACT_ITEM_MAX_WIDTH)

  if (sameRowItemWidth < COMPACT_ITEM_MIN_WIDTH) {
    const fieldsOnlyGapWidth = columnGap * (requiredCount - 1)
    const fieldsOnlyAvailableWidth = containerWidth - fieldsOnlyGapWidth
    const fieldsOnlyItemWidth = Math.floor(fieldsOnlyAvailableWidth / requiredCount)

    return Math.min(maxItemWidth, fieldsOnlyItemWidth)
  }

  return Math.max(
    COMPACT_ITEM_MIN_WIDTH,
    Math.min(maxItemWidth, sameRowItemWidth),
  )
}

function canItemsShareFirstRow(
  items: HTMLElement[],
  containerWidth: number,
  actionsWidth: number,
  columnGap: number,
) {
  if (!items.length) {
    return true
  }

  const itemsWidth = items.reduce((total, item) => {
    return total + item.getBoundingClientRect().width
  }, 0)
  const gapWidth = columnGap * items.length

  return itemsWidth + actionsWidth + gapWidth <= containerWidth
}

function getItemWidth(containerWidth: number) {
  const itemWidth = getNumericCssSize(props.itemWidth)

  if (itemWidth) {
    return itemWidth
  }

  const controlWidth = getNumericCssSize(props.controlWidth)
  const minimumWidth = controlWidth ? Math.max(props.minItemWidth, controlWidth + 66) : props.minItemWidth

  if (containerWidth < 760) {
    return Math.max(minimumWidth, 180)
  }

  if (containerWidth < 900) {
    return Math.max(minimumWidth, 190)
  }

  return Math.max(minimumWidth, 206)
}

function formatCssSize(value: QueryBarCssSize) {
  return typeof value === 'number' ? `${value}px` : value
}

function getNumericCssSize(value: QueryBarCssSize | undefined) {
  if (typeof value === 'number') {
    return value
  }

  if (typeof value !== 'string') {
    return undefined
  }

  const match = value.trim().match(/^(\d+(?:\.\d+)?)px$/)
  return match ? Number(match[1]) : undefined
}

function handleReset() {
  emit('reset')
}

function handleSearch() {
  emit('search')
}
</script>

<template>
  <section
    class="fox-query-bar"
    :class="{ 'fox-query-bar--expanded': mergedExpanded }"
    :style="queryBarStyle"
  >
    <div ref="fieldsRef" class="fox-query-bar__fields">
      <div ref="filtersRef" class="fox-query-bar__filters">
        <slot />
        <template v-if="showAdvanced">
          <slot name="advanced" />
        </template>

        <NSpace align="center" :size="12" class="fox-query-bar__actions">
          <slot
            name="actions"
            :can-expand="canExpand"
            :expanded="mergedExpanded"
            :reset="handleReset"
            :search="handleSearch"
            :toggle="toggleExpanded"
          >
            <NButton v-if="showReset" secondary type="primary" @click="handleReset">
              <template #icon>
                <NIcon>
                  <IconRefresh />
                </NIcon>
              </template>
              {{ resetText }}
            </NButton>
            <NButton v-if="showSearch" type="primary" @click="handleSearch">
              <template #icon>
                <NIcon>
                  <IconSearch />
                </NIcon>
              </template>
              {{ searchText }}
            </NButton>
            <NButton
              v-if="canExpand"
              text
              type="primary"
              class="fox-query-bar__expand-button"
              @click="toggleExpanded"
            >
              {{ mergedExpanded ? collapseText : expandText }}
              <NIcon
                class="fox-query-bar__expand-icon"
                :class="{ 'fox-query-bar__expand-icon--expanded': mergedExpanded }"
              >
                <IconChevronDown />
              </NIcon>
            </NButton>
          </slot>
        </NSpace>
      </div>
    </div>
  </section>
</template>

<style scoped>
.fox-query-bar {
  --fox-query-bar-accent: var(--shell-selected-bg, #2563eb);

  background:
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--shell-surface-bg, #fff) 96%, #fff) 0%,
      var(--shell-surface-bg, #fff) 100%
  );
  border: 1px solid var(--shell-surface-border, rgba(148, 163, 184, 0.18));
  border-radius: 8px;
  box-shadow:
    0 1px 0 rgba(255, 255, 255, 0.74) inset,
    0 8px 24px rgba(15, 23, 42, 0.05);
  container-type: inline-size;
  overflow: hidden;
  padding: 18px 18px 16px;
  position: relative;
}

.fox-query-bar__fields {
  align-items: start;
  display: block;
  min-width: 0;
  position: relative;
  z-index: 1;
}

.fox-query-bar__filters {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 18px;
  min-width: 0;
}

:deep(.fox-query-bar__item) {
  align-items: center;
  background: transparent;
  display: grid;
  flex: 0 0 var(
    --fox-query-bar-item-width,
    var(--fox-query-bar-responsive-item-width, min(236px, 100%))
  );
  gap: 10px;
  grid-template-columns: max-content minmax(0, 1fr);
  min-width: 0;
}

:deep(.fox-query-bar__item--auto-hidden) {
  display: none;
}

:deep(.fox-query-bar__label) {
  color: var(--shell-subtle-color, #475569);
  font-size: 13px;
  font-weight: 600;
  line-height: 1.3;
  white-space: nowrap;
}

:deep(.fox-query-bar__label--required::before) {
  color: #ef4444;
  content: '*';
  margin-right: 4px;
}

.fox-query-bar__actions {
  align-self: center;
  align-items: center;
  flex: 1 1 0;
  justify-content: flex-end !important;
  margin-left: auto;
  max-width: 100%;
  min-width: min(var(--fox-query-bar-actions-width, 246px), 100%);
  order: 1;
}

.fox-query-bar__filters > :deep(.fox-query-bar__item) {
  order: 0;
}

.fox-query-bar__actions :deep(.n-button) {
  font-weight: 600;
  min-width: 76px;
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease,
    transform 0.18s ease;
}

.fox-query-bar__actions :deep(.n-button:not(.fox-query-bar__expand-button):hover) {
  box-shadow: 0 8px 18px color-mix(in srgb, var(--fox-query-bar-accent) 18%, transparent);
  transform: translateY(-1px);
}

.fox-query-bar__expand-button {
  background: transparent !important;
  box-shadow: none !important;
  min-width: 62px !important;
  transform: none !important;
}

.fox-query-bar__expand-button:hover,
.fox-query-bar__expand-button:focus,
.fox-query-bar__expand-button:active {
  background: transparent !important;
  box-shadow: none !important;
  transform: none !important;
}

.fox-query-bar__expand-button :deep(.n-button__border),
.fox-query-bar__expand-button :deep(.n-button__state-border) {
  border-color: transparent !important;
}

.fox-query-bar__expand-button :deep(.n-button__content) {
  align-items: center;
  display: inline-flex;
  gap: 3px;
  line-height: 1;
}

.fox-query-bar__expand-icon {
  font-size: 16px;
  transition: transform 0.18s ease;
  transform-origin: center;
}

.fox-query-bar__expand-icon--expanded {
  transform: rotate(180deg);
}

:deep(.fox-query-bar__item .n-input),
:deep(.fox-query-bar__item .n-base-selection) {
  min-width: 0;
  width: 100%;
}

:deep(.fox-query-bar__item .n-input__border),
:deep(.fox-query-bar__item .n-input__state-border),
:deep(.fox-query-bar__item .n-base-selection__border),
:deep(.fox-query-bar__item .n-base-selection__state-border) {
  border-radius: 6px;
}

@media (prefers-reduced-motion: reduce) {
  .fox-query-bar__actions :deep(.n-button),
  .fox-query-bar__expand-icon,
  :deep(.fox-query-bar__item) {
    transition: none;
  }

  .fox-query-bar__actions :deep(.n-button:not(.fox-query-bar__expand-button):hover) {
    transform: none;
  }
}

@media (max-width: 1200px) {
  .fox-query-bar__actions {
    justify-content: flex-end;
  }
}

@media (max-width: 760px) {
  .fox-query-bar {
    padding: 16px 14px 14px;
  }

  .fox-query-bar__actions {
    justify-content: flex-end !important;
  }

  .fox-query-bar__fields {
    gap: 14px;
  }

  :deep(.fox-query-bar__item) {
    flex-basis: var(
      --fox-query-bar-item-width,
      var(--fox-query-bar-responsive-item-width, min(180px, 100%))
    );
    grid-template-columns: minmax(56px, max-content) minmax(0, 1fr);
  }
}
</style>
