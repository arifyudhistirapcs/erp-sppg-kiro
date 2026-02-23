import api from './api'

const supplierService = {
  // Get all suppliers with optional filters
  getSuppliers(params = {}) {
    return api.get('/suppliers', { params })
  },

  // Get single supplier by ID
  getSupplier(id) {
    return api.get(`/suppliers/${id}`)
  },

  // Create new supplier
  createSupplier(data) {
    return api.post('/suppliers', data)
  },

  // Update existing supplier
  updateSupplier(id, data) {
    return api.put(`/suppliers/${id}`, data)
  },

  // Delete supplier
  deleteSupplier(id) {
    return api.delete(`/suppliers/${id}`)
  },

  // Get supplier performance metrics
  getSupplierPerformance(id) {
    return api.get(`/suppliers/${id}/performance`)
  }
}

export default supplierService
