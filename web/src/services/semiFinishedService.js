import api from './api'

const semiFinishedService = {
  // Get all semi-finished goods
  getAllSemiFinishedGoods(params = {}) {
    return api.get('/semi-finished', { params })
  },

  // Get single semi-finished goods by ID
  getSemiFinishedGoods(id) {
    return api.get(`/semi-finished/${id}`)
  },

  // Create new semi-finished goods with recipe
  createSemiFinishedGoods(data) {
    return api.post('/semi-finished', data)
  },

  // Update semi-finished goods
  updateSemiFinishedGoods(id, data) {
    return api.put(`/semi-finished/${id}`, data)
  },

  // Delete semi-finished goods
  deleteSemiFinishedGoods(id) {
    return api.delete(`/semi-finished/${id}`)
  },

  // Produce semi-finished goods (deduct raw ingredients, add stock)
  produceSemiFinishedGoods(id, quantity, notes = '') {
    return api.post(`/semi-finished/${id}/produce`, {
      quantity,
      notes
    })
  },

  // Get semi-finished goods inventory
  getSemiFinishedInventory() {
    return api.get('/semi-finished/inventory')
  }
}

export default semiFinishedService
