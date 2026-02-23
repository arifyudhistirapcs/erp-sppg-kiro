import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { hasPermission, hasAnyPermission, hasAllPermissions, getRoleLabel } from '@/utils/permissions'

/**
 * Composable for checking user permissions
 * @returns {Object}
 */
export const usePermissions = () => {
  const authStore = useAuthStore()

  const userRole = computed(() => authStore.user?.role || '')
  const roleLabel = computed(() => getRoleLabel(userRole.value))

  /**
   * Check if current user has a specific permission
   * @param {string} permission - Permission key
   * @returns {boolean}
   */
  const can = (permission) => {
    return hasPermission(userRole.value, permission)
  }

  /**
   * Check if current user has any of the specified permissions
   * @param {string[]} permissions - Array of permission keys
   * @returns {boolean}
   */
  const canAny = (permissions) => {
    return hasAnyPermission(userRole.value, permissions)
  }

  /**
   * Check if current user has all of the specified permissions
   * @param {string[]} permissions - Array of permission keys
   * @returns {boolean}
   */
  const canAll = (permissions) => {
    return hasAllPermissions(userRole.value, permissions)
  }

  /**
   * Check if current user has a specific role
   * @param {string} role - Role key
   * @returns {boolean}
   */
  const isRole = (role) => {
    return userRole.value === role
  }

  /**
   * Check if current user has any of the specified roles
   * @param {string[]} roles - Array of role keys
   * @returns {boolean}
   */
  const isAnyRole = (roles) => {
    return roles.includes(userRole.value)
  }

  return {
    userRole,
    roleLabel,
    can,
    canAny,
    canAll,
    isRole,
    isAnyRole
  }
}
