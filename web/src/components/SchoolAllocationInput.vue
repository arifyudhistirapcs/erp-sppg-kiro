<template>
  <div class="school-allocation-input">
    <div class="allocation-header">
      <h4>Alokasi Sekolah</h4>
    </div>

    <div class="schools-list">
      <div
        v-for="school in schools"
        :key="school.id"
        class="school-row"
      >
        <div class="school-info">
          <span class="school-name">{{ school.name }}</span>
          <a-tag :color="getSchoolCategoryColor(school.category)" size="small">
            {{ school.category }}
          </a-tag>
          <span class="school-meta">
            <template v-if="school.category === 'SD'">
              Kelas 1-3: {{ school.student_count_grade_1_3 || 0 }} siswa | 
              Kelas 4-6: {{ school.student_count_grade_4_6 || 0 }} siswa
            </template>
            <template v-else>
              {{ school.student_count || 0 }} siswa
            </template>
          </span>
        </div>
        <div class="school-input">
          <!-- SD schools: show both small and large portion fields -->
          <template v-if="school.category === 'SD'">
            <div class="portion-inputs-wrapper">
              <div class="portion-input-row">
                <div class="portion-field">
                  <label class="portion-label">Kecil (Kelas 1-3)</label>
                  <a-input-number
                    v-model:value="allocations[school.id].portions_small"
                    :min="0"
                    placeholder="0"
                    :disabled="autoFillEnabled[school.id]"
                    @change="handleAllocationChange"
                    style="width: 120px"
                  />
                </div>
                <div class="portion-field">
                  <label class="portion-label">Besar (Kelas 4-6)</label>
                  <a-input-number
                    v-model:value="allocations[school.id].portions_large"
                    :min="0"
                    placeholder="0"
                    :disabled="autoFillEnabled[school.id]"
                    @change="handleAllocationChange"
                    style="width: 120px"
                  />
                </div>
                <span class="unit-label">porsi</span>
              </div>
              <div class="auto-fill-row">
                <a-checkbox
                  v-model:checked="autoFillEnabled[school.id]"
                  @change="handleAutoFillChange(school)"
                >
                  <span class="checkbox-text">Samakan seperti jumlah siswa</span>
                </a-checkbox>
              </div>
            </div>
          </template>
          <!-- SMP/SMA schools: show only large portion field -->
          <template v-else>
            <div class="portion-input-group">
              <div class="portion-field">
                <label class="portion-label">Besar</label>
                <a-input-number
                  v-model:value="allocations[school.id].portions_large"
                  :min="0"
                  placeholder="0"
                  @change="handleAllocationChange"
                  style="width: 100px"
                />
              </div>
              <span class="unit-label">porsi</span>
            </div>
          </template>
        </div>
      </div>
    </div>

    <!-- Portion Size Statistics -->
    <div v-if="totalAllocated > 0" class="statistics-section">
      <a-divider style="margin: 16px 0 12px 0">Statistik Alokasi</a-divider>
      <div class="statistics-grid">
        <div class="stat-item">
          <span class="stat-label">Total Porsi Kecil</span>
          <span class="stat-value">{{ totalSmallPortions }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Total Porsi Besar</span>
          <span class="stat-value">{{ totalLargePortions }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Persentase Kecil</span>
          <span class="stat-value">{{ smallPortionPercentage }}%</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Persentase Besar</span>
          <span class="stat-value">{{ largePortionPercentage }}%</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Sekolah SD</span>
          <span class="stat-value">{{ sdSchoolCount }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Sekolah SMP/SMA</span>
          <span class="stat-value">{{ smpSmaSchoolCount }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
  schools: {
    type: Array,
    required: true,
    default: () => []
  },
  totalPortions: {
    type: Number,
    required: true,
    default: 0,
    validator: (value) => {
      return typeof value === 'number' && value >= 0
    }
  },
  modelValue: {
    type: Object,
    default: () => ({})
  }
})

const emit = defineEmits(['update:modelValue', 'validation-change'])

// Initialize allocations from modelValue
// Structure: { school_id: { portions_small: 0, portions_large: 0 } }
const allocations = ref({})
const autoFillEnabled = ref({})
const isInternalUpdate = ref(false)

// Helper function to get school category color
const getSchoolCategoryColor = (category) => {
  switch (category) {
    case 'SD':
      return 'blue'
    case 'SMP':
      return 'green'
    case 'SMA':
      return 'purple'
    default:
      return 'default'
  }
}

// Initialize allocations structure
const initializeAllocations = () => {
  const newAllocations = {}
  const newAutoFill = {}
  props.schools.forEach(school => {
    newAllocations[school.id] = {
      portions_small: 0,
      portions_large: 0
    }
    newAutoFill[school.id] = false
  })
  autoFillEnabled.value = newAutoFill
  return newAllocations
}

// Initialize immediately with schools data
if (props.schools && props.schools.length > 0) {
  if (props.modelValue && Object.keys(props.modelValue).length > 0) {
    // Load from modelValue if provided (for edit mode)
    const loadedAllocations = {}
    props.schools.forEach(school => {
      if (props.modelValue[school.id]) {
        loadedAllocations[school.id] = {
          portions_small: props.modelValue[school.id].portions_small || 0,
          portions_large: props.modelValue[school.id].portions_large || 0
        }
      } else {
        loadedAllocations[school.id] = {
          portions_small: 0,
          portions_large: 0
        }
      }
    })
    allocations.value = loadedAllocations
  } else {
    // Initialize with zeros (for new mode)
    allocations.value = initializeAllocations()
  }
}

// Watch modelValue ONLY for external changes (not from our own emit)
watch(() => props.modelValue, (newValue) => {
  // Skip if this is our own update
  if (isInternalUpdate.value) {
    isInternalUpdate.value = false
    return
  }
  
  // Only update if modelValue is explicitly reset to empty (when opening new modal)
  if (newValue && Object.keys(newValue).length === 0 && Object.keys(allocations.value).length > 0) {
    allocations.value = initializeAllocations()
  }
}, { deep: false }) // Use shallow watch to avoid triggering on nested changes

const totalAllocated = computed(() => {
  let total = 0
  Object.values(allocations.value).forEach(alloc => {
    total += (alloc.portions_small || 0) + (alloc.portions_large || 0)
  })
  return total
})

const isValid = computed(() => {
  // Always valid - user can allocate any amount
  return totalAllocated.value > 0
})

// Statistics computed properties
const totalSmallPortions = computed(() => {
  let total = 0
  Object.values(allocations.value).forEach(alloc => {
    total += alloc.portions_small || 0
  })
  return total
})

const totalLargePortions = computed(() => {
  let total = 0
  Object.values(allocations.value).forEach(alloc => {
    total += alloc.portions_large || 0
  })
  return total
})

const smallPortionPercentage = computed(() => {
  if (totalAllocated.value === 0) return 0
  return ((totalSmallPortions.value / totalAllocated.value) * 100).toFixed(1)
})

const largePortionPercentage = computed(() => {
  if (totalAllocated.value === 0) return 0
  return ((totalLargePortions.value / totalAllocated.value) * 100).toFixed(1)
})

const sdSchoolCount = computed(() => {
  let count = 0
  props.schools.forEach(school => {
    const alloc = allocations.value[school.id]
    if (school.category === 'SD' && alloc && (alloc.portions_small > 0 || alloc.portions_large > 0)) {
      count++
    }
  })
  return count
})

const smpSmaSchoolCount = computed(() => {
  let count = 0
  props.schools.forEach(school => {
    const alloc = allocations.value[school.id]
    if ((school.category === 'SMP' || school.category === 'SMA') && alloc && alloc.portions_large > 0) {
      count++
    }
  })
  return count
})

const handleAllocationChange = () => {
  console.log('handleAllocationChange called')
  console.log('allocations:', allocations.value)
  console.log('totalAllocated:', totalAllocated.value)
  console.log('isValid:', isValid.value)
  
  // Set flag to indicate this is an internal update
  isInternalUpdate.value = true
  
  // Emit the updated allocations
  emit('update:modelValue', { ...allocations.value })
  
  // Emit validation status - always valid if there are allocations
  emit('validation-change', {
    isValid: isValid.value,
    totalAllocated: totalAllocated.value,
    totalPortions: totalAllocated.value // Set totalPortions same as totalAllocated
  })
  
  console.log('Emitted validation-change with totalAllocated:', totalAllocated.value)
}

const handleAutoFillChange = (school) => {
  if (autoFillEnabled.value[school.id]) {
    // Auto-fill enabled: set portions to match student counts
    allocations.value[school.id].portions_small = school.student_count_grade_1_3 || 0
    allocations.value[school.id].portions_large = school.student_count_grade_4_6 || 0
  } else {
    // Auto-fill disabled: reset to 0
    allocations.value[school.id].portions_small = 0
    allocations.value[school.id].portions_large = 0
  }
  handleAllocationChange()
}
</script>

<style scoped>
.school-allocation-input {
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  padding: 16px;
  background: #fafafa;
}

.allocation-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.allocation-header h4 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.allocation-summary {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  border-radius: 4px;
  background: white;
  border: 1px solid #d9d9d9;
}

.allocation-summary.summary-valid {
  border-color: #52c41a;
  background: #f6ffed;
}

.allocation-summary.summary-invalid {
  border-color: #faad14;
  background: #fffbe6;
}

.allocation-summary.summary-empty {
  border-color: #d9d9d9;
  background: white;
}

.summary-text {
  font-weight: 500;
  font-size: 14px;
}

.error-message {
  margin-bottom: 16px;
}

.schools-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 16px;
}

.school-row {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
  background: white;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  transition: all 0.3s;
}

.school-row:hover {
  border-color: #40a9ff;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.school-info {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.school-name {
  font-weight: 500;
  font-size: 14px;
  color: #262626;
}

.school-meta {
  font-size: 12px;
  color: #8c8c8c;
}

.school-input {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.portion-inputs-wrapper {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.portion-input-row {
  display: flex;
  gap: 12px;
  align-items: flex-end;
}

.auto-fill-row {
  display: flex;
  align-items: center;
  padding-left: 4px;
}

.checkbox-text {
  font-size: 13px;
  color: #595959;
}

.portion-input-group {
  display: flex;
  gap: 12px;
  align-items: flex-end;
}

.auto-fill-checkbox {
  margin-bottom: 4px;
  white-space: nowrap;
}

.portion-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.portion-label {
  font-size: 11px;
  color: #8c8c8c;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.unit-label {
  font-size: 13px;
  color: #8c8c8c;
  margin-left: 4px;
}

.validation-hint {
  margin-top: 12px;
}

.statistics-section {
  margin-top: 16px;
}

.statistics-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px;
  background: white;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
}

.stat-label {
  font-size: 11px;
  color: #8c8c8c;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.stat-value {
  font-size: 18px;
  font-weight: 600;
  color: #1890ff;
}
</style>
