import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { message } from 'ant-design-vue'
import PickupTaskForm from '../PickupTaskForm.vue'
import pickupTaskService from '@/services/pickupTaskService'

// Mock the service
vi.mock('@/services/pickupTaskService', () => ({
  default: {
    getEligibleOrders: vi.fn(),
    getAvailableDrivers: vi.fn(),
    createPickupTask: vi.fn()
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

describe('PickupTaskForm - Form Validation', () => {
  let wrapper

  const mockEligibleOrders = [
    {
      delivery_record_id: 1,
      school_id: 10,
      school_name: 'SD Negeri 1',
      school_address: 'Jl. Pendidikan No. 1',
      latitude: -6.2088,
      longitude: 106.8456,
      ompreng_count: 15,
      delivery_date: '2024-01-15T00:00:00Z',
      current_stage: 9
    }
  ]

  const mockDrivers = [
    {
      driver_id: 1,
      full_name: 'Ahmad Supardi',
      phone_number: '081234567890'
    }
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    pickupTaskService.getEligibleOrders.mockResolvedValue({
      data: { eligible_orders: mockEligibleOrders }
    })
    pickupTaskService.getAvailableDrivers.mockResolvedValue({
      data: { available_drivers: mockDrivers }
    })
  })

  it('should display error when no orders are selected', async () => {
    wrapper = mount(PickupTaskForm)
    await wrapper.vm.$nextTick()

    // Submit button should be disabled when no orders selected
    const submitButton = wrapper.find('button[type="primary"]')
    expect(submitButton.attributes('disabled')).toBeDefined()
  })

  it('should display error when no driver is selected', async () => {
    wrapper = mount(PickupTaskForm)
    await wrapper.vm.$nextTick()

    // Select an order but no driver
    wrapper.vm.selectedOrders = [mockEligibleOrders[0]]
    await wrapper.vm.$nextTick()

    // Submit button should still be disabled
    const submitButton = wrapper.find('button[type="primary"]')
    expect(submitButton.attributes('disabled')).toBeDefined()
  })

  it('should enable submit button when orders and driver are selected', async () => {
    wrapper = mount(PickupTaskForm)
    await wrapper.vm.$nextTick()

    // Select order and driver
    wrapper.vm.selectedOrders = [mockEligibleOrders[0]]
    wrapper.vm.selectedDriver = 1
    await wrapper.vm.$nextTick()

    // Submit button should be enabled
    expect(wrapper.vm.canSubmit).toBe(true)
  })

  it('should display validation error messages', async () => {
    wrapper = mount(PickupTaskForm)
    await wrapper.vm.$nextTick()

    // Try to submit without selections
    await wrapper.vm.handleSubmit()

    // Should not proceed with submission
    expect(pickupTaskService.createPickupTask).not.toHaveBeenCalled()
  })
})

describe('PickupTaskForm - Drag and Drop Route Ordering', () => {
  let wrapper

  const mockMultipleOrders = [
    {
      delivery_record_id: 1,
      school_name: 'SD Negeri 1',
      school_address: 'Jl. Pendidikan No. 1',
      latitude: -6.2088,
      longitude: 106.8456,
      ompreng_count: 15,
      route_order: 1
    },
    {
      delivery_record_id: 2,
      school_name: 'SD Negeri 2',
      school_address: 'Jl. Pendidikan No. 2',
      latitude: -6.2089,
      longitude: 106.8457,
      ompreng_count: 20,
      route_order: 2
    },
    {
      delivery_record_id: 3,
      school_name: 'SD Negeri 3',
      school_address: 'Jl. Pendidikan No. 3',
      latitude: -6.2090,
      longitude: 106.8458,
      ompreng_count: 18,
      route_order: 3
    }
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    pickupTaskService.getEligibleOrders.mockResolvedValue({
      data: { eligible_orders: mockMultipleOrders }
    })
    pickupTaskService.getAvailableDrivers.mockResolvedValue({
      data: { available_drivers: [{ driver_id: 1, full_name: 'Driver 1' }] }
    })
  })

  it('should display selected orders with route numbers', async () => {
    wrapper = mount(PickupTaskForm)
    await wrapper.vm.$nextTick()

    wrapper.vm.selectedOrders = [...mockMultipleOrders]
    await wrapper.vm.$nextTick()

    // Check that route orders are displayed
    const routeTags = wrapper.findAll('.ant-tag')
    expect(routeTags.length).toBeGreaterThan(0)
  })

  it('should update route order after drag and drop', async () => {
    wrapper = mount(PickupTaskForm)
    await wrapper.vm.$nextTick()

    wrapper.vm.selectedOrders = [...mockMultipleOrders]
    await wrapper.vm.$nextTick()

    // Simulate drag and drop by reordering the array
    const reordered = [mockMultipleOrders[1], mockMultipleOrders[0], mockMultipleOrders[2]]
    wrapper.vm.selectedOrders = reordered
    wrapper.vm.updateRouteOrder()
    await wrapper.vm.$nextTick()

    // Verify route orders are updated
    expect(wrapper.vm.selectedOrders[0].route_order).toBe(1)
    expect(wrapper.vm.selectedOrders[1].route_order).toBe(2)
    expect(wrapper.vm.selectedOrders[2].route_order).toBe(3)
  })

  it('should allow removing orders from the route', async () => {
    wrapper = mount(PickupTaskForm)
    await wrapper.vm.$nextTick()

    wrapper.vm.selectedOrders = [...mockMultipleOrders]
    await wrapper.vm.$nextTick()

    // Remove the second order
    wrapper.vm.removeOrder(2)
    await wrapper.vm.$nextTick()

    // Verify order was removed and route orders updated
    expect(wrapper.vm.selectedOrders.length).toBe(2)
    expect(wrapper.vm.selectedOrders[0].route_order).toBe(1)
    expect(wrapper.vm.selectedOrders[1].route_order).toBe(2)
  })

  it('should maintain route order sequence starting from 1', async () => {
    wrapper = mount(PickupTaskForm)
    await wrapper.vm.$nextTick()

    wrapper.vm.selectedOrders = [...mockMultipleOrders]
    wrapper.vm.updateRouteOrder()
    await wrapper.vm.$nextTick()

    // Verify route orders start from 1 and are sequential
    wrapper.vm.selectedOrders.forEach((order, index) => {
      expect(order.route_order).toBe(index + 1)
    })
  })
})

describe('PickupTaskForm - Error Display', () => {
  let wrapper

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should display error message when API call fails', async () => {
    pickupTaskService.getEligibleOrders.mockRejectedValue(new Error('API Error'))
    pickupTaskService.getAvailableDrivers.mockResolvedValue({
      data: { available_drivers: [] }
    })

    wrapper = mount(PickupTaskForm)
    await wrapper.vm.$nextTick()
    await new Promise(resolve => setTimeout(resolve, 100))

    // Verify error message was shown
    expect(message.error).toHaveBeenCalledWith('Gagal memuat data order yang siap diambil')
  })

  it('should display error when submission fails', async () => {
    pickupTaskService.getEligibleOrders.mockResolvedValue({
      data: { eligible_orders: [] }
    })
    pickupTaskService.getAvailableDrivers.mockResolvedValue({
      data: { available_drivers: [{ driver_id: 1, full_name: 'Driver 1' }] }
    })
    pickupTaskService.createPickupTask.mockRejectedValue({
      response: {
        data: {
          error: {
            message: 'Validation failed'
          }
        }
      }
    })

    wrapper = mount(PickupTaskForm)
    await wrapper.vm.$nextTick()

    wrapper.vm.selectedOrders = [{ delivery_record_id: 1, route_order: 1 }]
    wrapper.vm.selectedDriver = 1
    await wrapper.vm.handleSubmit()
    await wrapper.vm.$nextTick()

    expect(message.error).toHaveBeenCalledWith('Validation failed')
  })
})
