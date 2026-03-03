<template>
  <div>
    <!-- Header Actions -->
    <div class="kds-cleaning-header">
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
      class="cleaning-alert"
    >
      <template #action>
        <a-button size="small" type="primary" @click="retryLoad">
          Coba Lagi
        </a-button>
      </template>
    </a-alert>

    <!-- Kanban Board -->
    <a-spin :spinning="loading" tip="Memuat data...">
      <a-empty
        v-if="!loading && pendingOmpreng.length === 0"
        description="Tidak ada ompreng yang perlu dicuci"
      />

      <div v-else class="kanban-board">
        <!-- Menunggu Column -->
        <div class="kanban-column">
          <div class="kanban-column-header">
            <h3 class="kanban-column-title">
              <ClockCircleOutlined class="column-icon" />
              Menunggu
            </h3>
            <span class="kanban-column-count">{{ pendingItems.length }}</span>
          </div>
          <div class="kanban-column-content">
            <div v-for="record in pendingItems" :key="record.id" class="h-card cleaning-card status-pending">
              <div class="cleaning-card__header">
                <div class="cleaning-card__school">
                  {{ record.school_name || record.delivery_record?.school?.name || '-' }}
                </div>
                <div class="cleaning-card__status-badge status-pending">
                  <span class="status-dot"></span>
                  Menunggu
                </div>
              </div>
              <div class="cleaning-card__count">
                <div class="count-value">{{ record.ompreng_count }}</div>
                <div class="count-label">Unit Ompreng</div>
              </div>
              <div class="cleaning-card__info">
                <div class="info-item">
                  <span class="info-label">Tanggal Pengiriman</span>
                  <span class="info-value">{{ formatDate(record.delivery_date || record.delivery_record?.delivery_date) }}</span>
                </div>
              </div>
              <a-button type="primary" block @click="handleStartCleaning(record)" :loading="updatingId === record.id" class="action-button action-button--start">
                <template #icon><play-circle-outlined /></template>
                Mulai Cuci
              </a-button>
            </div>
          </div>
        </div>

        <!-- Sedang Dicuci Column -->
        <div class="kanban-column">
          <div class="kanban-column-header">
            <h3 class="kanban-column-title">
              <SyncOutlined class="column-icon" />
              Sedang Dicuci
            </h3>
            <span class="kanban-column-count">{{ inProgressItems.length }}</span>
          </div>
          <div class="kanban-column-content">
            <div v-for="record in inProgressItems" :key="record.id" class="h-card cleaning-card status-in_progress">
              <div class="cleaning-card__header">
                <div class="cleaning-card__school">
                  {{ record.school_name || record.delivery_record?.school?.name || '-' }}
                </div>
                <div class="cleaning-card__status-badge status-in_progress">
                  <span class="status-dot"></span>
                  Sedang Dicuci
                </div>
              </div>
              <div class="cleaning-card__count">
                <div class="count-value">{{ record.ompreng_count }}</div>
                <div class="count-label">Unit Ompreng</div>
              </div>
              <div class="cleaning-card__info">
                <div class="info-item">
                  <span class="info-label">Tanggal Pengiriman</span>
                  <span class="info-value">{{ formatDate(record.delivery_date || record.delivery_record?.delivery_date) }}</span>
                </div>
                <div v-if="record.started_at" class="info-item">
                  <span class="info-label">Mulai Cuci</span>
                  <span class="info-value">{{ formatDateTime(record.started_at) }}</span>
                </div>
              </div>
              <a-button type="primary" block @click="handleCompleteCleaning(record)" :loading="updatingId === record.id" class="action-button action-button--complete">
                <template #icon><check-circle-outlined /></template>
                Selesai
              </a-button>
            </div>
          </div>
        </div>

        <!-- Selesai Column -->
        <div class="kanban-column">
          <div class="kanban-column-header">
            <h3 class="kanban-column-title">
              <CheckCircleOutlined class="column-icon" />
              Selesai
            </h3>
            <span class="kanban-column-count">{{ completedItems.length }}</span>
          </div>
          <div class="kanban-column-content">
            <div v-for="record in completedItems" :key="record.id" class="h-card cleaning-card status-completed">
              <div class="cleaning-card__header">
                <div class="cleaning-card__school">
                  {{ record.school_name || record.delivery_record?.school?.name || '-' }}
                </div>
                <div class="cleaning-card__status-badge status-completed">
                  <span class="status-dot"></span>
                  Selesai
                </div>
              </div>
              <div class="cleaning-card__count">
                <div class="count-value">{{ record.ompreng_count }}</div>
                <div class="count-label">Unit Ompreng</div>
              </div>
              <div class="cleaning-card__info">
                <div class="info-item">
                  <span class="info-label">Tanggal Pengiriman</span>
                  <span class="info-value">{{ formatDate(record.delivery_date || record.delivery_record?.delivery_date) }}</span>
                </div>
                <div v-if="record.started_at" class="info-item">
                  <span class="info-label">Mulai Cuci</span>
                  <span class="info-value">{{ formatDateTime(record.started_at) }}</span>
                </div>
                <div v-if="record.completed_at" class="info-item">
                  <span class="info-label">Selesai</span>
                  <span class="info-value">{{ formatDateTime(record.completed_at) }}</span>
                </div>
              </div>
              <div class="completed-badge">
                <check-outlined />
                Sudah Selesai
              </div>
            </div>
          </div>
        </div>
      </div>
    </a-spin>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { message } from 'ant-design-vue'
import {
  WifiOutlined,
  DisconnectOutlined,
  ReloadOutlined,
  PlayCircleOutlined,
  CheckCircleOutlined,
  CheckOutlined,
  ClockCircleOutlined,
  SyncOutlined
} from '@ant-design/icons-vue'
import { getPendingOmpreng, startCleaning, completeCleaning } from '@/services/cleaningService'
import KDSDatePicker from '@/components/KDSDatePicker.vue'
import HStatCard from '@/components/horizon/HStatCard.vue'

let database = null
let dbRef = null
let onValue = null
let off = null

const initFirebase = async () => {
  try {
    const firebaseModule = await import('@/services/firebase')
    const firebaseDatabase = await import('firebase/database')
    database = firebaseModule.database
    dbRef = firebaseDatabase.ref
    onValue = firebaseDatabase.onValue
    off = firebaseDatabase.off
    return true
  } catch (error) {
    console.warn('[KDS Cleaning] Firebase not configured:', error.message)
    return false
  }
}

const pendingOmpreng = ref([])
const loading = ref(false)
const updatingId = ref(null)
const isConnected = ref(true)
const error = ref(null)
const selectedDate = ref(new Date().toISOString().split('T')[0])
let firebaseListener = null

// Filter by status
const pendingItems = computed(() => pendingOmpreng.value.filter(o => o.cleaning_status === 'pending'))
const inProgressItems = computed(() => pendingOmpreng.value.filter(o => o.cleaning_status === 'in_progress'))
const completedItems = computed(() => pendingOmpreng.value.filter(o => o.cleaning_status === 'completed'))

const statusCounts = computed(() => ({
  pending: pendingItems.value.length,
  in_progress: inProgressItems.value.length,
  completed: completedItems.value.length
}))

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  let date
  if (typeof dateStr === 'number') {
    date = new Date(dateStr < 10000000000 ? dateStr * 1000 : dateStr)
  } else {
    date = new Date(dateStr)
  }
  if (isNaN(date.getTime())) return '-'
  const dayNames = ['Minggu', 'Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu']
  const monthNames = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember']
  return `${dayNames[date.getUTCDay()]}, ${date.getUTCDate()} ${monthNames[date.getUTCMonth()]} ${date.getUTCFullYear()}`
}

const formatDateTime = (dateStr) => {
  if (!dateStr) return '-'
  let date
  if (typeof dateStr === 'number') {
    date = new Date(dateStr < 10000000000 ? dateStr * 1000 : dateStr)
  } else {
    date = new Date(dateStr)
  }
  if (isNaN(date.getTime())) return '-'
  const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des']
  return `${date.getUTCDate()} ${monthNames[date.getUTCMonth()]} ${date.getUTCFullYear()}, ${date.getUTCHours().toString().padStart(2, '0')}:${date.getUTCMinutes().toString().padStart(2, '0')}`
}

const loadData = async (date = null) => {
  loading.value = true
  error.value = null
  const dateToLoad = date || selectedDate.value
  try {
    const response = await getPendingOmpreng(dateToLoad)
    if (response.success) {
      pendingOmpreng.value = response.data || []
    } else {
      error.value = response.message || 'Gagal memuat data'
    }
  } catch (err) {
    console.error('Error loading cleaning data:', err)
    error.value = err.response?.data?.message || 'Gagal memuat data. Silakan coba lagi.'
  } finally {
    loading.value = false
  }
}

const handleDateChange = (newDate) => {
  selectedDate.value = newDate
  pendingOmpreng.value = []
  loadData(newDate)
}

const retryLoad = () => loadData()
const refreshData = () => loadData()

const handleStartCleaning = async (record) => {
  updatingId.value = record.id
  try {
    const response = await startCleaning(record.id)
    if (response.success) {
      message.success('Pencucian dimulai')
      await loadData()
    } else {
      message.error(response.message || 'Gagal memulai pencucian')
    }
  } catch (error) {
    console.error('Error starting cleaning:', error)
    message.error(error.response?.data?.message || 'Gagal memulai pencucian')
  } finally {
    updatingId.value = null
  }
}

const handleCompleteCleaning = async (record) => {
  updatingId.value = record.id
  try {
    const response = await completeCleaning(record.id)
    if (response.success) {
      message.success('Pencucian selesai')
      await loadData()
    } else {
      message.error(response.message || 'Gagal menyelesaikan pencucian')
    }
  } catch (error) {
    console.error('Error completing cleaning:', error)
    message.error(error.response?.data?.message || 'Gagal menyelesaikan pencucian')
  } finally {
    updatingId.value = null
  }
}

const setupFirebaseListener = () => {
  if (!database || !dbRef || !onValue) {
    isConnected.value = false
    return
  }
  try {
    cleanupFirebaseListener()
    const cleaningRef = dbRef(database, '/cleaning/pending')
    firebaseListener = onValue(
      cleaningRef,
      (snapshot) => {
        isConnected.value = true
        const data = snapshot.val()
        if (data) {
          const firebaseOmpreng = Object.values(data)
          pendingOmpreng.value = pendingOmpreng.value.map(ompreng => {
            const fr = firebaseOmpreng.find(fo => fo.id === ompreng.id)
            if (fr) {
              return { ...ompreng, cleaning_status: fr.status, started_at: fr.started_at, completed_at: fr.completed_at }
            }
            return ompreng
          })
        }
      },
      (error) => {
        console.warn('[KDS Cleaning] Firebase listener error:', error.code)
        isConnected.value = false
        cleanupFirebaseListener()
      }
    )
  } catch (error) {
    console.error('[KDS Cleaning] Failed to setup Firebase listener:', error)
    isConnected.value = false
  }
}

const cleanupFirebaseListener = () => {
  if (firebaseListener && database && dbRef && off) {
    try {
      const cleaningRef = dbRef(database, '/cleaning/pending')
      off(cleaningRef)
      firebaseListener = null
    } catch (error) { /* ignore */ }
  }
}

onMounted(async () => {
  loadData()
  await initFirebase()
  setupFirebaseListener()
})

onUnmounted(() => {
  cleanupFirebaseListener()
})
</script>

<style scoped>
/* Header */
.kds-cleaning-header {
  display: flex;
  justify-content: flex-end;
  align-items: center;
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

.cleaning-alert {
  margin-bottom: var(--h-spacing-5);
}

/* Stat Cards */
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

/* Cleaning Card */
.cleaning-card {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-4);
  transition: all var(--h-transition-base);
  border-left: 4px solid var(--h-border-color);
}

.cleaning-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.cleaning-card.status-pending { border-left-color: #FFB547; }
.cleaning-card.status-in_progress { border-left-color: #4481EB; }
.cleaning-card.status-completed { border-left-color: #05CD99; }

.cleaning-card__header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--h-spacing-3);
}

.cleaning-card__school {
  flex: 1;
  font-size: var(--h-text-base);
  font-weight: var(--h-font-bold);
  color: var(--h-text-primary);
  line-height: var(--h-leading-tight);
}

.cleaning-card__status-badge {
  display: flex;
  align-items: center;
  gap: var(--h-spacing-2);
  padding: 4px 12px;
  border-radius: var(--h-radius-sm);
  font-size: var(--h-text-xs);
  font-weight: var(--h-font-medium);
  flex-shrink: 0;
}

.cleaning-card__status-badge.status-pending { background: rgba(255, 181, 71, 0.1); color: #FFB547; }
.cleaning-card__status-badge.status-in_progress { background: rgba(68, 129, 235, 0.1); color: #4481EB; }
.cleaning-card__status-badge.status-completed { background: rgba(5, 205, 153, 0.1); color: #05CD99; }

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}

/* Count */
.cleaning-card__count {
  text-align: center;
  padding: var(--h-spacing-4) 0;
  background: var(--h-bg-light);
  border-radius: var(--h-radius-md);
}

.count-value {
  font-size: 36px;
  font-weight: var(--h-font-bold);
  color: var(--h-primary);
  line-height: 1;
  margin-bottom: var(--h-spacing-2);
}

.count-label {
  font-size: var(--h-text-sm);
  color: var(--h-text-secondary);
  font-weight: var(--h-font-medium);
}

/* Info */
.cleaning-card__info {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-3);
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: var(--h-spacing-3);
  border-bottom: 1px solid var(--h-border-light);
}

.info-item:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.info-label {
  font-size: var(--h-text-sm);
  color: var(--h-text-secondary);
  font-weight: var(--h-font-medium);
}

.info-value {
  font-size: var(--h-text-sm);
  color: var(--h-text-primary);
  font-weight: var(--h-font-semibold);
  text-align: right;
}

/* Action */
.action-button {
  margin-top: var(--h-spacing-2);
  height: var(--h-touch-target-min);
  font-weight: var(--h-font-semibold);
  border-radius: var(--h-radius-md);
}

.action-button--start {
  background: var(--h-primary) !important;
  border-color: var(--h-primary) !important;
}

.action-button--complete {
  background: #05CD99 !important;
  border-color: #05CD99 !important;
}

.action-button--complete:hover {
  background: #01B574 !important;
  border-color: #01B574 !important;
}

.completed-badge {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--h-spacing-2);
  padding: var(--h-spacing-3);
  background: rgba(5, 205, 153, 0.1);
  border-radius: var(--h-radius-md);
  color: #05CD99;
  font-size: var(--h-text-sm);
  font-weight: var(--h-font-semibold);
}

/* Dark Mode */
.dark .kanban-column-header { background: var(--h-bg-card); }
.dark .kanban-column-title { color: var(--h-text-primary); }
.dark .cleaning-card__school { color: var(--h-text-primary); }
.dark .cleaning-card__count { background: rgba(163, 174, 208, 0.05); }
.dark .count-value { color: var(--h-primary-light); }
.dark .info-value { color: var(--h-text-primary); }
.dark .info-item { border-bottom-color: var(--h-border-color); }
.dark .completed-badge { background: rgba(5, 205, 153, 0.2); }

/* Responsive */
@media (max-width: 1024px) {
  .stat-cards-row { grid-template-columns: repeat(3, 1fr); gap: var(--h-spacing-3); }
  .kanban-board { grid-template-columns: 1fr; gap: var(--h-spacing-6); }
  .kanban-column { min-height: auto; }
}

@media (max-width: 768px) {
  .header-controls { flex-wrap: wrap; gap: var(--h-spacing-2); }
  .stat-cards-row { grid-template-columns: 1fr; gap: var(--h-spacing-3); }
  .count-value { font-size: 28px; }
}
</style>
