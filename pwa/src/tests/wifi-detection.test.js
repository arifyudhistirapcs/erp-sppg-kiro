/**
 * Wi-Fi Detection Tests
 * Tests for the Wi-Fi detection functionality in the PWA
 */

import { describe, it, expect, beforeEach, vi } from 'vitest'
import wifiService from '../services/wifiService.js'

// Mock navigator APIs
const mockNavigator = {
  onLine: true,
  connection: {
    type: 'wifi',
    effectiveType: '4g',
    downlink: 10,
    rtt: 50
  },
  geolocation: {
    getCurrentPosition: vi.fn()
  }
}

// Mock global navigator
Object.defineProperty(global, 'navigator', {
  value: mockNavigator,
  writable: true
})

describe('WiFi Detection Service', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Network Detection', () => {
    it('should detect Wi-Fi connection when available', async () => {
      mockNavigator.connection.type = 'wifi'
      
      const isConnected = await wifiService.isConnectedToWiFi()
      expect(isConnected).toBe(true)
    })

    it('should detect non-Wi-Fi connection', async () => {
      mockNavigator.connection.type = 'cellular'
      
      const isConnected = await wifiService.isConnectedToWiFi()
      expect(isConnected).toBe(false)
    })

    it('should fallback to online status when connection API unavailable', async () => {
      const originalConnection = mockNavigator.connection
      delete mockNavigator.connection
      mockNavigator.onLine = true
      
      const isConnected = await wifiService.isConnectedToWiFi()
      expect(isConnected).toBe(true)
      
      // Restore connection
      mockNavigator.connection = originalConnection
    })
  })

  describe('Network Information', () => {
    it('should return network information when available', () => {
      const networkInfo = wifiService.getNetworkInfo()
      
      expect(networkInfo).toEqual({
        type: 'wifi',
        effectiveType: '4g',
        downlink: 10,
        rtt: 50
      })
    })

    it('should return unknown values when connection API unavailable', () => {
      const originalConnection = mockNavigator.connection
      delete mockNavigator.connection
      
      const networkInfo = wifiService.getNetworkInfo()
      
      expect(networkInfo).toEqual({
        type: 'unknown',
        effectiveType: 'unknown',
        downlink: null,
        rtt: null
      })
      
      // Restore connection
      mockNavigator.connection = originalConnection
    })
  })

  describe('GPS Location', () => {
    it('should get current location successfully', async () => {
      const mockPosition = {
        coords: {
          latitude: -6.2088,
          longitude: 106.8456,
          accuracy: 10
        },
        timestamp: Date.now()
      }

      mockNavigator.geolocation.getCurrentPosition.mockImplementation((success) => {
        success(mockPosition)
      })

      const location = await wifiService.getCurrentLocation()
      
      expect(location).toEqual({
        latitude: -6.2088,
        longitude: 106.8456,
        accuracy: 10,
        timestamp: mockPosition.timestamp
      })
    })

    it('should handle geolocation errors', async () => {
      const mockError = {
        code: 1, // PERMISSION_DENIED
        message: 'Permission denied'
      }

      mockNavigator.geolocation.getCurrentPosition.mockImplementation((success, error) => {
        error(mockError)
      })

      await expect(wifiService.getCurrentLocation()).rejects.toThrow('Akses lokasi ditolak')
    })
  })

  describe('Distance Calculation', () => {
    it('should calculate distance between two GPS coordinates', () => {
      // Simplified distance test with closer coordinates for faster calculation
      const distance = wifiService.calculateDistance(
        -6.2088, 106.8456, // Jakarta
        -6.2100, 106.8470  // Very close location
      )
      
      // Should be a small distance (less than 2km)
      expect(distance).toBeGreaterThan(0)
      expect(distance).toBeLessThan(2000)
    })

    it('should return 0 for same coordinates', () => {
      const distance = wifiService.calculateDistance(
        -6.2088, 106.8456,
        -6.2088, 106.8456
      )
      
      expect(distance).toBe(0)
    })
  })

  describe('Wi-Fi Validation', () => {
    const authorizedNetworks = [
      {
        ssid: 'SPPG-Office',
        bssid: '00:00:00:00:00:00',
        location: 'Kantor Pusat',
        gps_boundaries: {
          center_lat: -6.2088,
          center_lng: 106.8456,
          radius_meters: 100
        }
      }
    ]

    beforeEach(() => {
      mockNavigator.connection.type = 'wifi'
    })

    it('should validate authorized SSID', async () => {
      const result = await wifiService.validateWiFiConnection(
        'SPPG-Office',
        null,
        authorizedNetworks
      )
      
      expect(result.isValid).toBe(true)
      expect(result.method).toBe('manual_ssid')
      expect(result.ssid).toBe('SPPG-Office')
    })

    it('should reject unauthorized SSID', async () => {
      const result = await wifiService.validateWiFiConnection(
        'Unknown-WiFi',
        null,
        authorizedNetworks
      )
      
      expect(result.isValid).toBe(false)
      expect(result.method).toBe('manual_ssid')
      expect(result.error).toBe('SSID tidak diotorisasi')
    })

    it('should validate GPS location within office area', async () => {
      const officeLocation = {
        latitude: -6.2088,
        longitude: 106.8456
      }
      
      const result = await wifiService.validateWiFiConnection(
        null,
        officeLocation,
        authorizedNetworks
      )
      
      expect(result.isValid).toBe(true)
      expect(result.method).toBe('gps_validation')
    })
  })

  describe('GPS Location Validation', () => {
    const authorizedNetworks = [
      {
        ssid: 'SPPG-Office',
        gps_boundaries: {
          center_lat: -6.2088,
          center_lng: 106.8456,
          radius_meters: 100
        }
      }
    ]

    it('should validate location within radius', () => {
      const location = {
        latitude: -6.2089, // Very close to center
        longitude: 106.8457
      }
      
      const result = wifiService.validateGPSLocation(location, authorizedNetworks)
      
      expect(result.isValid).toBe(true)
      expect(result.network).toBe('SPPG-Office')
      expect(result.distance).toBeLessThan(100)
    })

    it('should reject location outside radius', () => {
      const location = {
        latitude: -6.2200, // Far from center
        longitude: 106.8600
      }
      
      const result = wifiService.validateGPSLocation(location, authorizedNetworks)
      
      expect(result.isValid).toBe(false)
      expect(result.error).toBe('Lokasi di luar area kantor')
    })
  })
})

describe('Attendance Service Integration', () => {
  // These would be integration tests that test the full flow
  // For now, we'll just test that the services can be imported
  it('should import attendance service successfully', async () => {
    const { default: attendanceService } = await import('../services/attendanceService.js')
    expect(attendanceService).toBeDefined()
    expect(typeof attendanceService.checkIn).toBe('function')
    expect(typeof attendanceService.checkOut).toBe('function')
  })
})