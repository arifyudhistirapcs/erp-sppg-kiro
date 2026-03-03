import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import HChartCard from '../HChartCard.vue'

describe('HChartCard', () => {
  it('should render title correctly', () => {
    const wrapper = mount(HChartCard, {
      props: {
        title: 'Weekly Revenue'
      }
    })

    expect(wrapper.text()).toContain('Weekly Revenue')
  })

  it('should render subtitle when provided', () => {
    const wrapper = mount(HChartCard, {
      props: {
        title: 'Daily Traffic',
        subtitle: 'Last 7 days'
      }
    })

    expect(wrapper.text()).toContain('Daily Traffic')
    expect(wrapper.text()).toContain('Last 7 days')
  })

  it('should not render subtitle when not provided', () => {
    const wrapper = mount(HChartCard, {
      props: {
        title: 'Chart Title'
      }
    })

    const subtitle = wrapper.find('.h-chart-card__subtitle')
    expect(subtitle.exists()).toBe(false)
  })

  it('should apply default height of 320px', () => {
    const wrapper = mount(HChartCard, {
      props: {
        title: 'Test Chart'
      }
    })

    const chartArea = wrapper.find('.h-chart-card__chart')
    expect(chartArea.attributes('style')).toContain('height: 320px')
  })

  it('should apply custom height', () => {
    const wrapper = mount(HChartCard, {
      props: {
        title: 'Test Chart',
        height: 400
      }
    })

    const chartArea = wrapper.find('.h-chart-card__chart')
    expect(chartArea.attributes('style')).toContain('height: 400px')
  })

  it('should render default slot content', () => {
    const wrapper = mount(HChartCard, {
      props: {
        title: 'Chart with Content'
      },
      slots: {
        default: '<div class="test-chart">Chart Content</div>'
      }
    })

    const chartContent = wrapper.find('.test-chart')
    expect(chartContent.exists()).toBe(true)
    expect(chartContent.text()).toBe('Chart Content')
  })

  it('should render header-right slot content', () => {
    const wrapper = mount(HChartCard, {
      props: {
        title: 'Chart with Actions'
      },
      slots: {
        'header-right': '<button class="test-action">Action</button>'
      }
    })

    const headerRight = wrapper.find('.h-chart-card__header-right')
    expect(headerRight.exists()).toBe(true)
    
    const actionButton = wrapper.find('.test-action')
    expect(actionButton.exists()).toBe(true)
    expect(actionButton.text()).toBe('Action')
  })

  it('should not render header-right when slot is not provided', () => {
    const wrapper = mount(HChartCard, {
      props: {
        title: 'Simple Chart'
      }
    })

    const headerRight = wrapper.find('.h-chart-card__header-right')
    expect(headerRight.exists()).toBe(false)
  })

  it('should show skeleton when loading', () => {
    const wrapper = mount(HChartCard, {
      props: {
        title: 'Loading Chart',
        loading: true
      }
    })

    const skeleton = wrapper.find('a-skeleton')
    expect(skeleton.exists()).toBe(true)
    
    const content = wrapper.find('.h-chart-card__content')
    expect(content.exists()).toBe(false)
  })

  it('should not show skeleton when not loading', () => {
    const wrapper = mount(HChartCard, {
      props: {
        title: 'Loaded Chart',
        loading: false
      }
    })

    const skeleton = wrapper.find('a-skeleton')
    expect(skeleton.exists()).toBe(false)
    
    const content = wrapper.find('.h-chart-card__content')
    expect(content.exists()).toBe(true)
  })

  it('should have h-card class for styling', () => {
    const wrapper = mount(HChartCard, {
      props: {
        title: 'Test'
      }
    })

    expect(wrapper.classes()).toContain('h-card')
    expect(wrapper.classes()).toContain('h-chart-card')
  })

  it('should render both slots together', () => {
    const wrapper = mount(HChartCard, {
      props: {
        title: 'Full Chart',
        subtitle: 'With all features'
      },
      slots: {
        default: '<div class="chart">Chart</div>',
        'header-right': '<button>Dropdown</button>'
      }
    })

    expect(wrapper.text()).toContain('Full Chart')
    expect(wrapper.text()).toContain('With all features')
    expect(wrapper.find('.chart').exists()).toBe(true)
    expect(wrapper.find('button').exists()).toBe(true)
  })
})
