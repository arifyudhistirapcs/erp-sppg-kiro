<template>
  <div class="sync-status-container">
    <!-- Sync Status Badge -->
    <van-badge 
      v-if="pendingCount > 0" 
      :content="pendingCount" 
      :max="99"
      class="sync-badge"
    >
      <van-icon 
        :name="syncIcon" 
        :class="syncIconClass"
        @click="showSyncDetails = true"
      />
    </van-badge>
    
    <!-- No pending items -->
    <van-icon 
      v-else
      :name="syncIcon" 
      :class="syncIconClass"
      @click="showSyncDetails = true"
    />

    <!-- Sync Progress Toast -->
    <van-toast 
      v-model:show="showSyncProgress"
      type="loading"
      :message="syncProgressMessage"
      :duration="0"
      :forbid-click="true"
    />

    <!-- Sync Details Dialog -->
    <van-dialog 
      v-model:show="showSyncDetails" 
      title="Status Sinkronisasi"
      :show-cancel-button="false"
      confirm-button-text="Tutup"
      class="sync-details-dialog"
    >
      <div class="sync-details-content">
        <!-- Network Status -->
        <div class="status-item">
          <div class="status-label">
            <van-icon :name="isOnline ? 'success' : 'warning-o'" />
            <span>Status Jaringan</span>
          </div>
          <div :class="['status-value', isOnline ? 'online' : 'offline']">
            {{ isOnline ? 'Online' : 'Offline' }}
          </div>
        </div>

        <!-- Pending Sync Count -->
        <div class="status-item">
          <div class="status-label">
            <van-icon name="clock-o" />
            <span>Menunggu Sinkronisasi</span>
          </div>
          <div class="status-value">
            {{ pendingCount }} item
          </div>
        </div>

        <!-- Last Sync Time -->
        <div class="status-item" v-if="lastSyncTime">
          <div class="status-label">
            <van-icon name="completed" />
            <span>Sinkronisasi Terakhir</span>
          </div>
          <div class="status-value">
            {{ formatLastSyncTime(lastSyncTime) }}
          </div>
        </div>

        <!-- Sync Progress -->
        <div v-if="syncProgress.status === 'syncing'" class="sync-progress-section">
          <div class="progress-label">
            <span>Menyinkronkan data...</span>
            <span>{{ syncProgress.completed }}/{{ syncProgress.total }}</span>
          </div>
          <van-progress 
            :percentage="syncProgressPercentage" 
            color="#1989fa"
            stroke-width="6"
          />
        </div>

        <!-- Sync Actions -->
        <div class="sync-actions">
          <van-button 
            v-if="isOnline && pendingCount > 0 && !isSyncing"
            type="primary" 
            size="small"
            @click="forceSyncNow"
            :loading="isSyncing"
            block
          >
            <van-icon name="refresh" />
            Sinkronkan Sekarang
          </van-button>
          
          <van-button 
            v-if="syncProgress.failed > 0"
            type="warning" 
            size="small"
            @click="retryFailedSync"
            block
            class="retry-button"
          >
            <van-icon name="replay" />
            Coba Lagi ({{ syncProgress.failed }} gagal)
          </van-button>
        </div>
      </div>
    </van-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useDeliveryTasksStore } from '@/stores/deliveryTasks'
import syncService from '@/services/syncService'
import { showToast, showSuccessToast } from 'vant'

const deliveryTasksStore = useDeliveryTasksStore()

// Reactive data
const isOnline = ref(navigator.onLine)
const pendingCount = ref(0)
const lastSyncTime = ref(null)
const showSyncDetails = ref(false)
const showSyncProgress = ref(false)
const syncProgress = ref({
  total: 0,
  completed: 0,
  failed: 0,
  status: 'idle'
})

// Computed properties
const syncIcon = computed(() => {
  if (syncProgress.value.status === 'syncing') return 'loading'
  if (pendingCount.value > 0) return 'clock-o'
  if (!isOnline.value) return 'warning-o'
  return 'success'
})

const syncIconClass = computed(() => {
  const baseClass = 'sync-icon'
  if (syncProgress.value.status === 'syncing') return `${baseClass} syncing`
  if (pendingCount.value > 0) return `${baseClass} pending`
  if (!isOnline.value) return `${baseClass} offline`
  return `${baseClass} synced`
})

const syncProgressMessage = computed(() => {
  if (syncProgress.value.status === 'syncing') {
    return `Menyinkronkan ${syncProgress.value.completed}/${syncProgress.value.total}`
  }
  return 'Menyinkronkan data...'
})

const syncProgressPercentage = computed(() => {
  if (syncProgress.value.total === 0) return 0
  return Math.round((syncProgress.value.completed / syncProgress.value.total) * 100)
})

const isSyncing = computed(() => {
  return syncService.isSyncInProgress()
})

// Methods
const updatePendingCount = async () => {
  try {
    pendingCount.value = await deliveryTasksStore.getPendingSyncCount()
  } catch (error) {
    console.error('Error getting pending sync count:', error)
  }
}

const updateSyncProgress = (progress) => {
  syncProgress.value = { ...progress }
  
  // Show/hide progress toast
  if (progress.status === 'syncing') {
    showSyncProgress.value = true
  } else {
    showSyncProgress.value = false
    
    // Update pending count after sync
    updatePendingCount()
    
    // Show completion message
    if (progress.status === 'completed') {
      showSuccessToast('Sinkronisasi selesai')
    } else if (progress.status === 'completed_with_errors') {
      showToast(`Sinkronisasi selesai dengan ${progress.failed} error`)
    }
  }
}

const forceSyncNow = async () => {
  try {
    await deliveryTasksStore.syncAllOfflineData()
  } catch (error) {
    console.error('Error forcing sync:', error)
    showToast('Gagal memulai sinkronisasi')
  }
}

const retryFailedSync = async () => {
  try {
    await syncService.retryFailedSyncItems()
    showToast('Mencoba ulang sinkronisasi...')
  } catch (error) {
    console.error('Error retrying failed sync:', error)
    showToast('Gagal mencoba ulang sinkronisasi')
  }
}

const formatLastSyncTime = (timestamp) => {
  if (!timestamp) return 'Belum pernah'
  
  const date = new Date(timestamp)
  const now = new Date()
  const diffMs = now - date
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMins / 60)
  const diffDays = Math.floor(diffHours / 24)
  
  if (diffMins < 1) return 'Baru saja'
  if (diffMins < 60) return `${diffMins} menit lalu`
  if (diffHours < 24) return `${diffHours} jam lalu`
  if (diffDays < 7) return `${diffDays} hari lalu`
  
  return date.toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// Network status handlers
const handleOnline = () => {
  isOnline.value = true
}

const handleOffline = () => {
  isOnline.value = false
}

// Lifecycle
onMounted(() => {
  // Initial data load
  updatePendingCount()
  
  // Set up listeners
  deliveryTasksStore.addSyncProgressListener(updateSyncProgress)
  window.addEventListener('online', handleOnline)
  window.addEventListener('offline', handleOffline)
  
  // Update pending count periodically
  const interval = setInterval(updatePendingCount, 30000) // Every 30 seconds
  
  // Store interval for cleanup
  onUnmounted(() => {
    clearInterval(interval)
  })
})

onUnmounted(() => {
  deliveryTasksStore.removeSyncProgressListener(updateSyncProgress)
  window.removeEventListener('online', handleOnline)
  window.removeEventListener('offline', handleOffline)
})
</script>

<style scoped>
.sync-status-container {
  position: relative;
}

.sync-badge {
  cursor: pointer;
}

.sync-icon {
  font-size: 20px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.sync-icon.synced {
  color: #07c160;
}

.sync-icon.pending {
  color: #ff976a;
}

.sync-icon.offline {
  color: #ee0a24;
}

.sync-icon.syncing {
  color: #1989fa;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.sync-details-dialog {
  width: 85%;
  max-width: 400px;
}

.sync-details-content {
  padding: 16px;
}

.status-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #ebedf0;
}

.status-item:last-child {
  border-bottom: none;
}

.status-label {
  display: flex;
  align-items: center;
  color: #646566;
  font-size: 14px;
}

.status-label .van-icon {
  margin-right: 8px;
  font-size: 16px;
}

.status-value {
  font-weight: 500;
  font-size: 14px;
}

.status-value.online {
  color: #07c160;
}

.status-value.offline {
  color: #ee0a24;
}

.sync-progress-section {
  margin: 16px 0;
  padding: 16px;
  background-color: #f7f8fa;
  border-radius: 8px;
}

.progress-label {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-size: 14px;
  color: #646566;
}

.sync-actions {
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.retry-button {
  margin-top: 8px;
}

/* Mobile responsive */
@media (max-width: 375px) {
  .sync-details-dialog {
    width: 95%;
  }
  
  .sync-details-content {
    padding: 12px;
  }
  
  .status-item {
    padding: 10px 0;
  }
  
  .status-label,
  .status-value {
    font-size: 13px;
  }
}
</style>