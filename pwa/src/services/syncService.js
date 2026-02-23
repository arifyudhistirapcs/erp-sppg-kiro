import api from './api'
import db, { storageService } from './db'

// Enhanced sync service for handling offline data synchronization
class SyncService {
  constructor() {
    this.isOnline = navigator.onLine
    this.isSyncing = false
    this.syncProgress = {
      total: 0,
      completed: 0,
      failed: 0,
      status: 'idle', // idle, syncing, completed, error, completed_with_errors
      currentItem: null,
      startTime: null,
      endTime: null
    }
    this.listeners = []
    this.conflictResolutionStrategy = 'server_wins' // server_wins, client_wins, merge
    
    // Listen for network status changes
    window.addEventListener('online', this.handleOnline.bind(this))
    window.addEventListener('offline', this.handleOffline.bind(this))
    
    // Initialize sync metadata
    this.initializeSyncMeta()
  }

  // Initialize sync metadata
  async initializeSyncMeta() {
    try {
      const lastSyncTime = await storageService.getSyncMeta('lastSyncTime')
      if (!lastSyncTime) {
        await storageService.setSyncMeta('lastSyncTime', new Date().toISOString())
      }
      
      const syncSettings = await storageService.getSyncMeta('syncSettings')
      if (!syncSettings) {
        await storageService.setSyncMeta('syncSettings', {
          autoSync: true,
          syncInterval: 300000, // 5 minutes
          maxRetries: 3,
          batchSize: 10
        })
      }
    } catch (error) {
      console.error('Error initializing sync metadata:', error)
    }
  }

  // === NETWORK STATUS DETECTION ===
  
  /**
   * Detect online/offline status with enhanced checking
   */
  async detectOnlineStatus() {
    // Basic navigator.onLine check
    if (!navigator.onLine) {
      return false
    }

    // Enhanced connectivity check by pinging the API
    try {
      const controller = new AbortController()
      const timeoutId = setTimeout(() => controller.abort(), 5000) // 5 second timeout
      
      const response = await fetch('/api/v1/health', {
        method: 'HEAD',
        signal: controller.signal,
        cache: 'no-cache'
      })
      
      clearTimeout(timeoutId)
      return response.ok
    } catch (error) {
      console.log('Enhanced connectivity check failed:', error.message)
      return false
    }
  }

  // Handle online event
  async handleOnline() {
    console.log('Network connection detected, verifying...')
    
    // Verify actual connectivity
    const isReallyOnline = await this.detectOnlineStatus()
    
    if (isReallyOnline) {
      this.isOnline = true
      console.log('Network connection verified, starting auto-sync...')
      
      // Get sync settings
      const settings = await storageService.getSyncMeta('syncSettings')
      if (settings?.autoSync !== false) {
        setTimeout(() => this.syncPendingData(), 1000) // Delay to allow UI to update
      }
    } else {
      console.log('Network connection not stable, staying offline')
    }
  }

  // Handle offline event
  handleOffline() {
    this.isOnline = false
    console.log('Network connection lost')
    
    // Update sync progress if currently syncing
    if (this.isSyncing) {
      this.syncProgress.status = 'error'
      this.syncProgress.currentItem = 'Koneksi terputus'
      this.notifyProgress()
    }
  }

  // === PROGRESS MANAGEMENT ===
  
  // Add sync progress listener
  addProgressListener(callback) {
    this.listeners.push(callback)
  }

  // Remove sync progress listener
  removeProgressListener(callback) {
    this.listeners = this.listeners.filter(listener => listener !== callback)
  }

  // Notify all listeners of progress changes
  notifyProgress() {
    this.listeners.forEach(callback => {
      try {
        callback({ ...this.syncProgress })
      } catch (error) {
        console.error('Error in progress listener:', error)
      }
    })
  }

  // === SYNC QUEUE MANAGEMENT ===
  
  /**
   * Queue item for sync with enhanced metadata
   * @param {string} type - Sync item type
   * @param {Object} data - Data to sync
   * @param {number} priority - Priority (1=highest, 5=lowest)
   * @param {Object} options - Additional options
   */
  async queueForSync(type, data, priority = 3, options = {}) {
    try {
      const queueItem = {
        type,
        data,
        status: 'pending',
        retryCount: 0,
        lastAttempt: null,
        createdAt: new Date().toISOString(),
        priority,
        errorMessage: null,
        conflictData: null,
        ...options
      }

      const itemId = await db.syncQueue.add(queueItem)
      console.log(`Queued ${type} for sync with ID ${itemId}`)

      // Try immediate sync if online and not currently syncing
      if (this.isOnline && !this.isSyncing) {
        // Small delay to allow UI updates
        setTimeout(() => this.syncPendingData(), 100)
      }

      return itemId
    } catch (error) {
      console.error('Error queuing item for sync:', error)
      throw error
    }
  }

  // === MAIN SYNC METHODS ===
  
  /**
   * Sync all pending offline data with enhanced progress tracking
   */
  async syncPendingData() {
    if (!this.isOnline) {
      console.log('Cannot sync: offline')
      return { success: false, reason: 'offline' }
    }

    if (this.isSyncing) {
      console.log('Sync already in progress')
      return { success: false, reason: 'already_syncing' }
    }

    this.isSyncing = true
    this.syncProgress = {
      total: 0,
      completed: 0,
      failed: 0,
      status: 'syncing',
      currentItem: 'Mempersiapkan sinkronisasi...',
      startTime: new Date().toISOString(),
      endTime: null
    }
    this.notifyProgress()

    try {
      // Verify connection before starting
      const isOnline = await this.detectOnlineStatus()
      if (!isOnline) {
        throw new Error('Koneksi internet tidak stabil')
      }

      // Get sync settings
      const settings = await storageService.getSyncMeta('syncSettings') || {}
      const batchSize = settings.batchSize || 10

      // Get all pending sync items ordered by priority and creation time
      const pendingItems = await db.syncQueue
        .where('status')
        .anyOf(['pending', 'failed'])
        .and(item => item.retryCount < (settings.maxRetries || 3))
        .orderBy(['priority', 'createdAt'])
        .toArray()

      this.syncProgress.total = pendingItems.length
      this.syncProgress.currentItem = `Ditemukan ${pendingItems.length} item untuk disinkronkan`
      this.notifyProgress()

      if (pendingItems.length === 0) {
        this.syncProgress.status = 'completed'
        this.syncProgress.currentItem = 'Semua data sudah tersinkronisasi'
        this.syncProgress.endTime = new Date().toISOString()
        this.notifyProgress()
        
        // Update last sync time
        await storageService.setSyncMeta('lastSyncTime', new Date().toISOString())
        return { success: true, synced: 0, failed: 0 }
      }

      console.log(`Starting sync of ${pendingItems.length} items`)

      // Process items in batches to avoid overwhelming the server
      const batches = []
      for (let i = 0; i < pendingItems.length; i += batchSize) {
        batches.push(pendingItems.slice(i, i + batchSize))
      }

      let totalSynced = 0
      let totalFailed = 0

      for (const batch of batches) {
        // Check connection before each batch
        if (!await this.detectOnlineStatus()) {
          throw new Error('Koneksi terputus selama sinkronisasi')
        }

        // Process batch items
        for (const item of batch) {
          this.syncProgress.currentItem = `Menyinkronkan ${this.getItemDisplayName(item)}...`
          this.notifyProgress()

          try {
            await this.syncItem(item)
            this.syncProgress.completed++
            totalSynced++
          } catch (error) {
            console.error(`Failed to sync item ${item.id}:`, error)
            this.syncProgress.failed++
            totalFailed++
            
            // Update retry count and status
            const maxRetries = settings.maxRetries || 3
            const newRetryCount = item.retryCount + 1
            
            await db.syncQueue.update(item.id, {
              retryCount: newRetryCount,
              lastAttempt: new Date().toISOString(),
              status: newRetryCount >= maxRetries ? 'failed' : 'pending',
              errorMessage: error.message
            })
          }
          
          this.notifyProgress()
        }

        // Small delay between batches to prevent overwhelming the server
        if (batches.indexOf(batch) < batches.length - 1) {
          await new Promise(resolve => setTimeout(resolve, 500))
        }
      }

      // Update final status
      this.syncProgress.status = totalFailed > 0 ? 'completed_with_errors' : 'completed'
      this.syncProgress.currentItem = `Selesai: ${totalSynced} berhasil, ${totalFailed} gagal`
      this.syncProgress.endTime = new Date().toISOString()
      
      // Update last sync time
      await storageService.setSyncMeta('lastSyncTime', new Date().toISOString())
      
      console.log(`Sync completed: ${totalSynced} successful, ${totalFailed} failed`)
      
      return { success: true, synced: totalSynced, failed: totalFailed }

    } catch (error) {
      console.error('Error during sync:', error)
      this.syncProgress.status = 'error'
      this.syncProgress.currentItem = `Error: ${error.message}`
      this.syncProgress.endTime = new Date().toISOString()
      
      return { success: false, error: error.message }
    } finally {
      this.isSyncing = false
      this.notifyProgress()
    }
  }

  /**
   * Get display name for sync item
   * @param {Object} item - Sync queue item
   */
  getItemDisplayName(item) {
    switch (item.type) {
      case 'epod':
        return `e-POD tugas ${item.data.delivery_task_id}`
      case 'epod_photo':
        return `Foto e-POD`
      case 'epod_signature':
        return `Tanda tangan e-POD`
      case 'delivery_status':
        return `Status pengiriman`
      case 'attendance':
        return `Data absensi`
      default:
        return item.type
    }
  }

  /**
   * Sync individual item with enhanced error handling
   * @param {Object} item - Sync queue item
   */
  async syncItem(item) {
    const startTime = Date.now()
    
    try {
      let result
      
      switch (item.type) {
        case 'epod':
          result = await this.syncEPOD(item)
          break
        case 'epod_photo':
          result = await this.syncEPODPhoto(item)
          break
        case 'epod_signature':
          result = await this.syncEPODSignature(item)
          break
        case 'delivery_status':
          result = await this.syncDeliveryStatus(item)
          break
        case 'attendance':
          result = await this.syncAttendance(item)
          break
        default:
          throw new Error(`Unknown sync type: ${item.type}`)
      }

      // Mark as completed and remove from queue
      await db.syncQueue.delete(item.id)

      // Log successful sync
      await this.logSyncActivity(item.type, 'sync_success', 'success', {
        itemId: item.id,
        duration: Date.now() - startTime,
        result: result
      })

      return result
    } catch (error) {
      // Log failed sync
      await this.logSyncActivity(item.type, 'sync_failed', 'error', {
        itemId: item.id,
        duration: Date.now() - startTime,
        error: error.message,
        retryCount: item.retryCount
      })

      throw error
    }
  }

  // === CONFLICT RESOLUTION ===
  
  /**
   * Handle sync conflicts with configurable resolution strategy
   * @param {Object} localData - Local data
   * @param {Object} serverData - Server data
   * @param {string} strategy - Resolution strategy
   */
  async handleSyncConflict(localData, serverData, strategy = null) {
    const resolveStrategy = strategy || this.conflictResolutionStrategy
    
    switch (resolveStrategy) {
      case 'server_wins':
        console.log('Conflict resolved: server data wins')
        return serverData
        
      case 'client_wins':
        console.log('Conflict resolved: client data wins')
        return localData
        
      case 'merge':
        console.log('Conflict resolved: merging data')
        return {
          ...serverData,
          ...localData,
          // Prefer server timestamps for consistency
          updated_at: serverData.updated_at,
          // Merge arrays if present
          ...(Array.isArray(localData.items) && Array.isArray(serverData.items) ? {
            items: [...new Set([...serverData.items, ...localData.items])]
          } : {})
        }
        
      default:
        console.log('Unknown conflict resolution strategy, defaulting to server wins')
        return serverData
    }
  }

  // === SPECIFIC SYNC METHODS ===
  
  /**
   * Sync e-POD data with enhanced conflict handling
   */
  async syncEPOD(item) {
    try {
      // Prepare e-POD data for submission
      const epodData = {
        delivery_task_id: item.data.delivery_task_id,
        latitude: item.data.latitude,
        longitude: item.data.longitude,
        accuracy: item.data.accuracy,
        recipient_name: item.data.recipient_name,
        ompreng_drop_off: item.data.ompreng_drop_off,
        ompreng_pick_up: item.data.ompreng_pick_up,
        completed_at: item.data.completed_at,
        device_info: item.data.device_info || {
          userAgent: navigator.userAgent,
          timestamp: new Date().toISOString()
        }
      }

      const response = await api.post('/epod', epodData)
      
      if (response.data.success) {
        const epodId = response.data.epod?.id
        
        // Update local e-POD record with server ID
        await storageService.updateePODSyncStatus(
          item.data.local_epod_id,
          'synced',
          { serverId: epodId }
        )

        // Update any pending photo/signature sync items with the e-POD ID
        if (epodId) {
          await db.syncQueue
            .where('type')
            .anyOf(['epod_photo', 'epod_signature'])
            .and(syncItem => syncItem.data.taskId === item.data.delivery_task_id)
            .modify(syncItem => {
              syncItem.data.epodId = epodId
              return syncItem
            })
        }
        
        console.log(`e-POD synced successfully for task ${item.data.delivery_task_id}, e-POD ID: ${epodId}`)
        return response.data
      } else {
        throw new Error(response.data.message || 'Failed to sync e-POD')
      }
    } catch (error) {
      // Enhanced error handling for e-POD sync
      if (error.response?.status === 400) {
        throw new Error('Data e-POD tidak valid: ' + (error.response.data?.message || 'Periksa kembali data'))
      } else if (error.response?.status === 404) {
        throw new Error('Tugas pengiriman tidak ditemukan di server')
      } else if (error.response?.status === 409) {
        // e-POD already exists - handle conflict
        const serverData = error.response.data?.existing_epod
        if (serverData) {
          const resolvedData = await this.handleSyncConflict(item.data, serverData)
          console.log(`e-POD conflict resolved for task ${item.data.delivery_task_id}`)
          
          // Update local record with resolved data
          await storageService.updateePODSyncStatus(
            item.data.local_epod_id,
            'synced',
            { serverId: serverData.id, conflictResolved: true }
          )
          
          return { success: true, message: 'e-POD conflict resolved', data: resolvedData }
        } else {
          throw new Error('e-POD sudah ada di server')
        }
      } else {
        throw new Error(error.message || 'Gagal menyinkronkan e-POD')
      }
    }
  }

  /**
   * Sync e-POD photo with retry logic
   */
  async syncEPODPhoto(item) {
    try {
      const { epodId, taskId, photoData } = item.data
      
      if (!epodId) {
        throw new Error('e-POD ID tidak tersedia untuk upload foto')
      }
      
      // Convert base64 to blob for upload
      const response = await fetch(photoData)
      const blob = await response.blob()
      
      // Check file size (max 5MB)
      if (blob.size > 5 * 1024 * 1024) {
        throw new Error('Ukuran foto terlalu besar. Maksimal 5MB.')
      }
      
      const formData = new FormData()
      formData.append('photo', blob, `epod-photo-${taskId}-${Date.now()}.jpg`)
      
      const uploadResponse = await api.post(`/epod/${epodId}/upload-photo`, formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        },
        timeout: 30000 // 30 second timeout for file uploads
      })
      
      if (uploadResponse.data.success) {
        // Update local photo record if exists
        await db.photos.where('taskId').equals(taskId).modify({
          synced: true,
          syncedAt: new Date().toISOString(),
          photoUrl: uploadResponse.data.photo_url
        })
        
        console.log(`e-POD photo synced successfully for e-POD ${epodId}`)
        return uploadResponse.data
      } else {
        throw new Error(uploadResponse.data.message || 'Failed to sync e-POD photo')
      }
    } catch (error) {
      if (error.response?.status === 413) {
        throw new Error('Ukuran foto terlalu besar. Maksimal 5MB.')
      } else if (error.response?.status === 415) {
        throw new Error('Format foto tidak didukung. Gunakan JPG atau PNG.')
      } else if (error.code === 'ECONNABORTED') {
        throw new Error('Upload foto timeout. Periksa koneksi internet.')
      } else {
        throw new Error(error.message || 'Gagal menyinkronkan foto e-POD')
      }
    }
  }

  /**
   * Sync e-POD signature with retry logic
   */
  async syncEPODSignature(item) {
    try {
      const { epodId, taskId, signatureData } = item.data
      
      if (!epodId) {
        throw new Error('e-POD ID tidak tersedia untuk upload tanda tangan')
      }
      
      // Convert base64 to blob for upload
      const response = await fetch(signatureData)
      const blob = await response.blob()
      
      const formData = new FormData()
      formData.append('signature', blob, `epod-signature-${taskId}-${Date.now()}.png`)
      
      const uploadResponse = await api.post(`/epod/${epodId}/upload-signature`, formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        },
        timeout: 30000 // 30 second timeout for file uploads
      })
      
      if (uploadResponse.data.success) {
        // Update local signature record if exists
        await db.signatures.where('taskId').equals(taskId).modify({
          synced: true,
          syncedAt: new Date().toISOString(),
          signatureUrl: uploadResponse.data.signature_url
        })
        
        console.log(`e-POD signature synced successfully for e-POD ${epodId}`)
        return uploadResponse.data
      } else {
        throw new Error(uploadResponse.data.message || 'Failed to sync e-POD signature')
      }
    } catch (error) {
      if (error.response?.status === 413) {
        throw new Error('Ukuran tanda tangan terlalu besar.')
      } else if (error.response?.status === 415) {
        throw new Error('Format tanda tangan tidak didukung.')
      } else if (error.code === 'ECONNABORTED') {
        throw new Error('Upload tanda tangan timeout. Periksa koneksi internet.')
      } else {
        throw new Error(error.message || 'Gagal menyinkronkan tanda tangan e-POD')
      }
    }
  }

  /**
   * Sync delivery status update
   */
  async syncDeliveryStatus(item) {
    const { taskId, status, timestamp } = item.data
    
    const response = await api.put(`/delivery-tasks/${taskId}/status`, { 
      status,
      updated_at: timestamp || new Date().toISOString()
    })
    
    if (response.data.success) {
      console.log(`Delivery status synced successfully for task ${taskId}`)
      return response.data
    } else {
      throw new Error(response.data.message || 'Failed to sync delivery status')
    }
  }

  /**
   * Sync attendance data
   */
  async syncAttendance(item) {
    const response = await api.post('/attendance', item.data)
    
    if (response.data.success) {
      // Update local attendance record
      await db.attendance.where('id').equals(item.data.local_id).modify({
        syncStatus: 'synced',
        serverId: response.data.attendance?.id
      })
      
      console.log(`Attendance synced successfully for employee ${item.data.employee_id}`)
      return response.data
    } else {
      throw new Error(response.data.message || 'Failed to sync attendance')
    }
  }

  // === UTILITY METHODS ===
  
  /**
   * Log sync activity with enhanced details
   */
  async logSyncActivity(type, action, status, details) {
    try {
      await db.syncLog.add({
        type,
        action,
        status,
        timestamp: new Date().toISOString(),
        details: JSON.stringify(details),
        duration: details.duration || 0,
        dataSize: details.dataSize || 0
      })
    } catch (error) {
      console.error('Error logging sync activity:', error)
      // Don't throw - logging failure shouldn't break sync
    }
  }

  /**
   * Get sync status for a specific item type
   */
  async getSyncStatus(type, identifier) {
    const items = await db.syncQueue
      .where('type')
      .equals(type)
      .and(item => {
        if (type === 'epod') {
          return item.data.delivery_task_id === identifier
        }
        return item.data.taskId === identifier || item.data.epodId === identifier
      })
      .toArray()

    if (items.length === 0) {
      return 'synced'
    }

    const hasFailedItems = items.some(item => item.status === 'failed')
    const hasPendingItems = items.some(item => item.status === 'pending')

    if (hasFailedItems) return 'failed'
    if (hasPendingItems) return 'pending'
    return 'syncing'
  }

  /**
   * Get pending sync count
   */
  async getPendingSyncCount() {
    return await db.syncQueue
      .where('status')
      .anyOf(['pending', 'failed'])
      .count()
  }

  /**
   * Get sync progress
   */
  getSyncProgress() {
    return { ...this.syncProgress }
  }

  /**
   * Check if currently syncing
   */
  isSyncInProgress() {
    return this.isSyncing
  }

  /**
   * Get sync statistics with enhanced metrics
   */
  async getSyncStatistics() {
    try {
      const now = new Date()
      const last24Hours = new Date(now.getTime() - 24 * 60 * 60 * 1000)
      
      const recentLogs = await db.syncLog
        .where('timestamp')
        .above(last24Hours.toISOString())
        .toArray()

      const stats = {
        total: recentLogs.length,
        successful: recentLogs.filter(log => log.status === 'success').length,
        failed: recentLogs.filter(log => log.status === 'error').length,
        averageDuration: 0,
        totalDataSize: 0,
        byType: {}
      }

      // Calculate averages and group by type
      let totalDuration = 0
      recentLogs.forEach(log => {
        const details = JSON.parse(log.details || '{}')
        totalDuration += details.duration || 0
        stats.totalDataSize += details.dataSize || 0
        
        if (!stats.byType[log.type]) {
          stats.byType[log.type] = { total: 0, successful: 0, failed: 0 }
        }
        stats.byType[log.type].total++
        if (log.status === 'success') {
          stats.byType[log.type].successful++
        } else if (log.status === 'error') {
          stats.byType[log.type].failed++
        }
      })

      stats.averageDuration = stats.total > 0 ? Math.round(totalDuration / stats.total) : 0

      return stats
    } catch (error) {
      console.error('Error getting sync statistics:', error)
      return { total: 0, successful: 0, failed: 0, averageDuration: 0, totalDataSize: 0, byType: {} }
    }
  }

  /**
   * Clear failed sync items
   */
  async clearFailedSyncItems() {
    const deletedCount = await db.syncQueue.where('status').equals('failed').delete()
    console.log(`Cleared ${deletedCount} failed sync items`)
    return deletedCount
  }

  /**
   * Retry failed sync items
   */
  async retryFailedSyncItems() {
    const updatedCount = await db.syncQueue.where('status').equals('failed').modify({
      status: 'pending',
      retryCount: 0,
      lastAttempt: null,
      errorMessage: null
    })

    console.log(`Reset ${updatedCount} failed items for retry`)

    if (this.isOnline && updatedCount > 0) {
      setTimeout(() => this.syncPendingData(), 1000)
    }

    return updatedCount
  }

  /**
   * Clean up old sync logs (keep last 7 days)
   */
  async cleanupOldLogs() {
    try {
      const cutoffDate = new Date()
      cutoffDate.setDate(cutoffDate.getDate() - 7)
      
      const deletedCount = await db.syncLog
        .where('timestamp')
        .below(cutoffDate.toISOString())
        .delete()

      console.log(`Cleaned up ${deletedCount} old sync log entries`)
      return deletedCount
    } catch (error) {
      console.error('Error cleaning up old logs:', error)
      return 0
    }
  }

  /**
   * Get recent sync errors for debugging
   */
  async getRecentSyncErrors(limit = 10) {
    try {
      return await db.syncLog
        .where('status')
        .equals('error')
        .reverse()
        .limit(limit)
        .toArray()
    } catch (error) {
      console.error('Error getting recent sync errors:', error)
      return []
    }
  }

  /**
   * Update sync settings
   */
  async updateSyncSettings(settings) {
    try {
      const currentSettings = await storageService.getSyncMeta('syncSettings') || {}
      const newSettings = { ...currentSettings, ...settings }
      await storageService.setSyncMeta('syncSettings', newSettings)
      
      // Update conflict resolution strategy if provided
      if (settings.conflictResolution) {
        this.conflictResolutionStrategy = settings.conflictResolution
      }
      
      console.log('Sync settings updated:', newSettings)
      return newSettings
    } catch (error) {
      console.error('Error updating sync settings:', error)
      throw error
    }
  }

  /**
   * Get current sync settings
   */
  async getSyncSettings() {
    try {
      return await storageService.getSyncMeta('syncSettings') || {
        autoSync: true,
        syncInterval: 300000,
        maxRetries: 3,
        batchSize: 10,
        conflictResolution: 'server_wins'
      }
    } catch (error) {
      console.error('Error getting sync settings:', error)
      return {}
    }
  }
}

// Create singleton instance
const syncService = new SyncService()

export default syncService