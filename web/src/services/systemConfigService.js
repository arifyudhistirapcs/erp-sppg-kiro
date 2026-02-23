import api from './api'

const systemConfigService = {
  // Get all system configurations
  async getConfigs(params = {}) {
    const response = await api.get('/system-config', { params })
    return response.data
  },

  // Get configurations grouped by category
  async getConfigsByCategory() {
    const response = await api.get('/system-config/by-category')
    return response.data
  },

  // Get specific configuration by key
  async getConfig(key) {
    const response = await api.get(`/system-config/${key}`)
    return response.data
  },

  // Set single configuration
  async setConfig(configData) {
    const response = await api.post('/system-config', configData)
    return response.data
  },

  // Set multiple configurations at once
  async setMultipleConfigs(configs) {
    const response = await api.post('/system-config/bulk', configs)
    return response.data
  },

  // Initialize default configurations
  async initializeDefaults() {
    const response = await api.post('/system-config/initialize-defaults')
    return response.data
  },

  // Delete configuration
  async deleteConfig(key) {
    const response = await api.delete(`/system-config/${key}`)
    return response.data
  }
}

export default systemConfigService