# HDataTable Component

A modern, responsive data table component that wraps Ant Design Vue Table with Horizon UI styling. Features mobile card view, status badges, progress bars, and customizable action buttons.

## Features

- ✅ Horizon UI styling with custom colors and typography
- ✅ Mobile-responsive with card view option
- ✅ Status badge rendering with colored dots
- ✅ Progress bar rendering with dynamic colors
- ✅ Action buttons with larger touch targets on mobile
- ✅ Horizontal scroll with indicator on mobile
- ✅ Dark mode support
- ✅ Row selection support
- ✅ Pagination support
- ✅ Custom cell rendering via slots

## Props

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `columns` | `Array` | **required** | Table columns configuration (Ant Design format) |
| `dataSource` | `Array` | `[]` | Table data source |
| `loading` | `Boolean` | `false` | Loading state |
| `pagination` | `Object\|Boolean` | `{ current: 1, pageSize: 10, ... }` | Pagination configuration |
| `rowSelection` | `Object` | `null` | Row selection configuration |
| `mobileCardView` | `Boolean` | `true` | Enable mobile card view (< 768px) |

## Column Types

The component supports special column types via the `type` property:

### Status Badge (`type: 'status'`)

Renders a colored badge with a dot indicator. Automatically detects status type:

- **Success**: "approved", "done", "completed", "success" → Green
- **Warning**: "pending", "waiting", "in progress" → Orange
- **Error**: "error", "failed", "rejected" → Red
- **Disabled**: "disable", "inactive" → Gray
- **Default**: Any other status → Default gray

```javascript
{
  title: 'Status',
  dataIndex: 'status',
  key: 'status',
  type: 'status'
}
```

### Progress Bar (`type: 'progress'`)

Renders a progress bar with dynamic colors:

- **Green**: >= 80%
- **Orange**: >= 50%
- **Red**: < 50%

```javascript
{
  title: 'Progress',
  dataIndex: 'progress',
  key: 'progress',
  type: 'progress'
}
```

### Action Buttons (`type: 'actions'`)

Renders action buttons using the `actions` slot:

```javascript
{
  title: 'Actions',
  key: 'actions',
  type: 'actions'
}
```

## Slots

### `actions`

Custom action buttons for each row.

**Scope**: `{ record }`

```vue
<HDataTable :columns="columns" :data-source="data">
  <template #actions="{ record }">
    <a-button type="primary" size="small" @click="edit(record)">
      Edit
    </a-button>
    <a-button danger size="small" @click="remove(record)">
      Delete
    </a-button>
  </template>
</HDataTable>
```

### `cell-{dataIndex}`

Custom cell rendering for specific columns.

**Scope**: `{ record, text }`

```vue
<HDataTable :columns="columns" :data-source="data">
  <template #cell-name="{ record, text }">
    <strong>{{ text }}</strong>
    <span v-if="record.isNew" class="badge">New</span>
  </template>
</HDataTable>
```

## Usage Examples

### Basic Table

```vue
<template>
  <HDataTable
    :columns="columns"
    :data-source="data"
  />
</template>

<script setup>
import { HDataTable } from '@/components/horizon'

const columns = [
  { title: 'Name', dataIndex: 'name', key: 'name' },
  { title: 'Age', dataIndex: 'age', key: 'age' },
  { title: 'Address', dataIndex: 'address', key: 'address' }
]

const data = [
  { key: '1', name: 'John', age: 32, address: 'New York' },
  { key: '2', name: 'Jane', age: 28, address: 'London' }
]
</script>
```

### Table with Status Badges

```vue
<template>
  <HDataTable
    :columns="columns"
    :data-source="data"
  />
</template>

<script setup>
import { HDataTable } from '@/components/horizon'

const columns = [
  { title: 'Task', dataIndex: 'task', key: 'task' },
  { title: 'Status', dataIndex: 'status', key: 'status', type: 'status' },
  { title: 'Date', dataIndex: 'date', key: 'date' }
]

const data = [
  { key: '1', task: 'Design', status: 'Completed', date: '2024-01-15' },
  { key: '2', task: 'Development', status: 'In Progress', date: '2024-01-16' },
  { key: '3', task: 'Testing', status: 'Pending', date: '2024-01-17' }
]
</script>
```

### Table with Progress Bars

```vue
<template>
  <HDataTable
    :columns="columns"
    :data-source="data"
  />
</template>

<script setup>
import { HDataTable } from '@/components/horizon'

const columns = [
  { title: 'Project', dataIndex: 'project', key: 'project' },
  { title: 'Progress', dataIndex: 'progress', key: 'progress', type: 'progress' },
  { title: 'Team', dataIndex: 'team', key: 'team' }
]

const data = [
  { key: '1', project: 'Website', progress: 85, team: 'Frontend' },
  { key: '2', project: 'Mobile App', progress: 60, team: 'Mobile' },
  { key: '3', project: 'API', progress: 30, team: 'Backend' }
]
</script>
```

### Table with Action Buttons

```vue
<template>
  <HDataTable
    :columns="columns"
    :data-source="data"
  >
    <template #actions="{ record }">
      <a-button type="primary" size="small" @click="handleEdit(record)">
        Edit
      </a-button>
      <a-button danger size="small" @click="handleDelete(record)">
        Delete
      </a-button>
    </template>
  </HDataTable>
</template>

<script setup>
import { HDataTable } from '@/components/horizon'

const columns = [
  { title: 'Name', dataIndex: 'name', key: 'name' },
  { title: 'Email', dataIndex: 'email', key: 'email' },
  { title: 'Actions', key: 'actions', type: 'actions' }
]

const data = [
  { key: '1', name: 'Alice', email: 'alice@example.com' },
  { key: '2', name: 'Bob', email: 'bob@example.com' }
]

const handleEdit = (record) => {
  console.log('Edit:', record)
}

const handleDelete = (record) => {
  console.log('Delete:', record)
}
</script>
```

### Table with Row Selection

```vue
<template>
  <HDataTable
    :columns="columns"
    :data-source="data"
    :row-selection="{
      selectedRowKeys: selectedKeys,
      onChange: onSelectChange
    }"
  />
</template>

<script setup>
import { ref } from 'vue'
import { HDataTable } from '@/components/horizon'

const selectedKeys = ref([])

const columns = [
  { title: 'Name', dataIndex: 'name', key: 'name' },
  { title: 'Age', dataIndex: 'age', key: 'age' }
]

const data = [
  { key: '1', name: 'John', age: 32 },
  { key: '2', name: 'Jane', age: 28 }
]

const onSelectChange = (keys) => {
  selectedKeys.value = keys
  console.log('Selected:', keys)
}
</script>
```

### Table with Pagination

```vue
<template>
  <HDataTable
    :columns="columns"
    :data-source="data"
    :pagination="{
      current: currentPage,
      pageSize: 10,
      total: totalItems,
      showSizeChanger: true,
      showTotal: (total) => `Total ${total} items`,
      onChange: handlePageChange
    }"
  />
</template>

<script setup>
import { ref } from 'vue'
import { HDataTable } from '@/components/horizon'

const currentPage = ref(1)
const totalItems = ref(100)

const handlePageChange = (page, pageSize) => {
  currentPage.value = page
  // Fetch data for new page
}
</script>
```

## Styling

### Custom Table Styling

The component uses Horizon UI design tokens:

- **Row Height**: 56px
- **Header Text**: 12px uppercase, #74788C, font-weight 700
- **Cell Text**: 14px, #322837
- **Hover Background**: #F8FDEA
- **Status Badge**: Radius 8px, padding 4px 12px
- **Touch Targets (Mobile)**: 44px minimum

### Dark Mode

The component automatically supports dark mode when the `.dark` class is applied to the document root:

```javascript
// Toggle dark mode
document.documentElement.classList.toggle('dark')
```

## Mobile Responsiveness

### Card View (< 768px)

When `mobileCardView` is enabled (default), the table automatically switches to a card layout on mobile devices:

- Each row becomes a card
- Fields are stacked vertically
- Labels are shown above values
- Touch-friendly spacing and targets

### Horizontal Scroll

When card view is disabled, the table uses horizontal scroll on mobile with a visible scroll indicator.

### Touch Targets

Action buttons automatically increase to 44x44px on mobile for better touch accessibility.

## Accessibility

- Semantic HTML structure
- Keyboard navigation support (via Ant Design)
- ARIA labels for interactive elements
- High contrast colors for readability
- Minimum touch target size (44px) on mobile

## Browser Support

- Chrome (latest 2 versions)
- Firefox (latest 2 versions)
- Safari (latest 2 versions)
- Edge (latest 2 versions)
- Safari iOS (latest 2 versions)
- Chrome Android (latest 2 versions)

## Performance Tips

1. **Use `key` prop**: Always provide a unique `key` for each data item
2. **Pagination**: Use pagination for large datasets (> 50 rows)
3. **Virtual scrolling**: For very large datasets, consider using Ant Design's virtual scroll feature
4. **Lazy loading**: Load data on demand rather than all at once
5. **Memoization**: Use `computed` or `useMemo` for expensive column calculations

## Related Components

- [HStatCard](./HStatCard.md) - Statistics card component
- [HChartCard](./HChartCard.md) - Chart card component
- [HKanbanCard](./HKanbanCard.md) - Kanban card component (coming soon)

## See Also

- [Ant Design Vue Table Documentation](https://antdv.com/components/table)
- [Horizon UI Design System](../../../docs/HORIZON_UI_GUIDE.md)
