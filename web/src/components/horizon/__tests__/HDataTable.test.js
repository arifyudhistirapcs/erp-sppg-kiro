import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import HDataTable from '../HDataTable.vue'

// Mock useBreakpoint composable
vi.mock('@/composables/useBreakpoint', () => ({
  useBreakpoint: () => ({
    isMobile: false,
    isTablet: false,
    isDesktop: true,
    breakpoint: 'xl'
  })
}))

describe('HDataTable', () => {
  const columns = [
    { title: 'Name', dataIndex: 'name', key: 'name' },
    { title: 'Age', dataIndex: 'age', key: 'age' }
  ]

  const dataSource = [
    { key: '1', name: 'John', age: 32 },
    { key: '2', name: 'Jane', age: 28 }
  ]

  it('renders table with data', () => {
    const wrapper = mount(HDataTable, {
      props: {
        columns,
        dataSource
      }
    })

    expect(wrapper.find('.h-data-table').exists()).toBe(true)
    expect(wrapper.find('.h-data-table__table').exists()).toBe(true)
  })

  it('applies h-card class', () => {
    const wrapper = mount(HDataTable, {
      props: {
        columns,
        dataSource
      }
    })

    expect(wrapper.find('.h-card').exists()).toBe(true)
  })

  it('accepts status column type', () => {
    const statusColumns = [
      { title: 'Status', dataIndex: 'status', key: 'status', type: 'status' }
    ]
    const statusData = [
      { key: '1', status: 'Completed' },
      { key: '2', status: 'Pending' }
    ]

    const wrapper = mount(HDataTable, {
      props: {
        columns: statusColumns,
        dataSource: statusData
      }
    })

    // Verify component renders with status columns
    expect(wrapper.find('.h-data-table').exists()).toBe(true)
    expect(wrapper.props('columns')[0].type).toBe('status')
  })

  it('accepts progress column type', () => {
    const progressColumns = [
      { title: 'Progress', dataIndex: 'progress', key: 'progress', type: 'progress' }
    ]
    const progressData = [
      { key: '1', progress: 75 }
    ]

    const wrapper = mount(HDataTable, {
      props: {
        columns: progressColumns,
        dataSource: progressData
      }
    })

    // Verify component renders with progress columns
    expect(wrapper.find('.h-data-table').exists()).toBe(true)
    expect(wrapper.props('columns')[0].type).toBe('progress')
  })

  it('accepts loading prop', () => {
    const wrapper = mount(HDataTable, {
      props: {
        columns,
        dataSource,
        loading: true
      }
    })

    expect(wrapper.props('loading')).toBe(true)
  })

  it('accepts pagination prop', () => {
    const pagination = {
      current: 1,
      pageSize: 10,
      total: 50
    }

    const wrapper = mount(HDataTable, {
      props: {
        columns,
        dataSource,
        pagination
      }
    })

    expect(wrapper.props('pagination')).toEqual(pagination)
  })

  it('accepts rowSelection prop', () => {
    const rowSelection = {
      selectedRowKeys: ['1'],
      onChange: vi.fn()
    }

    const wrapper = mount(HDataTable, {
      props: {
        columns,
        dataSource,
        rowSelection
      }
    })

    expect(wrapper.props('rowSelection')).toEqual(rowSelection)
  })

  it('accepts mobileCardView prop', () => {
    const wrapper = mount(HDataTable, {
      props: {
        columns,
        dataSource,
        mobileCardView: false
      }
    })

    expect(wrapper.props('mobileCardView')).toBe(false)
  })

  it('getStatusType returns correct status types', () => {
    const wrapper = mount(HDataTable, {
      props: {
        columns,
        dataSource
      }
    })

    const vm = wrapper.vm

    expect(vm.getStatusType('Completed')).toBe('success')
    expect(vm.getStatusType('Pending')).toBe('warning')
    expect(vm.getStatusType('Error')).toBe('error')
    expect(vm.getStatusType('Disabled')).toBe('disabled')
    expect(vm.getStatusType('Unknown')).toBe('default')
  })

  it('getProgressColor returns correct colors', () => {
    const wrapper = mount(HDataTable, {
      props: {
        columns,
        dataSource
      }
    })

    const vm = wrapper.vm

    expect(vm.getProgressColor(85)).toBe('var(--h-success)')
    expect(vm.getProgressColor(60)).toBe('var(--h-warning)')
    expect(vm.getProgressColor(30)).toBe('var(--h-error)')
  })
})
