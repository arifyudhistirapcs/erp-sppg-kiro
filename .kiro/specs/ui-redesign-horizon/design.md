# UI Redesign - Design Document

## 🎨 Design System

### Component Library Architecture

Kita akan membuat wrapper components yang mengadopsi Horizon UI style di atas Ant Design Vue yang sudah ada.

**Strategy**: Hybrid Approach
- Keep Ant Design Vue sebagai base
- Create Horizon-styled wrapper components
- Gradual migration per page
- No breaking changes

### Core Components to Build

#### 1. HStatCard.vue

Mini statistics card dengan icon, value, dan change indicator.

**Props**:
```javascript
{
  icon: String,           // Icon name atau component
  iconBg: String,         // Gradient background
  label: String,          // "Earnings", "Spend this month"
  value: String|Number,   // "$350.4", "154"
  change: String,         // "+23%"
  changeType: String,     // "increase" | "decrease"
  loading: Boolean
}
```

**Usage**:
```vue
<HStatCard
  icon="DollarOutlined"
  icon-bg="linear-gradient(135deg, #f82c17 0%, #ff4d38 100%)"
  label="Total Pendapatan"
  value="Rp 45.2M"
  change="+12.5%"
  change-type="increase"
/>
```

#### 2. HChartCard.vue

Card wrapper untuk charts dengan header dan actions.

**Props**:
```javascript
{
  title: String,
  subtitle: String,
  height: Number,         // Default 320
  loading: Boolean,
  actions: Array          // Dropdown actions
}
```

**Slots**:
- `default`: Chart content
- `header-right`: Custom header actions

#### 3. HDataTable.vue

Modern table dengan Horizon UI styling.

**Props**:
```javascript
{
  columns: Array,
  dataSource: Array,
  loading: Boolean,
  pagination: Object,
  rowSelection: Object,
  showHeader: Boolean
}
```

**Features**:
- Custom row hover
- Status badges
- Progress bars
- Action buttons
- Sorting & filtering

#### 4. HKanbanCard.vue

Card untuk kanban board items.

**Props**:
```javascript
{
  title: String,
  description: String,
  image: String,
  status: String,
  assignees: Array,
  dueDate: String
}
```

#### 5. HSidebar.vue

Sidebar navigation dengan Horizon UI style.

**Features**:
- Collapsible
- Icon + label
- Active state highlighting
- Nested menu support
- Role-based visibility

#### 6. HHeader.vue

Top header dengan search, notifications, theme toggle.

**Features**:
- Breadcrumb navigation
- Global search
- Notification dropdown
- Dark mode toggle
- User menu

#### 7. ThemeToggle.vue

Dark/Light mode toggle button.

**Features**:
- Smooth transition
- Icon animation
- Persist preference
- System preference detection

### Layout Components

#### HorizonLayout.vue

Main layout wrapper combining sidebar and header.

```vue
<template>
  <div class="horizon-layout" :class="{ 'dark': isDark }">
    <HSidebar />
    <div class="horizon-main">
      <HHeader />
      <div class="horizon-content">
        <slot />
      </div>
    </div>
  </div>
</template>
```

## 🎨 Styling Strategy

### CSS Architecture

```
styles/
├── horizon/
│   ├── variables.css      # CSS variables
│   ├── base.css           # Reset & base styles
│   ├── components.css     # Component styles
│   ├── utilities.css      # Utility classes
│   └── dark-mode.css      # Dark mode overrides
└── theme.css              # Keep existing (backward compat)
```

### CSS Variables Structure

```css
/* horizon/variables.css */
:root {
  /* Colors */
  --h-primary: #f82c17;
  --h-primary-hover: #e02915;
  --h-primary-light: #ff4d38;
  
  /* Backgrounds */
  --h-bg-primary: #F4F7FE;
  --h-bg-secondary: #FFFFFF;
  --h-bg-card: #FFFFFF;
  
  /* Text */
  --h-text-primary: #2B3674;
  --h-text-secondary: #A3AED0;
  
  /* Spacing */
  --h-spacing-1: 4px;
  --h-spacing-2: 8px;
  /* ... */
  
  /* Shadows */
  --h-shadow-card: 0px 18px 40px rgba(112, 144, 176, 0.12);
  
  /* Radius */
  --h-radius-md: 12px;
  --h-radius-lg: 16px;
}

/* Dark mode */
.dark {
  --h-bg-primary: #0B1437;
  --h-bg-secondary: #111C44;
  --h-bg-card: #111C44;
  --h-text-primary: #FFFFFF;
  --h-border-color: #1B254B;
}
```

### Utility Classes

```css
/* horizon/utilities.css */
.h-card {
  background: var(--h-bg-card);
  border-radius: var(--h-radius-lg);
  box-shadow: var(--h-shadow-card);
  padding: var(--h-spacing-6);
}

.h-card-hover {
  transition: all 0.2s ease;
}

.h-card-hover:hover {
  transform: translateY(-4px);
  box-shadow: var(--h-shadow-xl);
}

.h-gradient-primary {
  background: linear-gradient(135deg, #f82c17 0%, #ff4d38 100%);
}

.h-text-primary {
  color: var(--h-text-primary);
}

.h-text-secondary {
  color: var(--h-text-secondary);
}
```

## 📐 Page Layouts

### Dashboard Layout

```
┌─────────────────────────────────────────────────┐
│ Header (Breadcrumb, Search, Actions)            │
├─────────────────────────────────────────────────┤
│ Stats Row (4 cards)                             │
│ [Earnings] [Spend] [Sales] [Tasks]              │
├─────────────────────────────────────────────────┤
│ Charts Row                                       │
│ [Weekly Revenue - 60%] [Daily Traffic - 40%]    │
├─────────────────────────────────────────────────┤
│ Tables Row                                       │
│ [Check Table - 50%] [Complex Table - 50%]       │
└─────────────────────────────────────────────────┘
```

### KDS Layout (Cooking/Packing/Cleaning)

```
┌─────────────────────────────────────────────────┐
│ Header + Filters                                 │
├─────────────────────────────────────────────────┤
│ Status Cards Row                                 │
│ [Pending] [In Progress] [Completed]              │
├─────────────────────────────────────────────────┤
│ Kanban Board                                     │
│ [Backlog] [In Progress] [Done]                   │
│   Card 1     Card 3        Card 5                │
│   Card 2     Card 4        Card 6                │
└─────────────────────────────────────────────────┘
```

### Form Layout

```
┌─────────────────────────────────────────────────┐
│ Header (Title, Actions)                          │
├─────────────────────────────────────────────────┤
│ Form Card                                        │
│ ┌─────────────────────────────────────────────┐ │
│ │ Section 1                                   │ │
│ │ [Input Fields]                              │ │
│ ├─────────────────────────────────────────────┤ │
│ │ Section 2                                   │ │
│ │ [Input Fields]                              │ │
│ └─────────────────────────────────────────────┘ │
│ [Cancel] [Save]                                  │
└─────────────────────────────────────────────────┘
```

## 🎯 Implementation Phases

### Phase 1: Foundation (Week 1-2)

**Deliverables**:
1. CSS variables setup
2. Base utility classes
3. Theme toggle functionality
4. Dark mode implementation

**Files**:
- `styles/horizon/variables.css`
- `styles/horizon/base.css`
- `styles/horizon/utilities.css`
- `styles/horizon/dark-mode.css`
- `composables/useDarkMode.js`

### Phase 2: Core Components (Week 3-4)

**Deliverables**:
1. HStatCard component
2. HChartCard component
3. HDataTable component
4. HSidebar component
5. HHeader component
6. ThemeToggle component

**Files**:
- `components/horizon/HStatCard.vue`
- `components/horizon/HChartCard.vue`
- `components/horizon/HDataTable.vue`
- `components/layout/HSidebar.vue`
- `components/layout/HHeader.vue`
- `components/layout/ThemeToggle.vue`
- `layouts/HorizonLayout.vue`

### Phase 3: Dashboard Pages (Week 5-6)

**Pages to Redesign**:
1. Dashboard Kepala SPPG
2. Dashboard Kepala Yayasan
3. Monitoring Aktivitas

**Approach**:
- Create new versions alongside existing
- Feature flag to toggle between old/new
- A/B testing with select users

### Phase 4: Feature Pages (Week 7-8)

**Pages to Redesign**:
1. KDS pages (Cooking, Packing, Cleaning)
2. Menu Planning
3. Recipe Management
4. Inventory Management

### Phase 5: Forms & Tables (Week 9-10)

**Pages to Redesign**:
1. All form pages
2. All list/table pages
3. Login page
4. Profile page

## 🔧 Technical Implementation

### Dark Mode Implementation

```javascript
// composables/useDarkMode.js
import { ref, watch } from 'vue'

export function useDarkMode() {
  const isDark = ref(false)
  
  // Load from localStorage
  const stored = localStorage.getItem('theme')
  if (stored) {
    isDark.value = stored === 'dark'
  } else {
    // Detect system preference
    isDark.value = window.matchMedia('(prefers-color-scheme: dark)').matches
  }
  
  // Apply theme
  const applyTheme = (dark) => {
    if (dark) {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }
  
  // Watch for changes
  watch(isDark, (newValue) => {
    applyTheme(newValue)
    localStorage.setItem('theme', newValue ? 'dark' : 'light')
  }, { immediate: true })
  
  const toggle = () => {
    isDark.value = !isDark.value
  }
  
  return {
    isDark,
    toggle
  }
}
```

### Chart Integration

```javascript
// composables/useHorizonChart.js
import { ref, onMounted } from 'vue'
import * as echarts from 'echarts'

export function useHorizonChart(chartRef, options) {
  const chart = ref(null)
  
  const initChart = () => {
    if (!chartRef.value) return
    
    chart.value = echarts.init(chartRef.value)
    
    // Horizon UI theme
    const horizonTheme = {
      color: ['#f82c17', '#ff4d38', '#ff6b54', '#ff8970'],
      backgroundColor: 'transparent',
      textStyle: {
        fontFamily: 'DM Sans, sans-serif',
        fontSize: 14,
        color: '#2B3674'
      },
      // ... more config
    }
    
    chart.value.setOption({
      ...horizonTheme,
      ...options
    })
  }
  
  onMounted(() => {
    initChart()
  })
  
  return {
    chart,
    initChart
  }
}
```

## 📱 Responsive Design

### Breakpoint Strategy

```javascript
// composables/useBreakpoint.js
import { ref, onMounted, onUnmounted } from 'vue'

export function useBreakpoint() {
  const breakpoint = ref('xl')
  
  const updateBreakpoint = () => {
    const width = window.innerWidth
    if (width < 640) breakpoint.value = 'sm'
    else if (width < 768) breakpoint.value = 'md'
    else if (width < 1024) breakpoint.value = 'lg'
    else if (width < 1280) breakpoint.value = 'xl'
    else breakpoint.value = '2xl'
  }
  
  onMounted(() => {
    updateBreakpoint()
    window.addEventListener('resize', updateBreakpoint)
  })
  
  onUnmounted(() => {
    window.removeEventListener('resize', updateBreakpoint)
  })
  
  return {
    breakpoint,
    isMobile: computed(() => ['sm', 'md'].includes(breakpoint.value)),
    isTablet: computed(() => breakpoint.value === 'lg'),
    isDesktop: computed(() => ['xl', '2xl'].includes(breakpoint.value))
  }
}
```

## 🎭 Animation Guidelines

### Transitions

```css
/* Smooth transitions for theme switching */
* {
  transition: background-color 0.2s ease, color 0.2s ease, border-color 0.2s ease;
}

/* Card hover effect */
.h-card-hover {
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.h-card-hover:hover {
  transform: translateY(-4px);
}

/* Button hover */
.h-button {
  transition: all 0.15s ease;
}

.h-button:hover {
  transform: scale(1.02);
}

/* Page transitions */
.page-enter-active,
.page-leave-active {
  transition: opacity 0.2s ease, transform 0.3s ease;
}

.page-enter-from {
  opacity: 0;
  transform: translateX(20px);
}

.page-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}
```

## ✅ Quality Checklist

### Design Consistency
- [ ] All colors from design system
- [ ] Consistent spacing (8px grid)
- [ ] Consistent border radius
- [ ] Consistent shadows
- [ ] Consistent typography

### Functionality
- [ ] All existing features work
- [ ] No regressions
- [ ] Dark mode works everywhere
- [ ] Responsive on all devices
- [ ] Keyboard navigation works

### Performance
- [ ] Lighthouse score > 90
- [ ] No layout shifts
- [ ] Fast page transitions
- [ ] Optimized images
- [ ] Code splitting

### Accessibility
- [ ] WCAG 2.1 AA compliant
- [ ] Keyboard accessible
- [ ] Screen reader friendly
- [ ] Color contrast ratios met
- [ ] Focus indicators visible

---

**Status**: Design Defined
**Next Step**: Implementation (Tasks)
**Dependencies**: Requirements approved
