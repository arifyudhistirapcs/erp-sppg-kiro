import { ref, watch } from 'vue'

/**
 * Global dark mode state (singleton)
 * Shared across all components that use useDarkMode()
 */
const isDark = ref(false)
let initialized = false

function applyTheme(dark) {
  if (typeof document === 'undefined') return
  if (dark) {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
}

function initTheme() {
  if (initialized) return
  initialized = true

  // Load from localStorage or system preference
  const stored = localStorage.getItem('theme')
  if (stored) {
    isDark.value = stored === 'dark'
  } else {
    isDark.value = window.matchMedia('(prefers-color-scheme: dark)').matches
  }
  applyTheme(isDark.value)

  // Watch changes
  watch(isDark, (val) => {
    applyTheme(val)
    localStorage.setItem('theme', val ? 'dark' : 'light')
  })

  // Listen to system preference changes
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
    if (!localStorage.getItem('theme')) {
      isDark.value = e.matches
    }
  })
}

export function useDarkMode() {
  initTheme()

  const toggle = () => {
    isDark.value = !isDark.value
  }

  const setTheme = (dark) => {
    isDark.value = dark
  }

  return { isDark, toggle, setTheme }
}
