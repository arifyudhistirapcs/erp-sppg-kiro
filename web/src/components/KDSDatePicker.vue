<template>
  <div class="kds-date-picker">
    <a-space :size="12">
      <a-date-picker
        v-model:value="internalDate"
        :disabled="disabled"
        :loading="loading"
        format="DD/MM/YYYY"
        placeholder="Pilih tanggal"
        :allow-clear="false"
        @change="handleDateChange"
        @keydown="handleKeyDown"
        :get-popup-container="(trigger) => trigger.parentElement"
      >
        <template #suffixIcon>
          <calendar-outlined v-if="!loading" />
          <loading-outlined v-else spin />
        </template>
      </a-date-picker>
      
      <a-button
        type="primary"
        :disabled="disabled || loading || isToday"
        @click="selectToday"
      >
        <template #icon><home-outlined /></template>
        Hari Ini
      </a-button>
    </a-space>
    
    <div v-if="internalDate" class="selected-date-display">
      <span class="date-label">Tanggal Terpilih:</span>
      <strong class="date-value">{{ formatDisplayDate(internalDate) }}</strong>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import dayjs from 'dayjs'
import { CalendarOutlined, HomeOutlined, LoadingOutlined } from '@ant-design/icons-vue'

const props = defineProps({
  modelValue: {
    type: [Object, null],
    default: null
  },
  loading: {
    type: Boolean,
    default: false
  },
  disabled: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'change'])

const SESSION_STORAGE_KEY = 'kds_selected_date'
const internalDate = ref(null)

// Check if the selected date is today
const isToday = computed(() => {
  if (!internalDate.value) return true
  return dayjs(internalDate.value).isSame(dayjs(), 'day')
})

// Format date for display
const formatDisplayDate = (date) => {
  if (!date) return ''
  const dayjsDate = dayjs(date)
  const today = dayjs()
  
  if (dayjsDate.isSame(today, 'day')) {
    return `Hari Ini - ${dayjsDate.format('DD MMMM YYYY')}`
  } else if (dayjsDate.isSame(today.subtract(1, 'day'), 'day')) {
    return `Kemarin - ${dayjsDate.format('DD MMMM YYYY')}`
  } else if (dayjsDate.isSame(today.add(1, 'day'), 'day')) {
    return `Besok - ${dayjsDate.format('DD MMMM YYYY')}`
  } else {
    return dayjsDate.format('dddd, DD MMMM YYYY')
  }
}

// Handle date change from picker
const handleDateChange = (date) => {
  if (date) {
    const jsDate = date.toDate()
    internalDate.value = date
    persistDate(jsDate)
    emit('update:modelValue', jsDate)
    emit('change', jsDate)
  }
}

// Handle "Today" button click
const selectToday = () => {
  const today = dayjs()
  const todayDate = today.toDate()
  internalDate.value = today
  persistDate(todayDate)
  emit('update:modelValue', todayDate)
  emit('change', todayDate)
}

// Keyboard navigation support
const handleKeyDown = (event) => {
  if (disabled || loading) return
  
  const currentDate = internalDate.value ? dayjs(internalDate.value) : dayjs()
  let newDate = null
  
  switch (event.key) {
    case 'ArrowLeft':
      event.preventDefault()
      newDate = currentDate.subtract(1, 'day')
      break
    case 'ArrowRight':
      event.preventDefault()
      newDate = currentDate.add(1, 'day')
      break
    case 'ArrowUp':
      event.preventDefault()
      newDate = currentDate.subtract(7, 'day')
      break
    case 'ArrowDown':
      event.preventDefault()
      newDate = currentDate.add(7, 'day')
      break
    case 'Home':
      event.preventDefault()
      selectToday()
      return
    case 'Escape':
      event.preventDefault()
      // Close the picker by blurring
      event.target.blur()
      return
    default:
      return
  }
  
  if (newDate) {
    const jsDate = newDate.toDate()
    internalDate.value = newDate
    persistDate(jsDate)
    emit('update:modelValue', jsDate)
    emit('change', jsDate)
  }
}

// Persist date to session storage
const persistDate = (date) => {
  try {
    sessionStorage.setItem(SESSION_STORAGE_KEY, date.toISOString())
  } catch (error) {
    console.warn('Failed to persist date to session storage:', error)
  }
}

// Load date from session storage
const loadPersistedDate = () => {
  try {
    const stored = sessionStorage.getItem(SESSION_STORAGE_KEY)
    if (stored) {
      return new Date(stored)
    }
  } catch (error) {
    console.warn('Failed to load date from session storage:', error)
  }
  return null
}

// Initialize date on mount
onMounted(() => {
  // Priority: props.modelValue > session storage > today
  let initialDate = props.modelValue || loadPersistedDate() || new Date()
  internalDate.value = dayjs(initialDate)
  
  // If we loaded from storage or default, emit to parent
  if (!props.modelValue) {
    emit('update:modelValue', initialDate)
  }
})

// Watch for external changes to modelValue
watch(() => props.modelValue, (newValue) => {
  if (newValue) {
    internalDate.value = dayjs(newValue)
  }
})
</script>

<style scoped>
.kds-date-picker {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.selected-date-display {
  padding: 8px 12px;
  background: #f0f5ff;
  border: 1px solid #adc6ff;
  border-radius: 4px;
  font-size: 14px;
}

.date-label {
  color: #595959;
  margin-right: 8px;
}

.date-value {
  color: #1890ff;
  font-size: 15px;
}

/* Loading state styling */
.kds-date-picker :deep(.ant-picker-disabled) {
  background-color: #f5f5f5;
  cursor: not-allowed;
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .kds-date-picker {
    width: 100%;
  }
  
  .kds-date-picker :deep(.ant-space) {
    width: 100%;
    display: flex;
  }
  
  .kds-date-picker :deep(.ant-picker) {
    flex: 1;
  }
}
</style>
