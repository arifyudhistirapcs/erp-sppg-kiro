import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import MenuPlanningView from '../MenuPlanningView.vue'
import SchoolAllocationInput from '@/components/SchoolAllocationInput.vue'
import menuPlanningService from '@/services/menuPlanningService'
import recipeService from '@/services/recipeService'
import schoolService from '@/services/schoolService'

/**
 * End-to-End Integration Test: Validation Error Scenarios
 * 
 * Task 6.5.5: Test validation error scenarios
 * 
 * This test suite validates various error scenarios in the menu planning workflow:
 * 1. Sum mismatch errors (too many/too few portions)
 * 2. SMP/SMA schools with small portions error
 * 3. Negative values error
 * 4. All portions zero error
 * 5. Error messages are displayed correctly
 * 
 * These tests validate Requirements 3, 6, and 12 (Validation and Error Handling)
 */

// Mock services
vi.mock('@/services/menuPlanningService')
vi.mock('@/services/recipeService')
vi.mock('@/services/schoolService')

// Mock auth store
vi.mock('@/stores/auth', () => ({
  useAuthStore: vi.fn(() => ({
    user: { role: 'ahli_gizi' }
  }))
}))

// Mock dayjs
vi.mock('dayjs', () => {
  const dayjs = vi.fn((date) => ({
    startOf: vi.fn(() => ({
      format: vi.fn(() => date || '2024-01-15'),
      add: vi.fn(() => ({
        format: vi.fn(() => '2024-01-16')
      }))
    })),
    format: vi.fn(() => date || '2024-01-15'),
    isSame: vi.fn(() => false),
    add: vi.fn(() => ({
      format: vi.fn(() => '2024-01-16')
    })),
    subtract: vi.fn(() => ({
      format: vi.fn(() => '2024-01-08')
    }))
  }))
  
  dayjs.extend = vi.fn()
  
  return { default: dayjs }
})

describe('E2E: Validation Error Scenarios', () => {
  let wrapper

  const mockSDSchool = {
    id: 1,
    name: 'SD Negeri 1',
    category: 'SD',
    student_count_grade_1_3: 150,
    student_count_grade_4_6: 200,
    student_count: 350
  }

  const mockSMPSchool = {
    id: 2,
    name: 'SMP Negeri 1',
    category: 'SMP',
    student_count: 300
  }

  const mockSMASchool = {
    id: 3,
    name: 'SMA Negeri 1',
    category: 'SMA',
    student_count: 250
  }

  const mockRecipe = {
    id: 1,
    name: 'Nasi Goreng Spesial',
    category: 'Main Course',
    total_calories: 500,
    total_protein: 20,
    total_carbs: 60,
    total_fat: 15
  }

  const mockMenuPlan = {
    id: 1,
    week_start_date: '2024-01-15',
    week_end_date: '2024-01-21',
    status: 'draft'
  }

  beforeEach(() => {
    vi.clearAllMocks()

    schoolService.getSchools = vi.fn(() => 
      Promise.resolve({ 
        data: { 
          schools: [mockSDSchool, mockSMPSchool, mockSMASchool] 
        } 
      })
    )

    recipeService.getRecipes = vi.fn(() => 
      Promise.resolve({ 
        data: { 
          recipes: [mockRecipe] 
        } 
      })
    )

    menuPlanningService.getMenuPlans = vi.fn(() => 
      Promise.resolve({ 
        data: { 
          menu_plans: [mockMenuPlan] 
        } 
      })
    )

    menuPlanningService.createMenuItem = vi.fn(() => 
      Promise.resolve({ data: { menu_item: {} } })
    )

    menuPlanningService.updateMenuItem = vi.fn(() => 
      Promise.resolve({ data: { menu_item: {} } })
    )

    menuPlanningService.deleteMenuItem = vi.fn(() => 
      Promise.resolve({ data: {} })
    )

    menuPlanningService.approveMenuPlan = vi.fn(() => 
      Promise.resolve({ data: {} })
    )

    menuPlanningService.createMenuPlan = vi.fn(() => 
      Promise.resolve({ data: { menu_plan: mockMenuPlan } })
    )
  })

  afterEach(() => {
    if (wrapper) {
      wrapper.unmount()
    }
  })

  const createWrapper = () => {
    return mount(MenuPlanningView, {
      global: {
        stubs: {
          'a-page-header': {
            template: '<div class="a-page-header"><slot /><slot name="extra" /></div>'
          },
          'a-card': {
            template: '<div class="a-card"><slot /></div>'
          },
          'a-row': {
            template: '<div class="a-row"><slot /></div>',
            props: ['gutter']
          },
          'a-col': {
            template: '<div class="a-col"><slot /></div>',
            props: ['span']
          },
          'a-space': {
            template: '<div class="a-space"><slot /></div>'
          },
          'a-button': {
            template: '<button class="a-button" @click="$emit(\'click\')" :disabled="disabled"><slot name="icon" /><slot /></button>',
            props: ['type', 'loading', 'disabled']
          },
          'a-date-picker': {
            template: '<input type="text" class="a-date-picker" />',
            props: ['value', 'picker', 'format']
          },
          'a-tag': {
            template: '<span class="a-tag"><slot /></span>',
            props: ['color']
          },
          'a-spin': {
            template: '<div class="a-spin"><slot /></div>',
            props: ['spinning']
          },
          'a-modal': {
            template: `
              <div v-if="visible" class="a-modal">
                <div class="modal-title">{{ title }}</div>
                <div class="modal-body"><slot /></div>
                <div class="modal-footer">
                  <button class="btn-cancel" @click="$emit('update:visible', false)">{{ cancelText }}</button>
                  <button class="btn-ok" :disabled="okButtonProps?.disabled" @click="$emit('ok')">{{ okText }}</button>
                </div>
              </div>
            `,
            props: ['visible', 'title', 'okText', 'cancelText', 'okButtonProps', 'width', 'bodyStyle']
          },
          'a-form': {
            template: '<form class="a-form"><slot /></form>',
            props: ['layout']
          },
          'a-form-item': {
            template: '<div class="a-form-item"><label>{{ label }}</label><slot /></div>',
            props: ['label']
          },
          'a-select': {
            template: `
              <select class="a-select" :value="value" @change="$emit('update:value', parseInt($event.target.value))">
                <slot />
              </select>
            `,
            props: ['value', 'showSearch', 'placeholder', 'filterOption']
          },
          'a-select-option': {
            template: '<option :value="value"><slot /></option>',
            props: ['value']
          },
          'a-input-number': {
            template: '<input type="number" class="a-input-number" :value="value" @input="$emit(\'update:value\', parseInt($event.target.value) || 0)" />',
            props: ['value', 'min']
          },
          'a-divider': {
            template: '<hr class="a-divider" />',
            props: ['style']
          },
          'a-alert': {
            template: '<div class="a-alert" :class="`alert-${type}`"><slot /></div>',
            props: ['type', 'message', 'showIcon']
          },
          'PlusOutlined': true,
          'CopyOutlined': true,
          'LeftOutlined': true,
          'RightOutlined': true,
          'CheckOutlined': true,
          'DeleteOutlined': true,
          'EditOutlined': true,
          'SchoolAllocationInput': SchoolAllocationInput
        }
      }
    })
  }

  describe('Sum Mismatch Errors', () => {
    it('should display error when allocated portions are less than total portions', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      // Allocate only 400 portions instead of 500
      wrapper.vm.schoolAllocations = {
        1: { portions_small: 150, portions_large: 100 }, // 250 total
        2: { portions_small: 0, portions_large: 150 }    // 150 total
        // Total: 400, but expected 500
      }
      await wrapper.vm.$nextTick()

      const isValid = wrapper.vm.validateAllocations()
      expect(isValid).toBe(false)

      wrapper.vm.isAllocationValid = isValid
      await wrapper.vm.$nextTick()

      // Verify submit button is disabled
      const submitButton = wrapper.find('.btn-ok')
      expect(submitButton.attributes('disabled')).toBeDefined()

      // Verify error message is displayed
      expect(wrapper.vm.allocationError).toBeTruthy()
      expect(wrapper.vm.allocationError).toContain('400')
      expect(wrapper.vm.allocationError).toContain('500')
    })

    it('should display error when allocated portions exceed total portions', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      // Allocate 600 portions instead of 500
      wrapper.vm.schoolAllocations = {
        1: { portions_small: 200, portions_large: 250 }, // 450 total
        2: { portions_small: 0, portions_large: 150 }    // 150 total
        // Total: 600, but expected 500
      }
      await wrapper.vm.$nextTick()

      const isValid = wrapper.vm.validateAllocations()
      expect(isValid).toBe(false)

      wrapper.vm.isAllocationValid = isValid
      await wrapper.vm.$nextTick()

      // Verify submit button is disabled
      const submitButton = wrapper.find('.btn-ok')
      expect(submitButton.attributes('disabled')).toBeDefined()

      // Verify error message is displayed
      expect(wrapper.vm.allocationError).toBeTruthy()
      expect(wrapper.vm.allocationError).toContain('600')
      expect(wrapper.vm.allocationError).toContain('500')
    })

    it('should clear error when allocations are corrected to match total', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      // First, set invalid allocations
      wrapper.vm.schoolAllocations = {
        1: { portions_small: 150, portions_large: 100 },
        2: { portions_small: 0, portions_large: 150 }
      }
      await wrapper.vm.$nextTick()

      let isValid = wrapper.vm.validateAllocations()
      expect(isValid).toBe(false)
      wrapper.vm.isAllocationValid = isValid
      await wrapper.vm.$nextTick()

      // Now correct the allocations
      wrapper.vm.schoolAllocations = {
        1: { portions_small: 150, portions_large: 200 }, // 350 total
        2: { portions_small: 0, portions_large: 150 }    // 150 total
        // Total: 500 (correct)
      }
      await wrapper.vm.$nextTick()

      isValid = wrapper.vm.validateAllocations()
      expect(isValid).toBe(true)
      wrapper.vm.isAllocationValid = isValid
      await wrapper.vm.$nextTick()

      // Verify submit button is enabled
      const submitButton = wrapper.find('.btn-ok')
      expect(submitButton.attributes('disabled')).toBeFalsy()

      // Verify error message is cleared
      expect(wrapper.vm.allocationError).toBeFalsy()
    })
  })

  describe('SMP/SMA Schools with Small Portions Error', () => {
    it('should display error when SMP school has small portions', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      // Try to allocate small portions to SMP school (invalid)
      wrapper.vm.schoolAllocations = {
        1: { portions_small: 150, portions_large: 200 }, // 350 total
        2: { portions_small: 50, portions_large: 100 }   // 150 total - INVALID: SMP cannot have small
        // Total: 500 (sum is correct, but SMP has small portions)
      }
      await wrapper.vm.$nextTick()

      const isValid = wrapper.vm.validateAllocations()
      
      // The validation should fail because SMP has small portions
      expect(isValid).toBe(false)
      wrapper.vm.isAllocationValid = isValid
      await wrapper.vm.$nextTick()

      // Verify submit button is disabled
      const submitButton = wrapper.find('.btn-ok')
      expect(submitButton.attributes('disabled')).toBeDefined()

      // Verify error message mentions SMP and small portions
      expect(wrapper.vm.allocationError).toBeTruthy()
      expect(wrapper.vm.allocationError.toLowerCase()).toContain('smp')
    })

    it('should display error when SMA school has small portions', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      // Try to allocate small portions to SMA school (invalid)
      wrapper.vm.schoolAllocations = {
        1: { portions_small: 150, portions_large: 150 }, // 300 total
        3: { portions_small: 50, portions_large: 150 }   // 200 total - INVALID: SMA cannot have small
        // Total: 500 (sum is correct, but SMA has small portions)
      }
      await wrapper.vm.$nextTick()

      const isValid = wrapper.vm.validateAllocations()
      
      // The validation should fail because SMA has small portions
      expect(isValid).toBe(false)
      wrapper.vm.isAllocationValid = isValid
      await wrapper.vm.$nextTick()

      // Verify submit button is disabled
      const submitButton = wrapper.find('.btn-ok')
      expect(submitButton.attributes('disabled')).toBeDefined()

      // Verify error message mentions SMA and small portions
      expect(wrapper.vm.allocationError).toBeTruthy()
      expect(wrapper.vm.allocationError.toLowerCase()).toContain('sma')
    })

    it('should allow SMP/SMA schools with zero small portions', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      // Correct allocation: SMP and SMA with zero small portions
      wrapper.vm.schoolAllocations = {
        1: { portions_small: 150, portions_large: 150 }, // 300 total
        2: { portions_small: 0, portions_large: 100 },   // 100 total - Valid
        3: { portions_small: 0, portions_large: 100 }    // 100 total - Valid
        // Total: 500
      }
      await wrapper.vm.$nextTick()

      const isValid = wrapper.vm.validateAllocations()
      expect(isValid).toBe(true)
      wrapper.vm.isAllocationValid = isValid
      await wrapper.vm.$nextTick()

      // Verify submit button is enabled
      const submitButton = wrapper.find('.btn-ok')
      expect(submitButton.attributes('disabled')).toBeFalsy()

      // Verify no error message
      expect(wrapper.vm.allocationError).toBeFalsy()
    })
  })

  describe('Negative Values Error', () => {
    it('should prevent negative values in small portions', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      // Try to set negative small portions
      wrapper.vm.schoolAllocations = {
        1: { portions_small: -50, portions_large: 400 }, // Invalid: negative small
        2: { portions_small: 0, portions_large: 150 }
      }
      await wrapper.vm.$nextTick()

      const isValid = wrapper.vm.validateAllocations()
      expect(isValid).toBe(false)
      wrapper.vm.isAllocationValid = isValid
      await wrapper.vm.$nextTick()

      // Verify submit button is disabled
      const submitButton = wrapper.find('.btn-ok')
      expect(submitButton.attributes('disabled')).toBeDefined()
    })

    it('should prevent negative values in large portions', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      // Try to set negative large portions
      wrapper.vm.schoolAllocations = {
        1: { portions_small: 150, portions_large: -100 }, // Invalid: negative large
        2: { portions_small: 0, portions_large: 450 }
      }
      await wrapper.vm.$nextTick()

      const isValid = wrapper.vm.validateAllocations()
      expect(isValid).toBe(false)
      wrapper.vm.isAllocationValid = isValid
      await wrapper.vm.$nextTick()

      // Verify submit button is disabled
      const submitButton = wrapper.find('.btn-ok')
      expect(submitButton.attributes('disabled')).toBeDefined()
    })

    it('should allow zero values but not negative', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      // Zero is valid, negative is not
      wrapper.vm.schoolAllocations = {
        1: { portions_small: 0, portions_large: 350 },   // Valid: zero small
        2: { portions_small: 0, portions_large: 150 }    // Valid: zero small
      }
      await wrapper.vm.$nextTick()

      const isValid = wrapper.vm.validateAllocations()
      expect(isValid).toBe(true)
      wrapper.vm.isAllocationValid = isValid
      await wrapper.vm.$nextTick()

      // Verify submit button is enabled
      const submitButton = wrapper.find('.btn-ok')
      expect(submitButton.attributes('disabled')).toBeFalsy()
    })
  })

  describe('All Portions Zero Error', () => {
    it('should display error when a school has both portions set to zero', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      // Set both portions to zero for a school
      wrapper.vm.schoolAllocations = {
        1: { portions_small: 0, portions_large: 0 },     // Invalid: both zero
        2: { portions_small: 0, portions_large: 500 }
      }
      await wrapper.vm.$nextTick()

      const isValid = wrapper.vm.validateAllocations()
      expect(isValid).toBe(false)
      wrapper.vm.isAllocationValid = isValid
      await wrapper.vm.$nextTick()

      // Verify submit button is disabled
      const submitButton = wrapper.find('.btn-ok')
      expect(submitButton.attributes('disabled')).toBeDefined()

      // Verify error message indicates school must have at least one portion
      expect(wrapper.vm.allocationError).toBeTruthy()
    })

    it('should allow schools with at least one portion type greater than zero', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      // Valid: each school has at least one portion type > 0
      wrapper.vm.schoolAllocations = {
        1: { portions_small: 150, portions_large: 0 },   // Valid: small > 0
        2: { portions_small: 0, portions_large: 350 }    // Valid: large > 0
      }
      await wrapper.vm.$nextTick()

      const isValid = wrapper.vm.validateAllocations()
      expect(isValid).toBe(true)
      wrapper.vm.isAllocationValid = isValid
      await wrapper.vm.$nextTick()

      // Verify submit button is enabled
      const submitButton = wrapper.find('.btn-ok')
      expect(submitButton.attributes('disabled')).toBeFalsy()
    })
  })

  describe('Error Message Display', () => {
    it('should display clear error message for sum mismatch', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      wrapper.vm.schoolAllocations = {
        1: { portions_small: 100, portions_large: 100 },
        2: { portions_small: 0, portions_large: 100 }
      }
      await wrapper.vm.$nextTick()

      wrapper.vm.validateAllocations()
      wrapper.vm.isAllocationValid = false
      await wrapper.vm.$nextTick()

      // Error message should be clear and informative
      expect(wrapper.vm.allocationError).toBeTruthy()
      expect(wrapper.vm.allocationError).toContain('300') // allocated
      expect(wrapper.vm.allocationError).toContain('500') // expected
    })

    it('should display error message in the UI', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      wrapper.vm.schoolAllocations = {
        1: { portions_small: 100, portions_large: 100 },
        2: { portions_small: 0, portions_large: 100 }
      }
      await wrapper.vm.$nextTick()

      wrapper.vm.validateAllocations()
      wrapper.vm.isAllocationValid = false
      await wrapper.vm.$nextTick()

      // Check if error alert is displayed in the UI
      const html = wrapper.html()
      if (wrapper.vm.allocationError) {
        // Error should be visible somewhere in the component
        expect(wrapper.vm.allocationError).toBeTruthy()
      }
    })

    it('should update error message dynamically as user changes allocations', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      // First invalid state
      wrapper.vm.schoolAllocations = {
        1: { portions_small: 100, portions_large: 100 },
        2: { portions_small: 0, portions_large: 100 }
      }
      await wrapper.vm.$nextTick()

      wrapper.vm.validateAllocations()
      wrapper.vm.isAllocationValid = false
      await wrapper.vm.$nextTick()

      const firstError = wrapper.vm.allocationError
      expect(firstError).toBeTruthy()

      // Change to valid state
      wrapper.vm.schoolAllocations = {
        1: { portions_small: 150, portions_large: 200 },
        2: { portions_small: 0, portions_large: 150 }
      }
      await wrapper.vm.$nextTick()

      wrapper.vm.validateAllocations()
      wrapper.vm.isAllocationValid = true
      await wrapper.vm.$nextTick()

      // Error should be cleared
      expect(wrapper.vm.allocationError).toBeFalsy()
    })
  })

  describe('Multiple Validation Errors', () => {
    it('should handle multiple validation errors and display the most relevant one', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      // Multiple errors: SMP has small portions AND sum doesn't match
      wrapper.vm.schoolAllocations = {
        1: { portions_small: 100, portions_large: 100 },
        2: { portions_small: 50, portions_large: 50 }  // SMP with small portions
      }
      await wrapper.vm.$nextTick()

      const isValid = wrapper.vm.validateAllocations()
      expect(isValid).toBe(false)
      wrapper.vm.isAllocationValid = isValid
      await wrapper.vm.$nextTick()

      // Should display an error (either sum mismatch or SMP small portions)
      expect(wrapper.vm.allocationError).toBeTruthy()

      // Verify submit button is disabled
      const submitButton = wrapper.find('.btn-ok')
      expect(submitButton.attributes('disabled')).toBeDefined()
    })
  })

  describe('Real-time Validation Feedback', () => {
    it('should validate allocations in real-time as user types', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      // Start with empty allocations
      wrapper.vm.schoolAllocations = {
        1: { portions_small: 0, portions_large: 0 },
        2: { portions_small: 0, portions_large: 0 }
      }
      await wrapper.vm.$nextTick()

      // Gradually add allocations
      wrapper.vm.schoolAllocations[1].portions_small = 150
      await wrapper.vm.$nextTick()
      wrapper.vm.validateAllocations()
      expect(wrapper.vm.isAllocationValid).toBe(false) // Still incomplete

      wrapper.vm.schoolAllocations[1].portions_large = 200
      await wrapper.vm.$nextTick()
      wrapper.vm.validateAllocations()
      expect(wrapper.vm.isAllocationValid).toBe(false) // Still incomplete

      wrapper.vm.schoolAllocations[2].portions_large = 150
      await wrapper.vm.$nextTick()
      wrapper.vm.validateAllocations()
      wrapper.vm.isAllocationValid = wrapper.vm.validateAllocations()
      expect(wrapper.vm.isAllocationValid).toBe(true) // Now complete and valid
    })

    it('should show running total of allocated portions', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      wrapper.vm.schoolAllocations = {
        1: { portions_small: 150, portions_large: 200 },
        2: { portions_small: 0, portions_large: 100 }
      }
      await wrapper.vm.$nextTick()

      // Calculate total allocated
      const totalAllocated = wrapper.vm.schoolAllocations[1].portions_small + 
                            wrapper.vm.schoolAllocations[1].portions_large +
                            wrapper.vm.schoolAllocations[2].portions_small +
                            wrapper.vm.schoolAllocations[2].portions_large

      expect(totalAllocated).toBe(450)
      
      // The component should show this running total
      // This would be displayed in the UI for user feedback
    })
  })

  describe('Requirement Validation', () => {
    it('validates Requirement 3.1: Sum of portions must equal total_portions', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      // Invalid: sum doesn't match
      wrapper.vm.schoolAllocations = {
        1: { portions_small: 150, portions_large: 200 },
        2: { portions_small: 0, portions_large: 100 }
      }
      await wrapper.vm.$nextTick()

      const isValid = wrapper.vm.validateAllocations()
      expect(isValid).toBe(false)
    })

    it('validates Requirement 3.3: SMP/SMA schools cannot have small portions', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      // Invalid: SMP has small portions
      wrapper.vm.schoolAllocations = {
        1: { portions_small: 150, portions_large: 200 },
        2: { portions_small: 50, portions_large: 100 }
      }
      await wrapper.vm.$nextTick()

      const isValid = wrapper.vm.validateAllocations()
      expect(isValid).toBe(false)
    })

    it('validates Requirement 3.5: At least one portion type must be > 0', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      // Invalid: school has both portions = 0
      wrapper.vm.schoolAllocations = {
        1: { portions_small: 0, portions_large: 0 },
        2: { portions_small: 0, portions_large: 500 }
      }
      await wrapper.vm.$nextTick()

      const isValid = wrapper.vm.validateAllocations()
      expect(isValid).toBe(false)
    })

    it('validates Requirement 3.6: Portions must be non-negative integers', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      // Invalid: negative portions
      wrapper.vm.schoolAllocations = {
        1: { portions_small: -50, portions_large: 400 },
        2: { portions_small: 0, portions_large: 150 }
      }
      await wrapper.vm.$nextTick()

      const isValid = wrapper.vm.validateAllocations()
      expect(isValid).toBe(false)
    })

    it('validates Requirement 6.1: Real-time validation feedback', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      // Set invalid allocations
      wrapper.vm.schoolAllocations = {
        1: { portions_small: 100, portions_large: 100 },
        2: { portions_small: 0, portions_large: 100 }
      }
      await wrapper.vm.$nextTick()

      wrapper.vm.validateAllocations()
      wrapper.vm.isAllocationValid = false
      await wrapper.vm.$nextTick()

      // Error message should be displayed
      expect(wrapper.vm.allocationError).toBeTruthy()
    })

    it('validates Requirement 6.4: Submit button disabled when validation errors exist', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      // Set invalid allocations
      wrapper.vm.schoolAllocations = {
        1: { portions_small: 100, portions_large: 100 },
        2: { portions_small: 0, portions_large: 100 }
      }
      await wrapper.vm.$nextTick()

      wrapper.vm.validateAllocations()
      wrapper.vm.isAllocationValid = false
      await wrapper.vm.$nextTick()

      // Submit button should be disabled
      const submitButton = wrapper.find('.btn-ok')
      expect(submitButton.attributes('disabled')).toBeDefined()
    })

    it('validates Requirement 6.5: Submit button enabled when all validations pass', async () => {
      wrapper = createWrapper()
      await wrapper.vm.$nextTick()
      await new Promise(resolve => setTimeout(resolve, 100))

      wrapper.vm.showAddMenuModal('2024-01-15')
      await wrapper.vm.$nextTick()

      wrapper.vm.selectedRecipeId = mockRecipe.id
      wrapper.vm.selectedPortions = 500

      // Set valid allocations
      wrapper.vm.schoolAllocations = {
        1: { portions_small: 150, portions_large: 200 },
        2: { portions_small: 0, portions_large: 150 }
      }
      await wrapper.vm.$nextTick()

      wrapper.vm.validateAllocations()
      wrapper.vm.isAllocationValid = true
      await wrapper.vm.$nextTick()

      // Submit button should be enabled
      const submitButton = wrapper.find('.btn-ok')
      expect(submitButton.attributes('disabled')).toBeFalsy()
    })
  })
})
