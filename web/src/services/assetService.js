import api from './api'

const assetService = {
  // Get all assets with optional filters
  async getAssets(params = {}) {
    const response = await api.get('/assets', { params })
    return response.data
  },

  // Get asset by ID
  async getAssetById(id) {
    const response = await api.get(`/assets/${id}`)
    return response.data
  },

  // Create new asset
  async createAsset(assetData) {
    const response = await api.post('/assets', assetData)
    return response.data
  },

  // Update asset
  async updateAsset(id, assetData) {
    const response = await api.put(`/assets/${id}`, assetData)
    return response.data
  },

  // Delete asset
  async deleteAsset(id) {
    const response = await api.delete(`/assets/${id}`)
    return response.data
  },

  // Add maintenance record
  async addMaintenanceRecord(assetId, maintenanceData) {
    const response = await api.post(`/assets/${assetId}/maintenance`, maintenanceData)
    return response.data
  },

  // Get asset report
  async getAssetReport() {
    const response = await api.get('/assets/report')
    return response.data
  },

  // Get depreciation schedule
  async getDepreciationSchedule(assetId, years = 5) {
    const response = await api.get(`/assets/${assetId}/depreciation-schedule`, {
      params: { years }
    })
    return response.data
  },

  // Export asset report
  async exportAssetReport(format = 'excel') {
    const response = await api.post('/financial-reports/export', {
      start_date: '2020-01-01',
      end_date: new Date().toISOString().split('T')[0],
      format: format,
      include_assets: true
    }, {
      responseType: 'blob'
    })
    return response
  }
}

export default assetService