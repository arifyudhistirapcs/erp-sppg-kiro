import { describe, it, expect, vi, beforeEach } from 'vitest'
import employeeService from '../employeeService'
import api from '../api'

// Mock the api module
vi.mock('../api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn()
  }
}))

describe('employeeService', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getEmployees', () => {
    it('should fetch employees with default parameters', async () => {
      const mockResponse = {
        data: {
          data: [
            { id: 1, nik: '1234567890123456', full_name: 'John Doe', email: 'john@example.com' }
          ],
          total: 1
        }
      }
      
      api.get.mockResolvedValue(mockResponse)
      
      const result = await employeeService.getEmployees()
      
      expect(api.get).toHaveBeenCalledWith('/employees', { params: {} })
      expect(result).toEqual(mockResponse.data)
    })

    it('should fetch employees with search parameters', async () => {
      const params = {
        search: 'John',
        is_active: true,
        position: 'Chef'
      }
      
      const mockResponse = {
        data: {
          data: [],
          total: 0
        }
      }
      
      api.get.mockResolvedValue(mockResponse)
      
      await employeeService.getEmployees(params)
      
      expect(api.get).toHaveBeenCalledWith('/employees', { params })
    })
  })

  describe('getEmployeeById', () => {
    it('should fetch employee by ID', async () => {
      const mockResponse = {
        data: {
          data: { id: 1, nik: '1234567890123456', full_name: 'John Doe' }
        }
      }
      
      api.get.mockResolvedValue(mockResponse)
      
      const result = await employeeService.getEmployeeById(1)
      
      expect(api.get).toHaveBeenCalledWith('/employees/1')
      expect(result).toEqual(mockResponse.data)
    })
  })

  describe('createEmployee', () => {
    it('should create new employee', async () => {
      const employeeData = {
        nik: '1234567890123456',
        full_name: 'John Doe',
        email: 'john@example.com',
        position: 'Chef',
        role: 'chef'
      }
      
      const mockResponse = {
        data: {
          data: {
            user: { id: 1, nik: '1234567890123456', email: 'john@example.com' },
            credentials: { password: 'temp123' }
          }
        }
      }
      
      api.post.mockResolvedValue(mockResponse)
      
      const result = await employeeService.createEmployee(employeeData)
      
      expect(api.post).toHaveBeenCalledWith('/employees', employeeData)
      expect(result).toEqual(mockResponse.data)
    })
  })

  describe('updateEmployee', () => {
    it('should update employee', async () => {
      const employeeData = {
        full_name: 'John Smith',
        phone_number: '08123456789'
      }
      
      const mockResponse = {
        data: {
          data: { id: 1, full_name: 'John Smith' }
        }
      }
      
      api.put.mockResolvedValue(mockResponse)
      
      const result = await employeeService.updateEmployee(1, employeeData)
      
      expect(api.put).toHaveBeenCalledWith('/employees/1', employeeData)
      expect(result).toEqual(mockResponse.data)
    })
  })

  describe('deactivateEmployee', () => {
    it('should deactivate employee', async () => {
      const mockResponse = {
        data: { success: true, message: 'Employee deactivated' }
      }
      
      api.delete.mockResolvedValue(mockResponse)
      
      const result = await employeeService.deactivateEmployee(1)
      
      expect(api.delete).toHaveBeenCalledWith('/employees/1')
      expect(result).toEqual(mockResponse.data)
    })
  })

  describe('getEmployeeStats', () => {
    it('should fetch employee statistics', async () => {
      const mockResponse = {
        data: {
          data: {
            total_employees: 10,
            active_employees: 8,
            inactive_employees: 2,
            by_position: [
              { position: 'Chef', count: 3 },
              { position: 'Driver', count: 2 }
            ]
          }
        }
      }
      
      api.get.mockResolvedValue(mockResponse)
      
      const result = await employeeService.getEmployeeStats()
      
      expect(api.get).toHaveBeenCalledWith('/employees/stats')
      expect(result).toEqual(mockResponse.data)
    })
  })
})