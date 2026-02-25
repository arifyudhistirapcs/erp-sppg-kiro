<template>
  <div class="school-allocation-input">
    <div class="allocation-header">
      <h4>Alokasi Sekolah</h4>
      <div class="allocation-summary" :class="summaryClass">
        <span class="summary-text">
          Dialokasikan: {{ totalAllocated }} / {{ totalPortions }} porsi
        </span>
        <a-tag v-if="isValid" color="success">
          <CheckCircleOutlined /> Valid
        </a-tag>
        <a-tag v-else color="warning">
          <ExclamationCircleOutlined /> Belum Valid
        </a-tag>
      </div>
    </div>

    <div v-if="errorMessage" class="error-message">
      <a-alert :message="errorMessage" type="error" show-icon closable />
    </div>

    <div class="schools-list">
      <div
        v-for="school in schools"
        :key="school.id"
        class="school-row"
      >
        <div class="school-info">
          <span class="school-name">{{ school.name }}</span>
          <span class="school-meta">{{ school.student_count || 0 }} siswa</span>
        </div>
        <div class="school-input">
          <a-input-number
            v-model:value="allocations[school.id]"
            :min="0"
            :max="totalPortions"
            placeholder="0"
            @change="handleAllocationChange"
            style="width: 120px"
          />
          <span class="unit-label">porsi</span>
        </div>
      </div>
    </div>

    <div v-if="!isValid && totalAllocated > 0" class="validation-hint">
      <a-alert :message="validationHint" type="info" show-icon />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { CheckCircleOutlined, ExclamationCircleOutlined } from '@ant-design/icons-vue'

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
const allocations = ref({ ...props.modelValue })

// Watch for external changes to modelValue (including reset to empty object)
watch(() => props.modelValue, (newValue) => {
  // If modelValue is empty object, reset all allocations to 0
  if (Object.keys(newValue).length === 0) {
    const resetAllocations = {}
    props.schools.forEach(school => {
      resetAllocations[school.id] = 0
    })
    allocations.value = resetAllocations
  } else {
    allocations.value = { ...newValue }
  }
}, { deep: true })

// Watch for changes to schools to initialize allocations
watch(() => props.schools, (newSchools) => {
  if (newSchools && newSchools.length > 0) {
    // Only initialize if allocations is empty or doesn't have all schools
    const needsInit = Object.keys(allocations.value).length === 0
    
    if (needsInit) {
      // Initialize all schools with 0
      const newAllocations = {}
      newSchools.forEach(school => {
        newAllocations[school.id] = 0
      })
      allocations.value = newAllocations
    } else {
      // Just add missing schools
      newSchools.forEach(school => {
        if (!(school.id in allocations.value)) {
          allocations.value[school.id] = 0
        }
      })
    }
  }
}, { immediate: true })

const totalAllocated = computed(() => {
  return Object.values(allocations.value).reduce((sum, val) => sum + (val || 0), 0)
})

const isValid = computed(() => {
  if (props.totalPortions === 0) return false
  if (totalAllocated.value === 0) return false
  return totalAllocated.value === props.totalPortions
})

const summaryClass = computed(() => ({
  'summary-valid': isValid.value,
  'summary-invalid': !isValid.value && totalAllocated.value > 0,
  'summary-empty': totalAllocated.value === 0
}))

const errorMessage = computed(() => {
  if (totalAllocated.value === 0 && props.totalPortions > 0) {
    return 'Harap alokasikan porsi ke minimal satu sekolah'
  }
  if (totalAllocated.value > props.totalPortions) {
    const excess = totalAllocated.value - props.totalPortions
    return `Alokasi melebihi total porsi sebanyak ${excess} porsi`
  }
  return null
})

const validationHint = computed(() => {
  if (totalAllocated.value < props.totalPortions) {
    const remaining = props.totalPortions - totalAllocated.value
    return `Masih perlu mengalokasikan ${remaining} porsi lagi`
  }
  return ''
})

const handleAllocationChange = () => {
  // Emit the updated allocations
  emit('update:modelValue', { ...allocations.value })
  
  // Emit validation status
  emit('validation-change', {
    isValid: isValid.value,
    totalAllocated: totalAllocated.value,
    totalPortions: props.totalPortions
  })
}

// Initialize validation on mount
watch(() => props.totalPortions, () => {
  handleAllocationChange()
}, { immediate: true })
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
  justify-content: space-between;
  align-items: center;
  padding: 12px;
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
  flex-direction: column;
  gap: 4px;
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
  align-items: center;
  gap: 8px;
}

.unit-label {
  font-size: 13px;
  color: #8c8c8c;
}

.validation-hint {
  margin-top: 12px;
}
</style>
