import api from './api'

const purchaseOrderService = {
  // Get all purchase orders with optional filters
  getPurchaseOrders(params = {}) {
    return api.get('/purchase-orders', { params })
  },

  // Get single purchase order by ID
  getPurchaseOrder(id) {
    return api.get(`/purchase-orders/${id}`)
  },

  // Create new purchase order
  createPurchaseOrder(data) {
    return api.post('/purchase-orders', data)
  },

  // Update existing purchase order
  updatePurchaseOrder(id, data) {
    return api.put(`/purchase-orders/${id}`, data)
  },

  // Delete purchase order
  deletePurchaseOrder(id) {
    return api.delete(`/purchase-orders/${id}`)
  },

  // Approve purchase order
  approvePurchaseOrder(id) {
    return api.post(`/purchase-orders/${id}/approve`)
  }
}

export default purchaseOrderService
