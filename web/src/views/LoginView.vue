<template>
  <div class="horizon-login">
    <!-- Left Side: Login Form -->
    <div class="horizon-login__form-side">
      <div class="horizon-login__form-container">
        <!-- Title -->
        <div class="horizon-login__header">
          <h1 class="horizon-login__title">Sign In</h1>
          <p class="horizon-login__subtitle">Masuk ke Sistem ERP SPPG</p>
        </div>

        <!-- Form -->
        <a-form
          :model="formState"
          :rules="rules"
          @finish="handleLogin"
          layout="vertical"
          class="horizon-login__form"
        >
          <!-- NIK / Email Input -->
          <a-form-item
            label="NIK / Email"
            name="identifier"
            :validate-status="error ? 'error' : ''"
          >
            <a-input
              v-model:value="formState.identifier"
              placeholder="Masukkan NIK atau Email"
              class="horizon-login__input"
              :disabled="loading"
            >
              <template #prefix>
                <UserOutlined class="horizon-login__input-icon" />
              </template>
            </a-input>
          </a-form-item>

          <!-- Password Input -->
          <a-form-item
            label="Password"
            name="password"
            :validate-status="error ? 'error' : ''"
            :help="error"
          >
            <a-input-password
              v-model:value="formState.password"
              placeholder="Masukkan password"
              class="horizon-login__input"
              :disabled="loading"
            >
              <template #prefix>
                <LockOutlined class="horizon-login__input-icon" />
              </template>
            </a-input-password>
          </a-form-item>

          <!-- Keep me logged in -->
          <div class="horizon-login__options">
            <a-checkbox v-model:checked="formState.rememberMe">
              Keep me logged in
            </a-checkbox>
          </div>

          <!-- Submit Button -->
          <a-button
            type="primary"
            html-type="submit"
            class="horizon-login__submit"
            :loading="loading"
          >
            Masuk
          </a-button>
        </a-form>

        <!-- Footer -->
        <div class="horizon-login__footer">
          <p class="horizon-login__footer-text">
            Sistem Manajemen Operasional SPPG
          </p>
          <p class="horizon-login__footer-subtext">
            Satuan Pelayanan Pemenuhan Gizi
          </p>
        </div>
      </div>
    </div>

    <!-- Right Side: Branding -->
    <div class="horizon-login__brand-side">
      <div class="horizon-login__brand-content">
        <div class="horizon-login__brand-logo">
          <div class="horizon-login__logo-circle">
            <span class="horizon-login__logo-text">ERP</span>
          </div>
        </div>
        <h2 class="horizon-login__brand-title">ERP SPPG</h2>
        <p class="horizon-login__brand-subtitle">
          Sistem Manajemen Operasional Terpadu
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { UserOutlined, LockOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'

const router = useRouter()
const authStore = useAuthStore()

const formState = reactive({
  identifier: '',
  password: '',
  rememberMe: false
})

const loading = ref(false)
const error = ref(null)

const rules = {
  identifier: [
    { required: true, message: 'NIK atau Email harus diisi', trigger: 'blur' }
  ],
  password: [
    { required: true, message: 'Password harus diisi', trigger: 'blur' },
    { min: 6, message: 'Password minimal 6 karakter', trigger: 'blur' }
  ]
}

const handleLogin = async () => {
  loading.value = true
  error.value = null

  try {
    await authStore.login({
      identifier: formState.identifier,
      password: formState.password
    })

    message.success('Login berhasil!')
    
    // Redirect based on user role
    const user = authStore.user
    if (user) {
      // Redirect to appropriate dashboard based on role
      switch (user.role) {
        case 'kepala_sppg':
        case 'kepala_yayasan':
          router.push('/dashboard')
          break
        case 'ahli_gizi':
          router.push('/menu-planning')
          break
        case 'pengadaan':
          router.push('/purchase-orders')
          break
        case 'akuntan':
          router.push('/financial')
          break
        case 'chef':
        case 'packing':
          router.push('/kds')
          break
        default:
          router.push('/dashboard')
      }
    } else {
      router.push('/dashboard')
    }
  } catch (err) {
    console.error('Login error:', err)
    error.value = err.response?.data?.message || 'Login gagal. Periksa NIK/Email dan password Anda.'
    message.error(error.value)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
/* ========================================
   MAIN CONTAINER - Split Screen Layout
   ======================================== */

.horizon-login {
  display: flex;
  min-height: 100vh;
  background-color: var(--h-bg-secondary, #FFFFFF);
}

/* ========================================
   LEFT SIDE - Login Form
   ======================================== */

.horizon-login__form-side {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--h-spacing-12, 48px);
  background-color: var(--h-bg-secondary, #FFFFFF);
}

.horizon-login__form-container {
  width: 100%;
  max-width: 480px;
}

/* Header */
.horizon-login__header {
  margin-bottom: var(--h-spacing-8, 32px);
}

.horizon-login__title {
  font-size: var(--h-text-4xl, 36px);
  font-weight: var(--h-font-bold, 700);
  color: var(--h-text-primary, #322837);
  margin: 0 0 var(--h-spacing-2, 8px) 0;
  line-height: var(--h-leading-tight, 1.25);
}

.horizon-login__subtitle {
  font-size: var(--h-text-sm, 14px);
  color: var(--h-text-secondary, #74788C);
  margin: 0;
  line-height: var(--h-leading-normal, 1.5);
}

/* Form */
.horizon-login__form {
  margin-bottom: var(--h-spacing-6, 24px);
}

.horizon-login__form :deep(.ant-form-item-label > label) {
  font-size: var(--h-text-sm, 14px);
  font-weight: var(--h-font-medium, 500);
  color: var(--h-text-primary, #322837);
}

.horizon-login__input {
  height: 48px;
  border-radius: var(--h-radius-md, 12px);
  font-size: var(--h-text-base, 16px);
  border: 1px solid var(--h-border-color, #E9EDF7);
  transition: all var(--h-transition-base, 200ms);
}

.horizon-login__input:hover,
.horizon-login__input:focus {
  border-color: var(--h-primary, #5A4372);
}

.horizon-login__input-icon {
  color: var(--h-text-secondary, #74788C);
  font-size: var(--h-icon-sm, 20px);
}

/* Options (Keep me logged in) */
.horizon-login__options {
  margin-bottom: var(--h-spacing-6, 24px);
}

.horizon-login__options :deep(.ant-checkbox-wrapper) {
  font-size: var(--h-text-sm, 14px);
  color: var(--h-text-secondary, #74788C);
}

/* Submit Button */
.horizon-login__submit {
  width: 100%;
  height: 48px;
  border-radius: var(--h-radius-md, 12px);
  font-size: var(--h-text-base, 16px);
  font-weight: var(--h-font-semibold, 600);
  background: linear-gradient(135deg, #5A4372 0%, #3D2B53 100%);
  border: none;
  transition: all var(--h-transition-base, 200ms);
}

.horizon-login__submit:hover {
  transform: translateY(-2px);
  box-shadow: 0px 10px 20px rgba(90, 67, 114, 0.3);
}

.horizon-login__submit:active {
  transform: translateY(0);
}

/* Footer */
.horizon-login__footer {
  text-align: center;
  padding-top: var(--h-spacing-6, 24px);
  border-top: 1px solid var(--h-border-color, #E9EDF7);
}

.horizon-login__footer-text {
  font-size: var(--h-text-sm, 14px);
  color: var(--h-text-secondary, #74788C);
  margin: 0 0 var(--h-spacing-1, 4px) 0;
}

.horizon-login__footer-subtext {
  font-size: var(--h-text-xs, 12px);
  color: var(--h-text-light, #ACA9B0);
  margin: 0;
}

/* ========================================
   RIGHT SIDE - Branding
   ======================================== */

.horizon-login__brand-side {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #5A4372 0%, #3D2B53 100%);
  padding: var(--h-spacing-12, 48px);
  position: relative;
  overflow: hidden;
}

/* Decorative circles */
.horizon-login__brand-side::before,
.horizon-login__brand-side::after {
  content: '';
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.05);
}

.horizon-login__brand-side::before {
  width: 400px;
  height: 400px;
  top: -100px;
  right: -100px;
}

.horizon-login__brand-side::after {
  width: 300px;
  height: 300px;
  bottom: -80px;
  left: -80px;
}

.horizon-login__brand-content {
  text-align: center;
  position: relative;
  z-index: 1;
}

/* Logo */
.horizon-login__brand-logo {
  margin-bottom: var(--h-spacing-8, 32px);
}

.horizon-login__logo-circle {
  width: 120px;
  height: 120px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(10px);
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto;
  border: 3px solid rgba(255, 255, 255, 0.3);
}

.horizon-login__logo-text {
  font-size: var(--h-text-4xl, 36px);
  font-weight: var(--h-font-bold, 700);
  color: #FFFFFF;
  letter-spacing: 2px;
}

/* Brand Title */
.horizon-login__brand-title {
  font-size: var(--h-text-4xl, 36px);
  font-weight: var(--h-font-bold, 700);
  color: #FFFFFF;
  margin: 0 0 var(--h-spacing-4, 16px) 0;
  line-height: var(--h-leading-tight, 1.25);
}

.horizon-login__brand-subtitle {
  font-size: var(--h-text-lg, 18px);
  color: rgba(255, 255, 255, 0.8);
  margin: 0;
  line-height: var(--h-leading-normal, 1.5);
}

/* ========================================
   RESPONSIVE - Mobile & Tablet
   ======================================== */

/* Tablet (< 1024px) */
@media (max-width: 1024px) {
  .horizon-login__brand-side {
    display: none;
  }
  
  .horizon-login__form-side {
    flex: 1;
  }
}

/* Mobile (< 768px) */
@media (max-width: 768px) {
  .horizon-login__form-side {
    padding: var(--h-spacing-6, 24px);
  }
  
  .horizon-login__title {
    font-size: var(--h-text-3xl, 30px);
  }
  
  .horizon-login__form-container {
    max-width: 100%;
  }
}

/* Small Mobile (< 480px) */
@media (max-width: 480px) {
  .horizon-login__form-side {
    padding: var(--h-spacing-4, 16px);
  }
  
  .horizon-login__title {
    font-size: var(--h-text-2xl, 24px);
  }
  
  .horizon-login__input,
  .horizon-login__submit {
    height: 44px;
  }
}

/* ========================================
   DARK MODE SUPPORT
   ======================================== */

.dark .horizon-login__form-side {
  background-color: #322837;
}

.dark .horizon-login__title {
  color: var(--h-text-primary-dark, #F8FDEA);
}

.dark .horizon-login__subtitle,
.dark .horizon-login__footer-text {
  color: var(--h-text-secondary-dark, #ACA9B0);
}

.dark .horizon-login__input {
  background-color: #3D2B53;
  border-color: #5A4372;
  color: #F8FDEA;
}

.dark .horizon-login__input::placeholder {
  color: #ACA9B0;
}

.dark .horizon-login__footer {
  border-top-color: #5A4372;
}
</style>
