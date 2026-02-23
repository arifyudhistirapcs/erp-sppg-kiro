<template>
  <div class="login-container">
    <a-card class="login-card" title="Sistem ERP SPPG">
      <a-form
        :model="formState"
        :rules="rules"
        @finish="handleLogin"
        layout="vertical"
      >
        <a-form-item
          label="NIK / Email"
          name="identifier"
          :validate-status="error ? 'error' : ''"
        >
          <a-input
            v-model:value="formState.identifier"
            placeholder="Masukkan NIK atau Email"
            size="large"
            :disabled="loading"
          >
            <template #prefix>
              <UserOutlined />
            </template>
          </a-input>
        </a-form-item>

        <a-form-item
          label="Password"
          name="password"
          :validate-status="error ? 'error' : ''"
          :help="error"
        >
          <a-input-password
            v-model:value="formState.password"
            placeholder="Masukkan password"
            size="large"
            :disabled="loading"
          >
            <template #prefix>
              <LockOutlined />
            </template>
          </a-input-password>
        </a-form-item>

        <a-form-item>
          <a-button
            type="primary"
            html-type="submit"
            size="large"
            block
            :loading="loading"
          >
            Masuk
          </a-button>
        </a-form-item>
      </a-form>

      <div class="login-footer">
        <p>Sistem Manajemen Operasional SPPG</p>
        <p style="font-size: 12px; color: #999;">
          Satuan Pelayanan Pemenuhan Gizi
        </p>
      </div>
    </a-card>
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
  password: ''
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
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.login-card {
  width: 100%;
  max-width: 420px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
  border-radius: 8px;
}

.login-card :deep(.ant-card-head) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 8px 8px 0 0;
}

.login-card :deep(.ant-card-head-title) {
  color: white;
  font-size: 20px;
  font-weight: 600;
  text-align: center;
}

.login-footer {
  text-align: center;
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #f0f0f0;
}

.login-footer p {
  margin: 4px 0;
  color: #666;
}
</style>
