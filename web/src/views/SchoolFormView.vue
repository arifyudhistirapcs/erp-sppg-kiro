<template>
  <div class="school-form">
    <a-page-header
      :title="isEdit ? 'Edit Sekolah' : 'Tambah Sekolah'"
      :sub-title="isEdit ? 'Perbarui data sekolah' : 'Tambah sekolah baru'"
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
          <a-col :span="24">
            <a-form-item label="Nama Sekolah" name="name">
              <a-input 
                v-model:value="formData.name" 
                placeholder="Masukkan nama sekolah lengkap"
                size="large"
              />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="24">
          <a-col :span="24">
            <a-form-item label="Alamat Lengkap" name="address">
              <a-textarea 
                v-model:value="formData.address" 
                :rows="4" 
                placeholder="Masukkan alamat lengkap sekolah termasuk kecamatan dan kota"
              />
            </a-form-item>
          </a-col>
        </a-row>

        <a-divider>Koordinat GPS</a-divider>

        <a-alert
          message="Cara Mendapatkan Koordinat GPS"
          type="info"
          show-icon
          style="margin-bottom: 16px"
        >
          <template #description>
            <ol style="margin: 8px 0 0 0; padding-left: 20px;">
              <li>Buka Google Maps di browser</li>
              <li>Cari lokasi sekolah atau klik pada peta</li>
              <li>Klik kanan pada lokasi yang tepat</li>
              <li>Pilih koordinat yang muncul (contoh: -6.200000, 106.816666)</li>
              <li>Masukkan nilai Latitude (angka pertama) dan Longitude (angka kedua)</li>
            </ol>
          </template>
        </a-alert>

        <a-row :gutter="24">
          <a-col :span="12">
            <a-form-item label="Latitude" name="latitude">
              <a-input-number
                v-model:value="formData.latitude"
                :min="-90"
                :max="90"
                :precision="6"
                :step="0.000001"
                style="width: 100%"
                placeholder="-6.200000"
                size="large"
              />
              <div class="form-help">
                Rentang: -90 sampai 90 (contoh: -6.200000 untuk Jakarta)
              </div>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Longitude" name="longitude">
              <a-input-number
                v-model:value="formData.longitude"
                :min="-180"
                :max="180"
                :precision="6"
                :step="0.000001"
                style="width: 100%"
                placeholder="106.816666"
                size="large"
              />
              <div class="form-help">
                Rentang: -180 sampai 180 (contoh: 106.816666 untuk Jakarta)
              </div>
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="24" v-if="formData.latitude && formData.longitude">
          <a-col :span="24">
            <a-card size="small" title="Pratinjau Lokasi">
              <a-space>
                <a-button 
                  type="primary" 
                  @click="openMapsPreview"
                >
                  <template #icon><EnvironmentOutlined /></template>
                  Lihat di Google Maps
                </a-button>
                <a-button @click="copyCoordinates">
                  <template #icon><CopyOutlined /></template>
                  Salin Koordinat
                </a-button>
              </a-space>
              <div style="margin-top: 8px; color: #666;">
                Koordinat: {{ formData.latitude }}, {{ formData.longitude }}
              </div>
            </a-card>
          </a-col>
        </a-row>

        <a-divider>Informasi Kontak</a-divider>

        <a-row :gutter="24">
          <a-col :span="12">
            <a-form-item label="Nama Kontak Person" name="contact_person">
              <a-input 
                v-model:value="formData.contact_person" 
                placeholder="Nama kepala sekolah atau PIC"
                size="large"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Nomor Telepon" name="phone_number">
              <a-input 
                v-model:value="formData.phone_number" 
                placeholder="08xxxxxxxxxx"
                size="large"
              />
            </a-form-item>
          </a-col>
        </a-row>

        <a-divider>Informasi Tambahan</a-divider>

        <a-row :gutter="24">
          <a-col :span="12">
            <a-form-item label="Jumlah Siswa" name="student_count">
              <a-input-number
                v-model:value="formData.student_count"
                :min="0"
                :max="10000"
                style="width: 100%"
                placeholder="0"
                size="large"
              />
              <div class="form-help">
                Jumlah siswa yang akan menerima makanan
              </div>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Status Sekolah" name="is_active">
              <a-radio-group v-model:value="formData.is_active" size="large">
                <a-radio :value="true">
                  <a-tag color="green">Aktif</a-tag>
                  <span style="margin-left: 8px;">Menerima pengiriman makanan</span>
                </a-radio>
                <a-radio :value="false">
                  <a-tag color="red">Tidak Aktif</a-tag>
                  <span style="margin-left: 8px;">Tidak menerima pengiriman</span>
                </a-radio>
              </a-radio-group>
            </a-form-item>
          </a-col>
        </a-row>

        <a-divider />

        <a-row :gutter="16">
          <a-col :span="24" style="text-align: right;">
            <a-space>
              <a-button size="large" @click="handleBack">
                Batal
              </a-button>
              <a-button 
                type="primary" 
                size="large" 
                html-type="submit"
                :loading="submitting"
              >
                {{ isEdit ? 'Perbarui Sekolah' : 'Simpan Sekolah' }}
              </a-button>
            </a-space>
          </a-col>
        </a-row>
      </a-form>
    </a-card>

    <!-- GPS Validation Modal -->
    <a-modal
      v-model:open="gpsValidationVisible"
      title="Validasi Koordinat GPS"
      :footer="null"
      width="500px"
    >
      <a-result
        :status="gpsValidationStatus"
        :title="gpsValidationTitle"
        :sub-title="gpsValidationMessage"
      >
        <template #extra>
          <a-space>
            <a-button @click="gpsValidationVisible = false">
              Tutup
            </a-button>
            <a-button 
              v-if="gpsValidationStatus === 'success'" 
              type="primary"
              @click="proceedWithSave"
            >
              Lanjutkan Simpan
            </a-button>
          </a-space>
        </template>
      </a-result>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import { EnvironmentOutlined, CopyOutlined } from '@ant-design/icons-vue'
import schoolService from '@/services/schoolService'

const router = useRouter()
const route = useRoute()

const submitting = ref(false)
const gpsValidationVisible = ref(false)
const gpsValidationStatus = ref('success')
const gpsValidationTitle = ref('')
const gpsValidationMessage = ref('')
const formRef = ref()

const isEdit = computed(() => !!route.params.id)
const schoolId = computed(() => route.params.id)

const formData = reactive({
  name: '',
  address: '',
  latitude: null,
  longitude: null,
  contact_person: '',
  phone_number: '',
  student_count: 0,
  is_active: true
})

const rules = {
  name: [
    { required: true, message: 'Nama sekolah wajib diisi' },
    { min: 3, message: 'Nama sekolah minimal 3 karakter' },
    { max: 200, message: 'Nama sekolah maksimal 200 karakter' }
  ],
  address: [
    { required: true, message: 'Alamat wajib diisi' },
    { min: 10, message: 'Alamat minimal 10 karakter' }
  ],
  latitude: [
    { required: true, message: 'Latitude wajib diisi' },
    { type: 'number', min: -90, max: 90, message: 'Latitude harus antara -90 sampai 90' }
  ],
  longitude: [
    { required: true, message: 'Longitude wajib diisi' },
    { type: 'number', min: -180, max: 180, message: 'Longitude harus antara -180 sampai 180' }
  ],
  student_count: [
    { required: true, message: 'Jumlah siswa wajib diisi' },
    { type: 'number', min: 0, message: 'Jumlah siswa tidak boleh negatif' },
    { type: 'number', max: 10000, message: 'Jumlah siswa maksimal 10.000' }
  ],
  phone_number: [
    { pattern: /^(\+62|62|0)8[1-9][0-9]{6,9}$/, message: 'Format nomor telepon tidak valid' }
  ]
}

const validateGPSCoordinates = (lat, lng) => {
  // Basic validation
  if (lat < -90 || lat > 90) {
    return {
      valid: false,
      message: 'Latitude harus antara -90 sampai 90'
    }
  }
  
  if (lng < -180 || lng > 180) {
    return {
      valid: false,
      message: 'Longitude harus antara -180 sampai 180'
    }
  }

  // Check if coordinates are in Indonesia (rough bounds)
  const indonesiaBounds = {
    north: 6,
    south: -11,
    east: 141,
    west: 95
  }

  if (lat < indonesiaBounds.south || lat > indonesiaBounds.north ||
      lng < indonesiaBounds.west || lng > indonesiaBounds.east) {
    return {
      valid: true,
      warning: true,
      message: 'Koordinat berada di luar wilayah Indonesia. Pastikan koordinat sudah benar.'
    }
  }

  return { valid: true }
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    
    // Validate GPS coordinates
    const gpsValidation = validateGPSCoordinates(formData.latitude, formData.longitude)
    
    if (!gpsValidation.valid) {
      gpsValidationStatus.value = 'error'
      gpsValidationTitle.value = 'Koordinat GPS Tidak Valid'
      gpsValidationMessage.value = gpsValidation.message
      gpsValidationVisible.value = true
      return
    }

    if (gpsValidation.warning) {
      gpsValidationStatus.value = 'warning'
      gpsValidationTitle.value = 'Peringatan Koordinat GPS'
      gpsValidationMessage.value = gpsValidation.message
      gpsValidationVisible.value = true
      return
    }

    await saveSchool()
  } catch (error) {
    if (error.errorFields) {
      message.error('Mohon periksa kembali data yang diisi')
      return
    }
    console.error('Validation error:', error)
  }
}

const proceedWithSave = async () => {
  gpsValidationVisible.value = false
  await saveSchool()
}

const saveSchool = async () => {
  submitting.value = true
  try {
    if (isEdit.value) {
      await schoolService.updateSchool(schoolId.value, formData)
      message.success('Sekolah berhasil diperbarui')
    } else {
      await schoolService.createSchool(formData)
      message.success('Sekolah berhasil ditambahkan')
    }
    
    router.push('/schools')
  } catch (error) {
    message.error('Gagal menyimpan data sekolah')
    console.error(error)
  } finally {
    submitting.value = false
  }
}

const handleBack = () => {
  router.push('/schools')
}

const openMapsPreview = () => {
  if (formData.latitude && formData.longitude) {
    const url = `https://www.google.com/maps?q=${formData.latitude},${formData.longitude}`
    window.open(url, '_blank')
  }
}

const copyCoordinates = async () => {
  try {
    const coords = `${formData.latitude}, ${formData.longitude}`
    await navigator.clipboard.writeText(coords)
    message.success('Koordinat berhasil disalin')
  } catch (error) {
    message.error('Gagal menyalin koordinat')
  }
}

const loadSchoolData = async () => {
  if (!isEdit.value) return

  try {
    const response = await schoolService.getSchool(schoolId.value)
    const school = response.data
    
    Object.assign(formData, {
      name: school.name,
      address: school.address,
      latitude: school.latitude,
      longitude: school.longitude,
      contact_person: school.contact_person,
      phone_number: school.phone_number,
      student_count: school.student_count,
      is_active: school.is_active
    })
  } catch (error) {
    message.error('Gagal memuat data sekolah')
    console.error(error)
    router.push('/schools')
  }
}

onMounted(() => {
  loadSchoolData()
})
</script>

<style scoped>
.school-form {
  padding: 24px;
}

.form-help {
  font-size: 12px;
  color: #666;
  margin-top: 4px;
}

:deep(.ant-input-number) {
  width: 100%;
}

:deep(.ant-radio) {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}
</style>