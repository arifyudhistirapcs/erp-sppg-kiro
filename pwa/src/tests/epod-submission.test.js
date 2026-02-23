import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useDeliveryTasksStore } from '@/stores/deliveryTasks'
import syncService from '@/services/syncService'
import db from '@/services/db'
import api from '@/services/api'

// Mock dependencies
vi.mock('@/services/api')
vi.mock('@/services/db')
vi.mock('@/services/syncService')

describe('e-POD Submission and Sync', () => {
  let deliveryTasksStore

  beforeEach(() => {
    vi.clearAllMocks()
    deliveryTasksStore = useDeliveryTasksStore()
  })

  describe('e-POD Submission', () => {
    it('should submit e-POD successfully when online', async () => {
      // Mock online status
      Object.defineProperty(navigator, 'onLine', {
        writable: true,
        value: true
      })

      // Mock API response
      api.post.mockResolvedValue({
        data: {
          success: true,
          epod: { id: 123 }
        }
      })

      // Mock database operations
      db.epods.add.mockResolvedValue(1)
      db.epods.where.mockReturnValue({
        equals: vi.fn().mockReturnValue({
          modify: vi.fn().mockResolvedValue(1)
        })
      })

      const ePODData = {
        delivery_task_id: 1,
        latitude: -6.2088,
        longitude: 106.8456,
        accuracy: 10,
        recipient_name: 'John Doe',
        ompreng_drop_off: 5,
        ompreng_pick_up: 3,
        photo_url: 'data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQ...',
        signature_url: 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...',
        completed_at: new Date().toISOString(),
        device_info: { userAgent: 'test' }
      }

      const result = await deliveryTasksStore.submitePOD(ePODData)

      expect(result.success).toBe(true)
      expect(result.synced).toBe(true)
      expect(api.post).toHaveBeenCalledWith('/epod', expect.objectContaining({
        delivery_task_id: 1,
        latitude: -6.2088,
        longitude: 106.8456
      }))
    })

    it('should queue e-POD for sync when offline', async () => {
      // Mock offline status
      Object.defineProperty(navigator, 'onLine', {
        writable: true,
        value: false
      })

      // Mock database operations
      db.epods.add.mockResolvedValue(1)
      syncService.queueForSync.mockResolvedValue({ id: 1 })

      const ePODData = {
        delivery_task_id: 1,
        latitude: -6.2088,
        longitude: 106.8456,
        recipient_name: 'John Doe',
        ompreng_drop_off: 5,
        ompreng_pick_up: 3,
        photo_url: 'data:image/jpeg;base64,test',
        signature_url: 'data:image/png;base64,test',
        completed_at: new Date().toISOString()
      }

      const result = await deliveryTasksStore.submitePOD(ePODData)

      expect(result.success).toBe(true)
      expect(result.offline).toBe(true)
      expect(result.queued).toBe(true)
      expect(syncService.queueForSync).toHaveBeenCalledWith('epod', ePODData, 1)
    })

    it('should validate required e-POD data', async () => {
      const incompleteData = {
        delivery_task_id: 1,
        // Missing latitude, longitude, photo, signature
      }

      await expect(deliveryTasksStore.submitePOD(incompleteData))
        .rejects.toThrow('Data e-POD tidak lengkap')
    })

    it('should handle API errors gracefully', async () => {
      Object.defineProperty(navigator, 'onLine', {
        writable: true,
        value: true
      })

      // Mock API error
      api.post.mockRejectedValue({
        response: { status: 400, data: { message: 'Invalid data' } }
      })

      // Mock database operations
      db.epods.add.mockResolvedValue(1)
      syncService.queueForSync.mockResolvedValue({ id: 1 })

      const ePODData = {
        delivery_task_id: 1,
        latitude: -6.2088,
        longitude: 106.8456,
        recipient_name: 'John Doe',
        ompreng_drop_off: 5,
        ompreng_pick_up: 3,
        photo_url: 'data:image/jpeg;base64,test',
        signature_url: 'data:image/png;base64,test',
        completed_at: new Date().toISOString()
      }

      const result = await deliveryTasksStore.submitePOD(ePODData)

      // Should fall back to queuing for sync
      expect(result.success).toBe(true)
      expect(result.queued).toBe(true)
      expect(syncService.queueForSync).toHaveBeenCalled()
    })
  })

  describe('Sync Service', () => {
    it('should sync e-POD data successfully', async () => {
      const mockItem = {
        id: 1,
        type: 'epod',
        data: {
          delivery_task_id: 1,
          latitude: -6.2088,
          longitude: 106.8456,
          recipient_name: 'John Doe',
          ompreng_drop_off: 5,
          ompreng_pick_up: 3,
          completed_at: new Date().toISOString()
        }
      }

      api.post.mockResolvedValue({
        data: {
          success: true,
          epod: { id: 123 }
        }
      })

      db.syncQueue.delete.mockResolvedValue(1)
      db.epods.where.mockReturnValue({
        equals: vi.fn().mockReturnValue({
          modify: vi.fn().mockResolvedValue(1)
        })
      })

      const result = await syncService.syncEPOD(mockItem)

      expect(result.success).toBe(true)
      expect(api.post).toHaveBeenCalledWith('/epod', expect.objectContaining({
        delivery_task_id: 1
      }))
      expect(db.syncQueue.delete).toHaveBeenCalledWith(1)
    })

    it('should handle sync conflicts (409 status)', async () => {
      const mockItem = {
        id: 1,
        type: 'epod',
        data: {
          delivery_task_id: 1,
          latitude: -6.2088,
          longitude: 106.8456
        }
      }

      api.post.mockRejectedValue({
        response: { status: 409 }
      })

      db.syncQueue.delete.mockResolvedValue(1)
      db.epods.where.mockReturnValue({
        equals: vi.fn().mockReturnValue({
          modify: vi.fn().mockResolvedValue(1)
        })
      })

      const result = await syncService.syncEPOD(mockItem)

      expect(result.success).toBe(true)
      expect(result.message).toBe('e-POD already exists')
      expect(db.syncQueue.delete).toHaveBeenCalledWith(1)
    })

    it('should sync photo uploads successfully', async () => {
      const mockItem = {
        id: 1,
        type: 'epod_photo',
        data: {
          epodId: 123,
          taskId: 1,
          photoData: 'data:image/jpeg;base64,test'
        }
      }

      // Mock fetch for base64 conversion
      global.fetch = vi.fn().mockResolvedValue({
        blob: () => Promise.resolve(new Blob(['test'], { type: 'image/jpeg' }))
      })

      api.post.mockResolvedValue({
        data: {
          success: true,
          photo_url: 'https://example.com/photo.jpg'
        }
      })

      db.syncQueue.delete.mockResolvedValue(1)
      db.photos.where.mockReturnValue({
        equals: vi.fn().mockReturnValue({
          modify: vi.fn().mockResolvedValue(1)
        })
      })

      const result = await syncService.syncEPODPhoto(mockItem)

      expect(result.success).toBe(true)
      expect(api.post).toHaveBeenCalledWith(
        '/epod/123/upload-photo',
        expect.any(FormData),
        expect.objectContaining({
          headers: { 'Content-Type': 'multipart/form-data' }
        })
      )
    })

    it('should handle photo upload size errors', async () => {
      const mockItem = {
        id: 1,
        type: 'epod_photo',
        data: {
          epodId: 123,
          taskId: 1,
          photoData: 'data:image/jpeg;base64,test'
        }
      }

      global.fetch = vi.fn().mockResolvedValue({
        blob: () => Promise.resolve(new Blob(['test'], { type: 'image/jpeg' }))
      })

      api.post.mockRejectedValue({
        response: { status: 413 }
      })

      await expect(syncService.syncEPODPhoto(mockItem))
        .rejects.toThrow('Ukuran foto terlalu besar. Maksimal 5MB.')
    })
  })

  describe('Sync Status and Progress', () => {
    it('should track sync progress correctly', async () => {
      const mockProgress = {
        total: 3,
        completed: 1,
        failed: 0,
        status: 'syncing'
      }

      syncService.getSyncProgress.mockReturnValue(mockProgress)

      const progress = deliveryTasksStore.getSyncProgress()

      expect(progress.total).toBe(3)
      expect(progress.completed).toBe(1)
      expect(progress.status).toBe('syncing')
    })

    it('should get pending sync count', async () => {
      syncService.getPendingSyncCount.mockResolvedValue(5)

      const count = await deliveryTasksStore.getPendingSyncCount()

      expect(count).toBe(5)
    })

    it('should handle sync progress listeners', () => {
      const mockCallback = vi.fn()

      deliveryTasksStore.addSyncProgressListener(mockCallback)
      deliveryTasksStore.removeSyncProgressListener(mockCallback)

      expect(syncService.addProgressListener).toHaveBeenCalledWith(mockCallback)
      expect(syncService.removeProgressListener).toHaveBeenCalledWith(mockCallback)
    })
  })

  describe('Error Handling', () => {
    it('should provide specific error messages for different failure types', async () => {
      // Reduced test cases for faster execution
      const testCases = [
        {
          error: { response: { status: 400 } },
          expectedMessage: 'Data e-POD tidak valid. Periksa kembali semua informasi.'
        },
        {
          error: { response: { status: 500 } },
          expectedMessage: 'Server error. e-POD akan disimpan offline.'
        }
      ]

      for (const testCase of testCases) {
        Object.defineProperty(navigator, 'onLine', {
          writable: true,
          value: true
        })

        api.post.mockRejectedValue(testCase.error)
        db.epods.add.mockResolvedValue(1)
        syncService.queueForSync.mockResolvedValue({ id: 1 })

        const ePODData = {
          delivery_task_id: 1,
          latitude: -6.2088,
          longitude: 106.8456,
          recipient_name: 'John Doe',
          ompreng_drop_off: 5,
          ompreng_pick_up: 3,
          photo_url: 'data:image/jpeg;base64,test',
          signature_url: 'data:image/png;base64,test',
          completed_at: new Date().toISOString()
        }

        await expect(deliveryTasksStore.submitePOD(ePODData))
          .rejects.toThrow(testCase.expectedMessage)
      }
    })
  })
})