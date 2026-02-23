<template>
  <div class="school-list">
    <a-page-header
      title="Manajemen Sekolah"
      sub-title="Kelola data sekolah penerima manfaat"
    >
      <template #extra>
        <a-button type="primary" @click="showCreateModal">
          <template #icon><PlusOutlined /></template>
          Tambah Sekolah
        </a-button>
      </template>
    </a-page-header>

    <a-card>
      <a-space direction="vertical" style="width: 100%" :size="16">
        <!-- Search and Filter -->
        <a-row :gutter="16">
          <a-col :span="12">
            <a-input-search
              v-model:value="searchText"
              placeholder="Cari nama sekolah..."
              @search="handleSearch"
              allow-clear
            />
          </a-col>
          <a-col :span="6">
            <a-select
              v-model:value="filterStatus"
              placeholder="Status"
              style="width: 100%"
              @change="handleSearch"
              allow-clear
            >
              <a-select-option value="active">Aktif</a-select-option>
              <a-select-option value="inactive">Tidak Aktif</a-select-option>
            </a-select>
          </a-col>
        </a-row>

        <!-- Table -->
        <a-table
          :columns="columns"
          :data-source="schools"
          :loading="loading"
          :pagination="pagination"
          @change="handleTableChange"
          row-key="id"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'is_active'">
              <a-tag :color="record.is_active ? 'green' : 'red'">
                {{ record.is_active ? 'Aktif' : 'Tidak Aktif' }}
              </a-tag>
            </template>
            <template v-else-if="column.key === 'coordinates'">
              <a-space direction="vertical" size="small">
                <span>{{ record.latitude.toFixed(6) }}, {{ record.longitude.toFixed(6) }}</span>
                <a-button 
                  type="link" 
                  size="small" 
                  @click="openMaps(record.latitude, record.longitude)"
                >
                  <template #icon><EnvironmentOutlined /></template>
                  Lihat di Maps
                </a-button>
              </a-space>
            </template>
            <template v-else-if="column.key === 'student_count'">
              {{ formatNumber(record.student_count) }} siswa
            </template>
            <template v-else-if="column.key === 'actions'">
              <a-space>
                <a-button type="link" size="small" @click="viewSchool(record)">
                  Detail
                </a-button>
                <a-button type="link" size="small" @click="editSchool(record)">
                  Edit
                </a-button>
                <a-popconfirm
                  title="Yakin ingin menghapus sekolah ini?"
                  ok-text="Ya"
                  cancel-text="Tidak"
                  @confirm="deleteSchool(record.id)"
                >
                  <a-button type="link" size="small" danger>
                    Hapus
                  </a-button>
                </a-popconfirm>
              </a-space>
            </template>
          </template>
        </a-table>
      </a-space>
    </a-card>

    <!-- Create/Edit Modal -->
    <a-modal
      v-model:open="modalVisible"
      :title="editingSchool ? 'Edit Sekolah' : 'Tambah Sekolah'"
      :confirm-loading="submitting"
      @ok="handleSubmit"
      @cancel="handleCancel"
      width="700px"
    >
      <a-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        layout="vertical"
      >
        <a-form-item label="Nama Sekolah" name="name">
          <a-input v-model:value="formData.name" placeholder="Masukkan nama sekolah" />
        </a-form-item>

        <a-form-item label="Alamat" name="address">
          <a-textarea v-model:value="formData.address" :rows="3" placeholder="Alamat lengkap sekolah" />
        </a-form-item>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Latitude" name="latitude">
              <a-input-number
                v-model:value="formData.latitude"
                :min="-90"
                :max="90"
                :precision="6"
                style="width: 100%"
                placeholder="-6.200000"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Longitude" name="longitude">
              <a-input-number
                v-model:value="formData.longitude"
                :min="-180"
                :max="180"
                :precision="6"
                style="width: 100%"
                placeholder="106.816666"
              />
            </a-form-item>
          </a-col>
        </a-row>

        <a-alert
          message="Tips GPS Coordinates"
          description="Anda dapat mendapatkan koordinat GPS dari Google Maps dengan klik kanan pada lokasi sekolah dan pilih koordinat yang muncul."
          type="info"
          show-icon
          style="margin-bottom: 16px"
        />

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Nama Kontak" name="contact_person">
              <a-input v-model:value="formData.contact_person" placeholder="Nama kontak person" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Nomor Telepon" name="phone_number">
              <a-input v-model:value="formData.phone_number" placeholder="08xxxxxxxxxx" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Jumlah Siswa" name="student_count">
              <a-input-number
                v-model:value="formData.student_count"
                :min="0"
                style="width: 100%"
                placeholder="0"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Status" name="is_active">
              <a-switch 
                v-model:checked="formData.is_active" 
                checked-children="Aktif" 
                un-checked-children="Tidak Aktif" 
              />
            </a-form-item>
          </a-col>
        </a-row>
      </a-form>
    </a-modal>

    <!-- Detail Modal -->
    <a-modal
      v-model:open="detailModalVisible"
      title="Detail Sekolah"
      :footer="null"
      width="800px"
    >
      <a-descriptions v-if="selectedSchool" bordered :column="2">
        <a-descriptions-item label="Nama Sekolah" :span="2">
          {{ selectedSchool.name }}
        </a-descriptions-item>
        <a-descriptions-item label="Alamat" :span="2">
          {{ selectedSchool.address }}
        </a-descriptions-item>
        <a-descriptions-item label="Latitude">
          {{ selectedSchool.latitude?.toFixed(6) }}
        </a-descriptions-item>
        <a-descriptions-item label="Longitude">
          {{ selectedSchool.longitude?.toFixed(6) }}
        </a-descriptions-item>
        <a-descriptions-item label="Kontak Person">
          {{ selectedSchool.contact_person || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="Telepon">
          {{ selectedSchool.phone_number || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="Jumlah Siswa">
          {{ formatNumber(selectedSchool.student_count) }} siswa
        </a-descriptions-item>
        <a-descriptions-item label="Status">
          <a-tag :color="selectedSchool.is_active ? 'green' : 'red'">
            {{ selectedSchool.is_active ? 'Aktif' : 'Tidak Aktif' }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="Dibuat" :span="2">
          {{ formatDate(selectedSchool.created_at) }}
        </a-descriptions-item>
        <a-descriptions-item label="Diperbarui" :span="2">
          {{ formatDate(selectedSchool.updated_at) }}
        </a-descriptions-item>
      </a-descriptions>

      <a-divider>Lokasi GPS</a-divider>

      <a-space>
        <a-button 
          type="primary" 
          @click="openMaps(selectedSchool.latitude, selectedSchool.longitude)"
        >
          <template #icon><EnvironmentOutlined /></template>
          Buka di Google Maps
        </a-button>
        <a-button @click="copyCoordinates(selectedSchool.latitude, selectedSchool.longitude)">
          <template #icon><CopyOutlined /></template>
          Salin Koordinat
        </a-button>
      </a-space>

      <a-divider>Riwayat Perubahan</a-divider>

      <a-table
        :columns="historyColumns"
        :data-source="changeHistory"
        :loading="loadingHistory"
        :pagination="{ pageSize: 5 }"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'timestamp'">
            {{ formatDateTime(record.timestamp) }}
          </template>
          <template v-else-if="column.key === 'action'">
            <a-tag :color="getActionColor(record.action)">
              {{ getActionText(record.action) }}
            </a-tag>
          </template>
        </template>
      </a-table>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined, EnvironmentOutlined, CopyOutlined } from '@ant-design/icons-vue'
import schoolService from '@/services/schoolService'

const loading = ref(false)
const submitting = ref(false)
const modalVisible = ref(false)
const detailModalVisible = ref(false)
const editingSchool = ref(null)
const selectedSchool = ref(null)
const schools = ref([])
const changeHistory = ref([])
const loadingHistory = ref(false)
const searchText = ref('')
const filterStatus = ref(undefined)
const formRef = ref()

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0
})

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
  name: [{ required: true, message: 'Nama sekolah wajib diisi' }],
  address: [{ required: true, message: 'Alamat wajib diisi' }],
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
    { type: 'number', min: 0, message: 'Jumlah siswa tidak boleh negatif' }
  ]
}

const columns = [
  {
    title: 'Nama Sekolah',
    dataIndex: 'name',
    key: 'name',
    sorter: true
  },
  {
    title: 'Alamat',
    dataIndex: 'address',
    key: 'address',
    ellipsis: true
  },
  {
    title: 'Koordinat GPS',
    key: 'coordinates',
    width: 200
  },
  {
    title: 'Kontak',
    dataIndex: 'contact_person',
    key: 'contact_person'
  },
  {
    title: 'Jumlah Siswa',
    key: 'student_count',
    sorter: true,
    width: 120
  },
  {
    title: 'Status',
    key: 'is_active',
    width: 100
  },
  {
    title: 'Aksi',
    key: 'actions',
    width: 200
  }
]

const historyColumns = [
  {
    title: 'Waktu',
    key: 'timestamp',
    width: 150
  },
  {
    title: 'Aksi',
    key: 'action',
    width: 100
  },
  {
    title: 'Pengguna',
    dataIndex: 'user_name',
    key: 'user_name'
  },
  {
    title: 'Keterangan',
    dataIndex: 'description',
    key: 'description'
  }
]

const fetchSchools = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.current,
      page_size: pagination.pageSize,
      search: searchText.value || undefined,
      is_active: filterStatus.value
    }
    const response = await schoolService.getSchools(params)
    schools.value = response.data.data || []
    pagination.total = response.data.total || 0
  } catch (error) {
    message.error('Gagal memuat data sekolah')
    console.error(error)
  } finally {
    loading.value = false
  }
}

const handleTableChange = (pag, filters, sorter) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  fetchSchools()
}

const handleSearch = () => {
  pagination.current = 1
  fetchSchools()
}

const showCreateModal = () => {
  editingSchool.value = null
  resetForm()
  modalVisible.value = true
}

const editSchool = (school) => {
  editingSchool.value = school
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
  modalVisible.value = true
}

const viewSchool = async (school) => {
  selectedSchool.value = school
  detailModalVisible.value = true
  
  // Fetch change history (mock data for now)
  loadingHistory.value = true
  try {
    // TODO: Implement actual change history API
    changeHistory.value = [
      {
        timestamp: new Date(),
        action: 'update',
        user_name: 'Admin',
        description: 'Memperbarui data kontak'
      },
      {
        timestamp: new Date(Date.now() - 86400000),
        action: 'create',
        user_name: 'Admin',
        description: 'Membuat data sekolah'
      }
    ]
  } catch (error) {
    console.error('Gagal memuat riwayat perubahan:', error)
  } finally {
    loadingHistory.value = false
  }
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    submitting.value = true

    if (editingSchool.value) {
      await schoolService.updateSchool(editingSchool.value.id, formData)
      message.success('Sekolah berhasil diperbarui')
    } else {
      await schoolService.createSchool(formData)
      message.success('Sekolah berhasil ditambahkan')
    }

    modalVisible.value = false
    fetchSchools()
  } catch (error) {
    if (error.errorFields) {
      return
    }
    message.error('Gagal menyimpan data sekolah')
    console.error(error)
  } finally {
    submitting.value = false
  }
}

const deleteSchool = async (id) => {
  try {
    await schoolService.deleteSchool(id)
    message.success('Sekolah berhasil dihapus')
    fetchSchools()
  } catch (error) {
    message.error('Gagal menghapus sekolah')
    console.error(error)
  }
}

const handleCancel = () => {
  modalVisible.value = false
  resetForm()
}

const resetForm = () => {
  Object.assign(formData, {
    name: '',
    address: '',
    latitude: null,
    longitude: null,
    contact_person: '',
    phone_number: '',
    student_count: 0,
    is_active: true
  })
  formRef.value?.resetFields()
}

const openMaps = (lat, lng) => {
  const url = `https://www.google.com/maps?q=${lat},${lng}`
  window.open(url, '_blank')
}

const copyCoordinates = async (lat, lng) => {
  try {
    await navigator.clipboard.writeText(`${lat}, ${lng}`)
    message.success('Koordinat berhasil disalin')
  } catch (error) {
    message.error('Gagal menyalin koordinat')
  }
}

const formatNumber = (value) => {
  return new Intl.NumberFormat('id-ID').format(value)
}

const formatDate = (date) => {
  return new Date(date).toLocaleDateString('id-ID', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}

const formatDateTime = (date) => {
  return new Date(date).toLocaleString('id-ID', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const getActionColor = (action) => {
  const colors = {
    create: 'green',
    update: 'blue',
    delete: 'red'
  }
  return colors[action] || 'default'
}

const getActionText = (action) => {
  const texts = {
    create: 'Dibuat',
    update: 'Diperbarui',
    delete: 'Dihapus'
  }
  return texts[action] || action
}

onMounted(() => {
  fetchSchools()
})
</script>

<style scoped>
.school-list {
  padding: 24px;
}
</style>