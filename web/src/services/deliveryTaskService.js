import api from './api'

const deliveryTaskService = {
  // Get all delivery tasks with optional filters
  getDeliveryTasks(params = {}) {
    return api.get('/delivery-tasks', { params })
  },

  // Get single delivery task by ID
  getDeliveryTask(id) {
    return api.get(`/delivery-tasks/${id}`)
  },

  // Create new delivery task
  createDeliveryTask(data) {
    return api.post('/delivery-tasks', data)
  },

  // Update existing delivery task
  updateDeliveryTask(id, data) {
    return api.put(`/delivery-tasks/${id}`, data)
  },

  // Update delivery task status
  updateDeliveryTaskStatus(id, status) {
    return api.put(`/delivery-tasks/${id}/status`, { status })
  },

  // Get delivery tasks for a specific driver today
  getDriverTasksToday(driverId) {
    return api.get(`/delivery-tasks/driver/${driverId}/today`)
  },

  // Delete delivery task
  deleteDeliveryTask(id) {
    return api.delete(`/delivery-tasks/${id}`)
  },

  // Get all drivers (users with driver role)
  getDrivers() {
    return api.get('/employees', { 
      params: { 
        role: 'driver',
        is_active: true 
      } 
    })
  },

  // Get available recipes for delivery
  getAvailableRecipes() {
    return api.get('/recipes', { 
      params: { 
        is_active: true 
      } 
    })
  }
}

export default deliveryTaskService