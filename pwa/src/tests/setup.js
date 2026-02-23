// Test setup file
import { vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

// Setup Pinia for tests
beforeEach(() => {
  const pinia = createPinia()
  setActivePinia(pinia)
})

// Mock Pinia store
vi.mock('@/stores/auth.js', () => ({
  useAuthStore: () => ({
    user: { id: 1, name: 'Test User' },
    token: 'test-token',
    clearAuth: vi.fn()
  })
}))

// Mock environment variables
vi.mock('import.meta', () => ({
  env: {
    VITE_API_BASE_URL: 'http://localhost:8080/api/v1'
  }
}))

// Global test utilities
global.console = {
  ...console,
  // Suppress console.warn and console.error in tests unless needed
  warn: vi.fn(),
  error: vi.fn()
}