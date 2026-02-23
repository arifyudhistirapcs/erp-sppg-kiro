import api from './api'

const auditService = {
  // Get audit trail entries with filters
  async getAuditTrail(params = {}) {
    const response = await api.get('/audit-trail', { params })
    return response.data
  },

  // Get audit trail statistics
  async getAuditStats(params = {}) {
    const response = await api.get('/audit-trail/stats', { params })
    return response.data
  }
}

export default auditService