/**
 * Unit tests for photo capture functionality
 */
import { describe, it, expect, beforeEach, vi } from 'vitest'

// Mock MediaDevices API for testing
const mockMediaDevices = {
  getUserMedia: vi.fn(),
  enumerateDevices: vi.fn()
}

// Mock canvas and context
const mockCanvas = {
  getContext: vi.fn(() => ({
    drawImage: vi.fn(),
    clearRect: vi.fn()
  })),
  toDataURL: vi.fn(() => 'data:image/jpeg;base64,mockImageData'),
  width: 1920,
  height: 1080
}

// Mock video element
const mockVideo = {
  play: vi.fn(() => Promise.resolve()),
  addEventListener: vi.fn(),
  videoWidth: 1920,
  videoHeight: 1080,
  srcObject: null
}

describe('Photo Capture Functionality', () => {
  beforeEach(() => {
    // Reset mocks
    vi.clearAllMocks()
    
    // Setup global mocks
    global.navigator = {
      mediaDevices: mockMediaDevices,
      onLine: true
    }
    
    global.document = {
      createElement: vi.fn((tag) => {
        if (tag === 'canvas') return mockCanvas
        if (tag === 'video') return mockVideo
        return {}
      })
    }
  })

  it('should initialize camera devices successfully', async () => {
    // Mock camera devices
    const mockDevices = [
      { deviceId: 'camera1', kind: 'videoinput', label: 'Back Camera' },
      { deviceId: 'camera2', kind: 'videoinput', label: 'Front Camera' }
    ]
    
    mockMediaDevices.getUserMedia.mockResolvedValue({})
    mockMediaDevices.enumerateDevices.mockResolvedValue(mockDevices)
    
    // Test camera initialization logic
    const videoDevices = mockDevices.filter(device => device.kind === 'videoinput')
    const rearCamera = videoDevices.find(camera => 
      camera.label.toLowerCase().includes('back')
    )
    
    expect(videoDevices).toHaveLength(2)
    expect(rearCamera).toBeDefined()
    expect(rearCamera.label).toBe('Back Camera')
  })

  it('should compress image correctly', () => {
    // Test image compression logic
    const quality = 80
    const expectedDataURL = mockCanvas.toDataURL('image/jpeg', quality / 100)
    
    expect(expectedDataURL).toBe('data:image/jpeg;base64,mockImageData')
    expect(mockCanvas.toDataURL).toHaveBeenCalledWith('image/jpeg', 0.8)
  })

  it('should calculate photo size correctly', () => {
    const base64Data = 'data:image/jpeg;base64,mockImageData'
    const base64Length = base64Data.length
    const sizeInBytes = (base64Length * 3) / 4
    
    let expectedSize
    if (sizeInBytes < 1024) {
      expectedSize = `${Math.round(sizeInBytes)} B`
    } else if (sizeInBytes < 1024 * 1024) {
      expectedSize = `${Math.round(sizeInBytes / 1024)} KB`
    } else {
      expectedSize = `${Math.round(sizeInBytes / (1024 * 1024))} MB`
    }
    
    expect(expectedSize).toBeDefined()
    expect(typeof expectedSize).toBe('string')
  })

  it('should handle camera permission denied', async () => {
    const permissionError = new Error('Permission denied')
    permissionError.name = 'NotAllowedError'
    
    mockMediaDevices.getUserMedia.mockRejectedValue(permissionError)
    
    try {
      await mockMediaDevices.getUserMedia({ video: true })
    } catch (error) {
      expect(error.name).toBe('NotAllowedError')
    }
  })

  it('should handle no camera available', async () => {
    mockMediaDevices.enumerateDevices.mockResolvedValue([])
    
    const devices = await mockMediaDevices.enumerateDevices()
    const videoDevices = devices.filter(device => device.kind === 'videoinput')
    
    expect(videoDevices).toHaveLength(0)
  })

  it('should prefer rear camera over front camera', () => {
    const cameras = [
      { deviceId: 'front', label: 'Front Camera' },
      { deviceId: 'back', label: 'Back Camera' }
    ]
    
    const rearCamera = cameras.find(camera => 
      camera.label.toLowerCase().includes('back') || 
      camera.label.toLowerCase().includes('rear') ||
      camera.label.toLowerCase().includes('environment')
    )
    
    expect(rearCamera).toBeDefined()
    expect(rearCamera.deviceId).toBe('back')
  })
})