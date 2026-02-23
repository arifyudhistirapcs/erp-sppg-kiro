import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '@/services/api'
import db from '@/services/db'
import syncService from '@/services/syncService'

export const useDeliveryTasksStore = defineStore('deliveryTasks', () => {
  const tasks = ref([])
  const isLoading = ref(false)
  const lastSync = ref(null)

  // Fetch today's delivery tasks for a driver
  const fetchTodayTasks = async (driverId, forceRefresh = false) => {
    isLoading.value = true
    
    try {
      // Try to fetch from API first
      if (navigator.onLine || forceRefresh) {
        const response = await api.get(`/delivery-tasks/driver/${driverId}/today`)
        
        if (response.data.success) {
          const apiTasks = response.data.delivery_tasks || []
          tasks.value = apiTasks
          lastSync.value = new Date().toISOString()
          
          // Cache tasks in IndexedDB for offline access
          await cacheTasksOffline(apiTasks)
          
          return apiTasks
        }
      }
    } catch (error) {
      console.error('Error fetching tasks from API:', error)
      
      // If API fails, try to load from cache
      const cachedTasks = await loadTasksFromCache(driverId)
      if (cachedTasks.length > 0) {
        tasks.value = cachedTasks
        return cachedTasks
      }
      
      throw error
    } finally {
      isLoading.value = false
    }
  }

  // Cache tasks in IndexedDB
  const cacheTasksOffline = async (tasksData) => {
    try {
      // Clear existing tasks for today
      const today = new Date()
      const todayStr = today.toISOString().split('T')[0]
      
      await db.deliveryTasks.where('taskDate').startsWith(todayStr).delete()
      
      // Store new tasks
      const tasksToCache = tasksData.map(task => ({
        id: task.id,
        taskDate: task.task_date,
        driverId: task.driver_id,
        schoolId: task.school_id,
        portions: task.portions,
        status: task.status,
        routeOrder: task.route_order,
        school: task.school,
        menuItems: task.menu_items,
        cachedAt: new Date().toISOString()
      }))
      
      await db.deliveryTasks.bulkAdd(tasksToCache)
      console.log('Tasks cached successfully:', tasksToCache.length)
    } catch (error) {
      console.error('Error caching tasks:', error)
    }
  }

  // Load tasks from IndexedDB cache
  const loadTasksFromCache = async (driverId) => {
    try {
      const today = new Date()
      const todayStr = today.toISOString().split('T')[0]
      
      const cachedTasks = await db.deliveryTasks
        .where('taskDate')
        .startsWith(todayStr)
        .and(task => task.driverId === driverId)
        .toArray()
      
      // Convert back to API format
      const formattedTasks = cachedTasks.map(task => ({
        id: task.id,
        task_date: task.taskDate,
        driver_id: task.driverId,
        school_id: task.schoolId,
        portions: task.portions,
        status: task.status,
        route_order: task.routeOrder,
        school: task.school,
        menu_items: task.menuItems
      }))
      
      console.log('Loaded tasks from cache:', formattedTasks.length)
      return formattedTasks
    } catch (error) {
      console.error('Error loading tasks from cache:', error)
      return []
    }
  }

  // Update task status
  const updateTaskStatus = async (taskId, newStatus) => {
    try {
      // Update local state immediately for better UX
      const taskIndex = tasks.value.findIndex(task => task.id === taskId)
      if (taskIndex !== -1) {
        tasks.value[taskIndex].status = newStatus
      }

      if (navigator.onLine) {
        // Try to update on server
        const response = await api.put(`/delivery-tasks/${taskId}/status`, {
          status: newStatus
        })
        
        if (response.data.success) {
          // Update cache
          await updateTaskInCache(taskId, { status: newStatus })
        }
      } else {
        // Store update for later sync
        await storeOfflineUpdate(taskId, { status: newStatus })
      }
    } catch (error) {
      console.error('Error updating task status:', error)
      
      // Revert local state if API call failed
      const taskIndex = tasks.value.findIndex(task => task.id === taskId)
      if (taskIndex !== -1) {
        // You might want to revert to previous status here
        // For now, we'll keep the optimistic update
      }
      
      // Store for offline sync
      await storeOfflineUpdate(taskId, { status: newStatus })
      throw error
    }
  }

  // Update task in cache
  const updateTaskInCache = async (taskId, updates) => {
    try {
      await db.deliveryTasks.where('id').equals(taskId).modify(updates)
    } catch (error) {
      console.error('Error updating task in cache:', error)
    }
  }

  // Store offline updates for later sync
  const storeOfflineUpdate = async (taskId, updates) => {
    try {
      const offlineUpdate = {
        taskId,
        updates,
        timestamp: new Date().toISOString(),
        type: 'status_update'
      }
      
      // Check if offlineUpdates table exists, if not create it
      if (!db.offlineUpdates) {
        console.warn('offlineUpdates table not found, skipping offline storage')
        return
      }
      
      // Store in a separate table for offline updates
      await db.offlineUpdates.add(offlineUpdate)
    } catch (error) {
      console.error('Error storing offline update:', error)
    }
  }

  // Sync offline data when back online
  const syncOfflineData = async () => {
    if (!navigator.onLine) return

    try {
      // Check if offlineUpdates table exists
      if (!db.offlineUpdates) {
        console.warn('offlineUpdates table not found, skipping sync')
        return
      }
      
      // Get all pending offline updates
      const offlineUpdates = await db.offlineUpdates.toArray()
      
      for (const update of offlineUpdates) {
        try {
          if (update.type === 'status_update') {
            await api.put(`/delivery-tasks/${update.taskId}/status`, update.updates)
          }
          
          // Remove synced update
          await db.offlineUpdates.delete(update.id)
        } catch (error) {
          console.error('Error syncing update:', update, error)
          // Keep the update for next sync attempt
        }
      }
      
      console.log('Offline data sync completed')
    } catch (error) {
      console.error('Error syncing offline data:', error)
    }
  }

  // Get task by ID
  const getTaskById = (taskId) => {
    return tasks.value.find(task => task.id === taskId)
  }

  // Get tasks by status
  const getTasksByStatus = (status) => {
    return tasks.value.filter(task => task.status === status)
  }

  // Submit e-POD with enhanced sync functionality
  const submitePOD = async (ePODData) => {
    try {
      // Validate required data
      if (!ePODData.delivery_task_id || !ePODData.latitude || !ePODData.longitude) {
        throw new Error('Data e-POD tidak lengkap')
      }

      // Additional validation
      if (!ePODData.photo_url || !ePODData.signature_url) {
        throw new Error('Foto dan tanda tangan wajib dilengkapi')
      }

      if (!ePODData.recipient_name?.trim()) {
        throw new Error('Nama penerima wajib diisi')
      }

      // Store e-POD locally first for immediate UI update
      const localePOD = {
        deliveryTaskId: ePODData.delivery_task_id,
        latitude: ePODData.latitude,
        longitude: ePODData.longitude,
        accuracy: ePODData.accuracy || null,
        recipientName: ePODData.recipient_name,
        omprengDropOff: ePODData.ompreng_drop_off,
        omprengPickUp: ePODData.ompreng_pick_up,
        photoUrl: ePODData.photo_url,
        signatureUrl: ePODData.signature_url,
        completedAt: ePODData.completed_at,
        deviceInfo: ePODData.device_info || {},
        syncStatus: 'pending',
        retryCount: 0,
        createdAt: new Date().toISOString()
      }

      // Store in IndexedDB
      await db.epods.add(localePOD)

      // Update task status locally for immediate feedback
      await updateTaskStatus(ePODData.delivery_task_id, 'completed')

      if (navigator.onLine) {
        try {
          // Try immediate submission to backend
          const response = await api.post('/epod', {
            delivery_task_id: ePODData.delivery_task_id,
            latitude: ePODData.latitude,
            longitude: ePODData.longitude,
            accuracy: ePODData.accuracy,
            recipient_name: ePODData.recipient_name,
            ompreng_drop_off: ePODData.ompreng_drop_off,
            ompreng_pick_up: ePODData.ompreng_pick_up,
            completed_at: ePODData.completed_at,
            device_info: ePODData.device_info
          })
          
          if (response.data.success) {
            const epodId = response.data.epod?.id
            
            // Mark main e-POD as synced
            await db.epods.where('deliveryTaskId').equals(ePODData.delivery_task_id).modify({
              syncStatus: 'synced',
              lastSync: new Date().toISOString(),
              epodId: epodId
            })

            // Queue photo and signature uploads if they are base64 data
            if (ePODData.photo_url && ePODData.photo_url.startsWith('data:')) {
              await syncService.queueForSync('epod_photo', {
                epodId: epodId,
                taskId: ePODData.delivery_task_id,
                photoData: ePODData.photo_url
              }, 2)
            }

            if (ePODData.signature_url && ePODData.signature_url.startsWith('data:')) {
              await syncService.queueForSync('epod_signature', {
                epodId: epodId,
                taskId: ePODData.delivery_task_id,
                signatureData: ePODData.signature_url
              }, 2)
            }
            
            console.log('e-POD submitted successfully with ID:', epodId)
            return { success: true, synced: true, epodId }
          }
        } catch (error) {
          console.error('Failed to submit e-POD immediately:', error)
          // Fall through to queue for sync
        }
      }

      // Queue for sync (either offline or immediate submission failed)
      await syncService.queueForSync('epod', ePODData, 1) // High priority

      // Queue photo and signature separately for better handling
      if (ePODData.photo_url && ePODData.photo_url.startsWith('data:')) {
        await syncService.queueForSync('epod_photo', {
          epodId: null, // Will be updated after main e-POD is synced
          taskId: ePODData.delivery_task_id,
          photoData: ePODData.photo_url
        }, 2)
      }

      if (ePODData.signature_url && ePODData.signature_url.startsWith('data:')) {
        await syncService.queueForSync('epod_signature', {
          epodId: null, // Will be updated after main e-POD is synced
          taskId: ePODData.delivery_task_id,
          signatureData: ePODData.signature_url
        }, 2)
      }

      return { 
        success: true, 
        offline: !navigator.onLine,
        queued: true,
        message: navigator.onLine ? 'e-POD dalam antrian sinkronisasi' : 'e-POD disimpan offline dan akan disinkronkan saat online'
      }

    } catch (error) {
      console.error('Error submitting e-POD:', error)
      
      // Enhanced error handling
      if (error.message?.includes('tidak lengkap')) {
        throw new Error('Data e-POD tidak lengkap. Periksa GPS, foto, dan tanda tangan.')
      } else if (error.message?.includes('wajib')) {
        throw error // Pass validation errors as-is
      } else if (error.response?.status === 400) {
        throw new Error('Data e-POD tidak valid. Periksa kembali semua informasi.')
      } else if (error.response?.status === 404) {
        throw new Error('Tugas pengiriman tidak ditemukan.')
      } else if (error.response?.status >= 500) {
        throw new Error('Server error. e-POD akan disimpan offline.')
      } else {
        throw new Error(error.message || 'Gagal menyimpan e-POD. Akan dicoba lagi saat online.')
      }
    }
  }

  // Store e-POD for offline sync
  const storeePODOffline = async (ePODData) => {
    try {
      const offlineePOD = {
        ...ePODData,
        syncStatus: 'pending',
        createdAt: new Date().toISOString()
      }
      
      await db.epods.add(offlineePOD)
      console.log('e-POD stored for offline sync')
    } catch (error) {
      console.error('Error storing e-POD offline:', error)
    }
  }

  // Sync offline e-PODs
  const syncOfflineePODs = async () => {
    if (!navigator.onLine) return

    try {
      const pendingePODs = await db.epods.where('syncStatus').equals('pending').toArray()
      
      for (const epod of pendingePODs) {
        try {
          const response = await api.post('/epod', epod)
          
          if (response.data.success) {
            // Mark as synced
            await db.epods.update(epod.id, { syncStatus: 'synced' })
          }
        } catch (error) {
          console.error('Error syncing e-POD:', epod, error)
          // Keep as pending for next sync attempt
        }
      }
      
      console.log('e-POD sync completed')
    } catch (error) {
      console.error('Error syncing offline e-PODs:', error)
    }
  }

  // Enhanced sync that uses the new sync service
  const syncAllOfflineData = async () => {
    return await syncService.syncAllOfflineData()
  }

  // Get sync status for a delivery task
  const getTaskSyncStatus = async (taskId) => {
    return await syncService.getSyncStatus('epod', taskId)
  }

  // Get pending sync count
  const getPendingSyncCount = async () => {
    return await syncService.getPendingSyncCount()
  }

  // Get sync progress
  const getSyncProgress = () => {
    return syncService.getSyncProgress()
  }

  // Add sync progress listener
  const addSyncProgressListener = (callback) => {
    syncService.addProgressListener(callback)
  }

  // Remove sync progress listener
  const removeSyncProgressListener = (callback) => {
    syncService.removeProgressListener(callback)
  }

  // Clear all tasks
  const clearTasks = () => {
    tasks.value = []
  }

  return {
    // State
    tasks,
    isLoading,
    lastSync,
    
    // Actions
    fetchTodayTasks,
    updateTaskStatus,
    submitePOD,
    syncOfflineData,
    syncOfflineePODs,
    syncAllOfflineData,
    getTaskById,
    getTasksByStatus,
    clearTasks,
    
    // Enhanced sync functionality
    getTaskSyncStatus,
    getPendingSyncCount,
    getSyncProgress,
    addSyncProgressListener,
    removeSyncProgressListener
  }
})