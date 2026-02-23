<template>
  <div class="employee-list">
    <a-page-header
      title="Manajemen Karyawan"
      sub-title="Kelola data karyawan dan akun pengguna"
    >
      <template #extra>
        <a-button type="primary" @click="showCreateModal">
          <template #icon><PlusOutlined /></template>
          Tambah Karyawan
        </a-button>
      </template>
    </a-page-header>

    <a-card>
      <a-space direction="vertical" style="width: 100%" :size="16">
        <!-- Search and Filter -->
        <a-row :gutter="16">
          <a-col :span="8">
            <a-input-search
              v-model:value="searchText"
              placeholder="Cari NIK, nama, atau email..."
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
          <a-col :span="6">
            <a-select
              v-model:value="filterPosition"
              placeholder="Posisi"
              style="width: 100%"
              @change="handleSearch"
              allow-clear
            >
              <a-select-option value="Kepala SPPG">Kepala SPPG</a-select-option>
              <a-select-option value="Akuntan">Akuntan</a-select-option>
              <a-select-option value="Ahli Gizi">Ahli Gizi</a-select-option>
              <a-select-option value="Pengadaan">Pengadaan</a-select-option>
              <a-select-option value="Chef">Chef</a-select-option>
              <a-select-option value="Packing">Packing</a-select-option>
              <a-select-option value="Driver">Driver</a-select-option>
              <a-select-option value="Asisten Lapangan">Asisten Lapangan</a-select-option>
            </a-select>
          </a-col>
        </a-row>

        <!-- Statistics Cards -->
        <a-row :gutter="16">
          <a-col :span="6">
            <a-card size="small">
              <a-statistic
                title="Total Karyawan"
                :value="stats.total_employees || 0"
                :value-style="{ color: '#1890ff' }"
              />
            </a-card>
          </a-col>
          <a-col :span="6">
            <a-card size="small">
              <a-statistic
                title="Karyawan Aktif"
                :value="stats.active_employees || 0"
                :value-style="{ color: '#52c41a' }"
              />
            </a-card>
          </a-col>
          <a-col :span="6">
            <a-card size="small">
              <a-statistic
                title="Tidak Aktif"
                :value="stats.inactive_employees || 0"
                :value-style="{ color: '#ff4d4f' }"
              />
            </a-card>
          </a-col>
        </a-row>

        <!-- Table -->
        <a-table
          :columns="columns"
          :data-source="employees"
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
            <template v-else-if="column.key === 'role'">
              <a-tag color="blue">{{ getRoleLabel(record.user?.role) }}</a-tag>
            </template>
            <template v-else-if="column.key === 'join_date'">
              {{ formatDate(record.join_date) }}
            </template>
            <template v-else-if="column.key === 'actions'">
              <a-space>
                <a-button type="link" size="small" @click="viewEmployee(record)">
                  Detail
                </a-button>
                <a-button type="link" size="small" @click="editEmployee(record)">
                  Edit
                </a-button>
                <a-popconfirm
                  :title="record.is_active ? 'Yakin ingin menonaktifkan karyawan ini?' : 'Yakin ingin mengaktifkan karyawan ini?'"
                  ok-text="Ya"
                  cancel-text="Tidak"
                  @confirm="toggleEmployeeStatus(record)"
                >
                  <a-button type="link" size="small" :danger="record.is_active">
                    {{ record.is_active ? 'Nonaktifkan' : 'Aktifkan' }}
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
      :title="editingEmployee ? 'Edit Karyawan' : 'Tambah Karyawan'"
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
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="NIK" name="nik">
              <a-input v-model:value="formData.nik" placeholder="Nomor Induk Karyawan" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Nama Lengkap" name="full_name">
              <a-input v-model:value="formData.full_name" placeholder="Nama lengkap karyawan" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Email" name="email">
              <a-input v-model:value="formData.email" type="email" placeholder="email@example.com" />
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
          </a-col>
          <a-col :span="12">
            <a-form-item label="Role Sistem" name="role">
              <a-select v-model:value="formData.role" placeholder="Pilih role sistem">
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
          </a-col>
        </a-row>

        <a-form-item label="Tanggal Bergabung" name="join_date">
          <a-date-picker 
            v-model:value="formData.join_date" 
            style="width: 100%" 
            placeholder="Pilih tanggal bergabung"
            format="DD/MM/YYYY"
          />
        </a-form-item>

        <a-form-item label="Status" name="is_active">
          <a-switch v-model:checked="formData.is_active" checked-children="Aktif" un-checked-children="Tidak Aktif" />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- Detail Modal -->
    <a-modal
      v-model:open="detailModalVisible"
      title="Detail Karyawan"
      :footer="null"
      width="800px"
    >
      <a-descriptions v-if="selectedEmployee" bordered :column="2">
        <a-descriptions-item label="NIK">
          {{ selectedEmployee.nik }}
        </a-descriptions-item>
        <a-descriptions-item label="Nama Lengkap">
          {{ selectedEmployee.full_name }}
        </a-descriptions-item>
        <a-descriptions-item label="Email">
          {{ selectedEmployee.email }}
        </a-descriptions-item>
        <a-descriptions-item label="Nomor Telepon">
          {{ selectedEmployee.phone_number }}
        </a-descriptions-item>
        <a-descriptions-item label="Posisi">
          {{ selectedEmployee.position }}
        </a-descriptions-item>
        <a-descriptions-item label="Role Sistem">
          <a-tag color="blue">{{ getRoleLabel(selectedEmployee.user?.role) }}</a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="Tanggal Bergabung">
          {{ formatDate(selectedEmployee.join_date) }}
        </a-descriptions-item>
        <a-descriptions-item label="Status">
          <a-tag :color="selectedEmployee.is_active ? 'green' : 'red'">
            {{ selectedEmployee.is_active ? 'Aktif' : 'Tidak Aktif' }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="Dibuat Pada" :span="2">
          {{ formatDateTime(selectedEmployee.created_at) }}
        </a-descriptions-item>
      </a-descriptions>
    </a-modal>

    <!-- Credentials Modal -->
    <a-modal
      v-model:open="credentialsModalVisible"
      title="Kredensial Login Karyawan"
      :footer="null"
      width="500px"
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
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import employeeService from '@/services/employeeService'

const loading = ref(false)
const submitting = ref(false)
const modalVisible = ref(false)
const detailModalVisible = ref(false)
const credentialsModalVisible = ref(false)
const editingEmployee = ref(null)
const selectedEmployee = ref(null)
const employees = ref([])
const stats = ref({})
const searchText = ref('')
const filterStatus = ref(undefined)
const filterPosition = ref(undefined)
const formRef = ref()
const newCredentials = ref({})

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0
})

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
  nik: [{ required: true, message: 'NIK wajib diisi' }],
  full_name: [{ required: true, message: 'Nama lengkap wajib diisi' }],
  email: [
    { required: true, message: 'Email wajib diisi' },
    { type: 'email', message: 'Format email tidak valid' }
  ],
  phone_number: [{ required: true, message: 'Nomor telepon wajib diisi' }],
  position: [{ required: true, message: 'Posisi wajib dipilih' }],
  role: [{ required: true, message: 'Role sistem wajib dipilih' }],
  join_date: [{ required: true, message: 'Tanggal bergabung wajib diisi' }]
}

const columns = [
  {
    title: 'NIK',
    dataIndex: 'nik',
    key: 'nik',
    width: 120
  },
  {
    title: 'Nama Lengkap',
    dataIndex: 'full_name',
    key: 'full_name',
    sorter: true
  },
  {
    title: 'Email',
    dataIndex: 'email',
    key: 'email'
  },
  {
    title: 'Posisi',
    dataIndex: 'position',
    key: 'position'
  },
  {
    title: 'Role Sistem',
    key: 'role',
    width: 120
  },
  {
    title: 'Tanggal Bergabung',
    key: 'join_date',
    width: 140
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

const fetchEmployees = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.current,
      page_size: pagination.pageSize,
      search: searchText.value || undefined,
      is_active: filterStatus.value === 'active' ? true : filterStatus.value === 'inactive' ? false : undefined,
      position: filterPosition.value || undefined
    }
    const response = await employeeService.getEmployees(params)
    employees.value = response.data || []
    pagination.total = response.total || 0
  } catch (error) {
    message.error('Gagal memuat data karyawan')
    console.error(error)
  } finally {
    loading.value = false
  }
}

const fetchStats = async () => {
  try {
    const response = await employeeService.getEmployeeStats()
    stats.value = response.data || {}
  } catch (error) {
    console.error('Gagal memuat statistik karyawan:', error)
  }
}

const handleTableChange = (pag, filters, sorter) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  fetchEmployees()
}

const handleSearch = () => {
  pagination.current = 1
  fetchEmployees()
}

const showCreateModal = () => {
  editingEmployee.value = null
  resetForm()
  modalVisible.value = true
}

const editEmployee = (employee) => {
  editingEmployee.value = employee
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
  modalVisible.value = true
}

const viewEmployee = (employee) => {
  selectedEmployee.value = employee
  detailModalVisible.value = true
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    submitting.value = true

    const submitData = {
      ...formData,
      join_date: formData.join_date ? formData.join_date.format('YYYY-MM-DD') : null
    }

    if (editingEmployee.value) {
      await employeeService.updateEmployee(editingEmployee.value.id, submitData)
      message.success('Karyawan berhasil diperbarui')
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
      }
    }

    modalVisible.value = false
    fetchEmployees()
    fetchStats()
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

const toggleEmployeeStatus = async (employee) => {
  try {
    if (employee.is_active) {
      await employeeService.deactivateEmployee(employee.id)
      message.success('Karyawan berhasil dinonaktifkan')
    } else {
      // For reactivation, we need to call update with is_active: true
      await employeeService.updateEmployee(employee.id, { is_active: true })
      message.success('Karyawan berhasil diaktifkan')
    }
    fetchEmployees()
    fetchStats()
  } catch (error) {
    message.error('Gagal mengubah status karyawan')
    console.error(error)
  }
}

const handleCancel = () => {
  modalVisible.value = false
  resetForm()
}

const resetForm = () => {
  Object.assign(formData, {
    nik: '',
    full_name: '',
    email: '',
    phone_number: '',
    position: '',
    role: '',
    join_date: null,
    is_active: true
  })
  formRef.value?.resetFields()
}

const getRoleLabel = (role) => {
  const roleLabels = {
    kepala_sppg: 'Kepala SPPG/Yayasan',
    akuntan: 'Akuntan',
    ahli_gizi: 'Ahli Gizi',
    pengadaan: 'Pengadaan',
    chef: 'Chef',
    packing: 'Packing',
    driver: 'Driver',
    asisten: 'Asisten Lapangan'
  }
  return roleLabels[role] || role
}

const formatDate = (date) => {
  if (!date) return '-'
  return dayjs(date).format('DD/MM/YYYY')
}

const formatDateTime = (date) => {
  if (!date) return '-'
  return dayjs(date).format('DD/MM/YYYY HH:mm')
}

onMounted(() => {
  fetchEmployees()
  fetchStats()
})
</script>

<style scoped>
.employee-list {
  padding: 24px;
}
</style>