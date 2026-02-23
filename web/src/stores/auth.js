import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import authService from '@/services/authService'

export const useAuthStore = defineStore('auth', () => {
  const user = ref(null)
  const token = ref(localStorage.getItem('token') || null)
  const loading = ref(false)
  const error = ref(null)

  const isAuthenticated = computed(() => !!token.value)

  // Initialize user from localStorage if token exists
  const initAuth = () => {
    const storedUser = localStorage.getItem('user')
    if (storedUser && token.value) {
      try {
        user.value = JSON.parse(storedUser)
      } catch (e) {
        clearAuth()
      }
    }
  }

  const login = async (credentials) => {
    loading.value = true
    error.value = null
    try {
      const response = await authService.login(credentials)
      setAuth(response.user, response.token)
      return response
    } catch (err) {
      error.value = err.response?.data?.message || 'Login gagal. Silakan coba lagi.'
      throw err
    } finally {
      loading.value = false
    }
  }

  const logout = async () => {
    loading.value = true
    try {
      await authService.logout()
    } catch (err) {
      console.error('Logout error:', err)
    } finally {
      clearAuth()
      loading.value = false
    }
  }

  const refreshToken = async () => {
    try {
      const response = await authService.refreshToken()
      setAuth(response.user, response.token)
      return response
    } catch (err) {
      clearAuth()
      throw err
    }
  }

  const getCurrentUser = async () => {
    try {
      const response = await authService.getCurrentUser()
      user.value = response.user
      localStorage.setItem('user', JSON.stringify(response.user))
      return response.user
    } catch (err) {
      clearAuth()
      throw err
    }
  }

  function setAuth(userData, authToken) {
    user.value = userData
    token.value = authToken
    localStorage.setItem('token', authToken)
    localStorage.setItem('user', JSON.stringify(userData))
  }

  function clearAuth() {
    user.value = null
    token.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  // Initialize on store creation
  initAuth()

  return {
    user,
    token,
    loading,
    error,
    isAuthenticated,
    login,
    logout,
    refreshToken,
    getCurrentUser,
    setAuth,
    clearAuth
  }
})
