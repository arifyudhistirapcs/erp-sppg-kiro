import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { message } from 'ant-design-vue'
import PickupTaskList from '../PickupTaskList.vue'
import pickupTaskService from '@/services/pickupTaskService'

// Mock the service
vi.mock('@/services/pickupTaskService', () => ({
  default: {
    getPickupTasks: vi.fn(),
    getPickupTask: vi.fn()
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

describe('PickupTaskList - Table Display', () => {
  let wrapper

  const mockPickupTasks = [
    {
      id: 1,
      task_date: '2024-01-15T00:00:00Z',
      driver_id: 10,
      status: 'active',
      school_count: 3,
      completed_count: 1,
      created_at: '2024-01-15T08:00:00Z',
      driver: {
        id: 10,
        full_name: 'Ahmad Supardi',
        phone_number: '081234567890'
      }
    },
    {
      id: 2,
      task_date: '2024-01-15T00:00:00Z',
      driver_id: 11,
      status: 'active',
      school_count: 2,
      completed_count: 0,
      created_at: '2024-01-15T09:00:00Z',
      driver: {
        id: 11,
        full_name: 'Budi Santoso',
        phone_number: '081234567891'
      }
    }
  ]

  const mockDeliveryRecords = [
    {
      id: 123,
      school_id: 45,
      route_order: 1,
      current_stage: 11,
      current_status: 'driver_tiba_di_lokasi_pengambilan',
      ompreng_count: 15,
      school: {
        id: 45,
        name: 'SD Negeri 1',
        address: 'Jl. Pendidikan No. 1',
        latitude: -6.2088,
        longitude: 106.8456
      }
    },
    {
      id: 124,
      school_id: 46,
      route_order: 2,
      current_stage: 10,
      current_status: 'driver_menuju_lokasi_pengambilan',
      ompreng_count: 20,
      school: {
        id: 46,
        name: 'SD Negeri 2',
        address: 'Jl. Pendidikan No. 2',
        latitude: -6.2089,
        longitude: 106.8457
      }
    }
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    pickupTaskService.getPickupTasks.mockResolvedValue({
      data: { pickup_tasks: mockPickupTasks }
    })
    pickupTaskService.getPickupTask.mockResolvedValue({
      data: {
        pickup_task: {
          ...mockPickupTasks[0],
          delivery_records: mockDeliveryRecords
        }
      }
    })
  })

  it('should display pickup tasks in table', async () => {
    wrapper = mount(PickupTaskList)
    await wrapper.vm.$nextTick()
    await new Promise(resolve => setTimeout(resolve, 100))

    expect(wrapper.vm.pickupTasks.length).toBe(2)
  })

  it('should display driver information', async () => {
    wrapper = mount(PickupTaskList)
    await wrapper.vm.$nextTick()
    await new Promise(resolve => setTimeout(resolve, 100))

    const task = wrapper.vm.pickupTasks[0]
    expect(task.driver.full_name).toBe('Ahmad Supardi')
    expect(task.driver.phone_number).toBe('081234567890')
  })

  it('should display school count and progress', async () => {
    wrapper = mount(PickupTaskList)
    await wrapper.vm.$nextTick()
    await new Promise(resolve => setTimeout(resolve, 100))

    const task = wrapper.vm.pickupTasks[0]
    expect(task.school_count).toBe(3)
    expect(task.completed_count).toBe(1)
  })

  it('should calculate progress percentage correctly', async () => {
    wrapper = mount(PickupTaskList)
    await wrapper.vm.$nextTick()

    const progress = wrapper.vm.calculateProgress(mockPickupTasks[0])
    expect(progress).toBe(33) // 1/3 = 33%
  })
})

describe('PickupTaskList - Expandable Rows', () => {
  let wrapper

  const mockTaskWithDetails = {
    id: 1,
    status: 'active',
    school_count: 2,
    completed_count: 0,
    driver: { full_name: 'Driver 1' },
    delivery_records: [
      {
        id: 1,
        route_order: 1,
        current_stage: 10,
        school: {
          name: 'SD Negeri 1',
          address: 'Jl. Test 1',
          latitude: -6.2088,
          longitude: 106.8456
        },
        ompreng_count: 15
      },
      {
        id: 2,
        route_order: 2,
        current_stage: 11,
        school: {
          name: 'SD Negeri 2',
          address: 'Jl. Test 2',
          latitude: -6.2089,
          longitude: 106.8457
        },
        ompreng_count: 20
      }
    ]
  }

  beforeEach(() => {
    vi.clearAllMocks()
    pickupTaskService.getPickupTasks.mockResolvedValue({
      data: { pickup_tasks: [mockTaskWithDetails] }
    })
    pickupTaskService.getPickupTask.mockResolvedValue({
      data: { pickup_task: mockTaskWithDetails }
    })
  })

  it('should display delivery records when row is expanded', async () => {
    wrapper = mount(PickupTaskList)
    await wrapper.vm.$nextTick()
    await new Promise(resolve => setTimeout(resolve, 100))

    const task = wrapper.vm.pickupTasks[0]
    expect(task.delivery_records).toBeDefined()
    expect(task.delivery_records.length).toBe(2)
  })

  it('should display schools in route order', async () => {
    wrapper = mount(PickupTaskList)
    await wrapper.vm.$nextTick()
    await new Promise(resolve => setTimeout(resolve, 100))

    const records = wrapper.vm.pickupTasks[0].delivery_records
    expect(records[0].route_order).toBe(1)
    expect(records[1].route_order).toBe(2)
  })

  it('should display school information in expanded view', async () => {
    wrapper = mount(PickupTaskList)
    await wrapper.vm.$nextTick()
    await new Promise(resolve => setTimeout(resolve, 100))

    const record = wrapper.vm.pickupTasks[0].delivery_records[0]
    expect(record.school.name).toBe('SD Negeri 1')
    expect(record.school.address).toBe('Jl. Test 1')
    expect(record.school.latitude).toBe(-6.2088)
    expect(record.school.longitude).toBe(106.8456)
  })
})

describe('PickupTaskList - Stage Indicators', () => {
  let wrapper

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should display correct stage color for stage 10', async () => {
    wrapper = mount(PickupTaskList)
    const color = wrapper.vm.getStageColor(10)
    expect(color).toBe('blue')
  })

  it('should display correct stage color for stage 11', async () => {
    wrapper = mount(PickupTaskList)
    const color = wrapper.vm.getStageColor(11)
    expect(color).toBe('orange')
  })

  it('should display correct stage color for stage 12', async () => {
    wrapper = mount(PickupTaskList)
    const color = wrapper.vm.getStageColor(12)
    expect(color).toBe('purple')
  })

  it('should display correct stage color for stage 13', async () => {
    wrapper = mount(PickupTaskList)
    const color = wrapper.vm.getStageColor(13)
    expect(color).toBe('green')
  })

  it('should display correct stage text in Indonesian', async () => {
    wrapper = mount(PickupTaskList)
    
    expect(wrapper.vm.getStageText(10)).toBe('Menuju Lokasi')
    expect(wrapper.vm.getStageText(11)).toBe('Tiba di Sekolah')
    expect(wrapper.vm.getStageText(12)).toBe('Kembali ke SPPG')
    expect(wrapper.vm.getStageText(13)).toBe('Tiba di SPPG')
  })

  it('should display correct status color', async () => {
    wrapper = mount(PickupTaskList)
    
    expect(wrapper.vm.getStatusColor('active')).toBe('blue')
    expect(wrapper.vm.getStatusColor('completed')).toBe('green')
    expect(wrapper.vm.getStatusColor('cancelled')).toBe('red')
  })

  it('should display correct status text in Indonesian', async () => {
    wrapper = mount(PickupTaskList)
    
    expect(wrapper.vm.getStatusText('active')).toBe('Aktif')
    expect(wrapper.vm.getStatusText('completed')).toBe('Selesai')
    expect(wrapper.vm.getStatusText('cancelled')).toBe('Dibatalkan')
  })
})

describe('PickupTaskList - Filtering and Refresh', () => {
  let wrapper

  beforeEach(() => {
    vi.clearAllMocks()
    pickupTaskService.getPickupTasks.mockResolvedValue({
      data: { pickup_tasks: [] }
    })
  })

  it('should filter by date when date prop is provided', async () => {
    wrapper = mount(PickupTaskList, {
      props: {
        date: '2024-01-15'
      }
    })
    await wrapper.vm.$nextTick()
    await new Promise(resolve => setTimeout(resolve, 100))

    expect(pickupTaskService.getPickupTasks).toHaveBeenCalledWith(
      expect.objectContaining({
        status: 'active',
        date: '2024-01-15'
      })
    )
  })

  it('should refresh data when refresh method is called', async () => {
    wrapper = mount(PickupTaskList)
    await wrapper.vm.$nextTick()
    
    vi.clearAllMocks()
    await wrapper.vm.refresh()
    
    expect(pickupTaskService.getPickupTasks).toHaveBeenCalled()
  })
})
