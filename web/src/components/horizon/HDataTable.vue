<template>
  <div class="h-data-table h-card">
    <!-- Mobile Card View -->
    <div v-if="isMobile && mobileCardView" class="h-data-table__mobile-cards">
      <div
        v-for="(record, index) in dataSource"
        :key="record.key || index"
        class="h-data-table__mobile-card"
      >
        <div
          v-for="column in columns"
          :key="column.key || column.dataIndex"
          class="h-data-table__mobile-field"
        >
          <div class="h-data-table__mobile-label">
            {{ column.title }}
          </div>
          <div class="h-data-table__mobile-value">
            <slot
              v-if="column.customRender"
              :name="`cell-${column.dataIndex}`"
              :record="record"
              :text="record[column.dataIndex]"
            >
              {{ record[column.dataIndex] }}
            </slot>
            <template v-else>
              {{ record[column.dataIndex] }}
            </template>
          </div>
        </div>
      </div>
    </div>

    <!-- Desktop Table View -->
    <div v-else class="h-data-table__wrapper">
      <a-table
        :columns="columns"
        :data-source="dataSource"
        :loading="loading"
        :pagination="pagination"
        :row-selection="rowSelection"
        :scroll="{ x: 'max-content' }"
        :row-class-name="() => 'h-data-table__row'"
        class="h-data-table__table"
      >
        <!-- Custom Cell Rendering -->
        <template
          v-for="column in columns"
          :key="column.key || column.dataIndex"
          #[`bodyCell`]="{ column: col, record, text }"
        >
          <template v-if="col.dataIndex === column.dataIndex">
            <!-- Status Badge Rendering -->
            <span
              v-if="column.type === 'status'"
              class="h-data-table__status-badge"
              :class="`h-data-table__status-badge--${getStatusType(text)}`"
            >
              <span class="h-data-table__status-dot"></span>
              <span class="h-data-table__status-text">{{ text }}</span>
            </span>

            <!-- Progress Bar Rendering -->
            <div v-else-if="column.type === 'progress'" class="h-data-table__progress">
              <a-progress
                :percent="text"
                :stroke-color="getProgressColor(text)"
                :show-info="true"
                size="small"
              />
            </div>

            <!-- Action Buttons -->
            <div v-else-if="column.type === 'actions'" class="h-data-table__actions">
              <slot name="actions" :record="record" />
            </div>

            <!-- Custom Slot -->
            <slot
              v-else-if="$slots[`cell-${column.dataIndex}`]"
              :name="`cell-${column.dataIndex}`"
              :record="record"
              :text="text"
            />

            <!-- Default Text -->
            <span v-else>{{ text }}</span>
          </template>
        </template>
      </a-table>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useBreakpoint } from '@/composables/useBreakpoint'

const props = defineProps({
  /**
   * Table columns configuration
   * Example: [{ title: 'Name', dataIndex: 'name', key: 'name' }]
   */
  columns: {
    type: Array,
    required: true
  },

  /**
   * Table data source
   * Example: [{ key: '1', name: 'John', age: 32 }]
   */
  dataSource: {
    type: Array,
    default: () => []
  },

  /**
   * Loading state
   */
  loading: {
    type: Boolean,
    default: false
  },

  /**
   * Pagination configuration
   * Example: { current: 1, pageSize: 10, total: 50 }
   */
  pagination: {
    type: [Object, Boolean],
    default: () => ({
      current: 1,
      pageSize: 10,
      showSizeChanger: true,
      showTotal: (total) => `Total ${total} items`
    })
  },

  /**
   * Row selection configuration
   * Example: { selectedRowKeys: [], onChange: (keys) => {} }
   */
  rowSelection: {
    type: Object,
    default: null
  },

  /**
   * Enable mobile card view (stack rows as cards on mobile)
   */
  mobileCardView: {
    type: Boolean,
    default: true
  }
})

// Breakpoint detection
const { isMobile } = useBreakpoint()

/**
 * Get status type for badge styling
 */
const getStatusType = (status) => {
  const statusLower = String(status).toLowerCase()
  
  if (statusLower.includes('approved') || statusLower.includes('done') || statusLower.includes('completed') || statusLower.includes('success')) {
    return 'success'
  }
  if (statusLower.includes('pending') || statusLower.includes('waiting') || statusLower.includes('in progress')) {
    return 'warning'
  }
  if (statusLower.includes('error') || statusLower.includes('failed') || statusLower.includes('rejected')) {
    return 'error'
  }
  if (statusLower.includes('disable') || statusLower.includes('inactive')) {
    return 'disabled'
  }
  
  return 'default'
}

/**
 * Get progress bar color based on percentage
 */
const getProgressColor = (percent) => {
  if (percent >= 80) return 'var(--h-success)'
  if (percent >= 50) return 'var(--h-warning)'
  return 'var(--h-error)'
}
</script>

<style scoped>
/* ========================================
   TABLE CONTAINER
   ======================================== */

.h-data-table {
  /* Use h-card utility class for base styling */
  overflow: hidden;
}

.h-data-table__wrapper {
  width: 100%;
  overflow-x: auto;
  position: relative;
}

/* Horizontal scroll indicator on mobile */
@media (max-width: 767px) {
  .h-data-table__wrapper {
    /* Show scroll indicator */
    scrollbar-width: thin;
    scrollbar-color: var(--h-primary-light) var(--h-bg-light);
  }
  
  .h-data-table__wrapper::-webkit-scrollbar {
    height: 6px;
  }
  
  .h-data-table__wrapper::-webkit-scrollbar-track {
    background: var(--h-bg-light);
    border-radius: var(--h-radius-sm);
  }
  
  .h-data-table__wrapper::-webkit-scrollbar-thumb {
    background: var(--h-primary-light);
    border-radius: var(--h-radius-sm);
  }
}

/* ========================================
   TABLE STYLING
   ======================================== */

.h-data-table__table :deep(.ant-table) {
  background: transparent;
  font-family: var(--h-font-primary);
}

/* Table Header */
.h-data-table__table :deep(.ant-table-thead > tr > th) {
  background: transparent;
  border-bottom: 1px solid var(--h-border-color);
  padding: var(--h-spacing-4) var(--h-spacing-4);
  font-size: var(--h-text-xs);
  font-weight: var(--h-font-bold);
  color: var(--h-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  height: var(--h-table-header-height);
}

/* Table Body */
.h-data-table__table :deep(.ant-table-tbody > tr > td) {
  border-bottom: 1px solid var(--h-border-light);
  padding: var(--h-spacing-4) var(--h-spacing-4);
  font-size: var(--h-text-sm);
  color: var(--h-text-primary);
  height: var(--h-table-row-height);
  vertical-align: middle;
}

/* Row Hover Effect */
.h-data-table__table :deep(.ant-table-tbody > tr:hover > td) {
  background: #F8FDEA !important;
  transition: background var(--h-transition-fast);
}

/* Remove default Ant Design borders */
.h-data-table__table :deep(.ant-table-container) {
  border: none;
}

.h-data-table__table :deep(.ant-table-cell) {
  border-right: none;
}

/* ========================================
   STATUS BADGE
   ======================================== */

.h-data-table__status-badge {
  display: inline-flex;
  align-items: center;
  gap: var(--h-spacing-2);
  padding: 4px 12px;
  border-radius: var(--h-radius-sm);
  font-size: var(--h-text-xs);
  font-weight: var(--h-font-medium);
  line-height: 1;
}

.h-data-table__status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}

/* Status Badge Variants */
.h-data-table__status-badge--success {
  background: rgba(5, 205, 153, 0.1);
  color: var(--h-success-dark);
}

.h-data-table__status-badge--success .h-data-table__status-dot {
  background: var(--h-success);
}

.h-data-table__status-badge--warning {
  background: rgba(255, 181, 71, 0.1);
  color: var(--h-warning-dark);
}

.h-data-table__status-badge--warning .h-data-table__status-dot {
  background: var(--h-warning);
}

.h-data-table__status-badge--error {
  background: rgba(238, 93, 80, 0.1);
  color: var(--h-error-dark);
}

.h-data-table__status-badge--error .h-data-table__status-dot {
  background: var(--h-error);
}

.h-data-table__status-badge--disabled {
  background: rgba(163, 174, 208, 0.1);
  color: var(--h-text-secondary);
}

.h-data-table__status-badge--disabled .h-data-table__status-dot {
  background: var(--h-text-light);
}

.h-data-table__status-badge--default {
  background: var(--h-bg-light);
  color: var(--h-text-primary);
}

.h-data-table__status-badge--default .h-data-table__status-dot {
  background: var(--h-text-secondary);
}

/* ========================================
   PROGRESS BAR
   ======================================== */

.h-data-table__progress {
  width: 100%;
  max-width: 200px;
}

.h-data-table__progress :deep(.ant-progress-inner) {
  background: var(--h-bg-light);
}

.h-data-table__progress :deep(.ant-progress-text) {
  font-size: var(--h-text-xs);
  color: var(--h-text-primary);
  font-weight: var(--h-font-medium);
}

/* ========================================
   ACTION BUTTONS
   ======================================== */

.h-data-table__actions {
  display: flex;
  gap: var(--h-spacing-2);
  align-items: center;
}

.h-data-table__actions :deep(.ant-btn) {
  min-width: 32px;
  height: 32px;
  padding: 0 var(--h-spacing-3);
  border-radius: var(--h-radius-sm);
  font-size: var(--h-text-sm);
  transition: all var(--h-transition-fast);
}

.h-data-table__actions :deep(.ant-btn:hover) {
  transform: scale(1.05);
}

/* Mobile: Larger touch targets */
@media (max-width: 767px) {
  .h-data-table__actions :deep(.ant-btn) {
    min-width: var(--h-touch-target-min);
    height: var(--h-touch-target-min);
    padding: 0 var(--h-spacing-4);
  }
}

/* ========================================
   MOBILE CARD VIEW
   ======================================== */

.h-data-table__mobile-cards {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-4);
}

.h-data-table__mobile-card {
  background: var(--h-bg-secondary);
  border: 1px solid var(--h-border-color);
  border-radius: var(--h-radius-md);
  padding: var(--h-spacing-4);
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-3);
  transition: all var(--h-transition-base);
}

.h-data-table__mobile-card:active {
  transform: scale(0.98);
  background: #F8FDEA;
}

.h-data-table__mobile-field {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--h-spacing-3);
  padding: var(--h-spacing-2) 0;
  border-bottom: 1px solid var(--h-border-light);
}

.h-data-table__mobile-field:last-child {
  border-bottom: none;
}

.h-data-table__mobile-label {
  font-size: var(--h-text-xs);
  font-weight: var(--h-font-bold);
  color: var(--h-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  flex-shrink: 0;
  min-width: 100px;
}

.h-data-table__mobile-value {
  font-size: var(--h-text-sm);
  color: var(--h-text-primary);
  font-weight: var(--h-font-medium);
  text-align: right;
  flex: 1;
  word-break: break-word;
}

/* ========================================
   PAGINATION
   ======================================== */

.h-data-table__table :deep(.ant-pagination) {
  margin-top: var(--h-spacing-5);
  font-family: var(--h-font-primary);
}

.h-data-table__table :deep(.ant-pagination-item) {
  border-radius: var(--h-radius-sm);
  border-color: var(--h-border-color);
  font-weight: var(--h-font-medium);
}

.h-data-table__table :deep(.ant-pagination-item-active) {
  background: var(--h-primary);
  border-color: var(--h-primary);
}

.h-data-table__table :deep(.ant-pagination-item-active a) {
  color: #FFFFFF;
}

/* ========================================
   DARK MODE SUPPORT
   ======================================== */

.dark .h-data-table__table :deep(.ant-table-thead > tr > th) {
  color: var(--h-text-secondary);
  border-bottom-color: var(--h-border-color);
}

.dark .h-data-table__table :deep(.ant-table-tbody > tr > td) {
  color: var(--h-text-primary);
  border-bottom-color: var(--h-border-color);
}

.dark .h-data-table__table :deep(.ant-table-tbody > tr:hover > td) {
  background: rgba(90, 67, 114, 0.1) !important;
}

.dark .h-data-table__mobile-card {
  background: var(--h-bg-secondary);
  border-color: var(--h-border-color);
}

.dark .h-data-table__mobile-card:active {
  background: rgba(90, 67, 114, 0.1);
}

.dark .h-data-table__mobile-label {
  color: var(--h-text-secondary);
}

.dark .h-data-table__mobile-value {
  color: var(--h-text-primary);
}

.dark .h-data-table__progress :deep(.ant-progress-inner) {
  background: rgba(163, 174, 208, 0.1);
}

.dark .h-data-table__progress :deep(.ant-progress-text) {
  color: var(--h-text-primary);
}

/* ========================================
   LOADING STATE
   ======================================== */

.h-data-table__table :deep(.ant-spin-container) {
  min-height: 200px;
}

.h-data-table__table :deep(.ant-spin) {
  color: var(--h-primary);
}
</style>
