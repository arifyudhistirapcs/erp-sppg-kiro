<template>
  <div class="example-container">
    <h1>HDataTable Examples</h1>
    
    <!-- Example 1: Basic Table -->
    <section class="example-section">
      <h2>Basic Table</h2>
      <HDataTable
        :columns="basicColumns"
        :data-source="basicData"
        :loading="false"
      />
    </section>
    
    <!-- Example 2: Table with Status Badges -->
    <section class="example-section">
      <h2>Table with Status Badges</h2>
      <HDataTable
        :columns="statusColumns"
        :data-source="statusData"
      />
    </section>
    
    <!-- Example 3: Table with Progress Bars -->
    <section class="example-section">
      <h2>Table with Progress Bars</h2>
      <HDataTable
        :columns="progressColumns"
        :data-source="progressData"
      />
    </section>
    
    <!-- Example 4: Table with Actions -->
    <section class="example-section">
      <h2>Table with Action Buttons</h2>
      <HDataTable
        :columns="actionColumns"
        :data-source="actionData"
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
    </section>
    
    <!-- Example 5: Table with Row Selection -->
    <section class="example-section">
      <h2>Table with Row Selection</h2>
      <p>Selected: {{ selectedRowKeys.join(', ') }}</p>
      <HDataTable
        :columns="basicColumns"
        :data-source="basicData"
        :row-selection="{
          selectedRowKeys: selectedRowKeys,
          onChange: onSelectChange
        }"
      />
    </section>
    
    <!-- Example 6: Mobile Card View -->
    <section class="example-section">
      <h2>Mobile Card View (Resize to see)</h2>
      <HDataTable
        :columns="statusColumns"
        :data-source="statusData"
        :mobile-card-view="true"
      />
    </section>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { HDataTable } from '@/components/horizon'

// Basic Table Data
const basicColumns = [
  {
    title: 'Name',
    dataIndex: 'name',
    key: 'name'
  },
  {
    title: 'Age',
    dataIndex: 'age',
    key: 'age'
  },
  {
    title: 'Address',
    dataIndex: 'address',
    key: 'address'
  }
]

const basicData = [
  {
    key: '1',
    name: 'John Brown',
    age: 32,
    address: 'New York No. 1 Lake Park'
  },
  {
    key: '2',
    name: 'Jim Green',
    age: 42,
    address: 'London No. 1 Lake Park'
  },
  {
    key: '3',
    name: 'Joe Black',
    age: 32,
    address: 'Sidney No. 1 Lake Park'
  }
]

// Status Table Data
const statusColumns = [
  {
    title: 'Task',
    dataIndex: 'task',
    key: 'task'
  },
  {
    title: 'Status',
    dataIndex: 'status',
    key: 'status',
    type: 'status'
  },
  {
    title: 'Date',
    dataIndex: 'date',
    key: 'date'
  }
]

const statusData = [
  {
    key: '1',
    task: 'Design Homepage',
    status: 'Completed',
    date: '2024-01-15'
  },
  {
    key: '2',
    task: 'Implement API',
    status: 'In Progress',
    date: '2024-01-16'
  },
  {
    key: '3',
    task: 'Code Review',
    status: 'Pending',
    date: '2024-01-17'
  },
  {
    key: '4',
    task: 'Deploy to Production',
    status: 'Error',
    date: '2024-01-18'
  },
  {
    key: '5',
    task: 'Old Feature',
    status: 'Disabled',
    date: '2024-01-10'
  }
]

// Progress Table Data
const progressColumns = [
  {
    title: 'Project',
    dataIndex: 'project',
    key: 'project'
  },
  {
    title: 'Progress',
    dataIndex: 'progress',
    key: 'progress',
    type: 'progress'
  },
  {
    title: 'Team',
    dataIndex: 'team',
    key: 'team'
  }
]

const progressData = [
  {
    key: '1',
    project: 'Website Redesign',
    progress: 85,
    team: 'Frontend Team'
  },
  {
    key: '2',
    project: 'Mobile App',
    progress: 60,
    team: 'Mobile Team'
  },
  {
    key: '3',
    project: 'API Integration',
    progress: 30,
    team: 'Backend Team'
  }
]

// Action Table Data
const actionColumns = [
  {
    title: 'Name',
    dataIndex: 'name',
    key: 'name'
  },
  {
    title: 'Email',
    dataIndex: 'email',
    key: 'email'
  },
  {
    title: 'Actions',
    key: 'actions',
    type: 'actions'
  }
]

const actionData = [
  {
    key: '1',
    name: 'Alice Johnson',
    email: 'alice@example.com'
  },
  {
    key: '2',
    name: 'Bob Smith',
    email: 'bob@example.com'
  }
]

// Row Selection
const selectedRowKeys = ref([])

const onSelectChange = (keys) => {
  selectedRowKeys.value = keys
}

// Action Handlers
const handleEdit = (record) => {
  console.log('Edit:', record)
  alert(`Editing: ${record.name}`)
}

const handleDelete = (record) => {
  console.log('Delete:', record)
  alert(`Deleting: ${record.name}`)
}
</script>

<style scoped>
.example-container {
  padding: var(--h-spacing-6);
  background: var(--h-bg-primary);
  min-height: 100vh;
}

h1 {
  font-size: var(--h-text-3xl);
  font-weight: var(--h-font-bold);
  color: var(--h-text-primary);
  margin-bottom: var(--h-spacing-8);
}

.example-section {
  margin-bottom: var(--h-spacing-8);
}

.example-section h2 {
  font-size: var(--h-text-xl);
  font-weight: var(--h-font-semibold);
  color: var(--h-text-primary);
  margin-bottom: var(--h-spacing-4);
}

.example-section p {
  font-size: var(--h-text-sm);
  color: var(--h-text-secondary);
  margin-bottom: var(--h-spacing-3);
}
</style>
