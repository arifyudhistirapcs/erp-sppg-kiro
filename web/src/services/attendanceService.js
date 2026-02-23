import api from './api'

const attendanceService = {
  // Get attendance report with date range and optional employee filter
  async getAttendanceReport(params = {}) {
    const response = await api.get('/attendance/report', { params })
    return response.data
  },

  // Get attendance by date range for specific employee
  async getAttendanceByDateRange(employeeId, startDate, endDate) {
    const params = {
      employee_id: employeeId,
      start_date: startDate,
      end_date: endDate
    }
    const response = await api.get('/attendance/by-date-range', { params })
    return response.data
  },

  // Get attendance statistics for a specific date
  async getAttendanceStats(date) {
    const response = await api.get('/attendance/stats', { 
      params: { date } 
    })
    return response.data
  },

  // Export attendance report to Excel
  async exportToExcel(params = {}) {
    const response = await api.get('/attendance/export/excel', { 
      params,
      responseType: 'blob'
    })
    return response
  },

  // Export attendance report to PDF
  async exportToPDF(params = {}) {
    const response = await api.get('/attendance/export/pdf', { 
      params,
      responseType: 'blob'
    })
    return response
  }
}

export default attendanceService