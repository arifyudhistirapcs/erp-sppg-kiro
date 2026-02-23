import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor
api.interceptors.request.use(
  (config) => {
    const authStore = useAuthStore()
    if (authStore.token) {
      config.headers.Authorization = `Bearer ${authStore.token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor
api.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {
    if (error.response?.status === 401) {
      const authStore = useAuthStore()
      authStore.clearAuth()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// Attendance API methods
export const attendanceAPI = {
  // Get current day attendance
  getCurrentAttendance: (employeeId) => 
    api.get(`/attendance/today/${employeeId}`),
  
  // Check-in
  checkIn: (data) => 
    api.post('/attendance/check-in', data),
  
  // Check-out
  checkOut: (data) => 
    api.post('/attendance/check-out', data),
  
  // Get attendance history
  getHistory: (employeeId, days = 30) => 
    api.get(`/attendance/history/${employeeId}?days=${days}`),
  
  // Validate Wi-Fi only (for testing)
  validateWiFi: (data) => 
    api.post('/attendance/validate-wifi', data)
}

// Wi-Fi configuration API methods
export const wifiAPI = {
  // Get authorized Wi-Fi networks
  getAuthorizedNetworks: () => 
    api.get('/wifi-config'),
  
  // Add authorized network (admin only)
  addAuthorizedNetwork: (data) => 
    api.post('/wifi-config', data),
  
  // Update authorized network (admin only)
  updateAuthorizedNetwork: (id, data) => 
    api.put(`/wifi-config/${id}`, data),
  
  // Delete authorized network (admin only)
  deleteAuthorizedNetwork: (id) => 
    api.delete(`/wifi-config/${id}`)
}

export default api
