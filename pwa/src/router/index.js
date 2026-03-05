import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

// Allowed roles for tasks page
const TASK_ALLOWED_ROLES = ['driver', 'asisten_lapangan']

/**
 * Returns the default route path for a given user role.
 * Used for post-login redirect and unauthorized access fallback.
 */
function getDefaultRoute(role) {
  const roleStr = role?.toLowerCase()
  if (roleStr === 'kepala_sppg') return '/dashboard'
  if (roleStr === 'ahli_gizi') return '/menu-planning'
  if (roleStr === 'sekolah') return '/school-monitoring'
  if (TASK_ALLOWED_ROLES.includes(roleStr)) return '/tasks'
  return '/attendance'
}

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
      component: () => import('@/layouts/MobileLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          redirect: () => {
            const authStore = useAuthStore()
            return getDefaultRoute(authStore.user?.role)
          }
        },
        {
          path: 'dashboard',
          name: 'dashboard',
          component: () => import('@/views/DashboardView.vue'),
          meta: { roles: ['kepala_sppg'] }
        },
        {
          path: 'monitoring',
          name: 'monitoring',
          component: () => import('@/views/MonitoringView.vue'),
          meta: { roles: ['kepala_sppg'] }
        },
        {
          path: 'monitoring/:id',
          name: 'monitoring-detail',
          component: () => import('@/views/MonitoringDetailView.vue'),
          meta: { roles: ['kepala_sppg'] }
        },
        {
          path: 'menu-planning',
          name: 'menu-planning',
          component: () => import('@/views/MenuPlanningView.vue'),
          meta: { roles: ['kepala_sppg', 'ahli_gizi'] }
        },
        {
          path: 'school-monitoring',
          name: 'school-monitoring',
          component: () => import('@/views/SchoolMonitoringView.vue'),
          meta: { roles: ['sekolah'] }
        },
        {
          path: 'tasks',
          name: 'tasks',
          component: () => import('@/views/TasksView.vue'),
          meta: { roles: TASK_ALLOWED_ROLES }
        },
        {
          path: 'tasks/:id',
          name: 'task-detail',
          component: () => import('@/views/DeliveryTaskDetailView.vue'),
          meta: { roles: TASK_ALLOWED_ROLES }
        },
        {
          path: 'pickup-tasks/:id',
          name: 'pickup-task-detail',
          component: () => import('@/views/PickupTaskDetailView.vue'),
          meta: { roles: TASK_ALLOWED_ROLES }
        },
        {
          path: 'tasks/:taskId/epod',
          name: 'epod-form',
          component: () => import('@/views/ePODFormView.vue'),
          meta: { roles: TASK_ALLOWED_ROLES }
        },
        {
          path: 'review',
          name: 'review-form',
          component: () => import('@/views/ReviewFormView.vue'),
          meta: { roles: TASK_ALLOWED_ROLES }
        },
        {
          path: 'attendance',
          name: 'attendance',
          component: () => import('@/views/AttendanceView.vue')
        },
        {
          path: 'profile',
          name: 'profile',
          component: () => import('@/views/ProfileView.vue')
        }
      ]
    }
  ]
})

// Navigation guard
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()

  // Check if any matched route requires auth
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth)

  if (requiresAuth && !authStore.isAuthenticated) {
    next('/login')
    return
  }

  if (to.path === '/login' && authStore.isAuthenticated) {
    next(getDefaultRoute(authStore.user?.role))
    return
  }

  // Check role-based access: look through matched routes for meta.roles
  const requiredRoles = to.matched.reduce((roles, record) => {
    return record.meta.roles || roles
  }, null)

  if (requiredRoles) {
    const userRole = authStore.user?.role?.toLowerCase()
    if (!requiredRoles.includes(userRole)) {
      next(getDefaultRoute(authStore.user?.role))
      return
    }
  }

  next()
})

export { getDefaultRoute }
export default router
