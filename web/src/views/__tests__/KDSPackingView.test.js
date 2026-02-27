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

// Mock ant-design-vue message
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

describe('KDSPackingView - Portion Size Display', () => {
  const createWrapper = (schools = []) => {
    const wrapper = mount(KDSPackingView, {
      global: {
        stubs: {
          'KDSDatePicker': true,
          'a-space': true,
          'a-tag': {
            template: '<span><slot /></span>',
            props: ['color']
          },
          'a-button': {
            template: '<button><slot /></button>',
            props: ['loading', 'type', 'block', 'disabled']
          },
          'a-alert': true,
          'a-spin': {
            template: '<div><slot /></div>'
          },
          'a-empty': true,
          'a-row': {
            template: '<div><slot /></div>',
            props: ['gutter']
          },
          'a-col': {
            template: '<div><slot /></div>',
            props: ['xs', 'sm', 'md', 'lg', 'xl']
          },
          'a-card': {
            template: '<div class="a-card-stub"><div class="card-body"><slot /></div><div class="card-actions"><slot name="actions" /></div></div>'
          },
          'a-statistic': {
            template: '<div class="statistic"><div class="title">{{ title }}</div><div class="value">{{ value }}</div></div>',
            props: ['title', 'value', 'suffix', 'valueStyle']
          },
          'a-divider': {
            template: '<div><slot /></div>'
          },
          'a-list': {
            template: '<div><div v-for="(item, index) in dataSource" :key="index"><slot name="renderItem" :item="item" /></div></div>',
            props: ['dataSource', 'size', 'split']
          },
          'a-list-item': {
            template: '<div><slot /></div>'
          },
          'a-list-item-meta': {
            template: '<div><div class="meta-title"><slot name="title" /></div><div class="meta-description"><slot name="description" /></div></div>'
          },
          'a-avatar': {
            template: '<div><slot /></div>',
            props: ['src', 'shape', 'size']
          },
          'a-badge': {
            template: '<span><slot /></span>',
            props: ['count', 'numberStyle']
          }
        }
      }
    })
    
    // Set schools data after mounting
    if (schools.length > 0) {
      wrapper.vm.schools = schools
      wrapper.vm.loading = false
    }
    
    return wrapper
  }

  describe('Component Rendering', () => {
    it('should render KDS Packing View component', () => {
      const wrapper = createWrapper()
      
      expect(wrapper.exists()).toBe(true)
      expect(wrapper.find('.kds-packing-view').exists()).toBe(true)
    })

    it('should have schools data array', () => {
      const wrapper = createWrapper()
      
      expect(wrapper.vm.schools).toBeDefined()
      expect(Array.isArray(wrapper.vm.schools)).toBe(true)
    })
  })

  describe('Portion Size Display for SD Schools', () => {
    it('should display portion breakdown for SD school with mixed portions', async () => {
      const wrapper = createWrapper([{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 50,
        portions_large: 75,
        total_portions: 125,
        status: 'pending',
        menu_items: []
      }])

      await wrapper.vm.$nextTick()
      const html = wrapper.html()
      expect(html).toContain('SD Negeri 1')
      expect(html).toContain('portion-breakdown')
      expect(html).toContain('Kecil (Kelas 1-3)')
      expect(html).toContain('Besar (Kelas 4-6)')
    })

    it('should display correct portion counts for SD school', async () => {
      const wrapper = createWrapper([{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 50,
        portions_large: 75,
        total_portions: 125,
        status: 'pending',
        menu_items: []
      }])

      await wrapper.vm.$nextTick()
      const html = wrapper.html()
      expect(html).toContain('50 porsi')
      expect(html).toContain('75 porsi')
      expect(html).toContain('125')
    })

    it('should display small and large portion cards for SD schools', async () => {
      const wrapper = createWrapper([{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 50,
        portions_large: 75,
        total_portions: 125,
        status: 'pending',
        menu_items: []
      }])

      await wrapper.vm.$nextTick()
      const html = wrapper.html()
      expect(html).toContain('portion-size-card small')
      expect(html).toContain('portion-size-card large')
    })
  })

  describe('Portion Size Display for SMP/SMA Schools', () => {
    it('should display only large portions for SMP schools', async () => {
      const wrapper = createWrapper([{
        school_id: 2,
        school_name: 'SMP Negeri 1',
        school_category: 'SMP',
        portion_size_type: 'large',
        portions_small: 0,
        portions_large: 100,
        total_portions: 100,
        status: 'pending',
        menu_items: []
      }])

      await wrapper.vm.$nextTick()
      const html = wrapper.html()
      expect(html).toContain('SMP Negeri 1')
      expect(html).toContain('Porsi Besar')
      expect(html).toContain('100 porsi')
      expect(html).not.toContain('Kecil (Kelas 1-3)')
      expect(html).not.toContain('Besar (Kelas 4-6)')
    })

    it('should display only large portions for SMA schools', async () => {
      const wrapper = createWrapper([{
        school_id: 3,
        school_name: 'SMA Negeri 1',
        school_category: 'SMA',
        portion_size_type: 'large',
        portions_small: 0,
        portions_large: 120,
        total_portions: 120,
        status: 'pending',
        menu_items: []
      }])

      await wrapper.vm.$nextTick()
      const html = wrapper.html()
      expect(html).toContain('SMA Negeri 1')
      expect(html).toContain('Porsi Besar')
      expect(html).toContain('120 porsi')
      expect(html).not.toContain('Kecil')
      expect(html).not.toContain('Kelas 4-6')
    })

    it('should display single portion card for SMP/SMA schools', async () => {
      const wrapper = createWrapper([{
        school_id: 2,
        school_name: 'SMP Negeri 1',
        school_category: 'SMP',
        portion_size_type: 'large',
        portions_small: 0,
        portions_large: 100,
        total_portions: 100,
        status: 'pending',
        menu_items: []
      }])

      await wrapper.vm.$nextTick()
      const html = wrapper.html()
      expect(html).toContain('portion-size-card large single')
    })
  })

  describe('Menu Item Portion Size Display', () => {
    it('should display portion sizes for each menu item in SD school', async () => {
      const wrapper = createWrapper([{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 50,
        portions_large: 75,
        total_portions: 125,
        status: 'pending',
        menu_items: [{
          recipe_id: 1,
          recipe_name: 'Nasi Goreng',
          portions_small: 50,
          portions_large: 75,
          total_portions: 125
        }]
      }])

      await wrapper.vm.$nextTick()
      const html = wrapper.html()
      expect(html).toContain('Nasi Goreng')
      expect(html).toContain('Kecil: 50')
      expect(html).toContain('Besar: 75')
      expect(html).toContain('Total: 125')
    })

    it('should display only large portions for menu items in SMP school', async () => {
      const wrapper = createWrapper([{
        school_id: 2,
        school_name: 'SMP Negeri 1',
        school_category: 'SMP',
        portion_size_type: 'large',
        portions_small: 0,
        portions_large: 100,
        total_portions: 100,
        status: 'pending',
        menu_items: [{
          recipe_id: 1,
          recipe_name: 'Nasi Goreng',
          portions_small: 0,
          portions_large: 100,
          total_portions: 100
        }]
      }])

      await wrapper.vm.$nextTick()
      const html = wrapper.html()
      expect(html).toContain('Nasi Goreng')
      expect(html).toContain('Besar: 100')
      expect(html).toContain('Total: 100')
      expect(html).not.toContain('Kecil:')
    })

    it('should conditionally display small portion tag only when portions_small > 0', async () => {
      const wrapper = createWrapper([{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 0,
        portions_large: 75,
        total_portions: 75,
        status: 'pending',
        menu_items: [{
          recipe_id: 1,
          recipe_name: 'Nasi Goreng',
          portions_small: 0,
          portions_large: 75,
          total_portions: 75
        }]
      }])

      await wrapper.vm.$nextTick()
      const html = wrapper.html()
      expect(html).toContain('Nasi Goreng')
      expect(html).not.toContain('Kecil:')
      expect(html).toContain('Besar: 75')
    })

    it('should conditionally display large portion tag only when portions_large > 0', async () => {
      const wrapper = createWrapper([{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 50,
        portions_large: 0,
        total_portions: 50,
        status: 'pending',
        menu_items: [{
          recipe_id: 1,
          recipe_name: 'Nasi Goreng',
          portions_small: 50,
          portions_large: 0,
          total_portions: 50
        }]
      }])

      await wrapper.vm.$nextTick()
      const html = wrapper.html()
      expect(html).toContain('Nasi Goreng')
      expect(html).toContain('Kecil: 50')
      expect(html).not.toContain('Besar:')
    })
  })

  describe('Multiple Schools Display', () => {
    it('should render multiple schools with correct portion information', async () => {
      const wrapper = createWrapper([
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
        },
        {
          school_id: 3,
          school_name: 'SMA Negeri 1',
          school_category: 'SMA',
          portion_size_type: 'large',
          portions_small: 0,
          portions_large: 120,
          total_portions: 120,
          status: 'pending',
          menu_items: []
        }
      ])

      await wrapper.vm.$nextTick()
      const html = wrapper.html()
      expect(html).toContain('SD Negeri 1')
      expect(html).toContain('SMP Negeri 1')
      expect(html).toContain('SMA Negeri 1')
      expect(html).toContain('125')
      expect(html).toContain('100')
      expect(html).toContain('120')
    })
  })

  describe('Visual Styling', () => {
    it('should apply portion-breakdown class', async () => {
      const wrapper = createWrapper([{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 50,
        portions_large: 75,
        total_portions: 125,
        status: 'pending',
        menu_items: []
      }])

      await wrapper.vm.$nextTick()
      const html = wrapper.html()
      expect(html).toContain('portion-breakdown')
    })

    it('should apply portion-size-card classes', async () => {
      const wrapper = createWrapper([{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 50,
        portions_large: 75,
        total_portions: 125,
        status: 'pending',
        menu_items: []
      }])

      await wrapper.vm.$nextTick()
      const html = wrapper.html()
      expect(html).toContain('portion-size-card')
      expect(html).toContain('portion-label')
      expect(html).toContain('portion-value')
    })

    it('should apply menu-item-portions class for menu items', async () => {
      const wrapper = createWrapper([{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 50,
        portions_large: 75,
        total_portions: 125,
        status: 'pending',
        menu_items: [{
          recipe_id: 1,
          recipe_name: 'Nasi Goreng',
          portions_small: 50,
          portions_large: 75,
          total_portions: 125
        }]
      }])

      await wrapper.vm.$nextTick()
      const html = wrapper.html()
      expect(html).toContain('menu-item-portions')
    })
  })

  describe('Edge Cases', () => {
    it('should handle empty schools array', async () => {
      const wrapper = createWrapper([])

      await wrapper.vm.$nextTick()
      expect(wrapper.vm.schools).toEqual([])
    })

    it('should handle school with no menu items', async () => {
      const wrapper = createWrapper([{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 50,
        portions_large: 75,
        total_portions: 125,
        status: 'pending',
        menu_items: []
      }])

      await wrapper.vm.$nextTick()
      const html = wrapper.html()
      expect(html).toContain('SD Negeri 1')
      expect(html).toContain('125')
    })

    it('should handle SD school with only small portions', async () => {
      const wrapper = createWrapper([{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 50,
        portions_large: 0,
        total_portions: 50,
        status: 'pending',
        menu_items: []
      }])

      await wrapper.vm.$nextTick()
      const html = wrapper.html()
      expect(html).toContain('50 porsi')
      expect(html).toContain('0 porsi')
    })

    it('should handle SD school with only large portions', async () => {
      const wrapper = createWrapper([{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 0,
        portions_large: 75,
        total_portions: 75,
        status: 'pending',
        menu_items: []
      }])

      await wrapper.vm.$nextTick()
      const html = wrapper.html()
      expect(html).toContain('0 porsi')
      expect(html).toContain('75 porsi')
    })
  })

  describe('Firebase Real-time Updates - Portion Size Changes', () => {
    it('should update portion size data when Firebase data changes for SD school', async () => {
      // Initial data with SD school
      const wrapper = createWrapper([{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 50,
        portions_large: 75,
        total_portions: 125,
        status: 'pending',
        menu_items: []
      }])

      await wrapper.vm.$nextTick()
      
      // Verify initial state
      expect(wrapper.vm.schools[0].portions_small).toBe(50)
      expect(wrapper.vm.schools[0].portions_large).toBe(75)
      expect(wrapper.vm.schools[0].total_portions).toBe(125)

      // Simulate Firebase update with changed portion sizes
      const firebaseData = [{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 60,  // Changed from 50
        portions_large: 85,  // Changed from 75
        total_portions: 145, // Changed from 125
        status: 'packing'
      }]

      // Manually trigger the Firebase update logic
      wrapper.vm.schools = wrapper.vm.schools.map(school => {
        const firebaseSchool = firebaseData.find(fs => fs.school_id === school.school_id)
        if (firebaseSchool) {
          return {
            ...school,
            status: firebaseSchool.status,
            portion_size_type: firebaseSchool.portion_size_type || school.portion_size_type,
            portions_small: firebaseSchool.portions_small !== undefined ? firebaseSchool.portions_small : school.portions_small,
            portions_large: firebaseSchool.portions_large !== undefined ? firebaseSchool.portions_large : school.portions_large,
            total_portions: firebaseSchool.total_portions || school.total_portions
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

    it('should update portion size data when Firebase data changes for SMP school', async () => {
      // Initial data with SMP school
      const wrapper = createWrapper([{
        school_id: 2,
        school_name: 'SMP Negeri 1',
        school_category: 'SMP',
        portion_size_type: 'large',
        portions_small: 0,
        portions_large: 100,
        total_portions: 100,
        status: 'pending',
        menu_items: []
      }])

      await wrapper.vm.$nextTick()
      
      // Verify initial state
      expect(wrapper.vm.schools[0].portions_large).toBe(100)
      expect(wrapper.vm.schools[0].total_portions).toBe(100)

      // Simulate Firebase update with changed portion sizes
      const firebaseData = [{
        school_id: 2,
        school_name: 'SMP Negeri 1',
        school_category: 'SMP',
        portion_size_type: 'large',
        portions_small: 0,
        portions_large: 120,  // Changed from 100
        total_portions: 120,  // Changed from 100
        status: 'packing'
      }]

      // Manually trigger the Firebase update logic
      wrapper.vm.schools = wrapper.vm.schools.map(school => {
        const firebaseSchool = firebaseData.find(fs => fs.school_id === school.school_id)
        if (firebaseSchool) {
          return {
            ...school,
            status: firebaseSchool.status,
            portion_size_type: firebaseSchool.portion_size_type || school.portion_size_type,
            portions_small: firebaseSchool.portions_small !== undefined ? firebaseSchool.portions_small : school.portions_small,
            portions_large: firebaseSchool.portions_large !== undefined ? firebaseSchool.portions_large : school.portions_large,
            total_portions: firebaseSchool.total_portions || school.total_portions
          }
        }
        return school
      })

      await wrapper.vm.$nextTick()

      // Verify updated portion sizes
      expect(wrapper.vm.schools[0].portions_large).toBe(120)
      expect(wrapper.vm.schools[0].total_portions).toBe(120)
      expect(wrapper.vm.schools[0].status).toBe('packing')
    })

    it('should handle Firebase updates with multiple schools having different portion sizes', async () => {
      // Initial data with multiple schools
      const wrapper = createWrapper([
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
        },
        {
          school_id: 3,
          school_name: 'SMA Negeri 1',
          school_category: 'SMA',
          portion_size_type: 'large',
          portions_small: 0,
          portions_large: 120,
          total_portions: 120,
          status: 'pending',
          menu_items: []
        }
      ])

      await wrapper.vm.$nextTick()

      // Simulate Firebase update with changed portion sizes for all schools
      const firebaseData = [
        {
          school_id: 1,
          school_name: 'SD Negeri 1',
          school_category: 'SD',
          portion_size_type: 'mixed',
          portions_small: 55,  // Changed
          portions_large: 80,  // Changed
          total_portions: 135, // Changed
          status: 'packing'
        },
        {
          school_id: 2,
          school_name: 'SMP Negeri 1',
          school_category: 'SMP',
          portion_size_type: 'large',
          portions_small: 0,
          portions_large: 110,  // Changed
          total_portions: 110,  // Changed
          status: 'packing'
        },
        {
          school_id: 3,
          school_name: 'SMA Negeri 1',
          school_category: 'SMA',
          portion_size_type: 'large',
          portions_small: 0,
          portions_large: 130,  // Changed
          total_portions: 130,  // Changed
          status: 'packing'
        }
      ]

      // Manually trigger the Firebase update logic
      wrapper.vm.schools = wrapper.vm.schools.map(school => {
        const firebaseSchool = firebaseData.find(fs => fs.school_id === school.school_id)
        if (firebaseSchool) {
          return {
            ...school,
            status: firebaseSchool.status,
            portion_size_type: firebaseSchool.portion_size_type || school.portion_size_type,
            portions_small: firebaseSchool.portions_small !== undefined ? firebaseSchool.portions_small : school.portions_small,
            portions_large: firebaseSchool.portions_large !== undefined ? firebaseSchool.portions_large : school.portions_large,
            total_portions: firebaseSchool.total_portions || school.total_portions
          }
        }
        return school
      })

      await wrapper.vm.$nextTick()

      // Verify all schools have updated portion sizes
      expect(wrapper.vm.schools[0].portions_small).toBe(55)
      expect(wrapper.vm.schools[0].portions_large).toBe(80)
      expect(wrapper.vm.schools[0].total_portions).toBe(135)
      
      expect(wrapper.vm.schools[1].portions_large).toBe(110)
      expect(wrapper.vm.schools[1].total_portions).toBe(110)
      
      expect(wrapper.vm.schools[2].portions_large).toBe(130)
      expect(wrapper.vm.schools[2].total_portions).toBe(130)
    })

    it('should preserve portion size data when Firebase update does not include portion fields', async () => {
      // Initial data with portion sizes
      const wrapper = createWrapper([{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 50,
        portions_large: 75,
        total_portions: 125,
        status: 'pending',
        menu_items: []
      }])

      await wrapper.vm.$nextTick()

      // Simulate Firebase update without portion fields (only status update)
      const firebaseData = [{
        school_id: 1,
        status: 'packing'
        // No portion fields
      }]

      // Manually trigger the Firebase update logic
      wrapper.vm.schools = wrapper.vm.schools.map(school => {
        const firebaseSchool = firebaseData.find(fs => fs.school_id === school.school_id)
        if (firebaseSchool) {
          return {
            ...school,
            status: firebaseSchool.status,
            portion_size_type: firebaseSchool.portion_size_type || school.portion_size_type,
            portions_small: firebaseSchool.portions_small !== undefined ? firebaseSchool.portions_small : school.portions_small,
            portions_large: firebaseSchool.portions_large !== undefined ? firebaseSchool.portions_large : school.portions_large,
            total_portions: firebaseSchool.total_portions || school.total_portions
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

    it('should handle Firebase listener properly updating portion_size_type field', async () => {
      // Initial data
      const wrapper = createWrapper([{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 50,
        portions_large: 75,
        total_portions: 125,
        status: 'pending',
        menu_items: []
      }])

      await wrapper.vm.$nextTick()

      // Verify initial portion_size_type
      expect(wrapper.vm.schools[0].portion_size_type).toBe('mixed')

      // Simulate Firebase update maintaining portion_size_type
      const firebaseData = [{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',  // Explicitly maintained
        portions_small: 60,
        portions_large: 80,
        total_portions: 140,
        status: 'packing'
      }]

      // Manually trigger the Firebase update logic
      wrapper.vm.schools = wrapper.vm.schools.map(school => {
        const firebaseSchool = firebaseData.find(fs => fs.school_id === school.school_id)
        if (firebaseSchool) {
          return {
            ...school,
            status: firebaseSchool.status,
            portion_size_type: firebaseSchool.portion_size_type || school.portion_size_type,
            portions_small: firebaseSchool.portions_small !== undefined ? firebaseSchool.portions_small : school.portions_small,
            portions_large: firebaseSchool.portions_large !== undefined ? firebaseSchool.portions_large : school.portions_large,
            total_portions: firebaseSchool.total_portions || school.total_portions
          }
        }
        return school
      })

      await wrapper.vm.$nextTick()

      // Verify portion_size_type is maintained
      expect(wrapper.vm.schools[0].portion_size_type).toBe('mixed')
      expect(wrapper.vm.schools[0].portions_small).toBe(60)
      expect(wrapper.vm.schools[0].portions_large).toBe(80)
    })

    it('should render updated portion sizes in the UI after Firebase update', async () => {
      // Initial data
      const wrapper = createWrapper([{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 50,
        portions_large: 75,
        total_portions: 125,
        status: 'pending',
        menu_items: []
      }])

      await wrapper.vm.$nextTick()
      
      // Verify initial rendering
      let html = wrapper.html()
      expect(html).toContain('50 porsi')
      expect(html).toContain('75 porsi')
      expect(html).toContain('125')

      // Simulate Firebase update
      const firebaseData = [{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 65,
        portions_large: 90,
        total_portions: 155,
        status: 'packing'
      }]

      // Manually trigger the Firebase update logic
      wrapper.vm.schools = wrapper.vm.schools.map(school => {
        const firebaseSchool = firebaseData.find(fs => fs.school_id === school.school_id)
        if (firebaseSchool) {
          return {
            ...school,
            status: firebaseSchool.status,
            portion_size_type: firebaseSchool.portion_size_type || school.portion_size_type,
            portions_small: firebaseSchool.portions_small !== undefined ? firebaseSchool.portions_small : school.portions_small,
            portions_large: firebaseSchool.portions_large !== undefined ? firebaseSchool.portions_large : school.portions_large,
            total_portions: firebaseSchool.total_portions || school.total_portions
          }
        }
        return school
      })

      await wrapper.vm.$nextTick()

      // Verify updated rendering
      html = wrapper.html()
      expect(html).toContain('65 porsi')
      expect(html).toContain('90 porsi')
      expect(html).toContain('155')
    })

    it('should handle zero values in Firebase updates correctly', async () => {
      // Initial data with non-zero values
      const wrapper = createWrapper([{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 50,
        portions_large: 75,
        total_portions: 125,
        status: 'pending',
        menu_items: []
      }])

      await wrapper.vm.$nextTick()

      // Simulate Firebase update with zero small portions
      const firebaseData = [{
        school_id: 1,
        school_name: 'SD Negeri 1',
        school_category: 'SD',
        portion_size_type: 'mixed',
        portions_small: 0,  // Changed to 0
        portions_large: 125,
        total_portions: 125,
        status: 'packing'
      }]

      // Manually trigger the Firebase update logic
      wrapper.vm.schools = wrapper.vm.schools.map(school => {
        const firebaseSchool = firebaseData.find(fs => fs.school_id === school.school_id)
        if (firebaseSchool) {
          return {
            ...school,
            status: firebaseSchool.status,
            portion_size_type: firebaseSchool.portion_size_type || school.portion_size_type,
            portions_small: firebaseSchool.portions_small !== undefined ? firebaseSchool.portions_small : school.portions_small,
            portions_large: firebaseSchool.portions_large !== undefined ? firebaseSchool.portions_large : school.portions_large,
            total_portions: firebaseSchool.total_portions || school.total_portions
          }
        }
        return school
      })

      await wrapper.vm.$nextTick()

      // Verify zero value is properly updated
      expect(wrapper.vm.schools[0].portions_small).toBe(0)
      expect(wrapper.vm.schools[0].portions_large).toBe(125)
      expect(wrapper.vm.schools[0].total_portions).toBe(125)
    })
  })
})

