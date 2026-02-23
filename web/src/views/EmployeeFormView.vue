<template>
  <div class="employee-form">
    <a-page-header
      :title="isEdit ? 'Edit Karyawan' : 'Tambah Karyawan'"
      :sub-title="isEdit ? 'Perbarui informasi karyawan' : 'Tambah karyawan baru ke sistem'"
      @back="handleBack"
    />

    <a-card>
      <a-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        layout="vertical"
        @finish="handleSubmit"
      >
        <a-row :gutter="24">
          <a-col :span="12">
            <a-card title="Informasi Pribadi" size="small">
              <a-form-item label="NIK" name="nik">
                <a-input 
                  v-model:value="formData.nik" 
                  placeholder="Nomor Induk Karyawan"
                  :disabled="isEdit"
                />
              </a-form-item>

              <a-form-item label="Nama Lengkap" name="full_name">
                <a-input v-model:value="formData.full_name" placeholder="Nama lengkap karyawan" />
              </a-form-item>

              <a-form-item label="Email" name="email">
                <a-input 
                  v-model:value="formData.email" 
                  type="email" 
                  placeholder="email@example.com"
                  :disabled="isEdit"
                />
              </a-form-item>

              <a-form-item label="Nomor Telepon" name="phone_number">
                <a-input v-model:value="formData.phone_number" placeholder="08xxxxxxxxxx" />
              </a-form-item>
            </a-card>
          </a-col>

          <a-col :span="12">
            <a-card title="Informasi Pekerjaan" size="small">
              <a-form-item label="Posisi" name="position">
                <a-select v-model:value="formData.position" placeholder="Pilih posisi">
                  <a-select-option value="Kepala SPPG">Kepala SPPG</a-select-option>
                  <a-select-option value="Akuntan">Akuntan</a-select-option>
                  <a-select-option value="Ahli Gizi">Ahli Gizi</a-select-option>
                  <a-select-option value="Pengadaan">Pengadaan</a-select-option>
                  <a-select-option value="Chef">Chef</a-select-option>
                  <a-select-option value="Packing">Packing</a-select-option>
                  <a-select-option value="Driver">Driver</a-select-option>
                  <a-select-option value="Asisten Lapangan">Asisten Lapangan</a-select-option>
                </a-select>
              </a-form-item>

              <a-form-item label="Role Sistem" name="role">
                <a-select 
                  v-model:value="formData.role" 
                  placeholder="Pilih role sistem"
                  :disabled="isEdit"
                >
                  <a-select-option value="kepala_sppg">Kepala SPPG/Yayasan</a-select-option>
                  <a-select-option value="akuntan">Akuntan</a-select-option>
                  <a-select-option value="ahli_gizi">Ahli Gizi</a-select-option>
                  <a-select-option value="pengadaan">Pengadaan</a-select-option>
                  <a-select-option value="chef">Chef</a-select-option>
                  <a-select-option value="packing">Packing</a-select-option>
                  <a-select-option value="driver">Driver</a-select-option>
                  <a-select-option value="asisten">Asisten Lapangan</a-select-option>
                </a-select>
              </a-form-item>

              <a-form-item label="Tanggal Bergabung" name="join_date">
                <a-date-picker 
                  v-model:value="formData.join_date" 
                  style="width: 100%" 
                  placeholder="Pilih tanggal bergabung"
                  format="DD/MM/YYYY"
                />
              </a-form-item>

              <a-form-item label="Status" name="is_active">
                <a-switch 
                  v-model:checked="formData.is_active" 
                  checked-children="Aktif" 
                  un-checked-children="Tidak Aktif" 
                />
              </a-form-item>
            </a-card>
          </a-col>
        </a-row>

        <a-divider />

        <a-form-item>
          <a-space>
            <a-button type="primary" html-type="submit" :loading="submitting">
              {{ isEdit ? 'Perbarui' : 'Simpan' }}
            </a-button>
            <a-button @click="handleBack">
              Batal
            </a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </a-card>

    <!-- Credentials Modal -->
    <a-modal
      v-model:open="credentialsModalVisible"
      title="Kredensial Login Karyawan"
      :footer="null"
      width="500px"
      :closable="false"
      :mask-closable="false"
    >
      <a-alert
        message="Kredensial Berhasil Dibuat"
        description="Simpan informasi login berikut dan berikan kepada karyawan:"
        type="success"
        show-icon
        style="margin-bottom: 16px"
      />
      
      <a-descriptions bordered :column="1">
        <a-descriptions-item label="NIK/Email Login">
          <a-typography-text copyable>{{ newCredentials.nik }}</a-typography-text>
        </a-descriptions-item>
        <a-descriptions-item label="Email Login">
          <a-typography-text copyable>{{ newCredentials.email }}</a-typography-text>
        </a-descriptions-item>
        <a-descriptions-item label="Password Sementara">
          <a-typography-text copyable code>{{ newCredentials.password }}</a-typography-text>
        </a-descriptions-item>
      </a-descriptions>

      <a-alert
        message="Penting!"
        description="Password ini hanya ditampilkan sekali. Pastikan karyawan mengganti password setelah login pertama."
        type="warning"
        show-icon
        style="margin-top: 16px"
      />

      <div style="text-align: center; margin-top: 16px">
        <a-button type="primary" @click="handleCredentialsClose">
          Saya Sudah Menyimpan Kredensial
        </a-button>
      </div>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import dayjs from 'dayjs'
import employeeService from '@/services/employeeService'

const route = useRoute()
const router = useRouter()

const submitting = ref(false)
const credentialsModalVisible = ref(false)
const formRef = ref()
const newCredentials = ref({})

const isEdit = computed(() => !!route.params.id)

const formData = reactive({
  nik: '',
  full_name: '',
  email: '',
  phone_number: '',
  position: '',
  role: '',
  join_date: null,
  is_active: true
})

const rules = {
  nik: [
    { required: true, message: 'NIK wajib diisi' },
    { min: 16, max: 16, message: 'NIK harus 16 digit' },
    { pattern: /^\d+$/, message: 'NIK hanya boleh berisi angka' }
  ],
  full_name: [
    { required: true, message: 'Nama lengkap wajib diisi' },
    { min: 2, message: 'Nama lengkap minimal 2 karakter' }
  ],
  email: [
    { required: true, message: 'Email wajib diisi' },
    { type: 'email', message: 'Format email tidak valid' }
  ],
  phone_number: [
    { required: true, message: 'Nomor telepon wajib diisi' },
    { pattern: /^08\d{8,11}$/, message: 'Format nomor telepon tidak valid (08xxxxxxxxxx)' }
  ],
  position: [{ required: true, message: 'Posisi wajib dipilih' }],
  role: [{ required: true, message: 'Role sistem wajib dipilih' }],
  join_date: [{ required: true, message: 'Tanggal bergabung wajib diisi' }]
}

const fetchEmployee = async () => {
  if (!isEdit.value) return

  try {
    const response = await employeeService.getEmployeeById(route.params.id)
    const employee = response.data
    
    Object.assign(formData, {
      nik: employee.nik,
      full_name: employee.full_name,
      email: employee.email,
      phone_number: employee.phone_number,
      position: employee.position,
      role: employee.user?.role || '',
      join_date: employee.join_date ? dayjs(employee.join_date) : null,
      is_active: employee.is_active
    })
  } catch (error) {
    message.error('Gagal memuat data karyawan')
    console.error(error)
    router.push('/employees')
  }
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    submitting.value = true

    const submitData = {
      ...formData,
      join_date: formData.join_date ? formData.join_date.format('YYYY-MM-DD') : null
    }

    if (isEdit.value) {
      await employeeService.updateEmployee(route.params.id, submitData)
      message.success('Karyawan berhasil diperbarui')
      router.push('/employees')
    } else {
      const response = await employeeService.createEmployee(submitData)
      message.success('Karyawan berhasil ditambahkan')
      
      // Show credentials modal for new employee
      if (response.data && response.data.credentials) {
        newCredentials.value = {
          nik: response.data.user.nik,
          email: response.data.user.email,
          password: response.data.credentials.password
        }
        credentialsModalVisible.value = true
      } else {
        router.push('/employees')
      }
    }
  } catch (error) {
    if (error.errorFields) {
      return
    }
    
    const errorMessage = error.response?.data?.message || 'Gagal menyimpan data karyawan'
    message.error(errorMessage)
    console.error(error)
  } finally {
    submitting.value = false
  }
}

const handleBack = () => {
  router.push('/employees')
}

const handleCredentialsClose = () => {
  credentialsModalVisible.value = false
  router.push('/employees')
}

onMounted(() => {
  fetchEmployee()
})
</script>

<style scoped>
.employee-form {
  padding: 24px;
}
</style>