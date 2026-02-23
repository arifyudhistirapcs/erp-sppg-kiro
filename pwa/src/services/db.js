import Dexie from 'dexie'

// IndexedDB for offline storage
const db = new Dexie('erp_sppg_pwa')

db.version(1).stores({
  deliveryTasks: '++id, taskDate, driverId, schoolId, status, routeOrder, cachedAt',
  epods: '++id, deliveryTaskId, syncStatus',
  attendance: '++id, date, syncStatus',
  offlineUpdates: '++id, taskId, type, timestamp'
})

// Version 2: Add photo storage
db.version(2).stores({
  deliveryTasks: '++id, taskDate, driverId, schoolId, status, routeOrder, cachedAt',
  epods: '++id, deliveryTaskId, syncStatus',
  attendance: '++id, date, syncStatus',
  offlineUpdates: '++id, taskId, type, timestamp',
  photos: '++id, taskId, photoData, timestamp, synced'
})

// Version 3: Add signature storage
db.version(3).stores({
  deliveryTasks: '++id, taskDate, driverId, schoolId, status, routeOrder, cachedAt',
  epods: '++id, deliveryTaskId, syncStatus',
  attendance: '++id, date, syncStatus',
  offlineUpdates: '++id, taskId, type, timestamp',
  photos: '++id, taskId, photoData, timestamp, synced',
  signatures: '++id, taskId, signatureData, quality, timestamp, synced'
})

// Version 4: Enhanced e-POD queue management
db.version(4).stores({
  deliveryTasks: '++id, taskDate, driverId, schoolId, status, routeOrder, cachedAt',
  epods: '++id, deliveryTaskId, syncStatus, retryCount, lastAttempt, createdAt, epodId',
  attendance: '++id, date, syncStatus',
  offlineUpdates: '++id, taskId, type, timestamp',
  photos: '++id, taskId, photoData, timestamp, synced, syncedAt, photoUrl',
  signatures: '++id, taskId, signatureData, quality, timestamp, synced, syncedAt, signatureUrl',
  syncQueue: '++id, type, data, status, retryCount, lastAttempt, createdAt, priority'
})

// Version 5: Enhanced sync tracking and error handling
db.version(5).stores({
  deliveryTasks: '++id, taskDate, driverId, schoolId, status, routeOrder, cachedAt',
  epods: '++id, deliveryTaskId, syncStatus, retryCount, lastAttempt, createdAt, epodId, lastSync',
  attendance: '++id, date, syncStatus',
  offlineUpdates: '++id, taskId, type, timestamp',
  photos: '++id, taskId, photoData, timestamp, synced, syncedAt, photoUrl',
  signatures: '++id, taskId, signatureData, quality, timestamp, synced, syncedAt, signatureUrl',
  syncQueue: '++id, type, data, status, retryCount, lastAttempt, createdAt, priority, errorMessage',
  syncLog: '++id, type, action, status, timestamp, details'
})

// Version 6: Enhanced offline sync schema for PWA requirements
db.version(6).stores({
  // Delivery tasks cache with complete task information
  deliveryTasks: '++id, serverId, taskDate, driverId, schoolId, status, routeOrder, portions, menuItems, cachedAt, lastUpdated',
  
  // Schools master data cache
  schools: '++id, serverId, name, address, latitude, longitude, contactPerson, phoneNumber, studentCount, isActive, cachedAt',
  
  // e-POD data with enhanced tracking
  epods: '++id, deliveryTaskId, serverId, latitude, longitude, accuracy, recipientName, omprengDropOff, omprengPickUp, completedAt, syncStatus, retryCount, lastAttempt, createdAt, lastSync',
  
  // Attendance records
  attendance: '++id, employeeId, date, checkIn, checkOut, workHours, ssid, bssid, syncStatus, createdAt',
  
  // Legacy support
  offlineUpdates: '++id, taskId, type, timestamp',
  
  // Media files storage
  photos: '++id, taskId, epodId, photoData, timestamp, synced, syncedAt, photoUrl, fileSize',
  signatures: '++id, taskId, epodId, signatureData, quality, timestamp, synced, syncedAt, signatureUrl, fileSize',
  
  // Enhanced sync queue with conflict resolution
  syncQueue: '++id, type, data, status, retryCount, lastAttempt, createdAt, priority, errorMessage, conflictData',
  
  // Comprehensive sync logging
  syncLog: '++id, type, action, status, timestamp, details, duration, dataSize',
  
  // Sync metadata and configuration
  syncMeta: '++id, key, value, updatedAt'
})

// Storage service class for managing offline data
class OfflineStorageService {
  constructor() {
    this.db = db
  }

  // === DELIVERY TASKS METHODS ===
  
  /**
   * Save delivery task to cache
   * @param {Object} task - Delivery task data
   */
  async saveTask(task) {
    try {
      const taskData = {
        serverId: task.id,
        taskDate: task.task_date || task.taskDate,
        driverId: task.driver_id || task.driverId,
        schoolId: task.school_id || task.schoolId,
        status: task.status,
        routeOrder: task.route_order || task.routeOrder,
        portions: task.portions,
        menuItems: JSON.stringify(task.menu_items || task.menuItems || []),
        cachedAt: new Date().toISOString(),
        lastUpdated: task.updated_at || new Date().toISOString()
      }

      // Check if task already exists
      const existingTask = await this.db.deliveryTasks
        .where('serverId')
        .equals(task.id)
        .first()

      if (existingTask) {
        await this.db.deliveryTasks.update(existingTask.id, taskData)
        return existingTask.id
      } else {
        return await this.db.deliveryTasks.add(taskData)
      }
    } catch (error) {
      console.error('Error saving delivery task:', error)
      throw error
    }
  }

  /**
   * Get delivery tasks from cache
   * @param {Object} filters - Filter options
   */
  async getTasks(filters = {}) {
    try {
      let query = this.db.deliveryTasks.toCollection()

      // Apply filters
      if (filters.driverId) {
        query = query.filter(task => task.driverId === filters.driverId)
      }
      
      if (filters.taskDate) {
        query = query.filter(task => task.taskDate === filters.taskDate)
      }
      
      if (filters.status) {
        query = query.filter(task => task.status === filters.status)
      }

      const tasks = await query.toArray()
      
      // Parse menu items back to objects
      return tasks.map(task => ({
        ...task,
        menuItems: task.menuItems ? JSON.parse(task.menuItems) : []
      }))
    } catch (error) {
      console.error('Error getting delivery tasks:', error)
      throw error
    }
  }

  /**
   * Get single delivery task by server ID
   * @param {number} serverId - Server task ID
   */
  async getTaskByServerId(serverId) {
    try {
      const task = await this.db.deliveryTasks
        .where('serverId')
        .equals(serverId)
        .first()
      
      if (task && task.menuItems) {
        task.menuItems = JSON.parse(task.menuItems)
      }
      
      return task
    } catch (error) {
      console.error('Error getting task by server ID:', error)
      throw error
    }
  }

  // === SCHOOLS METHODS ===
  
  /**
   * Save school data to cache
   * @param {Object} school - School data
   */
  async saveSchool(school) {
    try {
      const schoolData = {
        serverId: school.id,
        name: school.name,
        address: school.address,
        latitude: school.latitude,
        longitude: school.longitude,
        contactPerson: school.contact_person || school.contactPerson,
        phoneNumber: school.phone_number || school.phoneNumber,
        studentCount: school.student_count || school.studentCount,
        isActive: school.is_active !== undefined ? school.is_active : school.isActive,
        cachedAt: new Date().toISOString()
      }

      // Check if school already exists
      const existingSchool = await this.db.schools
        .where('serverId')
        .equals(school.id)
        .first()

      if (existingSchool) {
        await this.db.schools.update(existingSchool.id, schoolData)
        return existingSchool.id
      } else {
        return await this.db.schools.add(schoolData)
      }
    } catch (error) {
      console.error('Error saving school:', error)
      throw error
    }
  }

  /**
   * Get schools from cache
   * @param {Object} filters - Filter options
   */
  async getSchools(filters = {}) {
    try {
      let query = this.db.schools.toCollection()

      if (filters.isActive !== undefined) {
        query = query.filter(school => school.isActive === filters.isActive)
      }

      return await query.toArray()
    } catch (error) {
      console.error('Error getting schools:', error)
      throw error
    }
  }

  /**
   * Get school by server ID
   * @param {number} serverId - Server school ID
   */
  async getSchoolByServerId(serverId) {
    try {
      return await this.db.schools
        .where('serverId')
        .equals(serverId)
        .first()
    } catch (error) {
      console.error('Error getting school by server ID:', error)
      throw error
    }
  }

  // === e-POD METHODS ===
  
  /**
   * Save e-POD data locally
   * @param {Object} epodData - e-POD data
   */
  async saveePOD(epodData) {
    try {
      const epod = {
        deliveryTaskId: epodData.delivery_task_id || epodData.deliveryTaskId,
        serverId: epodData.server_id || null,
        latitude: epodData.latitude,
        longitude: epodData.longitude,
        accuracy: epodData.accuracy,
        recipientName: epodData.recipient_name || epodData.recipientName,
        omprengDropOff: epodData.ompreng_drop_off || epodData.omprengDropOff,
        omprengPickUp: epodData.ompreng_pick_up || epodData.omprengPickUp,
        completedAt: epodData.completed_at || epodData.completedAt || new Date().toISOString(),
        syncStatus: epodData.syncStatus || 'pending',
        retryCount: 0,
        lastAttempt: null,
        createdAt: new Date().toISOString(),
        lastSync: null
      }

      return await this.db.epods.add(epod)
    } catch (error) {
      console.error('Error saving e-POD:', error)
      throw error
    }
  }

  /**
   * Get pending e-PODs for sync
   * @param {Object} filters - Filter options
   */
  async getPendingePODs(filters = {}) {
    try {
      let query = this.db.epods.where('syncStatus').anyOf(['pending', 'failed'])

      if (filters.deliveryTaskId) {
        query = query.and(epod => epod.deliveryTaskId === filters.deliveryTaskId)
      }

      return await query.toArray()
    } catch (error) {
      console.error('Error getting pending e-PODs:', error)
      throw error
    }
  }

  /**
   * Update e-POD sync status
   * @param {number} epodId - Local e-POD ID
   * @param {string} status - Sync status
   * @param {Object} additionalData - Additional data to update
   */
  async updateePODSyncStatus(epodId, status, additionalData = {}) {
    try {
      const updateData = {
        syncStatus: status,
        lastSync: new Date().toISOString(),
        ...additionalData
      }

      await this.db.epods.update(epodId, updateData)
    } catch (error) {
      console.error('Error updating e-POD sync status:', error)
      throw error
    }
  }

  // === MEDIA STORAGE METHODS ===
  
  /**
   * Save photo data
   * @param {Object} photoData - Photo data
   */
  async savePhoto(photoData) {
    try {
      const photo = {
        taskId: photoData.taskId,
        epodId: photoData.epodId || null,
        photoData: photoData.photoData,
        timestamp: new Date().toISOString(),
        synced: false,
        syncedAt: null,
        photoUrl: null,
        fileSize: photoData.fileSize || 0
      }

      return await this.db.photos.add(photo)
    } catch (error) {
      console.error('Error saving photo:', error)
      throw error
    }
  }

  /**
   * Save signature data
   * @param {Object} signatureData - Signature data
   */
  async saveSignature(signatureData) {
    try {
      const signature = {
        taskId: signatureData.taskId,
        epodId: signatureData.epodId || null,
        signatureData: signatureData.signatureData,
        quality: signatureData.quality || 'good',
        timestamp: new Date().toISOString(),
        synced: false,
        syncedAt: null,
        signatureUrl: null,
        fileSize: signatureData.fileSize || 0
      }

      return await this.db.signatures.add(signature)
    } catch (error) {
      console.error('Error saving signature:', error)
      throw error
    }
  }

  // === SYNC METADATA METHODS ===
  
  /**
   * Set sync metadata
   * @param {string} key - Metadata key
   * @param {any} value - Metadata value
   */
  async setSyncMeta(key, value) {
    try {
      const existing = await this.db.syncMeta.where('key').equals(key).first()
      const data = {
        key,
        value: JSON.stringify(value),
        updatedAt: new Date().toISOString()
      }

      if (existing) {
        await this.db.syncMeta.update(existing.id, data)
      } else {
        await this.db.syncMeta.add(data)
      }
    } catch (error) {
      console.error('Error setting sync metadata:', error)
      throw error
    }
  }

  /**
   * Get sync metadata
   * @param {string} key - Metadata key
   */
  async getSyncMeta(key) {
    try {
      const meta = await this.db.syncMeta.where('key').equals(key).first()
      return meta ? JSON.parse(meta.value) : null
    } catch (error) {
      console.error('Error getting sync metadata:', error)
      return null
    }
  }

  // === UTILITY METHODS ===
  
  /**
   * Clear all cached data
   */
  async clearCache() {
    try {
      await Promise.all([
        this.db.deliveryTasks.clear(),
        this.db.schools.clear(),
        this.db.epods.clear(),
        this.db.photos.clear(),
        this.db.signatures.clear(),
        this.db.syncQueue.clear(),
        this.db.syncLog.clear(),
        this.db.syncMeta.clear()
      ])
      console.log('Cache cleared successfully')
    } catch (error) {
      console.error('Error clearing cache:', error)
      throw error
    }
  }

  /**
   * Get cache statistics
   */
  async getCacheStats() {
    try {
      const [
        deliveryTasksCount,
        schoolsCount,
        epodsCount,
        photosCount,
        signaturesCount,
        syncQueueCount,
        syncLogCount
      ] = await Promise.all([
        this.db.deliveryTasks.count(),
        this.db.schools.count(),
        this.db.epods.count(),
        this.db.photos.count(),
        this.db.signatures.count(),
        this.db.syncQueue.count(),
        this.db.syncLog.count()
      ])

      return {
        deliveryTasks: deliveryTasksCount,
        schools: schoolsCount,
        epods: epodsCount,
        photos: photosCount,
        signatures: signaturesCount,
        syncQueue: syncQueueCount,
        syncLog: syncLogCount,
        total: deliveryTasksCount + schoolsCount + epodsCount + photosCount + signaturesCount
      }
    } catch (error) {
      console.error('Error getting cache stats:', error)
      return {}
    }
  }

  /**
   * Clean up old cached data
   * @param {number} daysToKeep - Number of days to keep data
   */
  async cleanupOldCache(daysToKeep = 7) {
    try {
      const cutoffDate = new Date()
      cutoffDate.setDate(cutoffDate.getDate() - daysToKeep)
      const cutoffISO = cutoffDate.toISOString()

      const deleteCounts = await Promise.all([
        this.db.deliveryTasks.where('cachedAt').below(cutoffISO).delete(),
        this.db.schools.where('cachedAt').below(cutoffISO).delete(),
        this.db.syncLog.where('timestamp').below(cutoffISO).delete()
      ])

      const totalDeleted = deleteCounts.reduce((sum, count) => sum + count, 0)
      console.log(`Cleaned up ${totalDeleted} old cache entries`)
      return totalDeleted
    } catch (error) {
      console.error('Error cleaning up old cache:', error)
      return 0
    }
  }
}

// Create storage service instance
const storageService = new OfflineStorageService()

export default db
export { storageService }
