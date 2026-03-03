# HBreadcrumb Component Usage Examples

## Overview

HBreadcrumb is a breadcrumb navigation component that automatically reads the current route and displays navigation path in "Pages / Page Name" format. It follows Horizon UI design specifications and is hidden on mobile devices.

## Features

- ✅ Auto-generates breadcrumbs from route meta
- ✅ Supports custom breadcrumb items
- ✅ Styled according to Horizon UI (14px, #74788C)
- ✅ Hidden on mobile (< 768px)
- ✅ Dark mode support
- ✅ Accessible (ARIA labels, semantic HTML)

## Basic Usage

### Auto-generated from Route

The simplest usage - component reads from current route:

```vue
<template>
  <HBreadcrumb />
</template>

<script setup>
import HBreadcrumb from '@/components/layout/HBreadcrumb.vue'
</script>
```

This will automatically generate breadcrumbs like:
- `Pages / Dashboard` (for route with meta.title: 'Dashboard')
- `Pages / Manajemen Resep` (for route with meta.title: 'Manajemen Resep')

### Custom Breadcrumb Items

For more control, provide custom items:

```vue
<template>
  <HBreadcrumb :items="breadcrumbItems" />
</template>

<script setup>
import HBreadcrumb from '@/components/layout/HBreadcrumb.vue'

const breadcrumbItems = [
  { label: 'Home', to: '/' },
  { label: 'Settings', to: '/settings' },
  { label: 'Profile' } // Last item without 'to' (current page)
]
</script>
```

### Custom Root Label

Change the default "Pages" label:

```vue
<template>
  <HBreadcrumb root-label="Home" />
</template>
```

Result: `Home / Dashboard`

### Hide Root Breadcrumb

Show only the current page:

```vue
<template>
  <HBreadcrumb :show-root="false" />
</template>
```

Result: `Dashboard` (no "Pages" prefix)

## Integration with HHeader

The HBreadcrumb component can be used standalone or integrated into HHeader:

```vue
<template>
  <header class="h-header">
    <div class="header-left">
      <!-- Use HBreadcrumb component -->
      <HBreadcrumb />
      
      <!-- Page Title -->
      <h1 class="page-title">{{ pageTitle }}</h1>
    </div>
    
    <!-- ... rest of header ... -->
  </header>
</template>

<script setup>
import HBreadcrumb from './HBreadcrumb.vue'
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const pageTitle = computed(() => route.meta?.title || 'Dashboard')
</script>
```

## Advanced Usage

### Dynamic Breadcrumbs Based on Data

```vue
<template>
  <HBreadcrumb :items="dynamicBreadcrumbs" />
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import HBreadcrumb from '@/components/layout/HBreadcrumb.vue'

const route = useRoute()

const dynamicBreadcrumbs = computed(() => {
  const crumbs = [
    { label: 'Pages', to: null }
  ]
  
  // Add parent page
  if (route.meta?.parent) {
    crumbs.push({
      label: route.meta.parent.title,
      to: route.meta.parent.path
    })
  }
  
  // Add current page
  crumbs.push({
    label: route.meta?.title || 'Current Page',
    to: null
  })
  
  return crumbs
})
</script>
```

### Breadcrumbs with Route Parameters

```vue
<template>
  <HBreadcrumb :items="breadcrumbsWithParams" />
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import HBreadcrumb from '@/components/layout/HBreadcrumb.vue'

const route = useRoute()

const breadcrumbsWithParams = computed(() => {
  const schoolId = route.params.id
  
  return [
    { label: 'Pages', to: null },
    { label: 'Sekolah', to: '/schools' },
    { label: `Edit Sekolah #${schoolId}`, to: null }
  ]
})
</script>
```

## Styling

The component uses CSS variables from the Horizon UI design system:

```css
/* Light mode */
--h-text-secondary: #74788C;  /* Breadcrumb text */
--h-text-primary: #322837;    /* Current page text */
--h-primary: #5A4372;         /* Link hover color */

/* Dark mode */
--h-text-secondary-dark: #ACA9B0;
--h-text-primary-dark: #F8FDEA;
--h-primary-light: #6a5382;
```

### Custom Styling

You can override styles if needed:

```vue
<template>
  <HBreadcrumb class="custom-breadcrumb" />
</template>

<style scoped>
.custom-breadcrumb {
  /* Your custom styles */
  font-size: 16px;
}
</style>
```

## Responsive Behavior

- **Desktop (≥ 768px)**: Breadcrumb is visible
- **Mobile (< 768px)**: Breadcrumb is hidden (both via v-if and CSS media query)

The component uses both JavaScript (`isMobile` from `useBreakpoint`) and CSS media queries for maximum compatibility.

## Accessibility

The component follows accessibility best practices:

- Uses semantic `<nav>` element with `aria-label="Breadcrumb"`
- Uses ordered list `<ol>` for breadcrumb items
- Separators have `aria-hidden="true"` to avoid screen reader clutter
- Current page is marked with appropriate styling

## Props Reference

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `items` | Array | `null` | Custom breadcrumb items. Format: `[{ label: String, to: String }]` |
| `rootLabel` | String | `'Pages'` | Label for the root breadcrumb item |
| `showRoot` | Boolean | `true` | Whether to show the root breadcrumb item |

## Route Meta Configuration

To get the best auto-generated breadcrumbs, configure your routes with `meta.title`:

```javascript
// router/index.js
{
  path: 'recipes',
  name: 'recipes',
  component: () => import('@/views/RecipeListView.vue'),
  meta: { 
    requiresAuth: true,
    roles: ['kepala_sppg', 'ahli_gizi'],
    title: 'Manajemen Resep' // Used by HBreadcrumb
  }
}
```

## Testing

The component includes comprehensive unit tests. Run them with:

```bash
npm test -- HBreadcrumb.test.js
```

Tests cover:
- Default breadcrumb rendering
- Custom breadcrumb items
- Separator rendering
- Current page styling
- Mobile hiding behavior
- Root label customization
- Show/hide root breadcrumb
