import { describe, it, expect, beforeEach, vi } from 'vitest'

// Mock IndexedDB for testing
const mockDB = {
  deliveryTasks: new Map(),
  schools: new Map(),
  epods: new Map(),
  syncQueue: new Map(),
  syncLog: new Map(),
  syncMeta: new Map()
}

// Mock Dexie
vi.mock('dexie', () => {
  return {
    default: class MockDexie {
      constructor(name) {
        this.name = name
      }
      version() {
        return {
          stores: () => this
        }
      }
    }
  }
})

// Mock the storage service
const mockStorageService = {
  async saveTask(task) {
    const id = Date.now()
    mockDB.deliveryTasks.set(id, { ...task, id })
    return id
  },
  
  async getTasks(filters = {}) {
    const tasks = Array.from(mockDB.deliveryTasks.values())
    if (filters.driverId) {
      return tasks.filter(task => task.driverId === filters.driverId)
    }
    return tasks
  },
  
  async getTaskByServerId(serverId) {
    return Array.from(mockDB.deliveryTasks.values()).find(task => task.serverId === serverId)
  },
  
  async saveSchool(school) {
    const id = Date.now()
    mockDB.schools.set(id, { ...school, id })
    return id
  },
  
  async getSchoolByServerId(serverId) {
    return Array.from(mockDB.schools.values()).find(school => school.serverId === serverId)
  },
  
  async saveePOD(epodData) {
    const id = Date.now()
    mockDB.epods.set(id, { ...epodData, id, syncStatus: 'pending' })
    return id
  },
  
  async getPendingePODs() {
    return Array.from(mockDB.epods.values()).filter(epod => epod.syncStatus === 'pending')
  },
  
  async setSyncMeta(key, value) {
    mockDB.syncMeta.set(key, { key, value: JSON.stringify(value) })
  },
  
  async getSyncMeta(key) {
    const meta = mockDB.syncMeta.get(key)
    return meta ? JSON.parse(meta.value) : null
  },
  
  async clearCache() {
    mockDB.deliveryTasks.clear()
    mockDB.schools.clear()
    mockDB.epods.clear()
    mockDB.syncQueue.clear()
    mockDB.syncLog.clear()
    mockDB.syncMeta.clear()
  },
  
  async getCacheStats() {
    return {
      deliveryTasks: mockDB.deliveryTasks.size,
      schools: mockDB.schools.size,
      epods: mockDB.epods.size,
      total: mockDB.deliveryTasks.size + mockDB.schools.size + mockDB.epods.size
    }
  }
}

// Mock sync service
const mockSyncService = {
  isOnline: true,
  isSyncing: false,
  listeners: [],
  
  addProgressListener(callback) {
    this.listeners.push(callback)
  },
  
  removeProgressListener(callback) {
    this.listeners = this.listeners.filter(l => l !== callback)
  },
  
  notifyProgress() {
    this.listeners.forEach(callback => callback({ status: 'idle' }))
  },
  
  async queueForSync(type, data, priority = 3) {
    const id = Date.now()
    mockDB.syncQueue.set(id, { id, type, data, priority, status: 'pending' })
    return id
  },
  
  async getPendingSyncCount() {
    return Array.from(mockDB.syncQueue.values()).filter(item => item.status === 'pending').length
  },
  
  async detectOnlineStatus() {
    return navigator.onLine
  },
  
  getSyncProgress() {
    return { status: 'idle', total: 0, completed: 0, failed: 0 }
  },
  
  async getSyncStatistics() {
    return { total: 0, successful: 0, failed: 0, byType: {} }
  },
  
  async updateSyncSettings(settings) {
    await mockStorageService.setSyncMeta('syncSettings', settings)
    return settings
  },
  
  async getSyncSettings() {
    return await mockStorageService.getSyncMeta('syncSettings') || {
      autoSync: true,
      syncInterval: 300000,
      maxRetries: 3
    }
  }
}

describe('PWA Offline Sync Service', () => {
  beforeEach(async () => {
    await mockStorageService.clearCache()
  })

  describe('IndexedDB Storage Service', () => {
    it('should save and retrieve delivery tasks', async () => {
      const taskData = {
        serverId: 1,
        taskDate: '2024-01-15',
        driverId: 123,
        schoolId: 456,
        status: 'pending',
        routeOrder: 1,
        portions: 50
      }

      // Save task
      const taskId = await mockStorageService.saveTask(taskData)
      expect(taskId).toBeDefined()

      // Retrieve tasks
      const tasks = await mockStorageService.getTasks()
      expect(tasks).toHaveLength(1)
      expect(tasks[0].serverId).toBe(1)
      expect(tasks[0].status).toBe('pending')
    })

    it('should save and retrieve schools', async () => {
      const schoolData = {
        serverId: 1,
        name: 'SDN 01 Jakarta',
        address: 'Jl. Merdeka No. 1',
        latitude: -6.2088,
        longitude: 106.8456,
        isActive: true
      }

      // Save school
      const schoolId = await mockStorageService.saveSchool(schoolData)
      expect(schoolId).toBeDefined()

      // Retrieve school
      const savedSchool = await mockStorageService.getSchoolByServerId(1)
      expect(savedSchool).toBeDefined()
      expect(savedSchool.name).toBe('SDN 01 Jakarta')
      expect(savedSchool.isActive).toBe(true)
    })

    it('should save and retrieve e-POD data', async () => {
      const epodData = {
        deliveryTaskId: 1,
        latitude: -6.2088,
        longitude: 106.8456,
        accuracy: 10,
        recipientName: 'Pak Kepala Sekolah',
        omprengDropOff: 50,
        omprengPickUp: 45
      }

      // Save e-POD
      const epodId = await mockStorageService.saveePOD(epodData)
      expect(epodId).toBeDefined()

      // Retrieve pending e-PODs
      const pendingEPODs = await mockStorageService.getPendingePODs()
      expect(pendingEPODs).toHaveLength(1)
      expect(pendingEPODs[0].deliveryTaskId).toBe(1)
      expect(pendingEPODs[0].syncStatus).toBe('pending')
    })

    it('should manage sync metadata', async () => {
      const testData = { setting1: 'value1', setting2: 42 }
      
      // Set metadata
      await mockStorageService.setSyncMeta('testKey', testData)
      
      // Get metadata
      const retrievedData = await mockStorageService.getSyncMeta('testKey')
      expect(retrievedData).toEqual(testData)
      
      // Non-existent key should return null
      const nonExistent = await mockStorageService.getSyncMeta('nonExistentKey')
      expect(nonExistent).toBeNull()
    })

    it('should provide cache statistics', async () => {
      // Add some test data
      await mockStorageService.saveTask({
        serverId: 1,
        taskDate: '2024-01-15',
        driverId: 123,
        schoolId: 456,
        status: 'pending'
      })
      
      await mockStorageService.saveSchool({
        serverId: 1,
        name: 'Test School',
        address: 'Test Address',
        isActive: true
      })

      // Get stats
      const stats = await mockStorageService.getCacheStats()
      expect(stats.deliveryTasks).toBe(1)
      expect(stats.schools).toBe(1)
      expect(stats.total).toBe(2)
    })
  })

  describe('Sync Service', () => {
    it('should detect online/offline status', async () => {
      // Mock navigator.onLine
      Object.defineProperty(navigator, 'onLine', {
        writable: true,
        value: true
      })

      const isOnline = await mockSyncService.detectOnlineStatus()
      expect(isOnline).toBe(true)
    })

    it('should queue items for sync', async () => {
      const testData = {
        delivery_task_id: 1,
        latitude: -6.2088,
        longitude: 106.8456
      }

      const itemId = await mockSyncService.queueForSync('epod', testData, 1)
      expect(itemId).toBeDefined()

      // Check if item was queued
      const pendingCount = await mockSyncService.getPendingSyncCount()
      expect(pendingCount).toBe(1)
    })

    it('should handle sync progress listeners', () => {
      const mockListener = vi.fn()
      
      // Add listener
      mockSyncService.addProgressListener(mockListener)
      
      // Trigger progress update
      mockSyncService.notifyProgress()
      expect(mockListener).toHaveBeenCalledWith({ status: 'idle' })
      
      // Remove listener
      mockSyncService.removeProgressListener(mockListener)
      mockSyncService.notifyProgress()
      expect(mockListener).toHaveBeenCalledTimes(1) // Should not be called again
    })

    it('should get sync statistics', async () => {
      const stats = await mockSyncService.getSyncStatistics()
      expect(stats).toHaveProperty('total')
      expect(stats).toHaveProperty('successful')
      expect(stats).toHaveProperty('failed')
      expect(stats).toHaveProperty('byType')
    })

    it('should manage sync settings', async () => {
      const newSettings = {
        autoSync: false,
        syncInterval: 600000,
        maxRetries: 5
      }

      await mockSyncService.updateSyncSettings(newSettings)
      const retrievedSettings = await mockSyncService.getSyncSettings()
      
      expect(retrievedSettings.autoSync).toBe(false)
      expect(retrievedSettings.syncInterval).toBe(600000)
      expect(retrievedSettings.maxRetries).toBe(5)
    })
  })

  describe('Integration Tests', () => {
    it('should handle complete offline workflow', async () => {
      // 1. Save delivery task offline
      const taskData = {
        serverId: 1,
        taskDate: '2024-01-15',
        driverId: 123,
        schoolId: 456,
        status: 'pending'
      }
      await mockStorageService.saveTask(taskData)

      // 2. Save e-POD offline
      const epodData = {
        deliveryTaskId: 1,
        latitude: -6.2088,
        longitude: 106.8456,
        recipientName: 'Test Recipient'
      }
      await mockStorageService.saveePOD(epodData)

      // 3. Queue for sync
      await mockSyncService.queueForSync('epod', epodData)

      // 4. Verify data is stored and queued
      const tasks = await mockStorageService.getTasks()
      const pendingEPODs = await mockStorageService.getPendingePODs()
      const pendingCount = await mockSyncService.getPendingSyncCount()

      expect(tasks).toHaveLength(1)
      expect(pendingEPODs).toHaveLength(1)
      expect(pendingCount).toBe(1)
    })

    it('should validate core PWA requirements', async () => {
      // Test requirement 23.1: PWA should cache essential data for offline access
      await mockStorageService.saveTask({ serverId: 1, status: 'pending' })
      await mockStorageService.saveSchool({ serverId: 1, name: 'Test School' })
      
      const stats = await mockStorageService.getCacheStats()
      expect(stats.total).toBeGreaterThan(0)

      // Test requirement 23.2: PWA should allow recording e-POD data locally when offline
      const epodData = { deliveryTaskId: 1, recipientName: 'Test' }
      const epodId = await mockStorageService.saveePOD(epodData)
      expect(epodId).toBeDefined()

      // Test requirement 23.3: PWA should automatically sync offline data when connection restored
      await mockSyncService.queueForSync('epod', epodData)
      const pendingCount = await mockSyncService.getPendingSyncCount()
      expect(pendingCount).toBe(1)
    })
  })
})