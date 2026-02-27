import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import KDSCookingView from '../KDSCookingView.vue'
import { getCookingToday } from '@/services/kdsService'

/**
 * End-to-End Integration Test: KDS Cooking View - Portion Size Display
 * 
 * Task 6.5.3: Test viewing allocations in KDS cooking view
 * 
 * This test simulates the complete workflow of viewing menu items in the KDS cooking view
 * with portion size information:
 * 1. Loading today's menu items from the API
 * 2. Displaying recipes with school allocations
 * 3. Showing portion size breakdown for SD schools (small and large)
 * 4. Showing only large portions for SMP/SMA schools
 * 5. Verifying labels and formatting are correct
 * 
 * This test validates Requirements 9 (Display Portion Sizes in KDS Cooking View)
 */

// Mock services
vi.mock('@/services/kdsService', () => ({
  getCookingToday: vi.fn(),
  updateCookingStatus: vi.fn()
}))

// Mock Firebase
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
    }
  }
})

describe('E2E: KDS Cooking View - Viewing Allocations with Portion Sizes', () => {
  let wrapper

  // Mock data representing realistic menu items with portion size allocations
  const mockRecipeWithSDSchool = {
    recipe_id: 1,
    name: 'Nasi Goreng Spesial',
    photo_url: 'https://example.com/nasi-goreng.jpg',
    portions_required: 125,
    status: 'pending',
    start_time: null,
    instructions: 'Tumis bumbu, masukkan nasi, aduk rata',
    items: [
      { name: 'Nasi', quantity: 500, unit: 'gram' },
      { name: 'Telur', quantity: 3, unit: 'butir' }
    ],
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

  const mockRecipeWithSMPSchool = {
    recipe_id: 2,
    name: 'Soto Ayam',
    photo_url: 'https://example.com/soto-ayam.jpg',
    portions_required: 100,
    status: 'cooking',
    start_time: 1705305600,
    instructions: 'Rebus ayam, buat kuah, sajikan dengan pelengkap',
    items: [
      { name: 'Ayam', quantity: 1, unit: 'kg' },
      { name: 'Kunyit', quantity: 2, unit: 'cm' }
    ],
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

  const mockRecipeWithSMASchool = {
    recipe_id: 3,
    name: 'Rendang Daging',
    photo_url: 'https://example.com/rendang.jpg',
    portions_required: 120,
    status: 'ready',
    start_time: 1705302000,
    instructions: 'Masak daging dengan bumbu rendang hingga empuk',
    items: [
      { name: 'Daging Sapi', quantity: 1.5, unit: 'kg' },
      { name: 'Santan', quantity: 500, unit: 'ml' }
    ],
    school_allocations: [
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

  const mockRecipeWithMultipleSchools = {
    recipe_id: 4,
    name: 'Ayam Goreng Kremes',
    photo_url: 'https://example.com/ayam-goreng.jpg',
    portions_required: 345,
    status: 'pending',
    start_time: null,
    instructions: 'Marinasi ayam, goreng hingga kecoklatan, taburi kremes',
    items: [
      { name: 'Ayam', quantity: 2, unit: 'kg' },
      { name: 'Tepung', quantity: 200, unit: 'gram' }
    ],
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

  const createWrapper = (recipes = []) => {
    // Mock API response
    getCookingToday.mockResolvedValue({
      success: true,
      data: recipes
    })

    return mount(KDSCookingView, {
      global: {
        stubs: {
          'KDSDatePicker': {
            template: '<div class="kds-date-picker-stub"></div>',
            props: ['modelValue', 'loading'],
            emits: ['change', 'update:modelValue']
          },
          'a-space': {
            template: '<div><slot /></div>',
            props: ['size']
          },
          'a-tag': {
            template: '<span class="a-tag-stub" :class="`color-${color}`"><slot /></span>',
            props: ['color']
          },
          'a-button': {
            template: '<button class="a-button-stub" :disabled="disabled" :loading="loading"><slot /></button>',
            props: ['loading', 'type', 'block', 'disabled']
          },
          'a-alert': {
            template: '<div class="a-alert-stub"><slot /></div>',
            props: ['type', 'message', 'closable', 'showIcon']
          },
          'a-spin': {
            template: '<div class="a-spin-stub"><slot /></div>',
            props: ['spinning', 'tip']
          },
          'a-empty': {
            template: '<div class="a-empty-stub">{{ description }}</div>',
            props: ['description']
          },
          'a-row': {
            template: '<div class="a-row-stub"><slot /></div>',
            props: ['gutter']
          },
          'a-col': {
            template: '<div class="a-col-stub"><slot /></div>',
            props: ['xs', 'sm', 'md', 'lg', 'xl']
          },
          'a-card': {
            template: '<div class="a-card-stub" :class="class"><div class="card-title">{{ title }}</div><div class="card-extra"><slot name="extra" /></div><div class="card-body"><slot /></div><div class="card-actions"><slot name="actions" /></div></div>',
            props: ['title', 'class']
          },
          'a-descriptions': {
            template: '<div class="a-descriptions-stub"><slot /></div>',
            props: ['column', 'size', 'bordered']
          },
          'a-descriptions-item': {
            template: '<div class="a-descriptions-item-stub"><span class="label">{{ label }}</span><span class="value"><slot /></span></div>',
            props: ['label']
          },
          'a-divider': {
            template: '<div class="a-divider-stub"><slot /></div>'
          },
          'a-list': {
            template: '<div class="a-list-stub"><div v-for="(item, index) in dataSource" :key="index" class="list-item"><slot name="renderItem" :item="item" /></div></div>',
            props: ['dataSource', 'size', 'split']
          },
          'a-list-item': {
            template: '<div class="a-list-item-stub"><slot /></div>'
          },
          'a-list-item-meta': {
            template: '<div class="a-list-item-meta-stub"><div class="meta-title"><slot name="title" /></div><div class="meta-description"><slot name="description" /></div></div>'
          },
          'a-badge': {
            template: '<span class="a-badge-stub" :data-count="count"><slot /></span>',
            props: ['count', 'numberStyle']
          },
          'WifiOutlined': { template: '<span>wifi</span>' },
          'DisconnectOutlined': { template: '<span>disconnect</span>' },
          'ReloadOutlined': { template: '<span>reload</span>' },
          'PlayCircleOutlined': { template: '<span>play</span>' },
          'CheckCircleOutlined': { template: '<span>check-circle</span>' },
          'CheckOutlined': { template: '<span>check</span>' }
        }
      }
    })
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('SD School Portion Size Display', () => {
    it('should display portion breakdown for SD school with both small and large portions', async () => {
      wrapper = createWrapper([mockRecipeWithSDSchool])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      const html = wrapper.html()
      
      // Verify school name and category are displayed
      expect(html).toContain('SD Negeri 1')
      expect(html).toContain('SD')
      
      // Verify portion breakdown section exists
      expect(html).toContain('portion-breakdown')
      
      // Verify small portion label and count
      expect(html).toContain('Kecil (Kelas 1-3)')
      expect(html).toContain('50')
      
      // Verify large portion label and count
      expect(html).toContain('Besar (Kelas 4-6)')
      expect(html).toContain('75')
      
      // Verify total portions
      expect(html).toContain('Total: 125 porsi')
    })

    it('should display visual indicators for small and large portions in SD schools', async () => {
      wrapper = createWrapper([mockRecipeWithSDSchool])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      const html = wrapper.html()
      
      // Verify portion size cards are present
      expect(html).toContain('portion-item portion-small')
      expect(html).toContain('portion-item portion-large')
      
      // Verify portion tags with appropriate styling
      expect(html).toContain('portion-tag portion-tag-small')
      expect(html).toContain('portion-tag portion-tag-large')
      
      // Verify portion icons
      expect(html).toContain('portion-icon portion-icon-small')
      expect(html).toContain('portion-icon portion-icon-large')
    })

    it('should format school allocation display correctly for SD schools', async () => {
      wrapper = createWrapper([mockRecipeWithSDSchool])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      // Test the formatSchoolAllocation function
      const allocation = mockRecipeWithSDSchool.school_allocations[0]
      const formatted = wrapper.vm.formatSchoolAllocation(allocation)
      
      expect(formatted).toContain('SD Negeri 1:')
      expect(formatted).toContain('Kecil (50)')
      expect(formatted).toContain('Besar (75)')
    })

    it('should handle SD school with only small portions', async () => {
      const recipeWithOnlySmall = {
        ...mockRecipeWithSDSchool,
        portions_required: 50,
        school_allocations: [{
          school_id: 1,
          school_name: 'SD Negeri 1',
          school_category: 'SD',
          portion_size_type: 'mixed',
          portions_small: 50,
          portions_large: 0,
          total_portions: 50
        }]
      }

      wrapper = createWrapper([recipeWithOnlySmall])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      const html = wrapper.html()
      
      // Should show small portion
      expect(html).toContain('Kecil (Kelas 1-3)')
      expect(html).toContain('50')
      
      // Should not show large portion badge (0 portions)
      const formatted = wrapper.vm.formatSchoolAllocation(recipeWithOnlySmall.school_allocations[0])
      expect(formatted).toContain('Kecil (50)')
      expect(formatted).not.toContain('Besar (0)')
    })

    it('should handle SD school with only large portions', async () => {
      const recipeWithOnlyLarge = {
        ...mockRecipeWithSDSchool,
        portions_required: 75,
        school_allocations: [{
          school_id: 1,
          school_name: 'SD Negeri 1',
          school_category: 'SD',
          portion_size_type: 'mixed',
          portions_small: 0,
          portions_large: 75,
          total_portions: 75
        }]
      }

      wrapper = createWrapper([recipeWithOnlyLarge])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      const html = wrapper.html()
      
      // Should show large portion
      expect(html).toContain('Besar (Kelas 4-6)')
      expect(html).toContain('75')
      
      // Should not show small portion badge (0 portions)
      const formatted = wrapper.vm.formatSchoolAllocation(recipeWithOnlyLarge.school_allocations[0])
      expect(formatted).toContain('Besar (75)')
      expect(formatted).not.toContain('Kecil (0)')
    })
  })

  describe('SMP School Portion Size Display', () => {
    it('should display only large portions for SMP schools', async () => {
      wrapper = createWrapper([mockRecipeWithSMPSchool])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      const html = wrapper.html()
      
      // Verify school name and category
      expect(html).toContain('SMP Negeri 1')
      expect(html).toContain('SMP')
      
      // Verify only large portion is shown
      expect(html).toContain('Besar')
      expect(html).toContain('100')
      
      // Verify small portion labels are NOT shown
      expect(html).not.toContain('Kecil (Kelas 1-3)')
      
      // Verify total portions
      expect(html).toContain('Total: 100 porsi')
    })

    it('should format school allocation display correctly for SMP schools', async () => {
      wrapper = createWrapper([mockRecipeWithSMPSchool])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      // Test the formatSchoolAllocation function
      const allocation = mockRecipeWithSMPSchool.school_allocations[0]
      const formatted = wrapper.vm.formatSchoolAllocation(allocation)
      
      expect(formatted).toContain('SMP Negeri 1:')
      expect(formatted).toContain('Besar (100)')
      expect(formatted).not.toContain('Kecil')
    })

    it('should display single portion card for SMP schools', async () => {
      wrapper = createWrapper([mockRecipeWithSMPSchool])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      const html = wrapper.html()
      
      // Should have large portion display
      expect(html).toContain('portion-item portion-large')
      
      // Should not have small portion display
      expect(html).not.toContain('portion-item portion-small')
    })
  })

  describe('SMA School Portion Size Display', () => {
    it('should display only large portions for SMA schools', async () => {
      wrapper = createWrapper([mockRecipeWithSMASchool])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      const html = wrapper.html()
      
      // Verify school name and category
      expect(html).toContain('SMA Negeri 1')
      expect(html).toContain('SMA')
      
      // Verify only large portion is shown
      expect(html).toContain('Besar')
      expect(html).toContain('120')
      
      // Verify small portion labels are NOT shown
      expect(html).not.toContain('Kecil (Kelas 1-3)')
      expect(html).not.toContain('Kelas 4-6')
      
      // Verify total portions
      expect(html).toContain('Total: 120 porsi')
    })

    it('should format school allocation display correctly for SMA schools', async () => {
      wrapper = createWrapper([mockRecipeWithSMASchool])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      // Test the formatSchoolAllocation function
      const allocation = mockRecipeWithSMASchool.school_allocations[0]
      const formatted = wrapper.vm.formatSchoolAllocation(allocation)
      
      expect(formatted).toContain('SMA Negeri 1:')
      expect(formatted).toContain('Besar (120)')
      expect(formatted).not.toContain('Kecil')
    })
  })

  describe('Multiple Schools Display', () => {
    it('should display allocations for multiple schools with correct portion information', async () => {
      wrapper = createWrapper([mockRecipeWithMultipleSchools])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      const html = wrapper.html()
      
      // Verify all schools are displayed
      expect(html).toContain('SD Negeri 1')
      expect(html).toContain('SMP Negeri 1')
      expect(html).toContain('SMA Negeri 1')
      
      // Verify SD school shows both portion sizes
      expect(html).toContain('Kecil (Kelas 1-3)')
      expect(html).toContain('Besar (Kelas 4-6)')
      
      // Verify portion counts
      expect(html).toContain('50')  // SD small
      expect(html).toContain('75')  // SD large
      expect(html).toContain('100') // SMP large
      expect(html).toContain('120') // SMA large
      
      // Verify total portions
      expect(html).toContain('Total: 125 porsi') // SD total
      expect(html).toContain('Total: 100 porsi') // SMP total
      expect(html).toContain('Total: 120 porsi') // SMA total
    })

    it('should display correct school categories for all schools', async () => {
      wrapper = createWrapper([mockRecipeWithMultipleSchools])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      const html = wrapper.html()
      
      // Verify school category tags are displayed
      expect(html).toContain('SD')
      expect(html).toContain('SMP')
      expect(html).toContain('SMA')
    })
  })

  describe('Recipe Information Display', () => {
    it('should display recipe details along with portion size allocations', async () => {
      wrapper = createWrapper([mockRecipeWithSDSchool])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      const html = wrapper.html()
      
      // Verify recipe name
      expect(html).toContain('Nasi Goreng Spesial')
      
      // Verify total portions required
      expect(html).toContain('125 porsi')
      
      // Verify ingredients are displayed
      expect(html).toContain('Nasi')
      expect(html).toContain('500')
      expect(html).toContain('gram')
      
      // Verify instructions are displayed
      expect(html).toContain('Tumis bumbu, masukkan nasi, aduk rata')
    })

    it('should display recipe status correctly', async () => {
      wrapper = createWrapper([
        mockRecipeWithSDSchool,
        mockRecipeWithSMPSchool,
        mockRecipeWithSMASchool
      ])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      const html = wrapper.html()
      
      // Verify different statuses are displayed
      expect(html).toContain('status-pending')
      expect(html).toContain('status-cooking')
      expect(html).toContain('status-ready')
    })
  })

  describe('Empty State and Error Handling', () => {
    it('should display empty message when no recipes are available', async () => {
      wrapper = createWrapper([])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      const html = wrapper.html()
      
      // Should show empty state
      expect(html).toContain('a-empty-stub')
    })

    it('should handle recipe with no school allocations', async () => {
      const recipeWithNoAllocations = {
        ...mockRecipeWithSDSchool,
        school_allocations: []
      }

      wrapper = createWrapper([recipeWithNoAllocations])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      const html = wrapper.html()
      
      // Should show recipe but indicate no allocations
      expect(html).toContain('Nasi Goreng Spesial')
      expect(html).toContain('Tidak ada alokasi sekolah')
    })

    it('should handle API errors gracefully', async () => {
      getCookingToday.mockResolvedValue({
        success: false,
        message: 'Failed to load data'
      })

      wrapper = createWrapper([])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      // Should set error state
      expect(wrapper.vm.error).toBe('Failed to load data')
    })
  })

  describe('Visual Styling and Formatting', () => {
    it('should apply correct CSS classes for portion size display', async () => {
      wrapper = createWrapper([mockRecipeWithSDSchool])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      const html = wrapper.html()
      
      // Verify CSS classes are applied
      expect(html).toContain('portion-breakdown')
      expect(html).toContain('portion-item')
      expect(html).toContain('portion-tag')
      expect(html).toContain('portion-icon')
      expect(html).toContain('portion-total')
    })

    it('should display school allocation title with proper formatting', async () => {
      wrapper = createWrapper([mockRecipeWithSDSchool])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      const html = wrapper.html()
      
      // Verify school allocation title section
      expect(html).toContain('school-allocation-title')
      expect(html).toContain('school-name-text')
    })
  })

  describe('Integration with Recipe Status', () => {
    it('should display portion sizes for recipes in pending status', async () => {
      wrapper = createWrapper([mockRecipeWithSDSchool])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      const html = wrapper.html()
      
      expect(wrapper.vm.recipes[0].status).toBe('pending')
      expect(html).toContain('SD Negeri 1')
      expect(html).toContain('50')
      expect(html).toContain('75')
    })

    it('should display portion sizes for recipes in cooking status', async () => {
      wrapper = createWrapper([mockRecipeWithSMPSchool])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      const html = wrapper.html()
      
      expect(wrapper.vm.recipes[0].status).toBe('cooking')
      expect(html).toContain('SMP Negeri 1')
      expect(html).toContain('100')
    })

    it('should display portion sizes for recipes in ready status', async () => {
      wrapper = createWrapper([mockRecipeWithSMASchool])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      const html = wrapper.html()
      
      expect(wrapper.vm.recipes[0].status).toBe('ready')
      expect(html).toContain('SMA Negeri 1')
      expect(html).toContain('120')
    })
  })

  describe('Requirement Validation', () => {
    it('validates Requirement 9.1: Display school allocations with portion size information', async () => {
      wrapper = createWrapper([mockRecipeWithSDSchool])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      const html = wrapper.html()
      
      // School allocations are displayed
      expect(html).toContain('SD Negeri 1')
      // With portion size information
      expect(html).toContain('portion-breakdown')
    })

    it('validates Requirement 9.2: Display both small and large portion counts for SD schools', async () => {
      wrapper = createWrapper([mockRecipeWithSDSchool])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      const html = wrapper.html()
      
      // Both small and large portions are displayed
      expect(html).toContain('Kecil (Kelas 1-3)')
      expect(html).toContain('Besar (Kelas 4-6)')
      expect(html).toContain('50')
      expect(html).toContain('75')
    })

    it('validates Requirement 9.3: Display only large portion count for SMP/SMA schools', async () => {
      wrapper = createWrapper([mockRecipeWithSMPSchool, mockRecipeWithSMASchool])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      const html = wrapper.html()
      
      // Only large portions are displayed for SMP/SMA
      expect(html).toContain('Besar')
      expect(html).toContain('100') // SMP
      expect(html).toContain('120') // SMA
      // Small portion labels should not appear
      const smpFormatted = wrapper.vm.formatSchoolAllocation(mockRecipeWithSMPSchool.school_allocations[0])
      const smaFormatted = wrapper.vm.formatSchoolAllocation(mockRecipeWithSMASchool.school_allocations[0])
      expect(smpFormatted).not.toContain('Kecil')
      expect(smaFormatted).not.toContain('Kecil')
    })

    it('validates Requirement 9.4: Display total portions as sum of all portions', async () => {
      wrapper = createWrapper([mockRecipeWithSDSchool])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      const html = wrapper.html()
      
      // Total portions displayed
      expect(html).toContain('Total: 125 porsi')
      
      // Verify it matches sum of small + large
      const allocation = mockRecipeWithSDSchool.school_allocations[0]
      expect(allocation.total_portions).toBe(allocation.portions_small + allocation.portions_large)
    })

    it('validates Requirement 9.5: Clearly label portion sizes with grade information', async () => {
      wrapper = createWrapper([mockRecipeWithSDSchool])
      
      await wrapper.vm.$nextTick()
      await wrapper.vm.loadData()
      await wrapper.vm.$nextTick()

      const html = wrapper.html()
      
      // Labels include grade information
      expect(html).toContain('Kecil (Kelas 1-3)')
      expect(html).toContain('Besar (Kelas 4-6)')
    })
  })
})
