import api from './api'

const financialReportService = {
  // Generate financial report
  async generateFinancialReport(startDate, endDate, options = {}) {
    const params = {
      start_date: startDate,
      end_date: endDate,
      include_budget: options.includeBudget || false,
      include_assets: options.includeAssets || false,
      include_trend: options.includeTrend || false
    }
    
    const response = await api.get('/financial-reports', { params })
    return response.data
  },

  // Export financial report
  async exportFinancialReport(startDate, endDate, format = 'excel', options = {}) {
    const response = await api.post('/financial-reports/export', {
      start_date: startDate,
      end_date: endDate,
      format: format,
      include_budget: options.includeBudget || false,
      include_assets: options.includeAssets || false,
      include_trend: options.includeTrend || false,
      include_charts: options.includeCharts || false
    }, {
      responseType: 'blob'
    })
    return response
  },

  // Get daily cash flow
  async getDailyCashFlow(date) {
    const response = await api.get('/financial-reports/daily', {
      params: { date }
    })
    return response.data
  },

  // Get weekly cash flow
  async getWeeklyCashFlow(year, week) {
    const response = await api.get('/financial-reports/weekly', {
      params: { year, week }
    })
    return response.data
  },

  // Get quarterly cash flow
  async getQuarterlyCashFlow(year, quarter) {
    const response = await api.get('/financial-reports/quarterly', {
      params: { year, quarter }
    })
    return response.data
  },

  // Get custom period report
  async getCustomPeriodReport(startDate, endDate) {
    const response = await api.get('/financial-reports/custom', {
      params: {
        start_date: startDate,
        end_date: endDate
      }
    })
    return response.data
  }
}

export default financialReportService