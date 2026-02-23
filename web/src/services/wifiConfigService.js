import api from './api'

const wifiConfigService = {
  // Get all Wi-Fi configurations
  async getWiFiConfigs(activeOnly = false) {
    const response = await api.get('/wifi-config', { 
      params: { active_only: activeOnly } 
    })
    return response.data
  },

  // Create new Wi-Fi configuration
  async createWiFiConfig(configData) {
    const response = await api.post('/wifi-config', configData)
    return response.data
  },

  // Update Wi-Fi configuration
  async updateWiFiConfig(id, configData) {
    const response = await api.put(`/wifi-config/${id}`, configData)
    return response.data
  },

  // Delete Wi-Fi configuration
  async deleteWiFiConfig(id) {
    const response = await api.delete(`/wifi-config/${id}`)
    return response.data
  }
}

export default wifiConfigService