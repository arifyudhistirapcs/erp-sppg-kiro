import api from './api'

const employeeService = {
  // Get all employees with optional filters
  async getEmployees(params = {}) {
    const response = await api.get('/employees', { params })
    return response.data
  },

  // Get employee by ID
  async getEmployeeById(id) {
    const response = await api.get(`/employees/${id}`)
    return response.data
  },

  // Create new employee
  async createEmployee(employeeData) {
    const response = await api.post('/employees', employeeData)
    return response.data
  },

  // Update employee
  async updateEmployee(id, employeeData) {
    const response = await api.put(`/employees/${id}`, employeeData)
    return response.data
  },

  // Deactivate employee
  async deactivateEmployee(id) {
    const response = await api.delete(`/employees/${id}`)
    return response.data
  },

  // Get employee statistics
  async getEmployeeStats() {
    const response = await api.get('/employees/stats')
    return response.data
  }
}

export default employeeService