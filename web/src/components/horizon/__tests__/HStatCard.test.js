import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import HStatCard from '../HStatCard.vue'
import { DollarOutlined } from '@ant-design/icons-vue'

describe('HStatCard', () => {
  it('should render label and value correctly', () => {
    const wrapper = mount(HStatCard, {
      props: {
        label: 'Total Pendapatan',
        value: 'Rp 45.2M'
      }
    })

    expect(wrapper.text()).toContain('Total Pendapatan')
    expect(wrapper.text()).toContain('Rp 45.2M')
  })

  it('should render icon when provided', () => {
    const wrapper = mount(HStatCard, {
      props: {
        icon: DollarOutlined,
        label: 'Earnings',
        value: '$350.4'
      }
    })

    const iconContainer = wrapper.find('.h-stat-card__icon')
    expect(iconContainer.exists()).toBe(true)
  })

  it('should apply custom icon background', () => {
    const customBg = 'linear-gradient(135deg, #05CD99 0%, #26d9a8 100%)'
    const wrapper = mount(HStatCard, {
      props: {
        icon: DollarOutlined,
        iconBg: customBg,
        label: 'Revenue',
        value: '$1000'
      }
    })

    const iconContainer = wrapper.find('.h-stat-card__icon')
    expect(iconContainer.attributes('style')).toContain(customBg)
  })

  it('should render change indicator with increase type', () => {
    const wrapper = mount(HStatCard, {
      props: {
        label: 'Sales',
        value: '100',
        change: '+23%',
        changeType: 'increase'
      }
    })

    const changeIndicator = wrapper.find('.h-stat-card__change')
    expect(changeIndicator.exists()).toBe(true)
    expect(changeIndicator.text()).toContain('+23%')
    expect(changeIndicator.classes()).toContain('h-stat-card__change--increase')
  })

  it('should render change indicator with decrease type', () => {
    const wrapper = mount(HStatCard, {
      props: {
        label: 'Sales',
        value: '100',
        change: '-5%',
        changeType: 'decrease'
      }
    })

    const changeIndicator = wrapper.find('.h-stat-card__change')
    expect(changeIndicator.exists()).toBe(true)
    expect(changeIndicator.text()).toContain('-5%')
    expect(changeIndicator.classes()).toContain('h-stat-card__change--decrease')
  })

  it('should not render change indicator when change is not provided', () => {
    const wrapper = mount(HStatCard, {
      props: {
        label: 'Total',
        value: '50'
      }
    })

    const changeIndicator = wrapper.find('.h-stat-card__change')
    expect(changeIndicator.exists()).toBe(false)
  })

  it('should show skeleton when loading', () => {
    const wrapper = mount(HStatCard, {
      props: {
        label: 'Loading',
        value: '0',
        loading: true
      }
    })

    // Check that skeleton is rendered (a-skeleton tag)
    const skeleton = wrapper.find('a-skeleton')
    expect(skeleton.exists()).toBe(true)
    
    const content = wrapper.find('.h-stat-card__content')
    expect(content.exists()).toBe(false)
  })

  it('should not show skeleton when not loading', () => {
    const wrapper = mount(HStatCard, {
      props: {
        label: 'Loaded',
        value: '100',
        loading: false
      }
    })

    const skeleton = wrapper.find('a-skeleton')
    expect(skeleton.exists()).toBe(false)
    
    const content = wrapper.find('.h-stat-card__content')
    expect(content.exists()).toBe(true)
  })

  it('should accept numeric value', () => {
    const wrapper = mount(HStatCard, {
      props: {
        label: 'Count',
        value: 42
      }
    })

    expect(wrapper.text()).toContain('42')
  })

  it('should have h-card class for styling', () => {
    const wrapper = mount(HStatCard, {
      props: {
        label: 'Test',
        value: '100'
      }
    })

    expect(wrapper.classes()).toContain('h-card')
    expect(wrapper.classes()).toContain('h-stat-card')
  })
})
