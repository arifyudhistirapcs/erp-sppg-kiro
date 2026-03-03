<template>
  <div>
    <!-- Header Controls -->
    <div class="packing-header">
      <div class="header-controls">
        <KDSDatePicker
          v-model="selectedDate"
          :loading="loading"
          @change="handleDateChange"
        />
        <a-tag :color="isConnected ? 'green' : 'red'" class="connection-tag">
          <template #icon>
            <wifi-outlined v-if="isConnected" />
            <disconnect-outlined v-else />
          </template>
          {{ isConnected ? 'Terhubung' : 'Terputus' }}
        </a-tag>
        <a-button @click="refreshData" :loading="loading" type="default">
          <template #icon><reload-outlined /></template>
          Refresh
        </a-button>
      </div>
    </div>



    <!-- Error Alert -->
    <a-alert
      v-if="error"
      type="error"
      :message="error"
      closable
      show-icon
      @close="error = null"
      class="error-alert"
    >
      <template #action>
        <a-button size="small" type="primary" @click="retryLoad">
          Coba Lagi
        </a-button>
      </template>
    </a-alert>



    <!-- Kanban Board -->
    <a-spin :spinning="loading" tip="Memuat data...">
      <a-empty v-if="!loading && schools.length === 0" :description="emptyMessage" />
      
      <div v-else class="kanban-board">
        <!-- Pending Column -->
        <div class="kanban-column">
          <div class="kanban-column-header">
            <h3 class="kanban-column-title">
              <ClockCircleOutlined class="column-icon" />
              Belum Mulai
            </h3>
            <span class="kanban-column-count">{{ pendingSchools.length }}</span>
          </div>
          <div class="kanban-column-content">
            <HKanbanCard
              v-for="school in pendingSchools"
              :key="school.school_id"
              :title="school.school_name"
              status="Belum Mulai"
              class="packing-card"
            >
              <div class="card-body">
                <!-- Total Portions -->
                <div class="total-portions">
                  <span class="portions-label">Total Porsi</span>
                  <span class="portions-value">{{ school.total_portions }}</span>
                </div>

                <!-- Portion Size Breakdown -->
                <div v-if="school.portion_size_type === 'mixed'" class="portion-breakdown">
                  <div class="portion-size-item portion-small">
                    <span class="size-badge">S</span>
                    <span class="size-label">Kecil</span>
                    <span class="size-value">{{ school.portions_small }}</span>
                  </div>
                  <div class="portion-size-item portion-large">
                    <span class="size-badge">L</span>
                    <span class="size-label">Besar</span>
                    <span class="size-value">{{ school.portions_large }}</span>
                  </div>
                </div>

                <!-- Menu Items -->
                <div class="menu-items-section">
                  <div class="section-title">Menu Items</div>
                  <div class="menu-items-list">
                    <div v-for="item in school.menu_items" :key="item.recipe_id" class="menu-item">
                      <img v-if="item.photo_url" :src="item.photo_url" :alt="item.recipe_name" class="menu-photo" />
                      <div v-else class="menu-photo-placeholder">
                        <picture-outlined />
                      </div>
                      <div class="menu-info">
                        <div class="menu-name">{{ item.recipe_name }}</div>
                        <div class="menu-portions">
                          <span v-if="item.portions_small > 0" class="portion-tag small">S: {{ item.portions_small }}</span>
                          <span v-if="item.portions_large > 0" class="portion-tag large">L: {{ item.portions_large }}</span>
                        </div>
                        
                        <!-- Ingredients/Components -->
                        <div v-if="item.items && item.items.length > 0" class="ingredients-list">
                          <div v-for="(ingredient, idx) in item.items" :key="idx" class="ingredient-item">
                            <span class="ingredient-name">{{ ingredient.name }}</span>
                            <div class="ingredient-quantities">
                              <span v-if="item.portions_small > 0 && ingredient.quantity_per_small > 0" class="ingredient-per-portion small">
                                S: {{ ingredient.quantity_per_small.toFixed(1) }} {{ ingredient.unit }}
                              </span>
                              <span v-if="item.portions_large > 0 && ingredient.quantity_per_large > 0" class="ingredient-per-portion large">
                                L: {{ ingredient.quantity_per_large.toFixed(1) }} {{ ingredient.unit }}
                              </span>
                              <span class="ingredient-total">
                                Total: {{ ingredient.quantity.toFixed(1) }} {{ ingredient.unit }}
                              </span>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Action Button -->
                <a-button
                  type="primary"
                  block
                  @click="startPacking(school)"
                  :loading="updatingSchoolId === school.school_id"
                  class="action-button"
                >
                  <template #icon><play-circle-outlined /></template>
                  Mulai Packing
                </a-button>
              </div>
            </HKanbanCard>
          </div>
        </div>

        <!-- Packing Column -->
        <div class="kanban-column">
          <div class="kanban-column-header">
            <h3 class="kanban-column-title">
              <InboxOutlined class="column-icon" />
              Siap Packing
            </h3>
            <span class="kanban-column-count">{{ packingSchools.length }}</span>
          </div>
          <div class="kanban-column-content">
            <HKanbanCard
              v-for="school in packingSchools"
              :key="school.school_id"
              :title="school.school_name"
              status="Siap Packing"
              class="packing-card"
            >
              <div class="card-body">
                <!-- Total Portions -->
                <div class="total-portions">
                  <span class="portions-label">Total Porsi</span>
                  <span class="portions-value">{{ school.total_portions }}</span>
                </div>

                <!-- Portion Size Breakdown -->
                <div v-if="school.portion_size_type === 'mixed'" class="portion-breakdown">
                  <div class="portion-size-item portion-small">
                    <span class="size-badge">S</span>
                    <span class="size-label">Kecil</span>
                    <span class="size-value">{{ school.portions_small }}</span>
                  </div>
                  <div class="portion-size-item portion-large">
                    <span class="size-badge">L</span>
                    <span class="size-label">Besar</span>
                    <span class="size-value">{{ school.portions_large }}</span>
                  </div>
                </div>

                <!-- Menu Items -->
                <div class="menu-items-section">
                  <div class="section-title">Menu Items</div>
                  <div class="menu-items-list">
                    <div v-for="item in school.menu_items" :key="item.recipe_id" class="menu-item">
                      <img v-if="item.photo_url" :src="item.photo_url" :alt="item.recipe_name" class="menu-photo" />
                      <div v-else class="menu-photo-placeholder">
                        <picture-outlined />
                      </div>
                      <div class="menu-info">
                        <div class="menu-name">{{ item.recipe_name }}</div>
                        <div class="menu-portions">
                          <span v-if="item.portions_small > 0" class="portion-tag small">S: {{ item.portions_small }}</span>
                          <span v-if="item.portions_large > 0" class="portion-tag large">L: {{ item.portions_large }}</span>
                        </div>
                        
                        <!-- Ingredients/Components -->
                        <div v-if="item.items && item.items.length > 0" class="ingredients-list">
                          <div v-for="(ingredient, idx) in item.items" :key="idx" class="ingredient-item">
                            <span class="ingredient-name">{{ ingredient.name }}</span>
                            <div class="ingredient-quantities">
                              <span v-if="item.portions_small > 0 && ingredient.quantity_per_small > 0" class="ingredient-per-portion small">
                                S: {{ ingredient.quantity_per_small.toFixed(1) }} {{ ingredient.unit }}
                              </span>
                              <span v-if="item.portions_large > 0 && ingredient.quantity_per_large > 0" class="ingredient-per-portion large">
                                L: {{ ingredient.quantity_per_large.toFixed(1) }} {{ ingredient.unit }}
                              </span>
                              <span class="ingredient-total">
                                Total: {{ ingredient.quantity.toFixed(1) }} {{ ingredient.unit }}
                              </span>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Action Button -->
                <a-button
                  type="primary"
                  block
                  @click="finishPacking(school)"
                  :loading="updatingSchoolId === school.school_id"
                  class="action-button action-button-success"
                >
                  <template #icon><check-circle-outlined /></template>
                  Selesai Packing
                </a-button>
              </div>
            </HKanbanCard>
          </div>
        </div>

        <!-- Ready Column -->
        <div class="kanban-column">
          <div class="kanban-column-header">
            <h3 class="kanban-column-title">
              <CheckCircleOutlined class="column-icon" />
              Selesai Packing
            </h3>
            <span class="kanban-column-count">{{ readySchools.length }}</span>
          </div>
          <div class="kanban-column-content">
            <HKanbanCard
              v-for="school in readySchools"
              :key="school.school_id"
              :title="school.school_name"
              status="Selesai Packing"
              class="packing-card"
            >
              <div class="card-body">
                <!-- Total Portions -->
                <div class="total-portions">
                  <span class="portions-label">Total Porsi</span>
                  <span class="portions-value">{{ school.total_portions }}</span>
                </div>

                <!-- Portion Size Breakdown -->
                <div v-if="school.portion_size_type === 'mixed'" class="portion-breakdown">
                  <div class="portion-size-item portion-small">
                    <span class="size-badge">S</span>
                    <span class="size-label">Kecil</span>
                    <span class="size-value">{{ school.portions_small }}</span>
                  </div>
                  <div class="portion-size-item portion-large">
                    <span class="size-badge">L</span>
                    <span class="size-label">Besar</span>
                    <span class="size-value">{{ school.portions_large }}</span>
                  </div>
                </div>

                <!-- Menu Items -->
                <div class="menu-items-section">
                  <div class="section-title">Menu Items</div>
                  <div class="menu-items-list">
                    <div v-for="item in school.menu_items" :key="item.recipe_id" class="menu-item">
                      <img v-if="item.photo_url" :src="item.photo_url" :alt="item.recipe_name" class="menu-photo" />
                      <div v-else class="menu-photo-placeholder">
                        <picture-outlined />
                      </div>
                      <div class="menu-info">
                        <div class="menu-name">{{ item.recipe_name }}</div>
                        <div class="menu-portions">
                          <span v-if="item.portions_small > 0" class="portion-tag small">S: {{ item.portions_small }}</span>
                          <span v-if="item.portions_large > 0" class="portion-tag large">L: {{ item.portions_large }}</span>
                        </div>
                        
                        <!-- Ingredients/Components -->
                        <div v-if="item.items && item.items.length > 0" class="ingredients-list">
                          <div v-for="(ingredient, idx) in item.items" :key="idx" class="ingredient-item">
                            <span class="ingredient-name">{{ ingredient.name }}</span>
                            <div class="ingredient-quantities">
                              <span v-if="item.portions_small > 0 && ingredient.quantity_per_small > 0" class="ingredient-per-portion small">
                                S: {{ ingredient.quantity_per_small.toFixed(1) }} {{ ingredient.unit }}
                              </span>
                              <span v-if="item.portions_large > 0 && ingredient.quantity_per_large > 0" class="ingredient-per-portion large">
                                L: {{ ingredient.quantity_per_large.toFixed(1) }} {{ ingredient.unit }}
                              </span>
                              <span class="ingredient-total">
                                Total: {{ ingredient.quantity.toFixed(1) }} {{ ingredient.unit }}
                              </span>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Completed Badge -->
                <div class="completed-badge">
                  <check-outlined />
                  Sudah Siap
                </div>
              </div>
            </HKanbanCard>
          </div>
        </div>
      </div>
    </a-spin>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { message, notification } from 'ant-design-vue'
import {
  WifiOutlined,
  DisconnectOutlined,
  ReloadOutlined,
  PlayCircleOutlined,
  CheckCircleOutlined,
  CheckOutlined,
  PictureOutlined,
  ClockCircleOutlined,
  InboxOutlined
} from '@ant-design/icons-vue'
import HStatCard from '@/components/horizon/HStatCard.vue'
import HKanbanCard from '@/components/horizon/HKanbanCard.vue'
import KDSDatePicker from '@/components/KDSDatePicker.vue'
import { getPackingToday, updatePackingStatus } from '@/services/kdsService'
import { database } from '@/services/firebase'
import { ref as dbRef, onValue, off } from 'firebase/database'

const schools = ref([])
const loading = ref(false)
const updatingSchoolId = ref(null)
const isConnected = ref(true)
const selectedDate = ref(new Date())
const error = ref(null)
let firebaseListener = null
let notificationListener = null

// Computed: Filter schools by status
const pendingSchools = computed(() => schools.value.filter(s => s.status === 'pending'))
const packingSchools = computed(() => schools.value.filter(s => s.status === 'packing'))
const readySchools = computed(() => schools.value.filter(s => s.status === 'ready'))

// Computed: Count by status
const pendingCount = computed(() => pendingSchools.value.length)
const packingCount = computed(() => packingSchools.value.length)
const readyCount = computed(() => readySchools.value.length)

// Computed: Check if all schools are ready
const allSchoolsReady = computed(() => {
  return schools.value.length > 0 && schools.value.every(s => s.status === 'ready')
})

// Compute empty message based on selected date
const emptyMessage = computed(() => {
  const today = new Date()
  const isToday = selectedDate.value.toDateString() === today.toDateString()
  return isToday ? 'Tidak ada alokasi packing untuk hari ini' : 'Tidak ada alokasi packing untuk tanggal ini'
})

// Load data from API
const loadData = async () => {
  loading.value = true
  error.value = null
  try {
    const response = await getPackingToday(selectedDate.value)
    if (response.success) {
      schools.value = response.data || []
    } else {
      error.value = response.message || 'Gagal memuat data'
    }
  } catch (err) {
    console.error('Error loading packing data:', err)
    error.value = err.response?.data?.message || 'Gagal memuat data alokasi packing. Silakan coba lagi.'
  } finally {
    loading.value = false
  }
}

// Retry loading data
const retryLoad = () => {
  loadData()
}

// Refresh data
const refreshData = () => {
  loadData()
}

// Start packing for a school
const startPacking = async (school) => {
  updatingSchoolId.value = school.school_id
  try {
    const response = await updatePackingStatus(school.school_id, 'packing')
    if (response.success) {
      message.success('Status berhasil diperbarui: Mulai Packing')
      await loadData()
    } else {
      message.error(response.message || 'Gagal memperbarui status')
    }
  } catch (error) {
    console.error('Error updating status:', error)
    message.error(error.response?.data?.message || 'Gagal memperbarui status')
  } finally {
    updatingSchoolId.value = null
  }
}

// Finish packing for a school
const finishPacking = async (school) => {
  updatingSchoolId.value = school.school_id
  try {
    const response = await updatePackingStatus(school.school_id, 'ready')
    if (response.success) {
      message.success(`${school.school_name} siap untuk pengiriman!`)
      await loadData()
    } else {
      message.error(response.message || 'Gagal memperbarui status')
    }
  } catch (error) {
    console.error('Error updating status:', error)
    message.error(error.response?.data?.message || 'Gagal memperbarui status')
  } finally {
    updatingSchoolId.value = null
  }
}

// Setup Firebase real-time listener for packing data
const setupFirebaseListener = () => {
  cleanupFirebaseListener()
  
  const dateStr = selectedDate.value.toISOString().split('T')[0]
  const packingRef = dbRef(database, `/kds/packing/${dateStr}`)
  
  firebaseListener = onValue(
    packingRef,
    (snapshot) => {
      isConnected.value = true
      const data = snapshot.val()
      
      if (data) {
        const firebaseSchools = Object.values(data)
        schools.value = schools.value.map(school => {
          const firebaseSchool = firebaseSchools.find(fs => fs.school_id === school.school_id)
          if (firebaseSchool) {
            return {
              ...school,
              status: firebaseSchool.status,
              portion_size_type: firebaseSchool.portion_size_type || school.portion_size_type,
              portions_small: firebaseSchool.portions_small !== undefined ? firebaseSchool.portions_small : school.portions_small,
              portions_large: firebaseSchool.portions_large !== undefined ? firebaseSchool.portions_large : school.portions_large,
              total_portions: firebaseSchool.total_portions || school.total_portions
            }
          }
          return school
        })
      }
    },
    (error) => {
      console.error('Firebase listener error:', error)
      isConnected.value = false
    }
  )
}

// Setup Firebase listener for notifications
const setupNotificationListener = () => {
  const notificationRef = dbRef(database, '/notifications/logistics/packing_complete')
  
  notificationListener = onValue(
    notificationRef,
    (snapshot) => {
      const data = snapshot.val()
      
      if (data) {
        const notifications = Object.values(data)
        const latest = notifications[notifications.length - 1]
        
        if (latest && latest.message) {
          notification.success({
            message: 'Notifikasi',
            description: latest.message,
            duration: 10,
            placement: 'topRight'
          })
        }
      }
    },
    (error) => {
      console.error('Notification listener error:', error)
    }
  )
}

// Cleanup Firebase listener
const cleanupFirebaseListener = () => {
  if (firebaseListener) {
    const dateStr = selectedDate.value.toISOString().split('T')[0]
    const packingRef = dbRef(database, `/kds/packing/${dateStr}`)
    off(packingRef)
    firebaseListener = null
  }
}

// Cleanup notification listener
const cleanupNotificationListener = () => {
  if (notificationListener) {
    const notificationRef = dbRef(database, '/notifications/logistics/packing_complete')
    off(notificationRef)
    notificationListener = null
  }
}

// Handle date change from date picker
const handleDateChange = (date) => {
  selectedDate.value = date
  loadData()
  setupFirebaseListener()
}

// Watch for date changes
watch(selectedDate, () => {
  setupFirebaseListener()
})

onMounted(() => {
  loadData()
  setupFirebaseListener()
  setupNotificationListener()
})

onUnmounted(() => {
  cleanupFirebaseListener()
  cleanupNotificationListener()
})
</script>

<style scoped>
/* Header Controls */
.packing-header {
  display: flex;
  justify-content: flex-end;
  margin-bottom: var(--h-spacing-4);
}

.header-controls {
  display: flex;
  align-items: center;
  gap: var(--h-spacing-3);
}

.connection-tag {
  font-size: var(--h-text-sm);
  padding: 4px 12px;
  border-radius: var(--h-radius-sm);
}

/* Alerts */
.ready-alert,
.error-alert {
  margin-bottom: var(--h-spacing-5);
}

/* Stat Cards Row */
.stat-cards-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--h-spacing-5);
  margin-bottom: var(--h-spacing-6);
}

/* Kanban Board */
.kanban-board {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--h-spacing-5);
  align-items: start;
}

/* Kanban Column */
.kanban-column {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-4);
  min-height: 400px;
}

.kanban-column-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--h-spacing-4);
  background: var(--h-bg-card);
  border-radius: var(--h-radius-lg);
  box-shadow: var(--h-shadow-sm);
}

.kanban-column-title {
  display: flex;
  align-items: center;
  gap: var(--h-spacing-2);
  margin: 0;
  font-size: var(--h-text-lg);
  font-weight: var(--h-font-bold);
  color: var(--h-text-primary);
}

.column-icon {
  font-size: var(--h-text-xl);
  color: var(--h-primary);
}

.kanban-column-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  height: 28px;
  padding: 0 var(--h-spacing-2);
  background: var(--h-primary);
  color: white;
  border-radius: var(--h-radius-full);
  font-size: var(--h-text-sm);
  font-weight: var(--h-font-bold);
}

.kanban-column-content {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-4);
}

/* Packing Card Content */
.packing-card {
  cursor: default;
}

.card-body {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-4);
}

/* Total Portions */
.total-portions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--h-spacing-4);
  background: linear-gradient(135deg, #5A4372 0%, #3D2B53 100%);
  border-radius: var(--h-radius-md);
  color: white;
}

.portions-label {
  font-size: var(--h-text-sm);
  font-weight: var(--h-font-medium);
}

.portions-value {
  font-size: var(--h-text-2xl);
  font-weight: var(--h-font-bold);
}

/* Portion Size Breakdown */
.portion-breakdown {
  display: flex;
  gap: var(--h-spacing-3);
}

.portion-size-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--h-spacing-2);
  padding: var(--h-spacing-3);
  border-radius: var(--h-radius-md);
  transition: all var(--h-transition-base);
}

.portion-size-item:hover {
  transform: translateY(-2px);
}

.portion-size-item.portion-small {
  background: linear-gradient(135deg, #FFF4E6 0%, #FFE7BA 100%);
  border: 2px solid #FFB547;
}

.portion-size-item.portion-large {
  background: linear-gradient(135deg, #E6F7FF 0%, #BAE7FF 100%);
  border: 2px solid #1890FF;
}

.portion-size-item.single {
  background: linear-gradient(135deg, #E6F7FF 0%, #BAE7FF 100%);
  border: 2px solid #1890FF;
}

.size-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: white;
  border-radius: var(--h-radius-full);
  font-size: var(--h-text-base);
  font-weight: var(--h-font-bold);
  color: var(--h-primary);
  box-shadow: var(--h-shadow-sm);
}

.size-label {
  font-size: var(--h-text-xs);
  font-weight: var(--h-font-medium);
  color: var(--h-text-secondary);
}

.size-value {
  font-size: var(--h-text-xl);
  font-weight: var(--h-font-bold);
  color: var(--h-text-primary);
}

/* Menu Items Section */
.menu-items-section {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-3);
}

.section-title {
  font-size: var(--h-text-sm);
  font-weight: var(--h-font-bold);
  color: var(--h-text-primary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.menu-items-list {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-2);
}

.menu-item {
  display: flex;
  gap: var(--h-spacing-3);
  padding: var(--h-spacing-3);
  background: var(--h-bg-light);
  border-radius: var(--h-radius-md);
  transition: all var(--h-transition-base);
}

.menu-item:hover {
  background: var(--h-border-color);
}

.menu-photo {
  width: 48px;
  height: 48px;
  border-radius: var(--h-radius-sm);
  object-fit: cover;
  flex-shrink: 0;
}

.menu-photo-placeholder {
  width: 48px;
  height: 48px;
  border-radius: var(--h-radius-sm);
  background: var(--h-border-color);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--h-text-light);
  font-size: var(--h-text-xl);
  flex-shrink: 0;
}

.menu-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-1);
  min-width: 0;
}

.menu-name {
  font-size: var(--h-text-sm);
  font-weight: var(--h-font-medium);
  color: var(--h-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.menu-portions {
  display: flex;
  gap: var(--h-spacing-1);
}

.portion-tag {
  font-size: var(--h-text-xs);
  font-weight: var(--h-font-medium);
  padding: 2px 8px;
  border-radius: var(--h-radius-sm);
}

.portion-tag.small {
  background: #FFF4E6;
  color: #FFB547;
  border: 1px solid #FFD591;
}

.portion-tag.large {
  background: #E6F7FF;
  color: #1890FF;
  border: 1px solid #91D5FF;
}

/* Ingredients List */
.ingredients-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: var(--h-spacing-2);
  padding-top: var(--h-spacing-2);
  border-top: 1px solid var(--h-border-color);
}

.ingredient-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: var(--h-text-xs);
  padding: 4px 0;
}

.ingredient-name {
  color: var(--h-text-secondary);
  font-weight: var(--h-font-semibold);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ingredient-quantities {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.ingredient-per-portion {
  font-size: 10px;
  font-weight: var(--h-font-medium);
  padding: 2px 6px;
  border-radius: 4px;
  white-space: nowrap;
}

.ingredient-per-portion.small {
  background: #FFF4E6;
  color: #FFB547;
  border: 1px solid #FFD591;
}

.ingredient-per-portion.large {
  background: #E6F7FF;
  color: #1890FF;
  border: 1px solid #91D5FF;
}

.ingredient-total {
  color: var(--h-text-primary);
  font-weight: var(--h-font-bold);
  font-size: 11px;
  white-space: nowrap;
}

/* Action Button */
.action-button {
  margin-top: var(--h-spacing-2);
  height: var(--h-touch-target-min);
  font-weight: var(--h-font-semibold);
  border-radius: var(--h-radius-md);
}

.action-button-success {
  background-color: #05CD99;
  border-color: #05CD99;
}

.action-button-success:hover {
  background-color: #04b888;
  border-color: #04b888;
}

/* Completed Badge */
.completed-badge {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--h-spacing-2);
  padding: var(--h-spacing-3);
  background: #E6FAF5;
  color: #05CD99;
  border-radius: var(--h-radius-md);
  font-size: var(--h-text-sm);
  font-weight: var(--h-font-semibold);
  margin-top: var(--h-spacing-2);
}

/* Dark Mode Support */
.dark .kanban-column-header {
  background: var(--h-bg-card);
}

.dark .kanban-column-title {
  color: var(--h-text-primary);
}

.dark .column-icon {
  color: var(--h-primary-light);
}

.dark .total-portions {
  background: linear-gradient(135deg, #6a5382 0%, #4d3b63 100%);
}

.dark .portion-size-item.portion-small {
  background: rgba(255, 181, 71, 0.15);
  border-color: #FFB547;
}

.dark .portion-size-item.portion-large,
.dark .portion-size-item.single {
  background: rgba(24, 144, 255, 0.15);
  border-color: #1890FF;
}

.dark .size-badge {
  background: var(--h-bg-card);
}

.dark .size-label {
  color: var(--h-text-secondary);
}

.dark .size-value {
  color: var(--h-text-primary);
}

.dark .menu-item {
  background: rgba(255, 255, 255, 0.05);
}

.dark .menu-item:hover {
  background: rgba(255, 255, 255, 0.08);
}

.dark .menu-photo-placeholder {
  background: rgba(255, 255, 255, 0.05);
  color: var(--h-text-light);
}

.dark .menu-name {
  color: var(--h-text-primary);
}

.dark .portion-tag.small {
  background: rgba(255, 181, 71, 0.2);
  color: #FFB547;
  border-color: rgba(255, 181, 71, 0.3);
}

.dark .portion-tag.large {
  background: rgba(24, 144, 255, 0.2);
  color: #1890FF;
  border-color: rgba(24, 144, 255, 0.3);
}

.dark .ingredients-list {
  border-top-color: rgba(255, 255, 255, 0.1);
}

.dark .ingredient-name {
  color: var(--h-text-secondary);
}

.dark .ingredient-per-portion.small {
  background: rgba(255, 181, 71, 0.2);
  color: #FFB547;
  border-color: rgba(255, 181, 71, 0.3);
}

.dark .ingredient-per-portion.large {
  background: rgba(24, 144, 255, 0.2);
  color: #1890FF;
  border-color: rgba(24, 144, 255, 0.3);
}

.dark .ingredient-total {
  color: var(--h-text-primary);
}

.dark .completed-badge {
  background: rgba(5, 205, 153, 0.2);
  color: #05CD99;
}

/* Mobile Responsive */
@media (max-width: 1024px) {
  .stat-cards-row {
    grid-template-columns: repeat(3, 1fr);
    gap: var(--h-spacing-3);
  }
  
  .kanban-board {
    grid-template-columns: 1fr;
    gap: var(--h-spacing-6);
  }
  
  .kanban-column {
    min-height: auto;
  }
}

@media (max-width: 768px) {
  .header-controls {
    flex-wrap: wrap;
    gap: var(--h-spacing-2);
  }
  
  .stat-cards-row {
    grid-template-columns: 1fr;
    gap: var(--h-spacing-3);
  }
  
  .portion-breakdown {
    flex-direction: column;
  }
  
  .portion-size-item {
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
  }
  
  .size-badge {
    width: 28px;
    height: 28px;
    font-size: var(--h-text-sm);
  }
  
  .menu-photo,
  .menu-photo-placeholder {
    width: 40px;
    height: 40px;
  }
}
</style>
