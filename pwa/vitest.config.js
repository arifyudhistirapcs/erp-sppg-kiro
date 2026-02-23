import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/tests/setup.js'],
    // Optimize test performance
    testTimeout: 5000, // Reduce timeout from default 10s to 5s
    hookTimeout: 5000, // Reduce hook timeout
    teardownTimeout: 1000, // Reduce teardown timeout
    // Run tests in parallel for better performance
    pool: 'threads',
    poolOptions: {
      threads: {
        singleThread: false,
        isolate: false // Faster but less isolated
      }
    },
    // Reduce reporter verbosity for faster output
    reporter: ['basic'],
    // Optimize coverage collection
    coverage: {
      enabled: false // Disable coverage for faster runs
    }
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, './src')
    }
  }
})