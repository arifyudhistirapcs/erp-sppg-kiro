import api from './api'

const authService = {
  /**
   * Login with NIK/Email and password
   * @param {Object} credentials - { identifier: string, password: string }
   * @returns {Promise<Object>} - { user, token }
   */
  async login(credentials) {
    const response = await api.post('/auth/login', {
      identifier: credentials.identifier, // NIK or Email
      password: credentials.password
    })
    return response.data
  },

  /**
   * Logout current user
   * @returns {Promise<void>}
   */
  async logout() {
    try {
      await api.post('/auth/logout')
    } catch (error) {
      console.error('Logout error:', error)
    }
  },

  /**
   * Refresh JWT token
   * @returns {Promise<Object>} - { user, token }
   */
  async refreshToken() {
    const response = await api.post('/auth/refresh')
    return response.data
  },

  /**
   * Get current user information
   * @returns {Promise<Object>} - { user }
   */
  async getCurrentUser() {
    const response = await api.get('/auth/me')
    return response.data
  }
}

export default authService
