import api from './api'

const menuPlanningService = {
  // Get all menu plans
  getMenuPlans(params = {}) {
    return api.get('/menu-plans', { params })
  },

  // Get single menu plan by ID
  getMenuPlan(id) {
    return api.get(`/menu-plans/${id}`)
  },

  // Get current week menu plan
  getCurrentWeekMenu() {
    return api.get('/menu-plans/current-week')
  },

  // Create new menu plan
  createMenuPlan(data) {
    return api.post('/menu-plans', data)
  },

  // Update existing menu plan
  updateMenuPlan(id, data) {
    return api.put(`/menu-plans/${id}`, data)
  },

  // Approve menu plan
  approveMenuPlan(id) {
    return api.post(`/menu-plans/${id}/approve`)
  },

  // Delete menu plan
  deleteMenuPlan(id) {
    return api.delete(`/menu-plans/${id}`)
  },

  // Create menu item with school allocations
  createMenuItem(menuPlanId, data) {
    return api.post(`/menu-plans/${menuPlanId}/items`, data)
  },

  // Update menu item with school allocations
  updateMenuItem(menuPlanId, itemId, data) {
    return api.put(`/menu-plans/${menuPlanId}/items/${itemId}`, data)
  },

  // Get menu item with school allocations
  getMenuItem(menuPlanId, itemId) {
    return api.get(`/menu-plans/${menuPlanId}/items/${itemId}`)
  },

  // Delete menu item
  deleteMenuItem(menuPlanId, itemId) {
    return api.delete(`/menu-plans/${menuPlanId}/items/${itemId}`)
  }
}

export default menuPlanningService
