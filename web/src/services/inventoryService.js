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
  },

  // Initialize inventory for all ingredients
  initializeInventory() {
    return api.post('/inventory/initialize')
  },

  // Initialize inventory for a specific ingredient
  initializeInventoryItem(ingredientId) {
    return api.post(`/inventory/initialize/${ingredientId}`)
  }
}

export default inventoryService
