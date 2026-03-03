import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { defineComponent, ref, h } from 'vue'
import { mount } from '@vue/test-utils'
import { useHorizonChart } from './useHorizonChart'
import * as echarts from 'echarts'

// Mock echarts
vi.mock('echarts', () => ({
  init: vi.fn(() => ({
    setOption: vi.fn(),
    getOption: vi.fn(() => ({})),
    resize: vi.fn(),
    dispose: vi.fn()
  }))
}))

// Mock useDarkMode
vi.mock('./useDarkMode', () => ({
  useDarkMode: vi.fn(() => ({
    isDark: ref(false)
  }))
}))

// Mock horizonChartTheme
vi.mock('@/utils/horizonChartTheme', () => ({
  getHorizonChartTheme: vi.fn((isDark) => ({
    color: ['#5A4372', '#3D2B53'],
    backgroundColor: 'transparent',
    textStyle: {
      fontFamily: 'DM Sans',
      fontSize: 14,
      color: isDark ? '#F8FDEA' : '#322837'
    }
  }))
}))

describe('useHorizonChart', () => {
  let mockChartInstance

  beforeEach(() => {
    // Reset mocks
    vi.clearAllMocks()
    
    // Get mock chart instance
    mockChartInstance = {
      setOption: vi.fn(),
      getOption: vi.fn(() => ({})),
      resize: vi.fn(),
      dispose: vi.fn()
    }
    
    echarts.init.mockReturnValue(mockChartInstance)
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  const createTestComponent = (options = {}) => {
    return defineComponent({
      setup() {
        const chartRef = ref(null)
        const composable = useHorizonChart(chartRef, options)
        return {
          chartRef,
          ...composable
        }
      },
      render() {
        return h('div', { ref: 'chartRef' })
      }
    })
  }

  it('should initialize chart instance on mount', async () => {
    const TestComponent = createTestComponent()
    const wrapper = mount(TestComponent)
    
    await wrapper.vm.$nextTick()
    
    expect(echarts.init).toHaveBeenCalled()
    expect(wrapper.vm.chart).toBeTruthy()
    
    wrapper.unmount()
  })

  it('should apply options with Horizon theme', async () => {
    const options = {
      title: { text: 'Test Chart' },
      series: [{ type: 'line', data: [1, 2, 3] }]
    }
    
    const TestComponent = createTestComponent(options)
    const wrapper = mount(TestComponent)
    
    await wrapper.vm.$nextTick()
    
    expect(mockChartInstance.setOption).toHaveBeenCalled()
    const callArgs = mockChartInstance.setOption.mock.calls[0][0]
    expect(callArgs).toHaveProperty('color')
    expect(callArgs).toHaveProperty('textStyle')
    expect(callArgs.title).toEqual({ text: 'Test Chart' })
    
    wrapper.unmount()
  })

  it('should handle resize events', async () => {
    const TestComponent = createTestComponent()
    const wrapper = mount(TestComponent)
    
    await wrapper.vm.$nextTick()
    
    wrapper.vm.resize()
    
    expect(mockChartInstance.resize).toHaveBeenCalled()
    
    wrapper.unmount()
  })

  it('should add window resize listener', async () => {
    const addEventListenerSpy = vi.spyOn(window, 'addEventListener')
    
    const TestComponent = createTestComponent()
    const wrapper = mount(TestComponent)
    
    await wrapper.vm.$nextTick()
    
    expect(addEventListenerSpy).toHaveBeenCalledWith('resize', expect.any(Function))
    
    wrapper.unmount()
    addEventListenerSpy.mockRestore()
  })

  it('should dispose chart on unmount', async () => {
    const TestComponent = createTestComponent()
    const wrapper = mount(TestComponent)
    
    await wrapper.vm.$nextTick()
    
    expect(wrapper.vm.chart).toBeTruthy()
    
    wrapper.unmount()
    
    expect(mockChartInstance.dispose).toHaveBeenCalled()
  })

  it('should update theme when setOption is called', async () => {
    const TestComponent = createTestComponent()
    const wrapper = mount(TestComponent)
    
    await wrapper.vm.$nextTick()
    
    const newOptions = {
      series: [{ type: 'bar', data: [4, 5, 6] }]
    }
    
    wrapper.vm.setOption(newOptions)
    
    expect(mockChartInstance.setOption).toHaveBeenCalled()
    const callArgs = mockChartInstance.setOption.mock.calls[mockChartInstance.setOption.mock.calls.length - 1][0]
    expect(callArgs).toHaveProperty('color')
    expect(callArgs).toHaveProperty('textStyle')
    expect(callArgs.series).toEqual([{ type: 'bar', data: [4, 5, 6] }])
    
    wrapper.unmount()
  })

  it('should return all expected functions', () => {
    const TestComponent = createTestComponent()
    const wrapper = mount(TestComponent)
    
    expect(wrapper.vm).toHaveProperty('chart')
    expect(wrapper.vm).toHaveProperty('setOption')
    expect(wrapper.vm).toHaveProperty('updateTheme')
    expect(wrapper.vm).toHaveProperty('resize')
    expect(wrapper.vm).toHaveProperty('initChart')
    
    wrapper.unmount()
  })
})
