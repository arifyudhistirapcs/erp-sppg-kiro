/**
 * Role-based access control utilities
 */

// Define role permissions for each feature/module
export const PERMISSIONS = {
  // Dashboard
  DASHBOARD_VIEW: ['kepala_sppg', 'kepala_yayasan', 'akuntan', 'ahli_gizi', 'pengadaan'],
  
  // Recipe & Menu Planning
  RECIPE_VIEW: ['kepala_sppg', 'ahli_gizi'],
  RECIPE_CREATE: ['kepala_sppg', 'ahli_gizi'],
  RECIPE_EDIT: ['kepala_sppg', 'ahli_gizi'],
  RECIPE_DELETE: ['kepala_sppg', 'ahli_gizi'],
  MENU_PLANNING_VIEW: ['kepala_sppg', 'ahli_gizi'],
  MENU_PLANNING_CREATE: ['kepala_sppg', 'ahli_gizi'],
  MENU_PLANNING_APPROVE: ['kepala_sppg', 'ahli_gizi'],
  
  // Kitchen Display System
  KDS_VIEW: ['kepala_sppg', 'ahli_gizi', 'chef', 'packing'],
  KDS_UPDATE_STATUS: ['chef', 'packing'],
  
  // Supply Chain
  SUPPLIER_VIEW: ['kepala_sppg', 'pengadaan'],
  SUPPLIER_MANAGE: ['kepala_sppg', 'pengadaan'],
  PO_VIEW: ['kepala_sppg', 'pengadaan'],
  PO_CREATE: ['kepala_sppg', 'pengadaan'],
  PO_APPROVE: ['kepala_sppg'],
  GRN_VIEW: ['kepala_sppg', 'pengadaan'],
  GRN_CREATE: ['kepala_sppg', 'pengadaan'],
  INVENTORY_VIEW: ['kepala_sppg', 'pengadaan', 'akuntan'],
  INVENTORY_MANAGE: ['kepala_sppg', 'pengadaan'],
  
  // Logistics
  SCHOOL_VIEW: ['kepala_sppg', 'pengadaan'],
  SCHOOL_MANAGE: ['kepala_sppg', 'pengadaan'],
  DELIVERY_VIEW: ['kepala_sppg', 'pengadaan'],
  DELIVERY_MANAGE: ['kepala_sppg', 'pengadaan'],
  OMPRENG_VIEW: ['kepala_sppg', 'pengadaan'],
  
  // HRM
  EMPLOYEE_VIEW: ['kepala_sppg', 'akuntan'],
  EMPLOYEE_MANAGE: ['kepala_sppg', 'akuntan'],
  ATTENDANCE_VIEW: ['kepala_sppg', 'akuntan'],
  ATTENDANCE_REPORT: ['kepala_sppg', 'akuntan'],
  
  // Financial
  ASSET_VIEW: ['kepala_sppg', 'akuntan'],
  ASSET_MANAGE: ['kepala_sppg', 'akuntan'],
  CASH_FLOW_VIEW: ['kepala_sppg', 'akuntan'],
  CASH_FLOW_MANAGE: ['kepala_sppg', 'akuntan'],
  FINANCIAL_REPORT_VIEW: ['kepala_sppg', 'kepala_yayasan', 'akuntan'],
  FINANCIAL_REPORT_EXPORT: ['kepala_sppg', 'kepala_yayasan', 'akuntan'],
  
  // System
  AUDIT_TRAIL_VIEW: ['kepala_sppg'],
  SYSTEM_CONFIG_VIEW: ['kepala_sppg'],
  SYSTEM_CONFIG_EDIT: ['kepala_sppg']
}

/**
 * Check if user has permission
 * @param {string} userRole - Current user's role
 * @param {string} permission - Permission key from PERMISSIONS
 * @returns {boolean}
 */
export const hasPermission = (userRole, permission) => {
  if (!userRole || !permission) return false
  const allowedRoles = PERMISSIONS[permission]
  if (!allowedRoles) return false
  return allowedRoles.includes(userRole)
}

/**
 * Check if user has any of the specified permissions
 * @param {string} userRole - Current user's role
 * @param {string[]} permissions - Array of permission keys
 * @returns {boolean}
 */
export const hasAnyPermission = (userRole, permissions) => {
  if (!userRole || !permissions || permissions.length === 0) return false
  return permissions.some(permission => hasPermission(userRole, permission))
}

/**
 * Check if user has all of the specified permissions
 * @param {string} userRole - Current user's role
 * @param {string[]} permissions - Array of permission keys
 * @returns {boolean}
 */
export const hasAllPermissions = (userRole, permissions) => {
  if (!userRole || !permissions || permissions.length === 0) return false
  return permissions.every(permission => hasPermission(userRole, permission))
}

/**
 * Get user role label in Indonesian
 * @param {string} role - Role key
 * @returns {string}
 */
export const getRoleLabel = (role) => {
  const roleLabels = {
    'kepala_sppg': 'Kepala SPPG',
    'kepala_yayasan': 'Kepala Yayasan',
    'akuntan': 'Akuntan',
    'ahli_gizi': 'Ahli Gizi',
    'pengadaan': 'Staff Pengadaan',
    'chef': 'Chef',
    'packing': 'Staff Packing',
    'driver': 'Driver',
    'asisten': 'Asisten Lapangan'
  }
  return roleLabels[role] || 'User'
}
