import axios from 'axios'
import { useAuthStore } from '@/stores/auth'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
    'ngrok-skip-browser-warning': 'true'
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
    // Only redirect to login if 401 and NOT on login page
    if (error.response?.status === 401 && !window.location.pathname.includes('/login')) {
      const authStore = useAuthStore()
      authStore.clearAuth()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// Auth API methods
export const authAPI = {
  // Login
  login: (credentials) => 
    api.post('/auth/login', credentials),
  
  // Logout
  logout: () => 
    api.post('/auth/logout'),
  
  // Get current user profile
  getProfile: () => 
    api.get('/auth/profile')
}

// Attendance API methods
export const attendanceAPI = {
  // Get current day attendance (uses current user from JWT)
  getCurrentAttendance: () => 
    api.get('/attendance/today'),
  
  // Check-in
  checkIn: (data) => 
    api.post('/attendance/check-in', data),
  
  // Check-out
  checkOut: (data) => 
    api.post('/attendance/check-out', data),
  
  // Get attendance history by date range
  getHistory: (days = 7) => {
    const endDate = new Date()
    const startDate = new Date()
    startDate.setDate(startDate.getDate() - days)
    
    const formatDate = (date) => date.toISOString().split('T')[0]
    
    return api.get('/attendance/by-date-range', {
      params: {
        start_date: formatDate(startDate),
        end_date: formatDate(endDate)
      }
    })
  },
  
  // Validate Wi-Fi only (for testing)
  validateWiFi: (data) => 
    api.post('/attendance/validate-wifi', data)
}

// Wi-Fi configuration API methods
export const wifiAPI = {
  // Get authorized Wi-Fi networks
  getAuthorizedNetworks: () => 
    api.get('/wifi-config?active_only=true'),
  
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
