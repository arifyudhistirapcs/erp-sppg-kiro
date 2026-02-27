import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import SchoolAllocationInput from '../SchoolAllocationInput.vue'

// Mock ant-design-vue components
vi.mock('ant-design-vue', async () => {
  const actual = await vi.importActual('ant-design-vue')
  return {
    ...actual
  }
})

describe('SchoolAllocationInput', () => {
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

  beforeEach(() => {
    vi.clearAllMocks()
  })

  const createWrapper = (props = {}) => {
    return mount(SchoolAllocationInput, {
      props: {
        schools: [],
        totalPortions: 500,
        modelValue: {},
        ...props
      },
      global: {
        stubs: {
          'a-input-number': {
            template: '<input type="number" :value="value" @input="$emit(\'update:value\', parseInt($event.target.value) || 0)" />',
            props: ['value', 'min', 'max', 'placeholder']
          },
          'a-tag': {
            template: '<span class="a-tag"><slot /></span>',
            props: ['color', 'size']
          },
          'a-alert': {
            template: '<div class="a-alert"><slot name="message">{{ message }}</slot></div>',
            props: ['message', 'type', 'showIcon', 'closable']
          },
          'a-divider': {
            template: '<hr class="a-divider"><slot /></hr>',
            props: ['style']
          },
          'CheckCircleOutlined': true,
          'ExclamationCircleOutlined': true
        }
      }
    })
  }

  describe('Component Rendering', () => {
    it('should render the component with basic structure', () => {
      wrapper = createWrapper()
      
      expect(wrapper.exists()).toBe(true)
      expect(wrapper.find('.school-allocation-input').exists()).toBe(true)
      expect(wrapper.find('.allocation-header').exists()).toBe(true)
      expect(wrapper.find('.allocation-summary').exists()).toBe(true)
    })

    it('should display allocation summary with correct initial values', () => {
      wrapper = createWrapper({ totalPortions: 500 })
      
      const summaryText = wrapper.find('.summary-text')
      expect(summaryText.exists()).toBe(true)
      expect(summaryText.text()).toContain('0 / 500 porsi')
    })

    it('should render schools list when schools are provided', () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool]
      })
      
      const schoolRows = wrapper.findAll('.school-row')
      expect(schoolRows).toHaveLength(2)
    })
  })

  describe('SD School Rendering', () => {
    it('should render both small and large portion fields for SD schools', () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      const portionFields = wrapper.findAll('.portion-field')
      expect(portionFields).toHaveLength(2)
      
      const labels = wrapper.findAll('.portion-label')
      expect(labels[0].text()).toContain('Kecil')
      expect(labels[1].text()).toContain('Besar')
    })

    it('should display student count context for SD schools', () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      const schoolMeta = wrapper.find('.school-meta')
      expect(schoolMeta.text()).toContain('Kelas 1-3: 150 siswa')
      expect(schoolMeta.text()).toContain('Kelas 4-6: 200 siswa')
    })

    it('should display SD category tag with correct color', () => {
      wrapper = createWrapper({
        schools: [mockSDSchool]
      })
      
      const tags = wrapper.findAll('.a-tag')
      // Find the tag that contains the school category (not the validation tag)
      const categoryTag = tags.find(tag => tag.text() === 'SD')
      expect(categoryTag).toBeTruthy()
    })

    it('should handle SD schools with missing student count data', () => {
      const schoolWithoutCounts = {
        ...mockSDSchool,
        student_count_grade_1_3: null,
        student_count_grade_4_6: null
      }
      
      wrapper = createWrapper({
        schools: [schoolWithoutCounts]
      })
      
      const schoolMeta = wrapper.find('.school-meta')
      expect(schoolMeta.text()).toContain('Kelas 1-3: 0 siswa')
      expect(schoolMeta.text()).toContain('Kelas 4-6: 0 siswa')
    })
  })

  describe('SMP School Rendering', () => {
    it('should render only large portion field for SMP schools', () => {
      wrapper = createWrapper({
        schools: [mockSMPSchool],
        totalPortions: 500
      })
      
      const portionFields = wrapper.findAll('.portion-field')
      expect(portionFields).toHaveLength(1)
      
      const label = wrapper.find('.portion-label')
      expect(label.text()).toBe('Besar')
    })

    it('should display student count context for SMP schools', () => {
      wrapper = createWrapper({
        schools: [mockSMPSchool],
        totalPortions: 500
      })
      
      const schoolMeta = wrapper.find('.school-meta')
      expect(schoolMeta.text()).toContain('300 siswa')
      expect(schoolMeta.text()).not.toContain('Kelas')
    })

    it('should display SMP category tag with correct color', () => {
      wrapper = createWrapper({
        schools: [mockSMPSchool]
      })
      
      const tags = wrapper.findAll('.a-tag')
      const categoryTag = tags.find(tag => tag.text() === 'SMP')
      expect(categoryTag).toBeTruthy()
    })
  })

  describe('SMA School Rendering', () => {
    it('should render only large portion field for SMA schools', () => {
      wrapper = createWrapper({
        schools: [mockSMASchool],
        totalPortions: 500
      })
      
      const portionFields = wrapper.findAll('.portion-field')
      expect(portionFields).toHaveLength(1)
      
      const label = wrapper.find('.portion-label')
      expect(label.text()).toBe('Besar')
    })

    it('should display student count context for SMA schools', () => {
      wrapper = createWrapper({
        schools: [mockSMASchool],
        totalPortions: 500
      })
      
      const schoolMeta = wrapper.find('.school-meta')
      expect(schoolMeta.text()).toContain('250 siswa')
      expect(schoolMeta.text()).not.toContain('Kelas')
    })

    it('should display SMA category tag with correct color', () => {
      wrapper = createWrapper({
        schools: [mockSMASchool]
      })
      
      const tags = wrapper.findAll('.a-tag')
      const categoryTag = tags.find(tag => tag.text() === 'SMA')
      expect(categoryTag).toBeTruthy()
    })
  })

  describe('Validation Messages', () => {
    it('should show error message when SMP school has small portions', async () => {
      wrapper = createWrapper({
        schools: [mockSMPSchool],
        totalPortions: 500
      })
      
      // Directly set allocations to simulate user input
      wrapper.vm.allocations[2] = { portions_small: 50, portions_large: 100 }
      await wrapper.vm.$nextTick()
      
      const errorMessage = wrapper.find('.error-message')
      expect(errorMessage.exists()).toBe(true)
      expect(errorMessage.text()).toContain('tidak boleh memiliki porsi kecil')
    })

    it('should show error message when SMA school has small portions', async () => {
      wrapper = createWrapper({
        schools: [mockSMASchool],
        totalPortions: 500
      })
      
      // Directly set allocations to simulate user input
      wrapper.vm.allocations[3] = { portions_small: 50, portions_large: 100 }
      await wrapper.vm.$nextTick()
      
      const errorMessage = wrapper.find('.error-message')
      expect(errorMessage.exists()).toBe(true)
      expect(errorMessage.text()).toContain('tidak boleh memiliki porsi kecil')
    })

    it('should show error message when allocations exceed total portions', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 100
      })
      
      // Directly set allocations to simulate user input
      wrapper.vm.allocations[1] = { portions_small: 80, portions_large: 50 }
      await wrapper.vm.$nextTick()
      
      const errorMessage = wrapper.find('.error-message')
      expect(errorMessage.exists()).toBe(true)
      expect(errorMessage.text()).toContain('Alokasi melebihi total porsi')
    })

    it('should show validation hint when allocations are less than total', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      // Directly set allocations to simulate user input
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 150 }
      await wrapper.vm.$nextTick()
      
      const validationHint = wrapper.find('.validation-hint')
      expect(validationHint.exists()).toBe(true)
      expect(validationHint.text()).toContain('Masih perlu mengalokasikan')
      expect(validationHint.text()).toContain('250 porsi lagi')
    })

    it('should show success indicator when allocations are valid', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      // Directly set allocations to simulate user input
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.isValid).toBe(true)
      expect(wrapper.find('.error-message').exists()).toBe(false)
    })

    it('should not show error when no allocations are made', () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      // With no allocations, there should be an error message about needing to allocate
      const errorMessage = wrapper.find('.error-message')
      expect(errorMessage.exists()).toBe(true)
      expect(errorMessage.text()).toContain('Harap alokasikan porsi')
    })
  })

  describe('Submit Button State', () => {
    it('should indicate invalid state when validation fails', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 150 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.isValid).toBe(false)
    })

    it('should indicate valid state when validation passes', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.isValid).toBe(true)
    })

    it('should be invalid when total portions is zero', () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 0
      })
      
      expect(wrapper.vm.isValid).toBe(false)
    })

    it('should be invalid when no allocations are made', () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      expect(wrapper.vm.isValid).toBe(false)
    })
  })

  describe('Statistics Display', () => {
    it('should display statistics section when allocations exist', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 150 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 250 }
      await wrapper.vm.$nextTick()
      
      const statisticsSection = wrapper.find('.statistics-section')
      expect(statisticsSection.exists()).toBe(true)
    })

    it('should not display statistics section when no allocations', () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      const statisticsSection = wrapper.find('.statistics-section')
      expect(statisticsSection.exists()).toBe(false)
    })

    it('should calculate total small portions correctly', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 150, portions_large: 200 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.totalSmallPortions).toBe(150)
    })

    it('should calculate total large portions correctly', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 150 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 250 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.totalLargePortions).toBe(400)
    })

    it('should calculate small portion percentage correctly', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.smallPortionPercentage).toBe('40.0')
    })

    it('should calculate large portion percentage correctly', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.largePortionPercentage).toBe('60.0')
    })

    it('should count SD schools with allocations', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, { ...mockSDSchool, id: 4, name: 'SD Negeri 2' }],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 150 }
      wrapper.vm.allocations[4] = { portions_small: 50, portions_large: 200 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.sdSchoolCount).toBe(2)
    })

    it('should count SMP/SMA schools with allocations', async () => {
      wrapper = createWrapper({
        schools: [mockSMPSchool, mockSMASchool],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 250 }
      wrapper.vm.allocations[3] = { portions_small: 0, portions_large: 250 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.smpSmaSchoolCount).toBe(2)
    })

    it('should handle zero allocations in percentage calculations', () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      expect(wrapper.vm.smallPortionPercentage).toBe(0)
      expect(wrapper.vm.largePortionPercentage).toBe(0)
    })
  })

  describe('Statistics Calculations - Total Small Portions', () => {
    it('should calculate total small portions across all SD schools', async () => {
      const sdSchool2 = { ...mockSDSchool, id: 4, name: 'SD Negeri 2' }
      const sdSchool3 = { ...mockSDSchool, id: 5, name: 'SD Negeri 3' }
      
      wrapper = createWrapper({
        schools: [mockSDSchool, sdSchool2, sdSchool3],
        totalPortions: 1000
      })
      
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 150 }
      wrapper.vm.allocations[4] = { portions_small: 75, portions_large: 125 }
      wrapper.vm.allocations[5] = { portions_small: 50, portions_large: 100 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.totalSmallPortions).toBe(225)
    })

    it('should calculate total small portions when some SD schools have zero small portions', async () => {
      const sdSchool2 = { ...mockSDSchool, id: 4, name: 'SD Negeri 2' }
      
      wrapper = createWrapper({
        schools: [mockSDSchool, sdSchool2],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 150, portions_large: 200 }
      wrapper.vm.allocations[4] = { portions_small: 0, portions_large: 150 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.totalSmallPortions).toBe(150)
    })

    it('should return zero when no small portions are allocated', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 0, portions_large: 250 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 250 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.totalSmallPortions).toBe(0)
    })

    it('should ignore SMP/SMA schools in small portion calculation', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool, mockSMASchool],
        totalPortions: 700
      })
      
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 200 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 200 }
      wrapper.vm.allocations[3] = { portions_small: 0, portions_large: 200 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.totalSmallPortions).toBe(100)
    })
  })

  describe('Statistics Calculations - Total Large Portions', () => {
    it('should calculate total large portions across all schools', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool, mockSMASchool],
        totalPortions: 1000
      })
      
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 200 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 300 }
      wrapper.vm.allocations[3] = { portions_small: 0, portions_large: 400 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.totalLargePortions).toBe(900)
    })

    it('should include large portions from SD schools', async () => {
      const sdSchool2 = { ...mockSDSchool, id: 4, name: 'SD Negeri 2' }
      
      wrapper = createWrapper({
        schools: [mockSDSchool, sdSchool2],
        totalPortions: 600
      })
      
      wrapper.vm.allocations[1] = { portions_small: 50, portions_large: 250 }
      wrapper.vm.allocations[4] = { portions_small: 100, portions_large: 200 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.totalLargePortions).toBe(450)
    })

    it('should include large portions from SMP/SMA schools only', async () => {
      wrapper = createWrapper({
        schools: [mockSMPSchool, mockSMASchool],
        totalPortions: 600
      })
      
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 300 }
      wrapper.vm.allocations[3] = { portions_small: 0, portions_large: 300 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.totalLargePortions).toBe(600)
    })

    it('should return zero when no large portions are allocated', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 200
      })
      
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 0 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.totalLargePortions).toBe(0)
    })
  })

  describe('Statistics Calculations - Percentage Breakdown', () => {
    it('should calculate percentage breakdown of small vs large portions', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 1000
      })
      
      wrapper.vm.allocations[1] = { portions_small: 250, portions_large: 750 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.smallPortionPercentage).toBe('25.0')
      expect(wrapper.vm.largePortionPercentage).toBe('75.0')
    })

    it('should calculate 50/50 percentage split correctly', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 400
      })
      
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 200 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.smallPortionPercentage).toBe('50.0')
      expect(wrapper.vm.largePortionPercentage).toBe('50.0')
    })

    it('should calculate percentage with multiple schools', async () => {
      const sdSchool2 = { ...mockSDSchool, id: 4, name: 'SD Negeri 2' }
      
      wrapper = createWrapper({
        schools: [mockSDSchool, sdSchool2, mockSMPSchool],
        totalPortions: 1000
      })
      
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 200 }
      wrapper.vm.allocations[4] = { portions_small: 150, portions_large: 250 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 300 }
      await wrapper.vm.$nextTick()
      
      // Total small: 250, Total large: 750, Total: 1000
      expect(wrapper.vm.smallPortionPercentage).toBe('25.0')
      expect(wrapper.vm.largePortionPercentage).toBe('75.0')
    })

    it('should handle 100% small portions', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 300
      })
      
      wrapper.vm.allocations[1] = { portions_small: 300, portions_large: 0 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.smallPortionPercentage).toBe('100.0')
      expect(wrapper.vm.largePortionPercentage).toBe('0.0')
    })

    it('should handle 100% large portions', async () => {
      wrapper = createWrapper({
        schools: [mockSMPSchool, mockSMASchool],
        totalPortions: 600
      })
      
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 300 }
      wrapper.vm.allocations[3] = { portions_small: 0, portions_large: 300 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.smallPortionPercentage).toBe('0.0')
      expect(wrapper.vm.largePortionPercentage).toBe('100.0')
    })

    it('should format percentages to one decimal place', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 300
      })
      
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 200 }
      await wrapper.vm.$nextTick()
      
      // 100/300 = 33.333...%
      expect(wrapper.vm.smallPortionPercentage).toBe('33.3')
      // 200/300 = 66.666...%
      expect(wrapper.vm.largePortionPercentage).toBe('66.7')
    })
  })

  describe('Statistics Calculations - School Count by Type', () => {
    it('should count schools by portion size type', async () => {
      const sdSchool2 = { ...mockSDSchool, id: 4, name: 'SD Negeri 2' }
      
      wrapper = createWrapper({
        schools: [mockSDSchool, sdSchool2, mockSMPSchool, mockSMASchool],
        totalPortions: 1000
      })
      
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 150 }
      wrapper.vm.allocations[4] = { portions_small: 50, portions_large: 200 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 250 }
      wrapper.vm.allocations[3] = { portions_small: 0, portions_large: 250 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.sdSchoolCount).toBe(2)
      expect(wrapper.vm.smpSmaSchoolCount).toBe(2)
    })

    it('should only count SD schools with allocations', async () => {
      const sdSchool2 = { ...mockSDSchool, id: 4, name: 'SD Negeri 2' }
      const sdSchool3 = { ...mockSDSchool, id: 5, name: 'SD Negeri 3' }
      
      wrapper = createWrapper({
        schools: [mockSDSchool, sdSchool2, sdSchool3],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 150 }
      wrapper.vm.allocations[4] = { portions_small: 50, portions_large: 200 }
      wrapper.vm.allocations[5] = { portions_small: 0, portions_large: 0 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.sdSchoolCount).toBe(2)
    })

    it('should only count SMP/SMA schools with allocations', async () => {
      const smpSchool2 = { ...mockSMPSchool, id: 6, name: 'SMP Negeri 2' }
      
      wrapper = createWrapper({
        schools: [mockSMPSchool, smpSchool2, mockSMASchool],
        totalPortions: 600
      })
      
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 300 }
      wrapper.vm.allocations[6] = { portions_small: 0, portions_large: 0 }
      wrapper.vm.allocations[3] = { portions_small: 0, portions_large: 300 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.smpSmaSchoolCount).toBe(2)
    })

    it('should count SD schools with only small portions', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 200
      })
      
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 0 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.sdSchoolCount).toBe(1)
    })

    it('should count SD schools with only large portions', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 300
      })
      
      wrapper.vm.allocations[1] = { portions_small: 0, portions_large: 300 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.sdSchoolCount).toBe(1)
    })

    it('should return zero counts when no allocations', () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool, mockSMASchool],
        totalPortions: 500
      })
      
      expect(wrapper.vm.sdSchoolCount).toBe(0)
      expect(wrapper.vm.smpSmaSchoolCount).toBe(0)
    })
  })

  describe('Statistics Calculations - Real-time Updates', () => {
    it('should update statistics in real-time as allocations change', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool],
        totalPortions: 500
      })
      
      // Initial state
      expect(wrapper.vm.totalSmallPortions).toBe(0)
      expect(wrapper.vm.totalLargePortions).toBe(0)
      
      // First allocation
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 150 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.totalSmallPortions).toBe(100)
      expect(wrapper.vm.totalLargePortions).toBe(150)
      expect(wrapper.vm.sdSchoolCount).toBe(1)
      
      // Second allocation
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 250 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.totalSmallPortions).toBe(100)
      expect(wrapper.vm.totalLargePortions).toBe(400)
      expect(wrapper.vm.smpSmaSchoolCount).toBe(1)
      
      // Update first allocation
      wrapper.vm.allocations[1] = { portions_small: 50, portions_large: 200 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.totalSmallPortions).toBe(50)
      expect(wrapper.vm.totalLargePortions).toBe(450)
    })

    it('should update percentages in real-time', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      // Start with 50/50 split
      wrapper.vm.allocations[1] = { portions_small: 250, portions_large: 250 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.smallPortionPercentage).toBe('50.0')
      expect(wrapper.vm.largePortionPercentage).toBe('50.0')
      
      // Change to 20/80 split
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 400 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.smallPortionPercentage).toBe('20.0')
      expect(wrapper.vm.largePortionPercentage).toBe('80.0')
    })

    it('should update school counts in real-time', async () => {
      const sdSchool2 = { ...mockSDSchool, id: 4, name: 'SD Negeri 2' }
      
      wrapper = createWrapper({
        schools: [mockSDSchool, sdSchool2, mockSMPSchool],
        totalPortions: 800
      })
      
      // Start with one SD school
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 200 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.sdSchoolCount).toBe(1)
      expect(wrapper.vm.smpSmaSchoolCount).toBe(0)
      
      // Add second SD school
      wrapper.vm.allocations[4] = { portions_small: 50, portions_large: 150 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.sdSchoolCount).toBe(2)
      
      // Add SMP school
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 300 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.smpSmaSchoolCount).toBe(1)
      
      // Remove allocation from first SD school
      wrapper.vm.allocations[1] = { portions_small: 0, portions_large: 0 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.sdSchoolCount).toBe(1)
    })

    it('should update all statistics simultaneously', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool],
        totalPortions: 600
      })
      
      wrapper.vm.allocations[1] = { portions_small: 150, portions_large: 250 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 200 }
      await wrapper.vm.$nextTick()
      
      // Verify all statistics are updated
      expect(wrapper.vm.totalSmallPortions).toBe(150)
      expect(wrapper.vm.totalLargePortions).toBe(450)
      expect(wrapper.vm.smallPortionPercentage).toBe('25.0')
      expect(wrapper.vm.largePortionPercentage).toBe('75.0')
      expect(wrapper.vm.sdSchoolCount).toBe(1)
      expect(wrapper.vm.smpSmaSchoolCount).toBe(1)
      
      // Change allocations
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 200 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 200 }
      await wrapper.vm.$nextTick()
      
      // Verify all statistics updated together
      expect(wrapper.vm.totalSmallPortions).toBe(200)
      expect(wrapper.vm.totalLargePortions).toBe(400)
      expect(wrapper.vm.smallPortionPercentage).toBe('33.3')
      expect(wrapper.vm.largePortionPercentage).toBe('66.7')
    })

    it('should handle rapid allocation changes', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      // Rapid changes
      wrapper.vm.allocations[1] = { portions_small: 50, portions_large: 100 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.totalSmallPortions).toBe(50)
      
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 200 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.totalSmallPortions).toBe(100)
      
      wrapper.vm.allocations[1] = { portions_small: 150, portions_large: 350 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.totalSmallPortions).toBe(150)
      expect(wrapper.vm.totalLargePortions).toBe(350)
      expect(wrapper.vm.smallPortionPercentage).toBe('30.0')
    })
  })

  describe('Mixed School Types', () => {
    it('should render correctly with mixed school types', () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool, mockSMASchool],
        totalPortions: 1000
      })
      
      const schoolRows = wrapper.findAll('.school-row')
      expect(schoolRows).toHaveLength(3)
    })

    it('should validate correctly with mixed school types', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool, mockSMASchool],
        totalPortions: 1000
      })
      
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 250 }
      wrapper.vm.allocations[3] = { portions_small: 0, portions_large: 250 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.isValid).toBe(true)
      expect(wrapper.vm.totalAllocated).toBe(1000)
    })
  })

  describe('Event Emissions', () => {
    it('should emit update:modelValue when allocations change', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      await wrapper.vm.handleAllocationChange()
      
      expect(wrapper.emitted('update:modelValue')).toBeTruthy()
    })

    it('should emit validation-change with correct data', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      await wrapper.vm.handleAllocationChange()
      
      const emitted = wrapper.emitted('validation-change')
      expect(emitted).toBeTruthy()
      expect(emitted[emitted.length - 1][0]).toEqual({
        isValid: true,
        totalAllocated: 500,
        totalPortions: 500
      })
    })
  })

  describe('Edge Cases', () => {
    it('should handle empty schools array', () => {
      wrapper = createWrapper({
        schools: [],
        totalPortions: 500
      })
      
      const schoolRows = wrapper.findAll('.school-row')
      expect(schoolRows).toHaveLength(0)
    })

    it('should handle schools with zero student counts', () => {
      const schoolWithZeroStudents = {
        ...mockSDSchool,
        student_count_grade_1_3: 0,
        student_count_grade_4_6: 0
      }
      
      wrapper = createWrapper({
        schools: [schoolWithZeroStudents]
      })
      
      const schoolMeta = wrapper.find('.school-meta')
      expect(schoolMeta.text()).toContain('0 siswa')
    })

    it('should handle very large portion numbers', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 10000
      })
      
      wrapper.vm.allocations[1] = { portions_small: 4000, portions_large: 6000 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.totalAllocated).toBe(10000)
      expect(wrapper.vm.isValid).toBe(true)
    })

    it('should handle school category color for unknown category', () => {
      const schoolWithUnknownCategory = {
        ...mockSDSchool,
        category: 'UNKNOWN'
      }
      
      wrapper = createWrapper({
        schools: [schoolWithUnknownCategory]
      })
      
      const color = wrapper.vm.getSchoolCategoryColor('UNKNOWN')
      expect(color).toBe('default')
    })
  })

  describe('Reactivity', () => {
    it('should update total allocated when allocations change', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      expect(wrapper.vm.totalAllocated).toBe(0)
      
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 150 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.totalAllocated).toBe(250)
      
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.totalAllocated).toBe(500)
    })

    it('should reset allocations when schools change', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.totalAllocated).toBe(500)
      
      // Reset allocations manually (simulating what happens when modelValue is reset)
      wrapper.vm.allocations = wrapper.vm.initializeAllocations()
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.totalAllocated).toBe(0)
    })

    it('should update validation state when totalPortions changes', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.isValid).toBe(true)
      
      await wrapper.setProps({
        totalPortions: 600
      })
      
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.isValid).toBe(false)
    })
  })

  describe('Validation Logic - Real-time Calculation', () => {
    it('should calculate sum of all portions as user types', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool],
        totalPortions: 500
      })
      
      // Initially zero
      expect(wrapper.vm.totalAllocated).toBe(0)
      
      // Add first allocation
      wrapper.vm.allocations[1] = { portions_small: 50, portions_large: 100 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.totalAllocated).toBe(150)
      
      // Add second allocation
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 200 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.totalAllocated).toBe(350)
      
      // Update first allocation
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 150 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.totalAllocated).toBe(450)
    })

    it('should display running total vs target total portions', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 150, portions_large: 200 }
      await wrapper.vm.$nextTick()
      
      const summaryText = wrapper.find('.summary-text')
      expect(summaryText.text()).toContain('350 / 500 porsi')
    })

    it('should update running total immediately on input change', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      const summaryText = wrapper.find('.summary-text')
      expect(summaryText.text()).toContain('0 / 500 porsi')
      
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 0 }
      await wrapper.vm.$nextTick()
      expect(summaryText.text()).toContain('100 / 500 porsi')
      
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 200 }
      await wrapper.vm.$nextTick()
      expect(summaryText.text()).toContain('300 / 500 porsi')
    })
  })

  describe('Validation Logic - Error Messages', () => {
    it('should display error when sum exceeds target', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 300
      })
      
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 200 }
      await wrapper.vm.$nextTick()
      
      const errorMessage = wrapper.find('.error-message')
      expect(errorMessage.exists()).toBe(true)
      expect(errorMessage.text()).toContain('Alokasi melebihi total porsi')
      expect(errorMessage.text()).toContain('100 porsi')
    })

    it('should display error when sum is less than target', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 150 }
      await wrapper.vm.$nextTick()
      
      const validationHint = wrapper.find('.validation-hint')
      expect(validationHint.exists()).toBe(true)
      expect(validationHint.text()).toContain('Masih perlu mengalokasikan')
      expect(validationHint.text()).toContain('250 porsi lagi')
    })

    it('should display error when no allocations are made', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      const errorMessage = wrapper.find('.error-message')
      expect(errorMessage.exists()).toBe(true)
      expect(errorMessage.text()).toContain('Harap alokasikan porsi')
    })

    it('should clear error when sum matches target', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      // First, create an error state
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 150 }
      await wrapper.vm.$nextTick()
      expect(wrapper.find('.validation-hint').exists()).toBe(true)
      
      // Now fix it
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      expect(wrapper.find('.error-message').exists()).toBe(false)
      expect(wrapper.find('.validation-hint').exists()).toBe(false)
    })

    it('should display specific error for SMP with small portions', async () => {
      wrapper = createWrapper({
        schools: [mockSMPSchool],
        totalPortions: 300
      })
      
      wrapper.vm.allocations[2] = { portions_small: 100, portions_large: 200 }
      await wrapper.vm.$nextTick()
      
      const errorMessage = wrapper.find('.error-message')
      expect(errorMessage.exists()).toBe(true)
      expect(errorMessage.text()).toContain('SMP Negeri 1')
      expect(errorMessage.text()).toContain('tidak boleh memiliki porsi kecil')
    })

    it('should display specific error for SMA with small portions', async () => {
      wrapper = createWrapper({
        schools: [mockSMASchool],
        totalPortions: 250
      })
      
      wrapper.vm.allocations[3] = { portions_small: 50, portions_large: 200 }
      await wrapper.vm.$nextTick()
      
      const errorMessage = wrapper.find('.error-message')
      expect(errorMessage.exists()).toBe(true)
      expect(errorMessage.text()).toContain('SMA Negeri 1')
      expect(errorMessage.text()).toContain('tidak boleh memiliki porsi kecil')
    })
  })

  describe('Validation Logic - Success Indicator', () => {
    it('should show success indicator when sum matches target', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      
      const successTag = wrapper.findAll('.a-tag').find(tag => tag.text().includes('Valid'))
      expect(successTag).toBeTruthy()
      expect(wrapper.vm.isValid).toBe(true)
    })

    it('should show success indicator with multiple schools', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool, mockSMASchool],
        totalPortions: 1000
      })
      
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 250 }
      wrapper.vm.allocations[3] = { portions_small: 0, portions_large: 250 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.isValid).toBe(true)
      const successTag = wrapper.findAll('.a-tag').find(tag => tag.text().includes('Valid'))
      expect(successTag).toBeTruthy()
    })

    it('should not show success when sum is incorrect', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 200 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.isValid).toBe(false)
      const warningTag = wrapper.findAll('.a-tag').find(tag => tag.text().includes('Belum Valid'))
      expect(warningTag).toBeTruthy()
    })

    it('should update success indicator dynamically', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      // Start invalid
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 200 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.isValid).toBe(false)
      
      // Make valid
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.isValid).toBe(true)
      
      // Make invalid again
      wrapper.vm.allocations[1] = { portions_small: 150, portions_large: 200 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.isValid).toBe(false)
    })
  })

  describe('Validation Logic - Submit Button State', () => {
    it('should indicate disabled state when validation fails (sum mismatch)', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 150 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.isValid).toBe(false)
      expect(wrapper.vm.totalAllocated).toBe(250)
      expect(wrapper.vm.totalAllocated).not.toBe(wrapper.vm.totalPortions)
    })

    it('should indicate disabled state when SMP has small portions', async () => {
      wrapper = createWrapper({
        schools: [mockSMPSchool],
        totalPortions: 300
      })
      
      wrapper.vm.allocations[2] = { portions_small: 100, portions_large: 200 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.isValid).toBe(false)
    })

    it('should indicate disabled state when no allocations', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      expect(wrapper.vm.isValid).toBe(false)
      expect(wrapper.vm.totalAllocated).toBe(0)
    })

    it('should indicate enabled state when all validations pass', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool],
        totalPortions: 600
      })
      
      wrapper.vm.allocations[1] = { portions_small: 150, portions_large: 200 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 250 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.isValid).toBe(true)
      expect(wrapper.vm.totalAllocated).toBe(600)
    })

    it('should transition from disabled to enabled when fixed', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      // Start disabled
      expect(wrapper.vm.isValid).toBe(false)
      
      // Add partial allocation - still disabled
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 200 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.isValid).toBe(false)
      
      // Complete allocation - now enabled
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.isValid).toBe(true)
    })
  })

  describe('Validation Logic - Non-negative Values', () => {
    it('should handle zero values correctly', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 0, portions_large: 500 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.isValid).toBe(true)
      expect(wrapper.vm.totalAllocated).toBe(500)
    })

    it('should validate SD school with only small portions', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 200
      })
      
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 0 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.isValid).toBe(true)
      expect(wrapper.vm.totalAllocated).toBe(200)
    })

    it('should validate SD school with only large portions', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 300
      })
      
      wrapper.vm.allocations[1] = { portions_small: 0, portions_large: 300 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.isValid).toBe(true)
      expect(wrapper.vm.totalAllocated).toBe(300)
    })

    it('should validate SMP school with only large portions', async () => {
      wrapper = createWrapper({
        schools: [mockSMPSchool],
        totalPortions: 300
      })
      
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 300 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.isValid).toBe(true)
      expect(wrapper.vm.totalAllocated).toBe(300)
    })
  })

  describe('Validation Logic - Complex Scenarios', () => {
    it('should validate multiple schools with mixed allocations', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool, mockSMASchool],
        totalPortions: 1000
      })
      
      wrapper.vm.allocations[1] = { portions_small: 150, portions_large: 250 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 300 }
      wrapper.vm.allocations[3] = { portions_small: 0, portions_large: 300 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.isValid).toBe(true)
      expect(wrapper.vm.totalAllocated).toBe(1000)
      expect(wrapper.vm.totalSmallPortions).toBe(150)
      expect(wrapper.vm.totalLargePortions).toBe(850)
    })

    it('should detect validation error in mixed school scenario', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 150 }
      wrapper.vm.allocations[2] = { portions_small: 50, portions_large: 200 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.isValid).toBe(false)
      const errorMessage = wrapper.find('.error-message')
      expect(errorMessage.text()).toContain('tidak boleh memiliki porsi kecil')
    })

    it('should validate when one school has zero allocation', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 0 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.isValid).toBe(true)
      expect(wrapper.vm.totalAllocated).toBe(500)
    })

    it('should handle rapid allocation changes', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      // Rapid changes
      wrapper.vm.allocations[1] = { portions_small: 50, portions_large: 100 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.totalAllocated).toBe(150)
      
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 200 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.totalAllocated).toBe(300)
      
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.totalAllocated).toBe(500)
      expect(wrapper.vm.isValid).toBe(true)
    })
  })

  // Task 6.4.4: Test error message display
  describe('Error Message Display (Task 6.4.4)', () => {
    it('should display error message when sum does not match total (too low)', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool],
        totalPortions: 1000
      })
      
      wrapper.vm.allocations[1] = { portions_small: 150, portions_large: 200 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 250 }
      await wrapper.vm.$nextTick()
      
      // Total allocated: 600, Total portions: 1000
      const validationHint = wrapper.find('.validation-hint')
      expect(validationHint.exists()).toBe(true)
      expect(validationHint.text()).toContain('Masih perlu mengalokasikan')
      expect(validationHint.text()).toContain('400 porsi lagi')
    })

    it('should display error message when sum does not match total (too high)', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 400
      })
      
      wrapper.vm.allocations[1] = { portions_small: 300, portions_large: 250 }
      await wrapper.vm.$nextTick()
      
      // Total allocated: 550, Total portions: 400
      const errorMessage = wrapper.find('.error-message')
      expect(errorMessage.exists()).toBe(true)
      expect(errorMessage.text()).toContain('Alokasi melebihi total porsi')
      expect(errorMessage.text()).toContain('150 porsi')
    })

    it('should display error message when SMP school has small portions', async () => {
      wrapper = createWrapper({
        schools: [mockSMPSchool],
        totalPortions: 300
      })
      
      wrapper.vm.allocations[2] = { portions_small: 100, portions_large: 200 }
      await wrapper.vm.$nextTick()
      
      const errorMessage = wrapper.find('.error-message')
      expect(errorMessage.exists()).toBe(true)
      expect(errorMessage.text()).toContain('SMP Negeri 1')
      expect(errorMessage.text()).toContain('SMP')
      expect(errorMessage.text()).toContain('tidak boleh memiliki porsi kecil')
    })

    it('should display error message when SMA school has small portions', async () => {
      wrapper = createWrapper({
        schools: [mockSMASchool],
        totalPortions: 250
      })
      
      wrapper.vm.allocations[3] = { portions_small: 75, portions_large: 175 }
      await wrapper.vm.$nextTick()
      
      const errorMessage = wrapper.find('.error-message')
      expect(errorMessage.exists()).toBe(true)
      expect(errorMessage.text()).toContain('SMA Negeri 1')
      expect(errorMessage.text()).toContain('SMA')
      expect(errorMessage.text()).toContain('tidak boleh memiliki porsi kecil')
    })

    it('should display error message when all portions are zero', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool],
        totalPortions: 500
      })
      
      // All allocations are zero (default state)
      const errorMessage = wrapper.find('.error-message')
      expect(errorMessage.exists()).toBe(true)
      expect(errorMessage.text()).toContain('Harap alokasikan porsi ke minimal satu sekolah')
    })

    it('should display error message when allocations are zero after having values', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      // Set allocations
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      expect(wrapper.find('.error-message').exists()).toBe(false)
      
      // Reset to zero
      wrapper.vm.allocations[1] = { portions_small: 0, portions_large: 0 }
      await wrapper.vm.$nextTick()
      
      const errorMessage = wrapper.find('.error-message')
      expect(errorMessage.exists()).toBe(true)
      expect(errorMessage.text()).toContain('Harap alokasikan porsi')
    })

    it('should clear error message when validation passes (sum matches)', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      // Start with error (partial allocation)
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 200 }
      await wrapper.vm.$nextTick()
      expect(wrapper.find('.validation-hint').exists()).toBe(true)
      
      // Fix allocation to match total
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.find('.error-message').exists()).toBe(false)
      expect(wrapper.find('.validation-hint').exists()).toBe(false)
      expect(wrapper.vm.isValid).toBe(true)
    })

    it('should clear error message when SMP/SMA small portions are removed', async () => {
      wrapper = createWrapper({
        schools: [mockSMPSchool],
        totalPortions: 300
      })
      
      // Start with error (SMP has small portions)
      wrapper.vm.allocations[2] = { portions_small: 100, portions_large: 200 }
      await wrapper.vm.$nextTick()
      expect(wrapper.find('.error-message').exists()).toBe(true)
      
      // Fix by removing small portions
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 300 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.find('.error-message').exists()).toBe(false)
      expect(wrapper.vm.isValid).toBe(true)
    })

    it('should clear error message when excess allocation is reduced', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 400
      })
      
      // Start with error (exceeds total)
      wrapper.vm.allocations[1] = { portions_small: 300, portions_large: 250 }
      await wrapper.vm.$nextTick()
      expect(wrapper.find('.error-message').exists()).toBe(true)
      expect(wrapper.find('.error-message').text()).toContain('melebihi')
      
      // Fix by reducing allocation
      wrapper.vm.allocations[1] = { portions_small: 150, portions_large: 250 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.find('.error-message').exists()).toBe(false)
      expect(wrapper.vm.isValid).toBe(true)
    })

    it('should display multiple error types correctly (priority: SMP/SMA small portions)', async () => {
      wrapper = createWrapper({
        schools: [mockSMPSchool],
        totalPortions: 300
      })
      
      // SMP with small portions AND exceeds total
      wrapper.vm.allocations[2] = { portions_small: 200, portions_large: 200 }
      await wrapper.vm.$nextTick()
      
      // Should show SMP/SMA error first (higher priority)
      const errorMessage = wrapper.find('.error-message')
      expect(errorMessage.exists()).toBe(true)
      expect(errorMessage.text()).toContain('tidak boleh memiliki porsi kecil')
    })

    it('should show correct error when only exceeding total (no SMP/SMA issue)', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 300
      })
      
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 200 }
      await wrapper.vm.$nextTick()
      
      const errorMessage = wrapper.find('.error-message')
      expect(errorMessage.exists()).toBe(true)
      expect(errorMessage.text()).toContain('Alokasi melebihi total porsi')
      expect(errorMessage.text()).toContain('100 porsi')
    })

    it('should update error message dynamically as allocations change', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      // No allocation error
      let errorMessage = wrapper.find('.error-message')
      expect(errorMessage.exists()).toBe(true)
      expect(errorMessage.text()).toContain('Harap alokasikan porsi')
      
      // Partial allocation error
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 150 }
      await wrapper.vm.$nextTick()
      let validationHint = wrapper.find('.validation-hint')
      expect(validationHint.exists()).toBe(true)
      expect(validationHint.text()).toContain('250 porsi lagi')
      
      // Excess allocation error
      wrapper.vm.allocations[1] = { portions_small: 300, portions_large: 300 }
      await wrapper.vm.$nextTick()
      errorMessage = wrapper.find('.error-message')
      expect(errorMessage.exists()).toBe(true)
      expect(errorMessage.text()).toContain('melebihi')
      
      // Valid state - no errors
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      expect(wrapper.find('.error-message').exists()).toBe(false)
      expect(wrapper.find('.validation-hint').exists()).toBe(false)
    })

    it('should display error for first SMP/SMA school with small portions in list', async () => {
      const smpSchool2 = { ...mockSMPSchool, id: 6, name: 'SMP Negeri 2' }
      
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool, smpSchool2],
        totalPortions: 1000
      })
      
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      wrapper.vm.allocations[2] = { portions_small: 50, portions_large: 200 }
      wrapper.vm.allocations[6] = { portions_small: 100, portions_large: 150 }
      await wrapper.vm.$nextTick()
      
      // Should show error for first SMP school encountered
      const errorMessage = wrapper.find('.error-message')
      expect(errorMessage.exists()).toBe(true)
      expect(errorMessage.text()).toContain('SMP Negeri 1')
      expect(errorMessage.text()).toContain('tidak boleh memiliki porsi kecil')
    })

    it('should handle error message with special characters in school name', async () => {
      const specialSchool = {
        ...mockSMPSchool,
        id: 10,
        name: 'SMP Negeri 1 "Pembangunan"'
      }
      
      wrapper = createWrapper({
        schools: [specialSchool],
        totalPortions: 300
      })
      
      wrapper.vm.allocations[10] = { portions_small: 100, portions_large: 200 }
      await wrapper.vm.$nextTick()
      
      const errorMessage = wrapper.find('.error-message')
      expect(errorMessage.exists()).toBe(true)
      expect(errorMessage.text()).toContain('SMP Negeri 1 "Pembangunan"')
      expect(errorMessage.text()).toContain('tidak boleh memiliki porsi kecil')
    })

    it('should show validation hint with exact remaining portions', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool],
        totalPortions: 1000
      })
      
      wrapper.vm.allocations[1] = { portions_small: 150, portions_large: 250 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 350 }
      await wrapper.vm.$nextTick()
      
      // Total: 750, Remaining: 250
      const validationHint = wrapper.find('.validation-hint')
      expect(validationHint.exists()).toBe(true)
      expect(validationHint.text()).toContain('250 porsi lagi')
    })

    it('should show error with exact excess amount', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 350, portions_large: 275 }
      await wrapper.vm.$nextTick()
      
      // Total: 625, Excess: 125
      const errorMessage = wrapper.find('.error-message')
      expect(errorMessage.exists()).toBe(true)
      expect(errorMessage.text()).toContain('125 porsi')
    })

    it('should not display error message when validation passes with multiple schools', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool, mockSMASchool],
        totalPortions: 1000
      })
      
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 250 }
      wrapper.vm.allocations[3] = { portions_small: 0, portions_large: 250 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.find('.error-message').exists()).toBe(false)
      expect(wrapper.find('.validation-hint').exists()).toBe(false)
      expect(wrapper.vm.isValid).toBe(true)
    })
  })

  // Task 6.4.6: Test statistics calculations
  describe('Statistics Calculations (Task 6.4.6)', () => {
    it('should verify total small portions calculation across all SD schools', async () => {
      const sdSchool2 = { ...mockSDSchool, id: 4, name: 'SD Negeri 2' }
      const sdSchool3 = { ...mockSDSchool, id: 5, name: 'SD Negeri 3' }
      
      wrapper = createWrapper({
        schools: [mockSDSchool, sdSchool2, sdSchool3, mockSMPSchool],
        totalPortions: 1500
      })
      
      // Allocate to all SD schools with different small portions
      wrapper.vm.allocations[1] = { portions_small: 120, portions_large: 180 }
      wrapper.vm.allocations[4] = { portions_small: 90, portions_large: 160 }
      wrapper.vm.allocations[5] = { portions_small: 75, portions_large: 125 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 750 }
      await wrapper.vm.$nextTick()
      
      // Total small portions should be sum of all SD schools only
      expect(wrapper.vm.totalSmallPortions).toBe(285) // 120 + 90 + 75
      
      // Verify it's displayed in the statistics section
      const statItems = wrapper.findAll('.stat-item')
      const smallPortionStat = statItems.find(item => 
        item.find('.stat-label').text().includes('Total Porsi Kecil')
      )
      expect(smallPortionStat).toBeTruthy()
      expect(smallPortionStat.find('.stat-value').text()).toBe('285')
    })

    it('should verify total large portions calculation across all schools', async () => {
      const sdSchool2 = { ...mockSDSchool, id: 4, name: 'SD Negeri 2' }
      
      wrapper = createWrapper({
        schools: [mockSDSchool, sdSchool2, mockSMPSchool, mockSMASchool],
        totalPortions: 2000
      })
      
      // Allocate to all schools
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 300 }
      wrapper.vm.allocations[4] = { portions_small: 150, portions_large: 350 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 600 }
      wrapper.vm.allocations[3] = { portions_small: 0, portions_large: 500 }
      await wrapper.vm.$nextTick()
      
      // Total large portions should include all schools (SD + SMP + SMA)
      expect(wrapper.vm.totalLargePortions).toBe(1750) // 300 + 350 + 600 + 500
      
      // Verify it's displayed in the statistics section
      const statItems = wrapper.findAll('.stat-item')
      const largePortionStat = statItems.find(item => 
        item.find('.stat-label').text().includes('Total Porsi Besar')
      )
      expect(largePortionStat).toBeTruthy()
      expect(largePortionStat.find('.stat-value').text()).toBe('1750')
    })

    it('should verify percentage breakdown of small vs large', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool],
        totalPortions: 1000
      })
      
      // 30% small, 70% large
      wrapper.vm.allocations[1] = { portions_small: 300, portions_large: 200 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 500 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.smallPortionPercentage).toBe('30.0')
      expect(wrapper.vm.largePortionPercentage).toBe('70.0')
      
      // Verify percentages are displayed
      const statItems = wrapper.findAll('.stat-item')
      const smallPercentStat = statItems.find(item => 
        item.find('.stat-label').text().includes('Persentase Kecil')
      )
      const largePercentStat = statItems.find(item => 
        item.find('.stat-label').text().includes('Persentase Besar')
      )
      
      expect(smallPercentStat).toBeTruthy()
      expect(smallPercentStat.find('.stat-value').text()).toBe('30.0%')
      expect(largePercentStat).toBeTruthy()
      expect(largePercentStat.find('.stat-value').text()).toBe('70.0%')
    })

    it('should verify count of schools by portion size type', async () => {
      const sdSchool2 = { ...mockSDSchool, id: 4, name: 'SD Negeri 2' }
      const sdSchool3 = { ...mockSDSchool, id: 5, name: 'SD Negeri 3' }
      const smpSchool2 = { ...mockSMPSchool, id: 6, name: 'SMP Negeri 2' }
      
      wrapper = createWrapper({
        schools: [mockSDSchool, sdSchool2, sdSchool3, mockSMPSchool, smpSchool2, mockSMASchool],
        totalPortions: 3000
      })
      
      // Allocate to 3 SD schools and 2 SMP/SMA schools (1 SMP/SMA has no allocation)
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      wrapper.vm.allocations[4] = { portions_small: 150, portions_large: 250 }
      wrapper.vm.allocations[5] = { portions_small: 100, portions_large: 200 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 800 }
      wrapper.vm.allocations[6] = { portions_small: 0, portions_large: 0 } // No allocation
      wrapper.vm.allocations[3] = { portions_small: 0, portions_large: 1000 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.sdSchoolCount).toBe(3)
      expect(wrapper.vm.smpSmaSchoolCount).toBe(2) // Only schools with allocations
      
      // Verify counts are displayed
      const statItems = wrapper.findAll('.stat-item')
      const sdCountStat = statItems.find(item => 
        item.find('.stat-label').text().includes('Sekolah SD')
      )
      const smpSmaCountStat = statItems.find(item => 
        item.find('.stat-label').text().includes('Sekolah SMP/SMA')
      )
      
      expect(sdCountStat).toBeTruthy()
      expect(sdCountStat.find('.stat-value').text()).toBe('3')
      expect(smpSmaCountStat).toBeTruthy()
      expect(smpSmaCountStat.find('.stat-value').text()).toBe('2')
    })

    it('should verify statistics update in real-time as allocations change', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool],
        totalPortions: 1000
      })
      
      // Initial state - no statistics displayed
      expect(wrapper.find('.statistics-section').exists()).toBe(false)
      
      // Add first allocation
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.find('.statistics-section').exists()).toBe(true)
      expect(wrapper.vm.totalSmallPortions).toBe(200)
      expect(wrapper.vm.totalLargePortions).toBe(300)
      expect(wrapper.vm.smallPortionPercentage).toBe('40.0')
      expect(wrapper.vm.largePortionPercentage).toBe('60.0')
      expect(wrapper.vm.sdSchoolCount).toBe(1)
      expect(wrapper.vm.smpSmaSchoolCount).toBe(0)
      
      // Add second allocation
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 500 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.totalSmallPortions).toBe(200)
      expect(wrapper.vm.totalLargePortions).toBe(800)
      expect(wrapper.vm.smallPortionPercentage).toBe('20.0')
      expect(wrapper.vm.largePortionPercentage).toBe('80.0')
      expect(wrapper.vm.sdSchoolCount).toBe(1)
      expect(wrapper.vm.smpSmaSchoolCount).toBe(1)
      
      // Modify first allocation
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 400 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.totalSmallPortions).toBe(100)
      expect(wrapper.vm.totalLargePortions).toBe(900)
      expect(wrapper.vm.smallPortionPercentage).toBe('10.0')
      expect(wrapper.vm.largePortionPercentage).toBe('90.0')
      
      // Remove allocation from SD school
      wrapper.vm.allocations[1] = { portions_small: 0, portions_large: 0 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.totalSmallPortions).toBe(0)
      expect(wrapper.vm.totalLargePortions).toBe(500)
      expect(wrapper.vm.smallPortionPercentage).toBe('0.0')
      expect(wrapper.vm.largePortionPercentage).toBe('100.0')
      expect(wrapper.vm.sdSchoolCount).toBe(0)
      expect(wrapper.vm.smpSmaSchoolCount).toBe(1)
    })

    it('should handle edge case with only small portions allocated', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 500, portions_large: 0 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.totalSmallPortions).toBe(500)
      expect(wrapper.vm.totalLargePortions).toBe(0)
      expect(wrapper.vm.smallPortionPercentage).toBe('100.0')
      expect(wrapper.vm.largePortionPercentage).toBe('0.0')
      expect(wrapper.vm.sdSchoolCount).toBe(1)
    })

    it('should handle edge case with only large portions allocated', async () => {
      wrapper = createWrapper({
        schools: [mockSMPSchool, mockSMASchool],
        totalPortions: 800
      })
      
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 400 }
      wrapper.vm.allocations[3] = { portions_small: 0, portions_large: 400 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.totalSmallPortions).toBe(0)
      expect(wrapper.vm.totalLargePortions).toBe(800)
      expect(wrapper.vm.smallPortionPercentage).toBe('0.0')
      expect(wrapper.vm.largePortionPercentage).toBe('100.0')
      expect(wrapper.vm.smpSmaSchoolCount).toBe(2)
    })

    it('should calculate statistics correctly with mixed school types and partial allocations', async () => {
      const sdSchool2 = { ...mockSDSchool, id: 4, name: 'SD Negeri 2' }
      
      wrapper = createWrapper({
        schools: [mockSDSchool, sdSchool2, mockSMPSchool, mockSMASchool],
        totalPortions: 2000
      })
      
      // Some schools have allocations, some don't
      wrapper.vm.allocations[1] = { portions_small: 250, portions_large: 350 }
      wrapper.vm.allocations[4] = { portions_small: 0, portions_large: 0 } // No allocation
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 700 }
      wrapper.vm.allocations[3] = { portions_small: 0, portions_large: 700 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.vm.totalSmallPortions).toBe(250)
      expect(wrapper.vm.totalLargePortions).toBe(1750)
      expect(wrapper.vm.totalAllocated).toBe(2000)
      expect(wrapper.vm.smallPortionPercentage).toBe('12.5')
      expect(wrapper.vm.largePortionPercentage).toBe('87.5')
      expect(wrapper.vm.sdSchoolCount).toBe(1) // Only SD school with allocation
      expect(wrapper.vm.smpSmaSchoolCount).toBe(2)
    })

    it('should display statistics section only when allocations exist', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool],
        totalPortions: 1000
      })
      
      // No allocations - statistics should not be displayed
      expect(wrapper.find('.statistics-section').exists()).toBe(false)
      
      // Add allocation - statistics should appear
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 200 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.find('.statistics-section').exists()).toBe(true)
      
      // Remove allocation - statistics should still be visible (totalAllocated > 0 check)
      wrapper.vm.allocations[1] = { portions_small: 0, portions_large: 0 }
      await wrapper.vm.$nextTick()
      
      expect(wrapper.find('.statistics-section').exists()).toBe(false)
    })

    it('should format percentage values to one decimal place', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 900
      })
      
      // Create a scenario that results in repeating decimals
      wrapper.vm.allocations[1] = { portions_small: 300, portions_large: 600 }
      await wrapper.vm.$nextTick()
      
      // 300/900 = 33.333...%, should be formatted as 33.3
      expect(wrapper.vm.smallPortionPercentage).toBe('33.3')
      // 600/900 = 66.666...%, should be formatted as 66.7
      expect(wrapper.vm.largePortionPercentage).toBe('66.7')
    })

    it('should handle zero allocations in percentage calculations without errors', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      // No allocations - should return 0 without errors
      expect(wrapper.vm.totalAllocated).toBe(0)
      expect(wrapper.vm.smallPortionPercentage).toBe(0)
      expect(wrapper.vm.largePortionPercentage).toBe(0)
      expect(() => wrapper.vm.smallPortionPercentage).not.toThrow()
      expect(() => wrapper.vm.largePortionPercentage).not.toThrow()
    })

    it('should count SD schools correctly when they have only small or only large portions', async () => {
      const sdSchool2 = { ...mockSDSchool, id: 4, name: 'SD Negeri 2' }
      const sdSchool3 = { ...mockSDSchool, id: 5, name: 'SD Negeri 3' }
      
      wrapper = createWrapper({
        schools: [mockSDSchool, sdSchool2, sdSchool3],
        totalPortions: 1000
      })
      
      // SD school 1: only small portions
      wrapper.vm.allocations[1] = { portions_small: 300, portions_large: 0 }
      // SD school 2: only large portions
      wrapper.vm.allocations[4] = { portions_small: 0, portions_large: 400 }
      // SD school 3: both portions
      wrapper.vm.allocations[5] = { portions_small: 150, portions_large: 150 }
      await wrapper.vm.$nextTick()
      
      // All three SD schools should be counted
      expect(wrapper.vm.sdSchoolCount).toBe(3)
      expect(wrapper.vm.totalSmallPortions).toBe(450)
      expect(wrapper.vm.totalLargePortions).toBe(550)
    })

    it('should update all statistics simultaneously when allocations change', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool],
        totalPortions: 1000
      })
      
      // Set initial allocation
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 500 }
      await wrapper.vm.$nextTick()
      
      // Capture initial statistics
      const initialSmall = wrapper.vm.totalSmallPortions
      const initialLarge = wrapper.vm.totalLargePortions
      const initialSmallPercent = wrapper.vm.smallPortionPercentage
      const initialLargePercent = wrapper.vm.largePortionPercentage
      const initialSdCount = wrapper.vm.sdSchoolCount
      const initialSmpSmaCount = wrapper.vm.smpSmaSchoolCount
      
      expect(initialSmall).toBe(200)
      expect(initialLarge).toBe(800)
      expect(initialSmallPercent).toBe('20.0')
      expect(initialLargePercent).toBe('80.0')
      expect(initialSdCount).toBe(1)
      expect(initialSmpSmaCount).toBe(1)
      
      // Change allocations
      wrapper.vm.allocations[1] = { portions_small: 400, portions_large: 100 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 500 }
      await wrapper.vm.$nextTick()
      
      // All statistics should update
      expect(wrapper.vm.totalSmallPortions).toBe(400)
      expect(wrapper.vm.totalLargePortions).toBe(600)
      expect(wrapper.vm.smallPortionPercentage).toBe('40.0')
      expect(wrapper.vm.largePortionPercentage).toBe('60.0')
      expect(wrapper.vm.sdSchoolCount).toBe(1)
      expect(wrapper.vm.smpSmaSchoolCount).toBe(1)
      
      // Verify all values changed
      expect(wrapper.vm.totalSmallPortions).not.toBe(initialSmall)
      expect(wrapper.vm.totalLargePortions).not.toBe(initialLarge)
      expect(wrapper.vm.smallPortionPercentage).not.toBe(initialSmallPercent)
      expect(wrapper.vm.largePortionPercentage).not.toBe(initialLargePercent)
    })
  })

  // Task 6.4.3: Test real-time validation calculations
  describe('Real-time Validation Calculations (Task 6.4.3)', () => {
    it('should recalculate total immediately when small portion changes', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      // Initial state
      expect(wrapper.vm.totalAllocated).toBe(0)
      
      // Change small portion
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 0 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.totalAllocated).toBe(100)
      
      // Change small portion again
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 0 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.totalAllocated).toBe(200)
    })

    it('should recalculate total immediately when large portion changes', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      // Change large portion
      wrapper.vm.allocations[1] = { portions_small: 0, portions_large: 150 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.totalAllocated).toBe(150)
      
      // Change large portion again
      wrapper.vm.allocations[1] = { portions_small: 0, portions_large: 300 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.totalAllocated).toBe(300)
    })

    it('should recalculate total when both portions change simultaneously', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      wrapper.vm.allocations[1] = { portions_small: 150, portions_large: 200 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.totalAllocated).toBe(350)
      
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.totalAllocated).toBe(500)
    })

    it('should update validation status from invalid to valid in real-time', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      // Start invalid (no allocations)
      expect(wrapper.vm.isValid).toBe(false)
      
      // Still invalid (partial allocation)
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 200 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.isValid).toBe(false)
      expect(wrapper.vm.totalAllocated).toBe(300)
      
      // Now valid (complete allocation)
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.isValid).toBe(true)
      expect(wrapper.vm.totalAllocated).toBe(500)
    })

    it('should update validation status from valid to invalid in real-time', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      // Start valid
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.isValid).toBe(true)
      
      // Make invalid by changing allocation
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 200 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.isValid).toBe(false)
      expect(wrapper.vm.totalAllocated).toBe(300)
    })

    it('should validate SMP/SMA small portion restriction in real-time', async () => {
      wrapper = createWrapper({
        schools: [mockSMPSchool],
        totalPortions: 300
      })
      
      // Valid state (no small portions)
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 300 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.isValid).toBe(true)
      
      // Invalid state (has small portions)
      wrapper.vm.allocations[2] = { portions_small: 50, portions_large: 250 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.isValid).toBe(false)
      
      // Back to valid
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 300 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.isValid).toBe(true)
    })

    it('should validate non-negative values in real-time', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      // Valid with zero small portions
      wrapper.vm.allocations[1] = { portions_small: 0, portions_large: 500 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.isValid).toBe(true)
      
      // Valid with zero large portions
      wrapper.vm.allocations[1] = { portions_small: 500, portions_large: 0 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.isValid).toBe(true)
      
      // Valid with both positive
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.isValid).toBe(true)
    })

    it('should validate at least one portion type > 0 in real-time', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool],
        totalPortions: 500
      })
      
      // Valid: SD school has portions, SMP has none
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 0 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.isValid).toBe(true)
      
      // Still valid: both schools have portions
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 100 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.isValid).toBe(false) // Now invalid because total is 600, not 500
      
      // Fix the total
      wrapper.vm.allocations[1] = { portions_small: 150, portions_large: 250 }
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 100 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.isValid).toBe(true)
    })

    it('should display running total correctly as allocations change', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool],
        totalPortions: 800
      })
      
      const summaryText = wrapper.find('.summary-text')
      
      // Initial state
      expect(summaryText.text()).toContain('0 / 800 porsi')
      
      // Add first school allocation
      wrapper.vm.allocations[1] = { portions_small: 150, portions_large: 250 }
      await wrapper.vm.$nextTick()
      expect(summaryText.text()).toContain('400 / 800 porsi')
      
      // Add second school allocation
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 200 }
      await wrapper.vm.$nextTick()
      expect(summaryText.text()).toContain('600 / 800 porsi')
      
      // Complete allocation
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 400 }
      await wrapper.vm.$nextTick()
      expect(summaryText.text()).toContain('800 / 800 porsi')
    })

    it('should update validation hint message in real-time', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      // Partial allocation - should show hint
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 150 }
      await wrapper.vm.$nextTick()
      
      let validationHint = wrapper.find('.validation-hint')
      expect(validationHint.exists()).toBe(true)
      expect(validationHint.text()).toContain('250 porsi lagi')
      
      // Update allocation - hint should update
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 200 }
      await wrapper.vm.$nextTick()
      
      validationHint = wrapper.find('.validation-hint')
      expect(validationHint.exists()).toBe(true)
      expect(validationHint.text()).toContain('100 porsi lagi')
      
      // Complete allocation - hint should disappear
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      
      validationHint = wrapper.find('.validation-hint')
      expect(validationHint.exists()).toBe(false)
    })

    it('should update error message in real-time when exceeding total', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 400
      })
      
      // Exceed by 100
      wrapper.vm.allocations[1] = { portions_small: 250, portions_large: 250 }
      await wrapper.vm.$nextTick()
      
      let errorMessage = wrapper.find('.error-message')
      expect(errorMessage.exists()).toBe(true)
      expect(errorMessage.text()).toContain('100 porsi')
      
      // Exceed by 200
      wrapper.vm.allocations[1] = { portions_small: 300, portions_large: 300 }
      await wrapper.vm.$nextTick()
      
      errorMessage = wrapper.find('.error-message')
      expect(errorMessage.exists()).toBe(true)
      expect(errorMessage.text()).toContain('200 porsi')
      
      // Fix to exact amount
      wrapper.vm.allocations[1] = { portions_small: 150, portions_large: 250 }
      await wrapper.vm.$nextTick()
      
      errorMessage = wrapper.find('.error-message')
      expect(errorMessage.exists()).toBe(false)
    })

    it('should handle multiple schools with real-time calculation', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool, mockSMPSchool, mockSMASchool],
        totalPortions: 1000
      })
      
      // Add allocations incrementally
      expect(wrapper.vm.totalAllocated).toBe(0)
      
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 200 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.totalAllocated).toBe(300)
      expect(wrapper.vm.isValid).toBe(false)
      
      wrapper.vm.allocations[2] = { portions_small: 0, portions_large: 350 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.totalAllocated).toBe(650)
      expect(wrapper.vm.isValid).toBe(false)
      
      wrapper.vm.allocations[3] = { portions_small: 0, portions_large: 350 }
      await wrapper.vm.$nextTick()
      expect(wrapper.vm.totalAllocated).toBe(1000)
      expect(wrapper.vm.isValid).toBe(true)
    })

    it('should emit validation-change event with updated data in real-time', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      // Change allocation and trigger event
      wrapper.vm.allocations[1] = { portions_small: 150, portions_large: 200 }
      await wrapper.vm.$nextTick()
      await wrapper.vm.handleAllocationChange()
      
      let emitted = wrapper.emitted('validation-change')
      expect(emitted).toBeTruthy()
      let lastEmit = emitted[emitted.length - 1][0]
      expect(lastEmit.isValid).toBe(false)
      expect(lastEmit.totalAllocated).toBe(350)
      expect(lastEmit.totalPortions).toBe(500)
      
      // Change to valid state
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      await wrapper.vm.handleAllocationChange()
      
      emitted = wrapper.emitted('validation-change')
      lastEmit = emitted[emitted.length - 1][0]
      expect(lastEmit.isValid).toBe(true)
      expect(lastEmit.totalAllocated).toBe(500)
      expect(lastEmit.totalPortions).toBe(500)
    })

    it('should update summary class in real-time based on validation state', async () => {
      wrapper = createWrapper({
        schools: [mockSDSchool],
        totalPortions: 500
      })
      
      const summary = wrapper.find('.allocation-summary')
      
      // Empty state
      expect(summary.classes()).toContain('summary-empty')
      
      // Invalid state (partial allocation)
      wrapper.vm.allocations[1] = { portions_small: 100, portions_large: 150 }
      await wrapper.vm.$nextTick()
      expect(summary.classes()).toContain('summary-invalid')
      
      // Valid state
      wrapper.vm.allocations[1] = { portions_small: 200, portions_large: 300 }
      await wrapper.vm.$nextTick()
      expect(summary.classes()).toContain('summary-valid')
    })
  })
})
