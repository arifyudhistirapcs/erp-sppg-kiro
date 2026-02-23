import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { message } from 'ant-design-vue'
import MainLayout from '@/layouts/MainLayout.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/',
      component: MainLayout,
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          redirect: '/dashboard'
        },
        {
          path: 'dashboard',
          name: 'dashboard',
          component: () => import('@/views/DashboardView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg', 'kepala_yayasan', 'akuntan', 'ahli_gizi', 'pengadaan']
          }
        },
        {
          path: 'dashboard/kepala-sppg',
          name: 'dashboard-kepala-sppg',
          component: () => import('@/views/DashboardKepalaSSPGView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg'],
            title: 'Dashboard Kepala SPPG'
          }
        },
        {
          path: 'dashboard/kepala-yayasan',
          name: 'dashboard-kepala-yayasan',
          component: () => import('@/views/DashboardKepalaYayasanView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_yayasan'],
            title: 'Dashboard Kepala Yayasan'
          }
        },
        {
          path: 'recipes',
          name: 'recipes',
          component: () => import('@/views/RecipeListView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg', 'ahli_gizi'],
            title: 'Manajemen Resep'
          }
        },
        {
          path: 'menu-planning',
          name: 'menu-planning',
          component: () => import('@/views/MenuPlanningView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg', 'ahli_gizi'],
            title: 'Perencanaan Menu'
          }
        },
        {
          path: 'kds/cooking',
          name: 'kds-cooking',
          component: () => import('@/views/KDSCookingView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg', 'ahli_gizi', 'chef'],
            title: 'KDS - Dapur Memasak'
          }
        },
        {
          path: 'kds/packing',
          name: 'kds-packing',
          component: () => import('@/views/KDSPackingView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg', 'ahli_gizi', 'chef', 'packing'],
            title: 'KDS - Packing'
          }
        },
        {
          path: 'suppliers',
          name: 'suppliers',
          component: () => import('@/views/SupplierListView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg', 'pengadaan'],
            title: 'Manajemen Supplier'
          }
        },
        {
          path: 'purchase-orders',
          name: 'purchase-orders',
          component: () => import('@/views/PurchaseOrderListView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg', 'pengadaan'],
            title: 'Purchase Order'
          }
        },
        {
          path: 'goods-receipts',
          name: 'goods-receipts',
          component: () => import('@/views/GoodsReceiptView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg', 'pengadaan'],
            title: 'Penerimaan Barang'
          }
        },
        {
          path: 'inventory',
          name: 'inventory',
          component: () => import('@/views/InventoryView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg', 'pengadaan', 'akuntan'],
            title: 'Manajemen Inventory'
          }
        },
        {
          path: 'schools',
          name: 'schools',
          component: () => import('@/views/SchoolListView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg', 'driver', 'asisten'],
            title: 'Manajemen Sekolah'
          }
        },
        {
          path: 'schools/create',
          name: 'school-create',
          component: () => import('@/views/SchoolFormView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg'],
            title: 'Tambah Sekolah'
          }
        },
        {
          path: 'schools/:id/edit',
          name: 'school-edit',
          component: () => import('@/views/SchoolFormView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg'],
            title: 'Edit Sekolah'
          }
        },
        {
          path: 'delivery-tasks',
          name: 'delivery-tasks',
          component: () => import('@/views/DeliveryTaskListView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg', 'driver', 'asisten'],
            title: 'Manajemen Tugas Pengiriman'
          }
        },
        {
          path: 'delivery-tasks/create',
          name: 'delivery-task-create',
          component: () => import('@/views/DeliveryTaskFormView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg'],
            title: 'Buat Tugas Pengiriman'
          }
        },
        {
          path: 'delivery-tasks/:id/edit',
          name: 'delivery-task-edit',
          component: () => import('@/views/DeliveryTaskFormView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg'],
            title: 'Edit Tugas Pengiriman'
          }
        },
        {
          path: 'ompreng-tracking',
          name: 'ompreng-tracking',
          component: () => import('@/views/OmprengTrackingView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg', 'driver', 'asisten'],
            title: 'Pelacakan Ompreng'
          }
        },
        {
          path: 'employees',
          name: 'employees',
          component: () => import('@/views/EmployeeListView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg', 'akuntan'],
            title: 'Manajemen Karyawan'
          }
        },
        {
          path: 'employees/create',
          name: 'employee-create',
          component: () => import('@/views/EmployeeFormView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg', 'akuntan'],
            title: 'Tambah Karyawan'
          }
        },
        {
          path: 'employees/:id/edit',
          name: 'employee-edit',
          component: () => import('@/views/EmployeeFormView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg', 'akuntan'],
            title: 'Edit Karyawan'
          }
        },
        {
          path: 'attendance-report',
          name: 'attendance-report',
          component: () => import('@/views/AttendanceReportView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg', 'akuntan'],
            title: 'Laporan Absensi'
          }
        },
        {
          path: 'wifi-config',
          name: 'wifi-config',
          component: () => import('@/views/WiFiConfigView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg', 'akuntan'],
            title: 'Konfigurasi Wi-Fi'
          }
        },
        {
          path: 'assets',
          name: 'assets',
          component: () => import('@/views/AssetListView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg', 'akuntan'],
            title: 'Manajemen Aset Dapur'
          }
        },
        {
          path: 'cash-flow',
          name: 'cash-flow',
          component: () => import('@/views/CashFlowListView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg', 'akuntan'],
            title: 'Manajemen Arus Kas'
          }
        },
        {
          path: 'financial-reports',
          name: 'financial-reports',
          component: () => import('@/views/FinancialReportView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg', 'akuntan'],
            title: 'Laporan Keuangan'
          }
        },
        {
          path: 'audit-trail',
          name: 'audit-trail',
          component: () => import('@/views/AuditTrailView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg'],
            title: 'Audit Trail'
          }
        },
        {
          path: 'system-config',
          name: 'system-config',
          component: () => import('@/views/SystemConfigView.vue'),
          meta: { 
            requiresAuth: true,
            roles: ['kepala_sppg'],
            title: 'Konfigurasi Sistem'
          }
        }
        // Additional routes will be added in subsequent tasks
      ]
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      redirect: '/dashboard'
    }
  ]
})

// Check if user has required role
const hasRequiredRole = (userRole, requiredRoles) => {
  if (!requiredRoles || requiredRoles.length === 0) {
    return true
  }
  return requiredRoles.includes(userRole)
}

// Navigation guard
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()
  
  // Check if route requires authentication
  if (to.meta.requiresAuth) {
    if (!authStore.isAuthenticated) {
      // Not authenticated, redirect to login
      next({
        path: '/login',
        query: { redirect: to.fullPath }
      })
      return
    }

    // Check if user data is loaded
    if (!authStore.user) {
      try {
        // Try to fetch current user data
        await authStore.getCurrentUser()
      } catch (error) {
        console.error('Failed to fetch user data:', error)
        authStore.clearAuth()
        next('/login')
        return
      }
    }

    // Check role-based access
    if (to.meta.roles) {
      const userRole = authStore.user?.role
      if (!hasRequiredRole(userRole, to.meta.roles)) {
        message.error('Anda tidak memiliki akses ke halaman ini')
        next('/dashboard')
        return
      }
    }

    next()
  } else {
    // Route doesn't require auth
    if (to.path === '/login' && authStore.isAuthenticated) {
      // Already logged in, redirect to dashboard
      next('/dashboard')
    } else {
      next()
    }
  }
})

// After navigation
router.afterEach((to, from) => {
  // Update page title
  const baseTitle = 'ERP SPPG'
  const pageTitle = to.meta.title || to.name
  document.title = pageTitle ? `${pageTitle} - ${baseTitle}` : baseTitle
})

export default router
