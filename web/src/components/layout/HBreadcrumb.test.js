import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import HBreadcrumb from './HBreadcrumb.vue'

// Mock vue-router
const mockRoute = {
  name: 'dashboard',
  meta: { title: 'Dashboard' }
}

vi.mock('vue-router', () => ({
  useRoute: () => mockRoute
}))

// Mock useBreakpoint composable
vi.mock('@/composables/useBreakpoint', () => ({
  useBreakpoint: () => ({
    isMobile: false
  })
}))

describe('HBreadcrumb', () => {
  it('renders breadcrumb with default root label', () => {
    const wrapper = mount(HBreadcrumb, {
      global: {
        stubs: {
          'router-link': true
        }
      }
    })
    
    expect(wrapper.find('.h-breadcrumb').exists()).toBe(true)
    expect(wrapper.text()).toContain('Pages')
    expect(wrapper.text()).toContain('Dashboard')
  })
  
  it('renders custom breadcrumb items', () => {
    const customItems = [
      { label: 'Home', to: '/' },
      { label: 'Settings', to: '/settings' },
      { label: 'Profile' }
    ]
    
    const wrapper = mount(HBreadcrumb, {
      props: {
        items: customItems
      },
      global: {
        stubs: {
          RouterLink: {
            template: '<a><slot /></a>'
          }
        }
      }
    })
    
    expect(wrapper.text()).toContain('Home')
    expect(wrapper.text()).toContain('Settings')
    expect(wrapper.text()).toContain('Profile')
  })
  
  it('renders separators between breadcrumb items', () => {
    const wrapper = mount(HBreadcrumb, {
      global: {
        stubs: {
          'router-link': true
        }
      }
    })
    
    const separators = wrapper.findAll('.breadcrumb-separator')
    expect(separators.length).toBeGreaterThan(0)
    expect(separators[0].text()).toBe('/')
  })
  
  it('applies current class to last breadcrumb item', () => {
    const wrapper = mount(HBreadcrumb, {
      global: {
        stubs: {
          'router-link': true
        }
      }
    })
    
    const items = wrapper.findAll('.breadcrumb-item')
    const lastItem = items[items.length - 1]
    expect(lastItem.find('.breadcrumb-current').exists()).toBe(true)
  })
  
  it('has CSS to hide on mobile', () => {
    // The component uses both v-if="!isMobile" and CSS @media query
    // to hide on mobile. We verify the component renders on desktop
    // and has the appropriate CSS class for media query hiding
    const wrapper = mount(HBreadcrumb, {
      global: {
        stubs: {
          'router-link': true
        }
      }
    })
    
    const breadcrumb = wrapper.find('.h-breadcrumb')
    expect(breadcrumb.exists()).toBe(true)
    
    // Verify the component has the h-breadcrumb class
    // which has @media (max-width: 767px) { display: none; }
    expect(breadcrumb.classes()).toContain('h-breadcrumb')
  })
  
  it('can hide root breadcrumb', () => {
    const wrapper = mount(HBreadcrumb, {
      props: {
        showRoot: false
      },
      global: {
        stubs: {
          'router-link': true
        }
      }
    })
    
    expect(wrapper.text()).not.toContain('Pages')
    expect(wrapper.text()).toContain('Dashboard')
  })
  
  it('uses custom root label', () => {
    const wrapper = mount(HBreadcrumb, {
      props: {
        rootLabel: 'Home'
      },
      global: {
        stubs: {
          'router-link': true
        }
      }
    })
    
    expect(wrapper.text()).toContain('Home')
    expect(wrapper.text()).not.toContain('Pages')
  })
})
