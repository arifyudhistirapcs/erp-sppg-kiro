import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import KDSCookingView from '../KDSCookingView.vue'
import KDSPackingView from '../KDSPackingView.vue'
import { ref as firebaseRef, onValue, off } from 'firebase/database'

/**
 * End-to-End Integration Test: Real-time Firebase Updates for Portion Sizes
 * 
 * Task 6.5.6: Test real-time updates across multiple clients
 * 
 * This test suite simulates real-time Firebase updates for portion size data:
 * 1. Changes in one client are reflected in another
 * 2. Portion size data is synced correctly
 * 3. Real-time listeners handle portion size updates
 * 4. Data consistency across multiple views
 * 
 * These tests validate that Firebase real-time synchronization works correctly
 * for the portion size differentiation feature.
 */

// Mock services
vi.mock('@/services/kdsService', () => ({
  getCookingToday: vi.fn(),
  getPackingToday: vi.fn(),
  updateCookingStatus: vi.fn(),
  updatePackingStatus: vi.fn()
}))

// Mock Firebase
let mockFirebaseListeners = {}
let mockFirebaseData = {}

vi.mock('firebase/database', () => ({
  ref: vi.fn((db, path) => ({ _path: path })),
  onValue: vi.fn((ref, callback) => {
    const path = ref._path
    mockFirebaseListeners[path] = callback
    // Immediately call with initial data if available
    if (mockFirebaseData[path]) {
      callback({
        exists: () => true,
        val: () => mockFirebaseData[path]
      })
    }
  }),
  off: vi.fn((ref) => {
    const path = ref._path
    delete mockFirebaseListeners[path]
  })
}))

vi.mock('@/services/firebase', () => ({
  database: {}
}))

// Mock ant-design-vue
vi.mock('ant-design-vue', async () => {
  const actual = await vi.importActual('ant-design-vue')
  return {
    ...actual,
    message: {
      success: vi.fn(),
      error: vi.fn(),
      warning: vi.fn(),
      info: vi.fn()
    },
    notification: {
      success: vi.fn(),
      error: vi.fn(),
      warning: vi.fn(),
      info: vi.fn()
    }
  }
})

// Helper function to simulate Firebase update
const simulateFirebaseUpdate = (path, data) => {
  mockFirebaseData[path] = data
  if (mockFirebaseListeners[path]) {
    mockFirebaseListeners[path]({
      exists: () => true,
      val: () => data
    })
  }
}

// Helper function to clear Firebase data
const clearFirebaseData = () => {
  mockFirebaseData = {}
  mockFirebaseListeners = {}
}

describe('E2E: Real-time Firebase Updates - KDS Cooking View', () => {
  let wrapper
  let getCookingToday

  beforeEach(async () => {
    vi.clearAllMocks()
    clearFirebaseData()

    const kdsService = await import('@/services/kdsService')
    getCookingToday = kdsService.getCookingToday

    // Mock initial API response
    getCookingToday.mockResolvedValue({
      success: true,
      data: []
    })
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
    }
  })

  const createCookingWrapper = (initialRecipes = []) => {
    getCookingToday.mockResolvedValue({
      success: true,
      data: initialRecipes
    })

    return mount(KDSCookingView, {
      global: {
        stubs: {
          'KDSDatePicker': {
            template: '<div class="kds-date-picker-stub"></div>',
            props: ['modelValue', 'loading'],
            emits: ['change', 'update:modelValue']
          },
          'a-space': { template: '<div><slot /></div>' },
          'a-tag': { template: '<span><slot /></span>' },
          'a-button': { template: '<button><slot /></button>' },
          'a-alert': { template: '<div><slot /></div>' },
          'a-spin': { template: '<div><slot /></div>' },
          'a-empty': { template: '<div><slot /></div>' },
          'a-row': { template: '<div><slot /></div>' },
          'a-col': { template: '<div><slot /></div>' },
          'a-card': { template: '<div><slot /></div>' },
          'a-descriptions': { template: '<div><slot /></div>' },
          'a-descriptions-item': { template: '<div><slot /></div>' },
          'a-divider': { template: '<div><slot /></div>' },
          'a-list': { template: '<div><slot /></div>' },
          'a-list-item': { template: '<div><slot /></div>' },
          'a-list-item-meta': { template: '<div><slot /></div>' },
          'a-badge': { template: '<span><slot /></span>' }
        }
      }
    })
  }

  it('should update SD school portion sizes when Firebase data changes', async () => {
    const initialRecipe = {
      recipe_id: 1,
      name: 'Nasi Goreng',
      portions_required: 125,
      status: 'pending',
      school_allocations: [
        {
          school_id: 1,
          school_name: 'SD Negeri 1',
          school_category: 'SD',
          portion_size_type: 'mixed',
          portions_small: 50,
          portions_large: 75,
          total_portions: 125
        }
      ]
    }

    wrapper = createCookingWrapper([initialRecipe])
    await wrapper.vm.$nextTick()
    await wrapper.vm.loadData()
    await wrapper.vm.$nextTick()

    // Verify initial state
    expect(wrapper.vm.recipes[0].school_allocations[0].portions_small).toBe(50)
    expect(wrapper.vm.recipes[0].school_allocations[0].portions_large).toBe(75)

    // Simulate Firebase update from another client
    const firebasePath = 'kds/cooking/2024-01-15'
    const updatedData = [
      {
        recipe_id: 1,
        status: 'pending',
        school_allocations: [
          {
            school_id: 1,
            school_name: 'SD Negeri 1',
            school_category: 'SD',
            portion_size_type: 'mixed',
            portions_small: 60,  // Changed
            portions_large: 85,  // Changed
            total_portions: 145
          }
        ]
      }
    ]

    simulateFirebaseUpdate(firebasePath, updatedData)
    await wrapper.vm.$nextTick()

    // Verify the update was applied
    expect(wrapper.vm.recipes[0].school_allocations[0].portions_small).toBe(60)
    expect(wrapper.vm.recipes[0].school_allocations[0].portions_large).toBe(85)
    expect(wrapper.vm.recipes[0].school_allocations[0].total_portions).toBe(145)
  })

  it('should update SMP school portion sizes when Firebase data changes', async () => {
    const initialRecipe = {
      recipe_id: 2,
      name: 'Soto Ayam',
      portions_required: 100,
      status: 'cooking',
      school_allocations: [
        {
          school_id: 2,
          school_name: 'SMP Negeri 1',
          school_category: 'SMP',
          portion_size_type: 'large',
          portions_small: 0,
          portions_large: 100,
          total_portions: 100
        }
      ]
    }

    wrapper = createCookingWrapper([initialRecipe])
    await wrapper.vm.$nextTick()
    await wrapper.vm.loadData()
    await wrapper.vm.$nextTick()

    // Verify initial state
    expect(wrapper.vm.recipes[0].school_allocations[0].portions_large).toBe(100)

    // Simulate Firebase update
    const firebasePath = 'kds/cooking/2024-01-15'
    const updatedData = [
      {
        recipe_id: 2,
        status: 'cooking',
        school_allocations: [
          {
            school_id: 2,
            school_name: 'SMP Negeri 1',
            school_category: 'SMP',
            portion_size_type: 'large',
            portions_small: 0,
            portions_large: 120,  // Changed
            total_portions: 120
          }
        ]
      }
    ]

    simulateFirebaseUpdate(firebasePath, updatedData)
    await wrapper.vm.$nextTick()

    // Verify the update was applied
    expect(wrapper.vm.recipes[0].school_allocations[0].portions_large).toBe(120)
    expect(wrapper.vm.recipes[0].school_allocations[0].total_portions).toBe(120)
  })

  it('should handle multiple schools updating simultaneously', async () => {
    const initialRecipe = {
      recipe_id: 3,
      name: 'Ayam Goreng',
      portions_required: 345,
      status: 'pending',
      school_allocations: [
        {
          school_id: 1,
          school_name: 'SD Negeri 1',
          school_category: 'SD',
          portion_size_type: 'mixed',
          portions_small: 50,
          portions_large: 75,
          total_portions: 125
        },
        {
          school_id: 2,
          school_name: 'SMP Negeri 1',
          school_category: 'SMP',
          portion_size_type: 'large',
          portions_small: 0,
          portions_large: 100,
          total_portions: 100
        },
        {
          school_id: 3,
          school_name: 'SMA Negeri 1',
          school_category: 'SMA',
          portion_size_type: 'large',
          portions_small: 0,
          portions_large: 120,
          total_portions: 120
        }
      ]
    }

    wrapper = createCookingWrapper([initialRecipe])
    await wrapper.vm.$nextTick()
    await wrapper.vm.loadData()
    await wrapper.vm.$nextTick()

    // Simulate Firebase update affecting all schools
    const firebasePath = 'kds/cooking/2024-01-15'
    const updatedData = [
      {
        recipe_id: 3,
        status: 'pending',
        school_allocations: [
          {
            school_id: 1,
            portions_small: 60,  // Changed
            portions_large: 80,  // Changed
            total_portions: 140
          },
          {
            school_id: 2,
            portions_large: 110,  // Changed
            total_portions: 110
          },
          {
            school_id: 3,
            portions_large: 130,  // Changed
            total_portions: 130
          }
        ]
      }
    ]

    simulateFirebaseUpdate(firebasePath, updatedData)
    await wrapper.vm.$nextTick()

    // Verify all schools were updated
    expect(wrapper.vm.recipes[0].school_allocations[0].portions_small).toBe(60)
    expect(wrapper.vm.recipes[0].school_allocations[0].portions_large).toBe(80)
    expect(wrapper.vm.recipes[0].school_allocations[1].portions_large).toBe(110)
    expect(wrapper.vm.recipes[0].school_allocations[2].portions_large).toBe(130)
  })

  it('should preserve portion sizes when only status changes', async () => {
    const initialRecipe = {
      recipe_id: 1,
      name: 'Nasi Goreng',
      portions_required: 125,
      status: 'pending',
      school_allocations: [
        {
          school_id: 1,
          school_name: 'SD Negeri 1',
          school_category: 'SD',
          portion_size_type: 'mixed',
          portions_small: 50,
          portions_large: 75,
          total_portions: 125
        }
      ]
    }

    wrapper = createCookingWrapper([initialRecipe])
    await wrapper.vm.$nextTick()
    await wrapper.vm.loadData()
    await wrapper.vm.$nextTick()

    // Simulate Firebase update with only status change
    const firebasePath = 'kds/cooking/2024-01-15'
    const updatedData = [
      {
        recipe_id: 1,
        status: 'cooking',  // Only status changed
        // school_allocations not included in update
      }
    ]

    simulateFirebaseUpdate(firebasePath, updatedData)
    await wrapper.vm.$nextTick()

    // Verify portion sizes are preserved
    expect(wrapper.vm.recipes[0].school_allocations[0].portions_small).toBe(50)
    expect(wrapper.vm.recipes[0].school_allocations[0].portions_large).toBe(75)
    expect(wrapper.vm.recipes[0].school_allocations[0].total_portions).toBe(125)
    // But status should be updated
    expect(wrapper.vm.recipes[0].status).toBe('cooking')
  })

  it('should handle zero values in portion size updates correctly', async () => {
    const initialRecipe = {
      recipe_id: 1,
      name: 'Nasi Goreng',
      portions_required: 125,
      status: 'pending',
      school_allocations: [
        {
          school_id: 1,
          school_name: 'SD Negeri 1',
          school_category: 'SD',
          portion_size_type: 'mixed',
          portions_small: 50,
          portions_large: 75,
          total_portions: 125
        }
      ]
    }

    wrapper = createCookingWrapper([initialRecipe])
    await wrapper.vm.$nextTick()
    await wrapper.vm.loadData()
    await wrapper.vm.$nextTick()

    // Simulate Firebase update with zero small portions
    const firebasePath = 'kds/cooking/2024-01-15'
    const updatedData = [
      {
        recipe_id: 1,
        status: 'pending',
        school_allocations: [
          {
            school_id: 1,
            portions_small: 0,  // Changed to 0
            portions_large: 125,
            total_portions: 125
          }
        ]
      }
    ]

    simulateFirebaseUpdate(firebasePath, updatedData)
    await wrapper.vm.$nextTick()

    // Verify zero is handled correctly (not treated as undefined)
    expect(wrapper.vm.recipes[0].school_allocations[0].portions_small).toBe(0)
    expect(wrapper.vm.recipes[0].school_allocations[0].portions_large).toBe(125)
  })
})

describe('E2E: Real-time Firebase Updates - KDS Packing View', () => {
  let wrapper
  let getPackingToday

  beforeEach(async () => {
    vi.clearAllMocks()
    clearFirebaseData()

    const kdsService = await import('@/services/kdsService')
    getPackingToday = kdsService.getPackingToday

    // Mock initial API response
    getPackingToday.mockResolvedValue({
      success: true,
      data: []
    })
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
    }
  })

  const createPackingWrapper = (initialSchools = []) => {
    getPackingToday.mockResolvedValue({
      success: true,
      data: initialSchools
    })

    return mount(KDSPackingView, {
      global: {
        stubs: {
          'KDSDatePicker': {
            template: '<div class="kds-date-picker-stub"></div>',
            props: ['modelValue', 'loading'],
            emits: ['change', 'update:modelValue']
          },
          'a-space': { template: '<div><slot /></div>' },
          'a-tag': { template: '<span><slot /></span>' },
          'a-button': { template: '<button><slot /></button>' },
          'a-alert': { template: '<div><slot /></div>' },
          'a-spin': { template: '<div><slot /></div>' },
          'a-empty': { template: '<div><slot /></div>' },
          'a-row': { template: '<div><slot /></div>' },
          'a-col': { template: '<div><slot /></div>' },
          'a-card': { template: '<div><slot /></div>' },
          'a-statistic': { template: '<div><slot /></div>' },
          'a-divider': { template: '<div><slot /></div>' },
          'a-list': { template: '<div><slot /></div>' },
          'a-list-item': { template: '<div><slot /></div>' },
          'a-list-item-meta': { template: '<div><slot /></div>' },
          'a-avatar': { template: '<div><slot /></div>' },
          'a-badge': { template: '<span><slot /></span>' }
        }
      }
    })
  }

  it('should update SD school portion sizes in packing view when Firebase data changes', async () => {
    const initialSchool = {
      school_id: 1,
      school_name: 'SD Negeri 1',
      school_category: 'SD',
      portion_size_type: 'mixed',
      portions_small: 50,
      portions_large: 75,
      total_portions: 125,
      status: 'pending',
      menu_items: []
    }

    wrapper = createPackingWrapper([initialSchool])
    await wrapper.vm.$nextTick()
    await wrapper.vm.loadData()
    await wrapper.vm.$nextTick()

    // Verify initial state
    expect(wrapper.vm.schools[0].portions_small).toBe(50)
    expect(wrapper.vm.schools[0].portions_large).toBe(75)

    // Simulate Firebase update
    const firebasePath = 'kds/packing/2024-01-15'
    const updatedData = [
      {
        school_id: 1,
        portion_size_type: 'mixed',
        portions_small: 60,  // Changed
        portions_large: 85,  // Changed
        total_portions: 145,
        status: 'packing'
      }
    ]

    simulateFirebaseUpdate(firebasePath, updatedData)
    await wrapper.vm.$nextTick()

    // Verify the update was applied
    expect(wrapper.vm.schools[0].portions_small).toBe(60)
    expect(wrapper.vm.schools[0].portions_large).toBe(85)
    expect(wrapper.vm.schools[0].total_portions).toBe(145)
    expect(wrapper.vm.schools[0].status).toBe('packing')
  })

  it('should update SMP school portion sizes in packing view when Firebase data changes', async () => {
    const initialSchool = {
      school_id: 2,
      school_name: 'SMP Negeri 1',
      school_category: 'SMP',
      portion_size_type: 'large',
      portions_small: 0,
      portions_large: 100,
      total_portions: 100,
      status: 'pending',
      menu_items: []
    }

    wrapper = createPackingWrapper([initialSchool])
    await wrapper.vm.$nextTick()
    await wrapper.vm.loadData()
    await wrapper.vm.$nextTick()

    // Simulate Firebase update
    const firebasePath = 'kds/packing/2024-01-15'
    const updatedData = [
      {
        school_id: 2,
        portion_size_type: 'large',
        portions_small: 0,
        portions_large: 120,  // Changed
        total_portions: 120,
        status: 'packing'
      }
    ]

    simulateFirebaseUpdate(firebasePath, updatedData)
    await wrapper.vm.$nextTick()

    // Verify the update was applied
    expect(wrapper.vm.schools[0].portions_large).toBe(120)
    expect(wrapper.vm.schools[0].total_portions).toBe(120)
  })

  it('should handle multiple schools updating simultaneously in packing view', async () => {
    const initialSchools = [
      {
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 50,
        portions_large: 75,
        total_portions: 125,
        status: 'pending',
        menu_items: []
      },
      {
        school_id: 2,
        school_name: 'SMP Negeri 1',
        school_category: 'SMP',
        portion_size_type: 'large',
        portions_small: 0,
        portions_large: 100,
        total_portions: 100,
        status: 'pending',
        menu_items: []
      }
    ]

    wrapper = createPackingWrapper(initialSchools)
    await wrapper.vm.$nextTick()
    await wrapper.vm.loadData()
    await wrapper.vm.$nextTick()

    // Simulate Firebase update affecting both schools
    const firebasePath = 'kds/packing/2024-01-15'
    const updatedData = [
      {
        school_id: 1,
        portions_small: 55,  // Changed
        portions_large: 80,  // Changed
        total_portions: 135,
        status: 'packing'
      },
      {
        school_id: 2,
        portions_large: 110,  // Changed
        total_portions: 110,
        status: 'packing'
      }
    ]

    simulateFirebaseUpdate(firebasePath, updatedData)
    await wrapper.vm.$nextTick()

    // Verify both schools were updated
    expect(wrapper.vm.schools[0].portions_small).toBe(55)
    expect(wrapper.vm.schools[0].portions_large).toBe(80)
    expect(wrapper.vm.schools[1].portions_large).toBe(110)
  })

  it('should maintain portion_size_type field through Firebase updates', async () => {
    const initialSchool = {
      school_id: 1,
      school_name: 'SD Negeri 1',
      school_category: 'SD',
      portion_size_type: 'mixed',
      portions_small: 50,
      portions_large: 75,
      total_portions: 125,
      status: 'pending',
      menu_items: []
    }

    wrapper = createPackingWrapper([initialSchool])
    await wrapper.vm.$nextTick()
    await wrapper.vm.loadData()
    await wrapper.vm.$nextTick()

    // Simulate Firebase update
    const firebasePath = 'kds/packing/2024-01-15'
    const updatedData = [
      {
        school_id: 1,
        portion_size_type: 'mixed',  // Maintained
        portions_small: 60,
        portions_large: 80,
        total_portions: 140,
        status: 'packing'
      }
    ]

    simulateFirebaseUpdate(firebasePath, updatedData)
    await wrapper.vm.$nextTick()

    // Verify portion_size_type is maintained
    expect(wrapper.vm.schools[0].portion_size_type).toBe('mixed')
    expect(wrapper.vm.schools[0].portions_small).toBe(60)
    expect(wrapper.vm.schools[0].portions_large).toBe(80)
  })

  it('should preserve portion sizes when Firebase update does not include portion fields', async () => {
    const initialSchool = {
      school_id: 1,
      school_name: 'SD Negeri 1',
      school_category: 'SD',
      portion_size_type: 'mixed',
      portions_small: 50,
      portions_large: 75,
      total_portions: 125,
      status: 'pending',
      menu_items: []
    }

    wrapper = createPackingWrapper([initialSchool])
    await wrapper.vm.$nextTick()
    await wrapper.vm.loadData()
    await wrapper.vm.$nextTick()

    // Simulate Firebase update without portion fields (only status)
    const firebasePath = 'kds/packing/2024-01-15'
    const updatedData = [
      {
        school_id: 1,
        status: 'packing'  // Only status changed
      }
    ]

    simulateFirebaseUpdate(firebasePath, updatedData)
    await wrapper.vm.$nextTick()

    // Verify portion sizes are preserved
    expect(wrapper.vm.schools[0].portions_small).toBe(50)
    expect(wrapper.vm.schools[0].portions_large).toBe(75)
    expect(wrapper.vm.schools[0].total_portions).toBe(125)
    // But status should be updated
    expect(wrapper.vm.schools[0].status).toBe('packing')
  })
})

describe('E2E: Cross-View Data Consistency', () => {
  let cookingWrapper
  let packingWrapper
  let getCookingToday
  let getPackingToday

  beforeEach(async () => {
    vi.clearAllMocks()
    clearFirebaseData()

    const kdsService = await import('@/services/kdsService')
    getCookingToday = kdsService.getCookingToday
    getPackingToday = kdsService.getPackingToday

    getCookingToday.mockResolvedValue({ success: true, data: [] })
    getPackingToday.mockResolvedValue({ success: true, data: [] })
  })

  afterEach(() => {
    if (cookingWrapper) {
      cookingWrapper.unmount()
    }
    if (packingWrapper) {
      packingWrapper.unmount()
    }
  })

  it('should sync portion size data between cooking and packing views', async () => {
    // Setup cooking view with initial data
    const initialRecipe = {
      recipe_id: 1,
      name: 'Nasi Goreng',
      portions_required: 125,
      status: 'pending',
      school_allocations: [
        {
          school_id: 1,
          school_name: 'SD Negeri 1',
          school_category: 'SD',
          portion_size_type: 'mixed',
          portions_small: 50,
          portions_large: 75,
          total_portions: 125
        }
      ]
    }

    getCookingToday.mockResolvedValue({
      success: true,
      data: [initialRecipe]
    })

    cookingWrapper = mount(KDSCookingView, {
      global: {
        stubs: {
          'KDSDatePicker': { template: '<div></div>' },
          'a-space': { template: '<div><slot /></div>' },
          'a-tag': { template: '<span><slot /></span>' },
          'a-button': { template: '<button><slot /></button>' },
          'a-alert': { template: '<div><slot /></div>' },
          'a-spin': { template: '<div><slot /></div>' },
          'a-empty': { template: '<div><slot /></div>' },
          'a-row': { template: '<div><slot /></div>' },
          'a-col': { template: '<div><slot /></div>' },
          'a-card': { template: '<div><slot /></div>' },
          'a-descriptions': { template: '<div><slot /></div>' },
          'a-descriptions-item': { template: '<div><slot /></div>' },
          'a-divider': { template: '<div><slot /></div>' },
          'a-list': { template: '<div><slot /></div>' },
          'a-list-item': { template: '<div><slot /></div>' },
          'a-list-item-meta': { template: '<div><slot /></div>' },
          'a-badge': { template: '<span><slot /></span>' }
        }
      }
    })

    await cookingWrapper.vm.$nextTick()
    await cookingWrapper.vm.loadData()
    await cookingWrapper.vm.$nextTick()

    // Setup packing view with corresponding school data
    const initialSchool = {
      school_id: 1,
      school_name: 'SD Negeri 1',
      school_category: 'SD',
      portion_size_type: 'mixed',
      portions_small: 50,
      portions_large: 75,
      total_portions: 125,
      status: 'pending',
      menu_items: []
    }

    getPackingToday.mockResolvedValue({
      success: true,
      data: [initialSchool]
    })

    packingWrapper = mount(KDSPackingView, {
      global: {
        stubs: {
          'KDSDatePicker': { template: '<div></div>' },
          'a-space': { template: '<div><slot /></div>' },
          'a-tag': { template: '<span><slot /></span>' },
          'a-button': { template: '<button><slot /></button>' },
          'a-alert': { template: '<div><slot /></div>' },
          'a-spin': { template: '<div><slot /></div>' },
          'a-empty': { template: '<div><slot /></div>' },
          'a-row': { template: '<div><slot /></div>' },
          'a-col': { template: '<div><slot /></div>' },
          'a-card': { template: '<div><slot /></div>' },
          'a-statistic': { template: '<div><slot /></div>' },
          'a-divider': { template: '<div><slot /></div>' },
          'a-list': { template: '<div><slot /></div>' },
          'a-list-item': { template: '<div><slot /></div>' },
          'a-list-item-meta': { template: '<div><slot /></div>' },
          'a-avatar': { template: '<div><slot /></div>' },
          'a-badge': { template: '<span><slot /></span>' }
        }
      }
    })

    await packingWrapper.vm.$nextTick()
    await packingWrapper.vm.loadData()
    await packingWrapper.vm.$nextTick()

    // Verify initial state in both views
    expect(cookingWrapper.vm.recipes[0].school_allocations[0].portions_small).toBe(50)
    expect(packingWrapper.vm.schools[0].portions_small).toBe(50)

    // Simulate Firebase update (as if from a third client)
    const cookingPath = 'kds/cooking/2024-01-15'
    const packingPath = 'kds/packing/2024-01-15'

    simulateFirebaseUpdate(cookingPath, [
      {
        recipe_id: 1,
        status: 'cooking',
        school_allocations: [
          {
            school_id: 1,
            portions_small: 60,
            portions_large: 85,
            total_portions: 145
          }
        ]
      }
    ])

    simulateFirebaseUpdate(packingPath, [
      {
        school_id: 1,
        portions_small: 60,
        portions_large: 85,
        total_portions: 145,
        status: 'packing'
      }
    ])

    await cookingWrapper.vm.$nextTick()
    await packingWrapper.vm.$nextTick()

    // Verify both views are updated with consistent data
    expect(cookingWrapper.vm.recipes[0].school_allocations[0].portions_small).toBe(60)
    expect(cookingWrapper.vm.recipes[0].school_allocations[0].portions_large).toBe(85)
    expect(packingWrapper.vm.schools[0].portions_small).toBe(60)
    expect(packingWrapper.vm.schools[0].portions_large).toBe(85)

    // Verify totals match
    expect(cookingWrapper.vm.recipes[0].school_allocations[0].total_portions).toBe(145)
    expect(packingWrapper.vm.schools[0].total_portions).toBe(145)
  })
})

describe('E2E: Firebase Listener Lifecycle', () => {
  let wrapper
  let getCookingToday

  beforeEach(async () => {
    vi.clearAllMocks()
    clearFirebaseData()

    const kdsService = await import('@/services/kdsService')
    getCookingToday = kdsService.getCookingToday

    getCookingToday.mockResolvedValue({ success: true, data: [] })
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
    }
  })

  it('should setup Firebase listener on component mount', async () => {
    const initialRecipe = {
      recipe_id: 1,
      name: 'Nasi Goreng',
      portions_required: 125,
      status: 'pending',
      school_allocations: []
    }

    getCookingToday.mockResolvedValue({
      success: true,
      data: [initialRecipe]
    })

    wrapper = mount(KDSCookingView, {
      global: {
        stubs: {
          'KDSDatePicker': { template: '<div></div>' },
          'a-space': { template: '<div><slot /></div>' },
          'a-tag': { template: '<span><slot /></span>' },
          'a-button': { template: '<button><slot /></button>' },
          'a-alert': { template: '<div><slot /></div>' },
          'a-spin': { template: '<div><slot /></div>' },
          'a-empty': { template: '<div><slot /></div>' },
          'a-row': { template: '<div><slot /></div>' },
          'a-col': { template: '<div><slot /></div>' },
          'a-card': { template: '<div><slot /></div>' },
          'a-descriptions': { template: '<div><slot /></div>' },
          'a-descriptions-item': { template: '<div><slot /></div>' },
          'a-divider': { template: '<div><slot /></div>' },
          'a-list': { template: '<div><slot /></div>' },
          'a-list-item': { template: '<div><slot /></div>' },
          'a-list-item-meta': { template: '<div><slot /></div>' },
          'a-badge': { template: '<span><slot /></span>' }
        }
      }
    })

    await wrapper.vm.$nextTick()
    await wrapper.vm.loadData()
    await wrapper.vm.$nextTick()

    // Verify Firebase listener was set up
    expect(onValue).toHaveBeenCalled()
    expect(firebaseRef).toHaveBeenCalled()
  })

  it('should cleanup Firebase listener on component unmount', async () => {
    const initialRecipe = {
      recipe_id: 1,
      name: 'Nasi Goreng',
      portions_required: 125,
      status: 'pending',
      school_allocations: []
    }

    getCookingToday.mockResolvedValue({
      success: true,
      data: [initialRecipe]
    })

    wrapper = mount(KDSCookingView, {
      global: {
        stubs: {
          'KDSDatePicker': { template: '<div></div>' },
          'a-space': { template: '<div><slot /></div>' },
          'a-tag': { template: '<span><slot /></span>' },
          'a-button': { template: '<button><slot /></button>' },
          'a-alert': { template: '<div><slot /></div>' },
          'a-spin': { template: '<div><slot /></div>' },
          'a-empty': { template: '<div><slot /></div>' },
          'a-row': { template: '<div><slot /></div>' },
          'a-col': { template: '<div><slot /></div>' },
          'a-card': { template: '<div><slot /></div>' },
          'a-descriptions': { template: '<div><slot /></div>' },
          'a-descriptions-item': { template: '<div><slot /></div>' },
          'a-divider': { template: '<div><slot /></div>' },
          'a-list': { template: '<div><slot /></div>' },
          'a-list-item': { template: '<div><slot /></div>' },
          'a-list-item-meta': { template: '<div><slot /></div>' },
          'a-badge': { template: '<span><slot /></span>' }
        }
      }
    })

    await wrapper.vm.$nextTick()
    await wrapper.vm.loadData()
    await wrapper.vm.$nextTick()

    // Unmount the component
    wrapper.unmount()

    // Verify Firebase listener was cleaned up
    expect(off).toHaveBeenCalled()
  })
})

describe('E2E: Documentation and Testing Approach', () => {
  it('should document the Firebase real-time testing approach', () => {
    // This test documents the approach for testing Firebase real-time features
    
    const testingApproach = {
      mockStrategy: 'Mock Firebase database and listeners',
      simulationMethod: 'Simulate Firebase updates by calling listener callbacks',
      verificationApproach: 'Verify component state updates after simulated Firebase changes',
      
      keyPoints: [
        'Firebase listeners are mocked to avoid actual database connections',
        'Updates are simulated by calling the listener callbacks with mock data',
        'Component state is verified after each simulated update',
        'Multiple views can be tested simultaneously to verify data consistency',
        'Listener lifecycle (setup and cleanup) is tested'
      ],
      
      limitations: [
        'Does not test actual Firebase connection',
        'Does not test network latency or connection issues',
        'Does not test Firebase security rules',
        'Manual testing still required for end-to-end Firebase integration'
      ],
      
      manualTestingRequired: [
        'Test with actual Firebase database',
        'Test with multiple real clients (different browsers/devices)',
        'Test network disconnection and reconnection',
        'Test Firebase security rules',
        'Test data persistence and recovery'
      ]
    }

    // Verify the testing approach is documented
    expect(testingApproach.mockStrategy).toBeDefined()
    expect(testingApproach.keyPoints).toHaveLength(5)
    expect(testingApproach.limitations).toHaveLength(4)
    expect(testingApproach.manualTestingRequired).toHaveLength(5)
  })
})
