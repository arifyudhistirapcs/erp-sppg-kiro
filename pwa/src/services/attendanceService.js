/**
 * Attendance Service for PWA
 * Handles check-in/out operations with Wi-Fi validation
 */

import { attendanceAPI, wifiAPI } from './api.js'
import wifiService from './wifiService.js'
import { useAuthStore } from '@/stores/auth.js'

class AttendanceService {
  constructor() {
    this.currentAttendance = null
    this.authorizedNetworks = []
  }

  /**
   * Initialize attendance service
   */
  async initialize() {
    try {
      await this.loadAuthorizedNetworks()
    } catch (error) {
      console.error('Failed to initialize attendance service:', error)
    }
  }

  /**
   * Load authorized Wi-Fi networks from backend
   */
  async loadAuthorizedNetworks() {
    try {
      const response = await wifiAPI.getAuthorizedNetworks()
      this.authorizedNetworks = response.data.networks || []
      return this.authorizedNetworks
    } catch (error) {
      console.error('Failed to load authorized networks:', error)
      // Fallback to default networks if API fails
      this.authorizedNetworks = [
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
      return this.authorizedNetworks
    }
  }

  /**
   * Get current attendance status
   */
  async getCurrentAttendance() {
    try {
      const authStore = useAuthStore()
      const response = await attendanceAPI.getCurrentAttendance(authStore.user.id)
      this.currentAttendance = response.data.attendance
      return this.currentAttendance
    } catch (error) {
      console.error('Failed to get current attendance:', error)
      return null
    }
  }

  /**
   * Perform check-in with Wi-Fi validation
   * @param {string} manualSSID - Optional manual SSID input
   * @param {boolean} useGPS - Whether to use GPS validation
   * @returns {Promise<Object>}
   */
  async checkIn(manualSSID = null, useGPS = false) {
    try {
      const authStore = useAuthStore()
      
      // Step 1: Get GPS location if requested
      let gpsLocation = null
      if (useGPS) {
        try {
          gpsLocation = await wifiService.getCurrentLocation()
        } catch (error) {
          return {
            success: false,
            error: 'GPS Error',
            message: error.message
          }
        }
      }

      // Step 2: Validate Wi-Fi connection
      const wifiValidation = await wifiService.validateWiFiConnection(
        manualSSID,
        gpsLocation,
        this.authorizedNetworks
      )

      if (!wifiValidation.isValid) {
        return {
          success: false,
          error: 'Wi-Fi Validation Failed',
          message: wifiValidation.error,
          details: wifiValidation.details,
          method: wifiValidation.method
        }
      }

      // Step 3: Submit check-in to backend
      const checkInData = {
        employee_id: authStore.user.id,
        check_in_time: new Date().toISOString(),
        wifi_validation: wifiValidation,
        location: gpsLocation
      }

      const response = await attendanceAPI.checkIn(checkInData)
      
      this.currentAttendance = response.data.attendance
      
      return {
        success: true,
        message: 'Check-in berhasil!',
        attendance: this.currentAttendance,
        validation: wifiValidation
      }

    } catch (error) {
      console.error('Check-in failed:', error)
      return {
        success: false,
        error: 'Check-in Failed',
        message: error.response?.data?.message || 'Terjadi kesalahan saat check-in',
        details: error.message
      }
    }
  }

  /**
   * Perform check-out
   * @returns {Promise<Object>}
   */
  async checkOut() {
    try {
      const authStore = useAuthStore()
      
      if (!this.currentAttendance || this.currentAttendance.check_out_time) {
        return {
          success: false,
          error: 'Invalid State',
          message: 'Tidak ada check-in aktif atau sudah check-out'
        }
      }

      const checkOutData = {
        employee_id: authStore.user.id,
        attendance_id: this.currentAttendance.id,
        check_out_time: new Date().toISOString()
      }

      const response = await attendanceAPI.checkOut(checkOutData)
      
      this.currentAttendance = response.data.attendance
      
      return {
        success: true,
        message: 'Check-out berhasil!',
        attendance: this.currentAttendance,
        workHours: response.data.work_hours
      }

    } catch (error) {
      console.error('Check-out failed:', error)
      return {
        success: false,
        error: 'Check-out Failed',
        message: error.response?.data?.message || 'Terjadi kesalahan saat check-out',
        details: error.message
      }
    }
  }

  /**
   * Get attendance history
   * @param {number} days - Number of days to fetch
   * @returns {Promise<Array>}
   */
  async getAttendanceHistory(days = 30) {
    try {
      const authStore = useAuthStore()
      const response = await attendanceAPI.getHistory(authStore.user.id, days)
      return response.data.attendance_history || []
    } catch (error) {
      console.error('Failed to get attendance history:', error)
      return []
    }
  }

  /**
   * Validate Wi-Fi without checking in (for testing)
   * @param {string} manualSSID 
   * @param {boolean} useGPS 
   * @returns {Promise<Object>}
   */
  async validateWiFiOnly(manualSSID = null, useGPS = false) {
    try {
      let gpsLocation = null
      if (useGPS) {
        gpsLocation = await wifiService.getCurrentLocation()
      }

      return await wifiService.validateWiFiConnection(
        manualSSID,
        gpsLocation,
        this.authorizedNetworks
      )
    } catch (error) {
      return {
        isValid: false,
        error: 'Validation Error',
        details: error.message
      }
    }
  }

  /**
   * Get authorized networks (for UI display)
   * @returns {Array}
   */
  getAuthorizedNetworks() {
    return this.authorizedNetworks.map(network => ({
      ssid: network.ssid,
      location: network.location || 'Kantor'
    }))
  }

  /**
   * Format work hours for display
   * @param {number} hours 
   * @returns {string}
   */
  formatWorkHours(hours) {
    if (!hours) return '0 jam 0 menit'
    
    const wholeHours = Math.floor(hours)
    const minutes = Math.round((hours - wholeHours) * 60)
    
    return `${wholeHours} jam ${minutes} menit`
  }

  /**
   * Check if user can check in
   * @returns {boolean}
   */
  canCheckIn() {
    return !this.currentAttendance || 
           (this.currentAttendance && this.currentAttendance.check_out_time)
  }

  /**
   * Check if user can check out
   * @returns {boolean}
   */
  canCheckOut() {
    return this.currentAttendance && 
           this.currentAttendance.check_in_time && 
           !this.currentAttendance.check_out_time
  }
}

export default new AttendanceService()