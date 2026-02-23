<template>
  <van-dialog 
    v-model:show="visible" 
    title="Status Sinkronisasi"
    :show-cancel-button="false"
    confirm-button-text="Tutup"
    class="sync-status-modal"
  >
    <div class="sync-status-content">
      <!-- Overall Status -->
      <div class="status-section">
        <div class="status-header">
          <van-icon :name="getStatusIcon()" :color="getStatusColor()" size="20" />
          <span class="status-text">{{ getStatusText() }}</span>
        </div>
        
        <div v-if="syncProgress.status === 'syncing'" class="progress-section">
          <van-progress 
            :percentage="progressPercentage" 
            :color="getStatusColor()"
            stroke-width="6"
            :show-pivot="false"
          />
          <div class="progress-text">
            {{ syncProgress.completed }} / {{ syncProgress.total }} item
          </div>
        </div>
      </div>

      <!-- Pending Items -->
      <div v-if="pendingCount > 0" class="pending-section">
        <div class="section-title">
          <van-icon name="clock-o" color="#ff976a" />
          <span>{{ pendingCount }} Item Menunggu</span>
        </div>
        
        <div class="pending-items">
          <div v-for="item in pendingItems" :key="item.id" class="pending-item">
            <div class="item-info">
              <span class="item-type">{{ getItemTypeText(item.type) }}</span>
              <span class="item-time">{{ formatTime(item.createdAt) }}</span>
            </div>
            <div v-if="item.retryCount > 0" class="retry-info">
              <van-tag type="warning" size="small">
                Percobaan ke-{{ item.retryCount + 1 }}
              </van-tag>
            </div>
          </div>
        </div>
      </div>

      <!-- Failed Items -->
      <div v-if="failedItems.length > 0" class="failed-section">
        <div class="section-title">
          <van-icon name="warning-o" color="#ee0a24" />
          <span>{{ failedItems.length }} Item Gagal</span>
        </div>
        
        <div class="failed-items">
          <div v-for="item in failedItems" :key="item.id" class="failed-item">
            <div class="item-info">
              <span class="item-type">{{ getItemTypeText(item.type) }}</span>
              <span class="error-message">{{ item.errorMessage || 'Error tidak diketahui' }}</span>
            </div>
            <van-button 
              type="primary" 
              size="mini"
              @click="retryItem(item)"
              :loading="retryingItems.includes(item.id)"
            >
              Coba Lagi
            </van-button>
          </div>
        </div>
      </div>

      <!-- Statistics -->
      <div v-if="statistics.total > 0" class="stats-section">
        <div class="section-title">
          <van-icon name="bar-chart-o" color="#1989fa" />
          <span>Statistik 24 Jam Terakhir</span>
        </div>
        
        <div class="stats-grid">
          <div class="stat-item">
            <div class="stat-value">{{ statistics.total }}</div>
            <div class="stat-label">Total</div>
          </div>
          <div class="stat-item success">
            <div class="stat-value">{{ statistics.successful }}</div>
            <div class="stat-label">Berhasil</div>
          </div>
          <div class="stat-item failed">
            <div class="stat-value">{{ statistics.failed }}</div>
            <div class="stat-label">Gagal</div>
          </div>
        </div>
      </div>

      <!-- Actions -->
      <div class="actions-section">
        <van-button 
          v-if="pendingCount > 0 && isOnline"
          type="primary" 
          size="small"
          @click="forceSyncAll"
          :loading="isForceSyncing"
          block
        >
          <van-icon name="refresh" />
          Paksa Sinkronisasi Semua
        </van-button>
        
        <van-button 
          v-if="failedItems.length > 0"
          type="warning" 
          size="small"
          @click="retryAllFailed"
          :loading="isRetryingAll"
          block
          class="retry-all-btn"
        >
          <van-icon name="replay" />
          Coba Lagi Semua yang Gagal
        </van-button>
      </div>
    </div>
  </van-dialog>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useDeliveryTasksStore } from '@/stores/deliveryTasks'
import { showToast, showSuccessToast } from 'vant'
import syncService from '@/services/syncService'
import db from '@/services/db'

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:show'])

const deliveryTasksStore = useDeliveryTasksStore()

// Reactive data
const visible = ref(false)
const pendingCount = ref(0)
const pendingItems = ref([])
const failedItems = ref([])
const statistics = ref({ total: 0, successful: 0, failed: 0, byType: {} })
const syncProgress = ref({ status: 'idle', total: 0, completed: 0, failed: 0 })
const isOnline = ref(navigator.onLine)
const isForceSyncing = ref(false)
const isRetryingAll = ref(false)
const retryingItems = ref([])

// Computed
const progressPercentage = computed(() => {
  if (syncProgress.value.total === 0) return 0
  return Math.round((syncProgress.value.completed / syncProgress.value.total) * 100)
})

// Watch props
watch(() => props.show, (newValue) => {
  visible.value = newValue
  if (newValue) {
    loadSyncStatus()
  }
})

watch(visible, (newValue) => {
  emit('update:show', newValue)
})

// Methods
const loadSyncStatus = async () => {
  try {
    // Get pending items
    pendingCount.value = await syncService.getPendingSyncCount()
    
    // Get pending and failed items details
    const allPendingItems = await db.syncQueue
      .where('status')
      .anyOf(['pending', 'failed'])
      .toArray()
    
    pendingItems.value = allPendingItems.filter(item => item.status === 'pending')
    failedItems.value = allPendingItems.filter(item => item.status === 'failed')
    
    // Get sync progress
    syncProgress.value = syncService.getSyncProgress()
    
    // Get statistics
    statistics.value = await syncService.getSyncStatistics()
  } catch (error) {
    console.error('Error loading sync status:', error)
  }
}

const getStatusIcon = () => {
  switch (syncProgress.value.status) {
    case 'syncing': return 'loading'
    case 'completed': return 'success'
    case 'completed_with_errors': return 'warning-o'
    case 'error': return 'close'
    default: return 'clock-o'
  }
}

const getStatusColor = () => {
  switch (syncProgress.value.status) {
    case 'syncing': return '#1989fa'
    case 'completed': return '#07c160'
    case 'completed_with_errors': return '#ff976a'
    case 'error': return '#ee0a24'
    default: return '#969799'
  }
}

const getStatusText = () => {
  if (!isOnline.value) {
    return 'Offline - Menunggu koneksi internet'
  }
  
  switch (syncProgress.value.status) {
    case 'syncing': return 'Sedang menyinkronkan...'
    case 'completed': return 'Sinkronisasi selesai'
    case 'completed_with_errors': return 'Sinkronisasi selesai dengan error'
    case 'error': return 'Error sinkronisasi'
    default: 
      if (pendingCount.value > 0) {
        return `${pendingCount.value} item menunggu sinkronisasi`
      }
      return 'Semua data tersinkronisasi'
  }
}

const getItemTypeText = (type) => {
  switch (type) {
    case 'epod': return 'Bukti Pengiriman'
    case 'epod_photo': return 'Foto e-POD'
    case 'epod_signature': return 'Tanda Tangan e-POD'
    case 'delivery_status': return 'Status Pengiriman'
    default: return type
  }
}

const formatTime = (timestamp) => {
  const date = new Date(timestamp)
  const now = new Date()
  const diffMs = now - date
  const diffMins = Math.floor(diffMs / 60000)
  
  if (diffMins < 1) return 'Baru saja'
  if (diffMins < 60) return `${diffMins} menit lalu`
  
  const diffHours = Math.floor(diffMins / 60)
  if (diffHours < 24) return `${diffHours} jam lalu`
  
  const diffDays = Math.floor(diffHours / 24)
  return `${diffDays} hari lalu`
}

const retryItem = async (item) => {
  if (retryingItems.value.includes(item.id)) return
  
  retryingItems.value.push(item.id)
  
  try {
    // Reset item status to pending
    await db.syncQueue.update(item.id, {
      status: 'pending',
      retryCount: 0,
      lastAttempt: null,
      errorMessage: null
    })
    
    // Trigger sync
    if (isOnline.value) {
      syncService.syncAllOfflineData()
    }
    
    showSuccessToast('Item akan dicoba lagi')
    await loadSyncStatus()
  } catch (error) {
    console.error('Error retrying item:', error)
    showToast('Gagal mengatur ulang item')
  } finally {
    retryingItems.value = retryingItems.value.filter(id => id !== item.id)
  }
}

const retryAllFailed = async () => {
  if (isRetryingAll.value) return
  
  isRetryingAll.value = true
  
  try {
    await syncService.retryFailedSyncItems()
    showSuccessToast('Semua item gagal akan dicoba lagi')
    await loadSyncStatus()
  } catch (error) {
    console.error('Error retrying all failed items:', error)
    showToast('Gagal mengatur ulang item yang gagal')
  } finally {
    isRetryingAll.value = false
  }
}

const forceSyncAll = async () => {
  if (isForceSyncing.value || !isOnline.value) return
  
  isForceSyncing.value = true
  
  try {
    await syncService.syncAllOfflineData()
    showSuccessToast('Sinkronisasi dipaksa dimulai')
  } catch (error) {
    console.error('Error forcing sync:', error)
    showToast('Gagal memaksa sinkronisasi')
  } finally {
    isForceSyncing.value = false
  }
}

// Sync progress listener
const onSyncProgress = (progress) => {
  syncProgress.value = progress
  
  if (progress.status === 'completed' || progress.status === 'completed_with_errors') {
    loadSyncStatus() // Refresh data when sync completes
  }
}

// Network status handlers
const handleOnline = () => {
  isOnline.value = true
  loadSyncStatus()
}

const handleOffline = () => {
  isOnline.value = false
}

// Lifecycle
onMounted(() => {
  // Add sync progress listener
  deliveryTasksStore.addSyncProgressListener(onSyncProgress)
  
  // Listen for network status changes
  window.addEventListener('online', handleOnline)
  window.addEventListener('offline', handleOffline)
})

onUnmounted(() => {
  // Remove sync progress listener
  deliveryTasksStore.removeSyncProgressListener(onSyncProgress)
  
  // Remove event listeners
  window.removeEventListener('online', handleOnline)
  window.removeEventListener('offline', handleOffline)
})
</script>

<style scoped>
.sync-status-modal {
  width: 90%;
  max-width: 400px;
}

.sync-status-content {
  padding: 16px;
  max-height: 60vh;
  overflow-y: auto;
}

.status-section {
  margin-bottom: 20px;
  text-align: center;
}

.status-header {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 12px;
  font-size: 16px;
  font-weight: 600;
}

.status-header .van-icon {
  margin-right: 8px;
}

.progress-section {
  margin-top: 12px;
}

.progress-text {
  margin-top: 8px;
  font-size: 14px;
  color: #646566;
}

.section-title {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  font-size: 14px;
  font-weight: 600;
  color: #323233;
}

.section-title .van-icon {
  margin-right: 8px;
}

.pending-section,
.failed-section {
  margin-bottom: 20px;
}

.pending-items,
.failed-items {
  background-color: #f7f8fa;
  border-radius: 8px;
  padding: 12px;
}

.pending-item,
.failed-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid #ebedf0;
}

.pending-item:last-child,
.failed-item:last-child {
  border-bottom: none;
}

.item-info {
  flex: 1;
}

.item-type {
  display: block;
  font-size: 14px;
  font-weight: 500;
  color: #323233;
}

.item-time {
  display: block;
  font-size: 12px;
  color: #969799;
  margin-top: 2px;
}

.error-message {
  display: block;
  font-size: 12px;
  color: #ee0a24;
  margin-top: 2px;
}

.retry-info {
  margin-left: 8px;
}

.stats-section {
  margin-bottom: 20px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-top: 12px;
}

.stat-item {
  text-align: center;
  padding: 12px;
  background-color: #f7f8fa;
  border-radius: 8px;
}

.stat-item.success {
  background-color: rgba(7, 193, 96, 0.1);
}

.stat-item.failed {
  background-color: rgba(238, 10, 36, 0.1);
}

.stat-value {
  font-size: 20px;
  font-weight: 600;
  color: #323233;
}

.stat-item.success .stat-value {
  color: #07c160;
}

.stat-item.failed .stat-value {
  color: #ee0a24;
}

.stat-label {
  font-size: 12px;
  color: #646566;
  margin-top: 4px;
}

.actions-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.retry-all-btn {
  margin-top: 8px;
}

/* Responsive adjustments */
@media (max-width: 375px) {
  .sync-status-content {
    padding: 12px;
  }
  
  .stats-grid {
    grid-template-columns: repeat(3, 1fr);
    gap: 8px;
  }
  
  .stat-item {
    padding: 8px;
  }
  
  .stat-value {
    font-size: 18px;
  }
}
</style>