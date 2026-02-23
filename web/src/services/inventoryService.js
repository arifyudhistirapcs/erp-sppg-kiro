import api from './api'

const inventoryService = {
  // Get all inventory items with optional filters
  getInventory(params = {}) {
    return api.get('/inventory', { params })
  },

  // Get single inventory item by ID
  getInventoryItem(id) {
    return api.get(`/inventory/${id}`)
  },

  // Get low stock alerts
  getLowStockAlerts() {
    return api.get('/inventory/alerts')
  },

  // Get inventory movements with filters
  getInventoryMovements(params = {}) {
    return api.get('/inventory/movements', { params })
  }
}

export default inventoryService
