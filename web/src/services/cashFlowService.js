import api from './api'

const cashFlowService = {
  // Get all cash flow entries with optional filters
  async getCashFlowEntries(params = {}) {
    const response = await api.get('/cash-flow', { params })
    return response.data
  },

  // Create new cash flow entry
  async createCashFlow(cashFlowData) {
    const response = await api.post('/cash-flow', cashFlowData)
    return response.data
  },

  // Get cash flow summary for date range
  async getCashFlowSummary(startDate, endDate) {
    const response = await api.get('/cash-flow/summary', {
      params: {
        start_date: startDate,
        end_date: endDate
      }
    })
    return response.data
  },

  // Calculate running balance for category (client-side calculation)
  async getRunningBalance(category, upToDate) {
    try {
      // Get all entries for the category up to the specified date
      const response = await api.get('/cash-flow', {
        params: {
          category: category,
          end_date: upToDate
        }
      })
      
      const entries = response.data.cash_flows || []
      let balance = 0
      
      entries.forEach(entry => {
        if (entry.type === 'income') {
          balance += entry.amount
        } else if (entry.type === 'expense') {
          balance -= entry.amount
        }
      })
      
      return { balance }
    } catch (error) {
      console.error('Error calculating running balance:', error)
      return { balance: 0 }
    }
  },

  // Export cash flow report
  async exportCashFlowReport(startDate, endDate, format = 'excel') {
    const response = await api.post('/financial-reports/export', {
      start_date: startDate,
      end_date: endDate,
      format: format,
      include_cash_flow: true
    }, {
      responseType: 'blob'
    })
    return response
  },

  // Note: Individual CRUD operations (get by ID, update, delete) are not implemented 
  // in the backend API yet, so we'll remove them for now
  // These can be added when the backend handlers are implemented
}

export default cashFlowService