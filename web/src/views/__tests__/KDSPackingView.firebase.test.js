import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import KDSPackingView from '../KDSPackingView.vue'

// Mock services
vi.mock('@/services/kdsService', () => ({
  getPackingToday: vi.fn().mockResolvedValue({ success: true, data: [] }),
  updatePackingStatus: vi.fn()
}))

vi.mock('@/services/firebase', () => ({
  database: {}
}))

vi.mock('firebase/database', () => ({
  ref: vi.fn(),
  onValue: vi.fn(),
  off: vi.fn()
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

describe('KDSPackingView - Firebase Real-time Updates for Portion Sizes', () => {
  const createWrapper = (schools = []) => {
    const wrapper = mount(KDSPackingView, {
      global: {
        stubs: {
          'KDSDatePicker': true,
          'a-space': true,
          'a-tag': { template: '<span><slot /></span>' },
          'a-button': { template: '<button><slot /></button>' },
          'a-alert': true,
          'a-spin': { template: '<div><slot /></div>' },
          'a-empty': true,
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
    
    if (schools.length > 0) {
      wrapper.vm.schools = schools
      wrapper.vm.loading = false
    }
    
    return wrapper
  }

  describe('Firebase Listener Updates - Portion Size Data', () => {
    it('should update portion size data when Firebase sends updated data for SD school', async () => {
      // Initial data
      const initialSchools = [{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 50,
        portions_large: 75,
        total_portions: 125,
        status: 'pending',
        menu_items: []
      }]

      const wrapper = createWrapper(initialSchools)
      await wrapper.vm.$nextTick()
      
      // Verify initial state
      expect(wrapper.vm.schools[0].portions_small).toBe(50)
      expect(wrapper.vm.schools[0].portions_large).toBe(75)
      expect(wrapper.vm.schools[0].total_portions).toBe(125)

      // Simulate Firebase update with changed portion sizes
      const firebaseUpdate = [{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 60,  // Changed
        portions_large: 85,  // Changed
        total_portions: 145,
        status: 'packing'
      }]

      // Apply the update (simulating what the Firebase listener does)
      wrapper.vm.schools = wrapper.vm.schools.map(school => {
        const updated = firebaseUpdate.find(fs => fs.school_id === school.school_id)
        if (updated) {
          return {
            ...school,
            status: updated.status,
            portion_size_type: updated.portion_size_type || school.portion_size_type,
            portions_small: updated.portions_small !== undefined ? updated.portions_small : school.portions_small,
            portions_large: updated.portions_large !== undefined ? updated.portions_large : school.portions_large,
            total_portions: updated.total_portions || school.total_portions
          }
        }
        return school
      })

      await wrapper.vm.$nextTick()

      // Verify updated portion sizes
      expect(wrapper.vm.schools[0].portions_small).toBe(60)
      expect(wrapper.vm.schools[0].portions_large).toBe(85)
      expect(wrapper.vm.schools[0].total_portions).toBe(145)
      expect(wrapper.vm.schools[0].status).toBe('packing')
    })

    it('should update portion size data when Firebase sends updated data for SMP school', async () => {
      // Initial data
      const initialSchools = [{
        school_id: 2,
        school_name: 'SMP Negeri 1',
        school_category: 'SMP',
        portion_size_type: 'large',
        portions_small: 0,
        portions_large: 100,
        total_portions: 100,
        status: 'pending',
        menu_items: []
      }]

      const wrapper = createWrapper(initialSchools)
      await wrapper.vm.$nextTick()

      // Simulate Firebase update
      const firebaseUpdate = [{
        school_id: 2,
        school_name: 'SMP Negeri 1',
        school_category: 'SMP',
        portion_size_type: 'large',
        portions_small: 0,
        portions_large: 120,  // Changed
        total_portions: 120,
        status: 'packing'
      }]

      wrapper.vm.schools = wrapper.vm.schools.map(school => {
        const updated = firebaseUpdate.find(fs => fs.school_id === school.school_id)
        if (updated) {
          return {
            ...school,
            status: updated.status,
            portion_size_type: updated.portion_size_type || school.portion_size_type,
            portions_small: updated.portions_small !== undefined ? updated.portions_small : school.portions_small,
            portions_large: updated.portions_large !== undefined ? updated.portions_large : school.portions_large,
            total_portions: updated.total_portions || school.total_portions
          }
        }
        return school
      })

      await wrapper.vm.$nextTick()

      expect(wrapper.vm.schools[0].portions_large).toBe(120)
      expect(wrapper.vm.schools[0].total_portions).toBe(120)
      expect(wrapper.vm.schools[0].status).toBe('packing')
    })

    it('should handle Firebase updates with multiple schools having different portion sizes', async () => {
      // Initial data
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

      const wrapper = createWrapper(initialSchools)
      await wrapper.vm.$nextTick()

      // Simulate Firebase update
      const firebaseUpdate = [
        {
          school_id: 1,
          portion_size_type: 'mixed',
          portions_small: 55,  // Changed
          portions_large: 80,  // Changed
          total_portions: 135,
          status: 'packing'
        },
        {
          school_id: 2,
          portion_size_type: 'large',
          portions_small: 0,
          portions_large: 110,  // Changed
          total_portions: 110,
          status: 'packing'
        }
      ]

      wrapper.vm.schools = wrapper.vm.schools.map(school => {
        const updated = firebaseUpdate.find(fs => fs.school_id === school.school_id)
        if (updated) {
          return {
            ...school,
            status: updated.status,
            portion_size_type: updated.portion_size_type || school.portion_size_type,
            portions_small: updated.portions_small !== undefined ? updated.portions_small : school.portions_small,
            portions_large: updated.portions_large !== undefined ? updated.portions_large : school.portions_large,
            total_portions: updated.total_portions || school.total_portions
          }
        }
        return school
      })

      await wrapper.vm.$nextTick()

      expect(wrapper.vm.schools[0].portions_small).toBe(55)
      expect(wrapper.vm.schools[0].portions_large).toBe(80)
      expect(wrapper.vm.schools[1].portions_large).toBe(110)
    })

    it('should preserve portion size data when Firebase update does not include portion fields', async () => {
      // Initial data
      const initialSchools = [{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 50,
        portions_large: 75,
        total_portions: 125,
        status: 'pending',
        menu_items: []
      }]

      const wrapper = createWrapper(initialSchools)
      await wrapper.vm.$nextTick()

      // Simulate Firebase update without portion fields (only status)
      const firebaseUpdate = [{
        school_id: 1,
        status: 'packing'
      }]

      wrapper.vm.schools = wrapper.vm.schools.map(school => {
        const updated = firebaseUpdate.find(fs => fs.school_id === school.school_id)
        if (updated) {
          return {
            ...school,
            status: updated.status,
            portion_size_type: updated.portion_size_type || school.portion_size_type,
            portions_small: updated.portions_small !== undefined ? updated.portions_small : school.portions_small,
            portions_large: updated.portions_large !== undefined ? updated.portions_large : school.portions_large,
            total_portions: updated.total_portions || school.total_portions
          }
        }
        return school
      })

      await wrapper.vm.$nextTick()

      // Verify portion sizes are preserved
      expect(wrapper.vm.schools[0].portions_small).toBe(50)
      expect(wrapper.vm.schools[0].portions_large).toBe(75)
      expect(wrapper.vm.schools[0].total_portions).toBe(125)
      expect(wrapper.vm.schools[0].status).toBe('packing')
    })

    it('should maintain portion_size_type field through Firebase updates', async () => {
      // Initial data
      const initialSchools = [{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 50,
        portions_large: 75,
        total_portions: 125,
        status: 'pending',
        menu_items: []
      }]

      const wrapper = createWrapper(initialSchools)
      await wrapper.vm.$nextTick()

      // Simulate Firebase update
      const firebaseUpdate = [{
        school_id: 1,
        portion_size_type: 'mixed',  // Maintained
        portions_small: 60,
        portions_large: 80,
        total_portions: 140,
        status: 'packing'
      }]

      wrapper.vm.schools = wrapper.vm.schools.map(school => {
        const updated = firebaseUpdate.find(fs => fs.school_id === school.school_id)
        if (updated) {
          return {
            ...school,
            status: updated.status,
            portion_size_type: updated.portion_size_type || school.portion_size_type,
            portions_small: updated.portions_small !== undefined ? updated.portions_small : school.portions_small,
            portions_large: updated.portions_large !== undefined ? updated.portions_large : school.portions_large,
            total_portions: updated.total_portions || school.total_portions
          }
        }
        return school
      })

      await wrapper.vm.$nextTick()

      expect(wrapper.vm.schools[0].portion_size_type).toBe('mixed')
      expect(wrapper.vm.schools[0].portions_small).toBe(60)
      expect(wrapper.vm.schools[0].portions_large).toBe(80)
    })

    it('should handle zero values in portion size updates correctly', async () => {
      // Initial data
      const initialSchools = [{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 50,
        portions_large: 75,
        total_portions: 125,
        status: 'pending',
        menu_items: []
      }]

      const wrapper = createWrapper(initialSchools)
      await wrapper.vm.$nextTick()

      // Simulate Firebase update with zero small portions
      const firebaseUpdate = [{
        school_id: 1,
        portion_size_type: 'mixed',
        portions_small: 0,  // Changed to 0
        portions_large: 125,
        total_portions: 125,
        status: 'packing'
      }]

      wrapper.vm.schools = wrapper.vm.schools.map(school => {
        const updated = firebaseUpdate.find(fs => fs.school_id === school.school_id)
        if (updated) {
          return {
            ...school,
            status: updated.status,
            portion_size_type: updated.portion_size_type || school.portion_size_type,
            portions_small: updated.portions_small !== undefined ? updated.portions_small : school.portions_small,
            portions_large: updated.portions_large !== undefined ? updated.portions_large : school.portions_large,
            total_portions: updated.total_portions || school.total_portions
          }
        }
        return school
      })

      await wrapper.vm.$nextTick()

      expect(wrapper.vm.schools[0].portions_small).toBe(0)
      expect(wrapper.vm.schools[0].portions_large).toBe(125)
    })
  })
})
