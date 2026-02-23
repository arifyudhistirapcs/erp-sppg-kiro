import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createWebHistory } from 'vue-router'
import { createPinia } from 'pinia'
import EmployeeListView from '../EmployeeListView.vue'
import employeeService from '@/services/employeeService'

// Mock the employee service
vi.mock('@/services/employeeService', () => ({
  default: {
    getEmployees: vi.fn(),
    getEmployeeStats: vi.fn(),
    deactivateEmployee: vi.fn(),
    updateEmployee: vi.fn()
  }
}))

// Mock ant-design-vue message
vi.mock('ant-design-vue', async () => {
  const actual = await vi.importActual('ant-design-vue')
  return {
    ...actual,
    message: {
      success: vi.fn(),
      error: vi.fn()
    }
  }
})

describe('EmployeeListView', () => {
  let wrapper
  let router
  let pinia

  beforeEach(() => {
    vi.clearAllMocks()
    
    // Setup router
    router = createRouter({
      history: createWebHistory(),
      routes: [
        { path: '/', component: { template: '<div>Home</div>' } },
        { path: '/employees', component: EmployeeListView }
      ]
    })
    
    // Setup pinia
    pinia = createPinia()
    
    // Mock service responses
    employeeService.getEmployees.mockResolvedValue({
      data: [
        {
          id: 1,
          nik: '1234567890123456',
          full_name: 'John Doe',
          email: 'john@example.com',
          position: 'Chef',
          is_active: true,
          join_date: '2024-01-01',
          user: { role: 'chef' }
        }
      ],
      total: 1
    })
    
    employeeService.getEmployeeStats.mockResolvedValue({
      data: {
        total_employees: 1,
        active_employees: 1,
        inactive_employees: 0,
        by_position: [{ position: 'Chef', count: 1 }]
      }
    })
  })

  const createWrapper = () => {
    return mount(EmployeeListView, {
      global: {
        plugins: [router, pinia],
        stubs: {
          'a-page-header': true,
          'a-card': true,
          'a-space': true,
          'a-row': true,
          'a-col': true,
          'a-input-search': true,
          'a-select': true,
          'a-select-option': true,
          'a-statistic': true,
          'a-table': true,
          'a-button': true,
          'a-modal': true,
          'a-form': true,
          'a-form-item': true,
          'a-input': true,
          'a-date-picker': true,
          'a-switch': true,
          'a-descriptions': true,
          'a-descriptions-item': true,
          'a-tag': true,
          'a-alert': true,
          'a-typography-text': true,
          'a-popconfirm': true,
          'PlusOutlined': true
        }
      }
    })
  }

  it('should render employee list view', async () => {
    wrapper = createWrapper()
    await wrapper.vm.$nextTick()
    
    expect(wrapper.exists()).toBe(true)
    expect(employeeService.getEmployees).toHaveBeenCalled()
    expect(employeeService.getEmployeeStats).toHaveBeenCalled()
  })

  it('should handle search functionality', async () => {
    wrapper = createWrapper()
    await wrapper.vm.$nextTick()
    
    // Set search text
    wrapper.vm.searchText = 'John'
    await wrapper.vm.handleSearch()
    
    expect(employeeService.getEmployees).toHaveBeenCalledWith({
      page: 1,
      page_size: 10,
      search: 'John',
      is_active: undefined,
      position: undefined
    })
  })

  it('should handle filter by status', async () => {
    wrapper = createWrapper()
    await wrapper.vm.$nextTick()
    
    // Set filter status
    wrapper.vm.filterStatus = 'active'
    await wrapper.vm.handleSearch()
    
    expect(employeeService.getEmployees).toHaveBeenCalledWith({
      page: 1,
      page_size: 10,
      search: undefined,
      is_active: true,
      position: undefined
    })
  })

  it('should handle filter by position', async () => {
    wrapper = createWrapper()
    await wrapper.vm.$nextTick()
    
    // Set filter position
    wrapper.vm.filterPosition = 'Chef'
    await wrapper.vm.handleSearch()
    
    expect(employeeService.getEmployees).toHaveBeenCalledWith({
      page: 1,
      page_size: 10,
      search: undefined,
      is_active: undefined,
      position: 'Chef'
    })
  })

  it('should format role labels correctly', async () => {
    wrapper = createWrapper()
    await wrapper.vm.$nextTick()
    
    expect(wrapper.vm.getRoleLabel('kepala_sppg')).toBe('Kepala SPPG/Yayasan')
    expect(wrapper.vm.getRoleLabel('chef')).toBe('Chef')
    expect(wrapper.vm.getRoleLabel('unknown')).toBe('unknown')
  })

  it('should format dates correctly', async () => {
    wrapper = createWrapper()
    await wrapper.vm.$nextTick()
    
    expect(wrapper.vm.formatDate('2024-01-01')).toBe('01/01/2024')
    expect(wrapper.vm.formatDate(null)).toBe('-')
  })
})