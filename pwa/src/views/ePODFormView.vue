<template>
  <div class="epod-form-container">
    <!-- Navigation Bar -->
    <van-nav-bar 
      title="Bukti Pengiriman Digital" 
      left-arrow 
      fixed
      @click-left="goBack"
    >
      <template #right>
        <van-icon 
          name="info-o" 
          @click="showHelp"
        />
      </template>
    </van-nav-bar>

    <!-- Offline Indicator -->
    <van-notice-bar 
      v-if="!isOnline" 
      type="warning" 
      text="Mode offline - Data akan disinkronkan saat online"
      left-icon="warning-o"
    />

    <!-- Sync Status Indicator -->
    <van-notice-bar 
      v-if="syncStatus.pending > 0"
      type="primary" 
      :text="`${syncStatus.pending} item menunggu sinkronisasi`"
      left-icon="clock-o"
      @click="showSyncDetails"
    />

    <!-- Sync Progress Indicator -->
    <van-notice-bar 
      v-if="syncStatus.syncing"
      type="primary" 
      :text="`Menyinkronkan... ${syncStatus.progress}%`"
      left-icon="loading"
    />

    <!-- Loading State -->
    <van-loading v-if="isLoading" type="spinner" vertical>
      Memuat data pengiriman...
    </van-loading>

    <!-- Error State -->
    <van-empty 
      v-else-if="!isLoading && !deliveryTask"
      image="error"
      description="Data pengiriman tidak ditemukan"
    >
      <van-button type="primary" @click="goBack">
        Kembali
      </van-button>
    </van-empty>

    <!-- e-POD Form Content -->
    <div v-else-if="deliveryTask" class="form-content">
      <!-- Delivery Task Summary -->
      <van-card class="task-summary">
        <template #title>
          <div class="task-header">
            <span>{{ deliveryTask.school?.name }}</span>
            <van-tag type="primary" size="large">
              {{ deliveryTask.portions }} Porsi
            </van-tag>
          </div>
        </template>
        
        <div class="task-info">
          <div class="info-item">
            <van-icon name="location-o" />
            <span>{{ deliveryTask.school?.address }}</span>
          </div>
          <div class="info-item">
            <van-icon name="contact" />
            <span>{{ deliveryTask.school?.contact_person }}</span>
          </div>
        </div>
      </van-card>

      <!-- GPS Location Section -->
      <van-cell-group title="Lokasi GPS" class="form-section">
        <van-cell 
          title="Status GPS" 
          :value="gpsStatus"
          :icon="gpsIcon"
          :class="gpsStatusClass"
        />
        <van-cell 
          v-if="currentLocation.latitude"
          title="Latitude" 
          :value="currentLocation.latitude.toFixed(6)"
          icon="aim"
        />
        <van-cell 
          v-if="currentLocation.longitude"
          title="Longitude" 
          :value="currentLocation.longitude.toFixed(6)"
          icon="aim"
        />
        <van-cell 
          v-if="currentLocation.accuracy"
          title="Akurasi" 
          :value="`±${currentLocation.accuracy.toFixed(0)}m`"
          icon="location-o"
          :class="getAccuracyClass()"
        />
        
        <van-button 
          v-if="!isCapturingGPS && (!currentLocation.latitude || currentLocation.accuracy > 50)"
          type="primary" 
          size="small"
          @click="captureGPS"
          :loading="isCapturingGPS"
          class="gps-button"
        >
          <van-icon name="aim" />
          {{ currentLocation.latitude ? 'Perbarui GPS' : 'Ambil Lokasi GPS' }}
        </van-button>
      </van-cell-group>

      <!-- Ompreng Tracking Section -->
      <van-cell-group title="Pelacakan Ompreng" class="form-section">
        <van-field
          v-model="formData.omprengDropOff"
          type="number"
          label="Ompreng Diantar"
          placeholder="Masukkan jumlah ompreng yang diantar"
          :rules="[{ required: true, message: 'Jumlah ompreng diantar wajib diisi' }]"
          :error="errors.omprengDropOff"
          :error-message="errors.omprengDropOff"
        >
          <template #left-icon>
            <van-icon name="shopping-cart-o" />
          </template>
        </van-field>
        
        <van-field
          v-model="formData.omprengPickUp"
          type="number"
          label="Ompreng Diambil"
          placeholder="Masukkan jumlah ompreng yang diambil"
          :rules="[{ required: true, message: 'Jumlah ompreng diambil wajib diisi' }]"
          :error="errors.omprengPickUp"
          :error-message="errors.omprengPickUp"
        >
          <template #left-icon>
            <van-icon name="shopping-cart-o" />
          </template>
        </van-field>
      </van-cell-group>

      <!-- Photo Section -->
      <van-cell-group title="Foto Bukti Pengiriman" class="form-section">
        <van-cell 
          title="Foto Serah Terima" 
          is-link
          @click="takePhoto"
          :value="getPhotoStatus()"
          :icon="formData.photo ? 'success' : 'photograph'"
        />
        
        <!-- Camera Selection -->
        <van-cell 
          v-if="availableCameras.length > 1"
          title="Kamera" 
          is-link
          @click="showCameraSelection = true"
          :value="selectedCamera?.label || 'Pilih kamera'"
          icon="video-o"
        />
        
        <div v-if="formData.photo" class="photo-preview">
          <img :src="formData.photo" alt="Foto bukti pengiriman" />
          <div class="photo-info">
            <span class="photo-size">{{ getPhotoSize() }}</span>
            <span class="photo-quality">{{ photoQuality }}% kualitas</span>
          </div>
          <div class="photo-actions">
            <van-button 
              type="primary" 
              size="small" 
              @click="retakePhoto"
              class="retake-photo-btn"
            >
              <van-icon name="photograph" />
              Ambil Ulang
            </van-button>
            <van-button 
              type="danger" 
              size="small" 
              @click="removePhoto"
              class="remove-photo-btn"
            >
              <van-icon name="delete-o" />
              Hapus Foto
            </van-button>
          </div>
        </div>
      </van-cell-group>

      <!-- Digital Signature Section -->
      <van-cell-group title="Tanda Tangan Digital" class="form-section">
        <van-field
          v-model="formData.recipientName"
          label="Nama Penerima"
          placeholder="Masukkan nama penerima"
          :rules="[{ required: true, message: 'Nama penerima wajib diisi' }]"
          :error="errors.recipientName"
          :error-message="errors.recipientName"
        >
          <template #left-icon>
            <van-icon name="contact" />
          </template>
        </van-field>
        
        <van-cell 
          title="Tanda Tangan" 
          is-link
          @click="openSignaturePad"
          :value="getSignatureStatus()"
          :icon="formData.signature ? 'success' : 'edit'"
        />
        
        <div v-if="formData.signature" class="signature-preview">
          <div class="signature-image-container">
            <img :src="formData.signature" alt="Tanda tangan digital" />
            <div class="signature-overlay">
              <van-icon name="success" size="20" />
            </div>
          </div>
          <div class="signature-info">
            <div class="signature-quality-badge" :class="getQualityClass()">
              {{ getQualityText() }}
            </div>
            <span class="signature-timestamp">
              {{ new Date().toLocaleString('id-ID') }}
            </span>
          </div>
          <div class="signature-actions-inline">
            <van-button 
              type="primary" 
              size="small" 
              @click="openSignaturePad"
              class="re-sign-btn"
            >
              <van-icon name="edit" />
              Tanda Tangan Ulang
            </van-button>
            <van-button 
              type="danger" 
              size="small" 
              @click="removeSignature"
              class="remove-signature-btn"
            >
              <van-icon name="delete-o" />
              Hapus
            </van-button>
          </div>
        </div>
      </van-cell-group>

      <!-- Submit Button -->
      <div class="submit-section">
        <van-button 
          type="success" 
          size="large"
          block 
          @click="submitePOD"
          :loading="isSubmitting"
          :disabled="!canSubmit"
          class="submit-button"
        >
          <van-icon name="success" />
          Kirim Bukti Pengiriman
        </van-button>
        
        <div v-if="!canSubmit" class="validation-info">
          <van-icon name="info-o" />
          <span>Lengkapi semua data yang diperlukan</span>
        </div>
      </div>
    </div>

    <!-- Camera Action Sheet -->
    <van-action-sheet 
      v-model:show="showCameraSheet" 
      :actions="cameraActions"
      @select="onCameraAction"
      cancel-text="Batal"
      description="Pilih cara mengambil foto"
    />

    <!-- Camera Selection Sheet -->
    <van-action-sheet 
      v-model:show="showCameraSelection" 
      :actions="cameraSelectionActions"
      @select="onCameraSelection"
      cancel-text="Batal"
      description="Pilih kamera yang akan digunakan"
    />

    <!-- Camera Preview Dialog -->
    <van-dialog 
      v-model:show="showCameraPreview" 
      title="Ambil Foto"
      :show-cancel-button="false"
      :show-confirm-button="false"
      class="camera-dialog"
    >
      <div class="camera-container">
        <video 
          ref="videoElement"
          class="camera-video"
          autoplay
          playsinline
          muted
        ></video>
        
        <canvas 
          ref="photoCanvas"
          class="photo-canvas"
          style="display: none;"
        ></canvas>
        
        <div class="camera-overlay">
          <div class="camera-frame"></div>
        </div>
        
        <div class="camera-controls">
          <van-button 
            type="default" 
            @click="closeCameraPreview"
            size="large"
            round
            class="camera-control-btn"
          >
            <van-icon name="cross" size="24" />
          </van-button>
          
          <van-button 
            type="primary" 
            @click="capturePhoto"
            size="large"
            round
            class="camera-control-btn capture-btn"
            :loading="isCapturingPhoto"
          >
            <van-icon name="photograph" size="28" />
          </van-button>
          
          <van-button 
            v-if="availableCameras.length > 1"
            type="default" 
            @click="switchCamera"
            size="large"
            round
            class="camera-control-btn"
          >
            <van-icon name="replay" size="24" />
          </van-button>
        </div>
        
        <div class="camera-info">
          <span>{{ selectedCamera?.label || 'Kamera' }}</span>
        </div>
      </div>
    </van-dialog>

    <!-- Signature Pad Dialog -->
    <van-dialog 
      v-model:show="showSignatureDialog" 
      title="Tanda Tangan Digital"
      :show-cancel-button="false"
      :show-confirm-button="false"
      class="signature-dialog"
    >
      <div class="signature-pad-container">
        <div class="signature-instructions">
          <van-icon name="info-o" />
          <span>Silakan buat tanda tangan di area di bawah ini</span>
        </div>
        
        <div class="signature-canvas-wrapper">
          <canvas 
            ref="signatureCanvas"
            class="signature-canvas"
            @touchstart="startDrawing"
            @touchmove="draw"
            @touchend="stopDrawing"
            @mousedown="startDrawing"
            @mousemove="draw"
            @mouseup="stopDrawing"
            @mouseleave="stopDrawing"
          ></canvas>
          
          <div v-if="!hasSignature" class="signature-placeholder">
            <van-icon name="edit" size="24" />
            <span>Tanda tangan di sini</span>
          </div>
        </div>
        
        <div v-if="hasSignature" class="signature-quality">
          <div class="quality-label">
            <span>Kualitas tanda tangan:</span>
            <span :class="getQualityClass()">{{ getQualityText() }}</span>
          </div>
          <van-progress 
            :percentage="signatureQuality" 
            :color="getQualityColor()"
            stroke-width="4"
          />
        </div>
        
        <div class="signature-actions">
          <van-button 
            type="default" 
            @click="clearSignature"
            size="small"
            :disabled="!hasSignature"
          >
            <van-icon name="delete-o" />
            Hapus
          </van-button>
          <van-button 
            type="primary" 
            @click="previewSignatureBeforeSave"
            size="small"
            :disabled="!hasSignature || signatureQuality < 40"
          >
            <van-icon name="eye-o" />
            Preview
          </van-button>
          <van-button 
            type="default" 
            @click="closeSignaturePad"
            size="small"
          >
            Batal
          </van-button>
        </div>
      </div>
    </van-dialog>

    <!-- Signature Preview Dialog -->
    <van-dialog 
      v-model:show="showSignaturePreview" 
      title="Konfirmasi Tanda Tangan"
      :show-cancel-button="false"
      :show-confirm-button="false"
      class="signature-preview-dialog"
    >
      <div class="signature-preview-container">
        <div class="preview-instructions">
          <van-icon name="success" />
          <span>Apakah tanda tangan ini sudah benar?</span>
        </div>
        
        <div class="signature-preview-image">
          <img v-if="previewSignature" :src="previewSignature" alt="Preview tanda tangan" />
        </div>
        
        <div class="signature-preview-info">
          <div class="info-item">
            <span>Kualitas:</span>
            <span :class="getQualityClass()">{{ getQualityText() }}</span>
          </div>
          <div class="info-item">
            <span>Goresan:</span>
            <span>{{ signatureStrokes.length }} goresan</span>
          </div>
        </div>
        
        <div class="signature-preview-actions">
          <van-button 
            type="default" 
            @click="cancelSignaturePreview"
            size="large"
          >
            <van-icon name="arrow-left" />
            Kembali
          </van-button>
          <van-button 
            type="success" 
            @click="confirmSignature"
            size="large"
          >
            <van-icon name="success" />
            Simpan
          </van-button>
        </div>
      </div>
    </van-dialog>

    <!-- Help Dialog -->
    <van-dialog 
      v-model:show="showHelpDialog" 
      title="Panduan e-POD"
      :show-cancel-button="false"
      confirm-button-text="Mengerti"
    >
      <div class="help-content">
        <h4>Langkah-langkah:</h4>
        <ol>
          <li>Pastikan GPS aktif dan akurat (≤50m)</li>
          <li>Isi jumlah ompreng yang diantar dan diambil</li>
          <li>Ambil foto bukti serah terima</li>
          <li>Minta tanda tangan penerima</li>
          <li>Kirim bukti pengiriman</li>
        </ol>
        
        <h4>Tips:</h4>
        <ul>
          <li>Foto harus jelas dan menunjukkan proses serah terima</li>
          <li>Tanda tangan harus dibuat oleh penerima</li>
          <li>Data akan tersimpan meski offline</li>
        </ul>
      </div>
    </van-dialog>

    <!-- Sync Status Modal -->
    <SyncStatusModal v-model:show="showSyncStatusModal" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useDeliveryTasksStore } from '@/stores/deliveryTasks'
import { showToast, showConfirmDialog, showSuccessToast } from 'vant'
import { validateSignatureQuality, compressSignature, getSignatureSize } from '@/utils/signatureValidator'
import db from '@/services/db'
import SyncStatusModal from '@/components/SyncStatusModal.vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const deliveryTasksStore = useDeliveryTasksStore()

// Reactive data
const isLoading = ref(false)
const isSubmitting = ref(false)
const isCapturingGPS = ref(false)
const isOnline = ref(navigator.onLine)
const deliveryTask = ref(null)

// Sync status tracking
const syncStatus = ref({
  pending: 0,
  syncing: false,
  progress: 0,
  lastSync: null
})

// GPS and location
const currentLocation = ref({
  latitude: null,
  longitude: null,
  accuracy: null,
  timestamp: null
})

// Form data
const formData = ref({
  omprengDropOff: '',
  omprengPickUp: '',
  photo: null,
  signature: null,
  recipientName: ''
})

// Form validation
const errors = ref({
  omprengDropOff: '',
  omprengPickUp: '',
  recipientName: ''
})

// UI state
const showCameraSheet = ref(false)
const showCameraSelection = ref(false)
const showCameraPreview = ref(false)
const showSignatureDialog = ref(false)
const showHelpDialog = ref(false)
const showSyncStatusModal = ref(false)

// Camera functionality
const videoElement = ref(null)
const photoCanvas = ref(null)
const currentStream = ref(null)
const availableCameras = ref([])
const selectedCamera = ref(null)
const isCapturingPhoto = ref(false)
const photoQuality = ref(80) // JPEG quality percentage

// Camera actions
const cameraActions = [
  { name: 'camera', text: 'Ambil Foto dengan Kamera', icon: 'photograph' },
  { name: 'gallery', text: 'Pilih dari Galeri', icon: 'photo-o' }
]

// Camera selection actions (computed)
const cameraSelectionActions = computed(() => {
  return availableCameras.value.map((camera, index) => ({
    name: camera.deviceId,
    text: camera.label || `Kamera ${index + 1}`,
    icon: camera.label?.toLowerCase().includes('front') ? 'user-o' : 'photograph'
  }))
})

// Signature pad
const signatureCanvas = ref(null)
const isDrawing = ref(false)
const hasSignature = ref(false)
const signatureContext = ref(null)
const lastPoint = ref({ x: 0, y: 0 })
const signatureStrokes = ref([])
const signatureQuality = ref(0)
const showSignaturePreview = ref(false)
const previewSignature = ref(null)

// Computed properties
const gpsStatus = computed(() => {
  if (isCapturingGPS.value) return 'Mengambil lokasi...'
  if (!currentLocation.value.latitude) return 'GPS belum diambil'
  if (currentLocation.value.accuracy > 50) return 'Akurasi rendah'
  return 'GPS siap'
})

const gpsIcon = computed(() => {
  if (isCapturingGPS.value) return 'loading'
  if (!currentLocation.value.latitude) return 'location-o'
  if (currentLocation.value.accuracy > 50) return 'warning-o'
  return 'success'
})

const gpsStatusClass = computed(() => {
  if (!currentLocation.value.latitude || currentLocation.value.accuracy > 50) {
    return 'gps-warning'
  }
  return 'gps-success'
})

const canSubmit = computed(() => {
  return (
    currentLocation.value.latitude &&
    currentLocation.value.longitude &&
    currentLocation.value.accuracy <= 50 &&
    formData.value.omprengDropOff &&
    formData.value.omprengPickUp &&
    formData.value.photo &&
    formData.value.signature &&
    formData.value.recipientName.trim()
  )
})

const getQualityClass = () => {
  if (signatureQuality.value >= 70) return 'quality-excellent'
  if (signatureQuality.value >= 50) return 'quality-good'
  if (signatureQuality.value >= 40) return 'quality-fair'
  return 'quality-poor'
}

const getQualityText = () => {
  if (signatureQuality.value >= 70) return 'Sangat Baik'
  if (signatureQuality.value >= 50) return 'Baik'
  if (signatureQuality.value >= 40) return 'Cukup'
  return 'Kurang'
}

const getQualityColor = () => {
  if (signatureQuality.value >= 70) return '#07c160'
  if (signatureQuality.value >= 50) return '#1989fa'
  if (signatureQuality.value >= 40) return '#ff976a'
  return '#ee0a24'
}

// Methods
const loadDeliveryTask = async () => {
  const taskId = route.params.taskId
  if (!taskId) {
    showToast('ID tugas tidak valid')
    goBack()
    return
  }

  isLoading.value = true
  try {
    // Get task from store
    let task = deliveryTasksStore.getTaskById(parseInt(taskId))
    
    if (!task) {
      // Fetch today's tasks if not in store
      await deliveryTasksStore.fetchTodayTasks(authStore.user.id)
      task = deliveryTasksStore.getTaskById(parseInt(taskId))
    }
    
    if (task) {
      deliveryTask.value = task
      // Initialize camera devices
      await initializeCameraDevices()
      // Auto-capture GPS when form loads
      await captureGPS()
    } else {
      showToast('Tugas pengiriman tidak ditemukan')
      goBack()
    }
  } catch (error) {
    console.error('Error loading delivery task:', error)
    showToast('Gagal memuat data pengiriman')
  } finally {
    isLoading.value = false
  }
}

const initializeCameraDevices = async () => {
  try {
    // Request camera permission first
    await navigator.mediaDevices.getUserMedia({ video: true })
    
    // Get available camera devices
    const devices = await navigator.mediaDevices.enumerateDevices()
    availableCameras.value = devices.filter(device => device.kind === 'videoinput')
    
    // Select rear camera by default if available
    const rearCamera = availableCameras.value.find(camera => 
      camera.label.toLowerCase().includes('back') || 
      camera.label.toLowerCase().includes('rear') ||
      camera.label.toLowerCase().includes('environment')
    )
    
    selectedCamera.value = rearCamera || availableCameras.value[0]
    
    console.log('Available cameras:', availableCameras.value.length)
  } catch (error) {
    console.error('Error initializing cameras:', error)
    // Camera access denied or not available
    availableCameras.value = []
  }
}

const captureGPS = async () => {
  if (!navigator.geolocation) {
    showToast('GPS tidak tersedia di perangkat ini')
    return
  }

  isCapturingGPS.value = true
  
  try {
    const position = await new Promise((resolve, reject) => {
      navigator.geolocation.getCurrentPosition(
        resolve,
        reject,
        {
          enableHighAccuracy: true,
          timeout: 15000,
          maximumAge: 60000
        }
      )
    })

    currentLocation.value = {
      latitude: position.coords.latitude,
      longitude: position.coords.longitude,
      accuracy: position.coords.accuracy,
      timestamp: new Date().toISOString()
    }

    if (position.coords.accuracy > 50) {
      showToast('Akurasi GPS rendah. Coba lagi di tempat terbuka')
    } else {
      showSuccessToast('Lokasi GPS berhasil diambil')
    }
  } catch (error) {
    console.error('GPS Error:', error)
    let errorMessage = 'Gagal mengambil lokasi GPS'
    
    switch (error.code) {
      case error.PERMISSION_DENIED:
        errorMessage = 'Akses GPS ditolak. Aktifkan izin lokasi'
        break
      case error.POSITION_UNAVAILABLE:
        errorMessage = 'Lokasi tidak tersedia. Pastikan GPS aktif'
        break
      case error.TIMEOUT:
        errorMessage = 'Timeout GPS. Coba lagi'
        break
    }
    
    showToast(errorMessage)
  } finally {
    isCapturingGPS.value = false
  }
}

const getAccuracyClass = () => {
  if (!currentLocation.value.accuracy) return ''
  return currentLocation.value.accuracy <= 20 ? 'accuracy-good' : 
         currentLocation.value.accuracy <= 50 ? 'accuracy-fair' : 'accuracy-poor'
}

const takePhoto = () => {
  if (availableCameras.value.length === 0) {
    showToast('Kamera tidak tersedia di perangkat ini')
    return
  }
  showCameraSheet.value = true
}

const onCameraAction = async (action) => {
  showCameraSheet.value = false
  
  if (action.name === 'camera') {
    await openCameraPreview()
  } else if (action.name === 'gallery') {
    await selectFromGallery()
  }
}

const onCameraSelection = (action) => {
  const camera = availableCameras.value.find(cam => cam.deviceId === action.name)
  if (camera) {
    selectedCamera.value = camera
    showCameraSelection.value = false
    showSuccessToast(`Kamera ${camera.label || 'dipilih'}`)
  }
}

const openCameraPreview = async () => {
  if (!selectedCamera.value) {
    showToast('Pilih kamera terlebih dahulu')
    return
  }

  try {
    showCameraPreview.value = true
    await nextTick()
    
    const constraints = {
      video: {
        deviceId: selectedCamera.value.deviceId,
        width: { ideal: 1920 },
        height: { ideal: 1080 },
        facingMode: selectedCamera.value.label?.toLowerCase().includes('front') ? 'user' : 'environment'
      }
    }
    
    currentStream.value = await navigator.mediaDevices.getUserMedia(constraints)
    
    if (videoElement.value) {
      videoElement.value.srcObject = currentStream.value
      await videoElement.value.play()
    }
  } catch (error) {
    console.error('Camera error:', error)
    showToast('Gagal mengakses kamera')
    closeCameraPreview()
  }
}

const capturePhoto = async () => {
  if (!videoElement.value || !photoCanvas.value) {
    showToast('Kamera tidak siap')
    return
  }

  isCapturingPhoto.value = true
  
  try {
    const video = videoElement.value
    const canvas = photoCanvas.value
    
    // Set canvas dimensions to match video
    canvas.width = video.videoWidth
    canvas.height = video.videoHeight
    
    // Draw video frame to canvas
    const context = canvas.getContext('2d')
    context.drawImage(video, 0, 0, canvas.width, canvas.height)
    
    // Compress and convert to JPEG
    const compressedDataURL = canvas.toDataURL('image/jpeg', photoQuality.value / 100)
    
    // Store photo data
    formData.value.photo = compressedDataURL
    
    // Store in IndexedDB for offline capability
    await storePhotoOffline(compressedDataURL)
    
    closeCameraPreview()
    showSuccessToast('Foto berhasil diambil')
  } catch (error) {
    console.error('Photo capture error:', error)
    showToast('Gagal mengambil foto')
  } finally {
    isCapturingPhoto.value = false
  }
}

const switchCamera = async () => {
  if (availableCameras.value.length <= 1) return
  
  // Find next camera
  const currentIndex = availableCameras.value.findIndex(cam => cam.deviceId === selectedCamera.value?.deviceId)
  const nextIndex = (currentIndex + 1) % availableCameras.value.length
  selectedCamera.value = availableCameras.value[nextIndex]
  
  // Restart camera with new device
  closeCameraPreview()
  await openCameraPreview()
}

const closeCameraPreview = () => {
  if (currentStream.value) {
    currentStream.value.getTracks().forEach(track => track.stop())
    currentStream.value = null
  }
  
  if (videoElement.value) {
    videoElement.value.srcObject = null
  }
  
  showCameraPreview.value = false
}

const storePhotoOffline = async (photoData) => {
  try {
    // Store in IndexedDB using Dexie for offline access
    const photoRecord = {
      taskId: deliveryTask.value.id,
      photoData: photoData,
      timestamp: new Date().toISOString(),
      synced: false
    }
    
    await db.photos.add(photoRecord)
    console.log('Photo stored offline successfully')
  } catch (error) {
    console.error('Error storing photo offline:', error)
    // Continue without offline storage if it fails
  }
}

const openIndexedDB = () => {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('ePODDatabase', 1)
    
    request.onerror = () => reject(request.error)
    request.onsuccess = () => resolve(request.result)
    
    request.onupgradeneeded = (event) => {
      const db = event.target.result
      
      // Create photos store if it doesn't exist
      if (!db.objectStoreNames.contains('photos')) {
        const photosStore = db.createObjectStore('photos', { keyPath: 'id' })
        photosStore.createIndex('taskId', 'taskId', { unique: false })
        photosStore.createIndex('synced', 'synced', { unique: false })
      }
    }
  })
}

const retakePhoto = () => {
  openCameraPreview()
}

const captureFromCamera = async () => {
  // Legacy method - now redirects to new camera preview
  await openCameraPreview()
}

const selectFromGallery = () => {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = 'image/*'
  
  input.onchange = async (event) => {
    const file = event.target.files[0]
    if (file) {
      try {
        // Compress the selected image
        const compressedDataURL = await compressImage(file, photoQuality.value)
        formData.value.photo = compressedDataURL
        
        // Store in IndexedDB for offline capability
        await storePhotoOffline(compressedDataURL)
        
        showSuccessToast('Foto berhasil dipilih')
      } catch (error) {
        console.error('Error processing selected image:', error)
        showToast('Gagal memproses foto')
      }
    }
  }
  
  input.click()
}

const compressImage = (file, quality) => {
  return new Promise((resolve, reject) => {
    const canvas = document.createElement('canvas')
    const context = canvas.getContext('2d')
    const img = new Image()
    
    img.onload = () => {
      // Calculate new dimensions (max 1920x1080)
      let { width, height } = img
      const maxWidth = 1920
      const maxHeight = 1080
      
      if (width > maxWidth || height > maxHeight) {
        const ratio = Math.min(maxWidth / width, maxHeight / height)
        width *= ratio
        height *= ratio
      }
      
      canvas.width = width
      canvas.height = height
      
      // Draw and compress
      context.drawImage(img, 0, 0, width, height)
      const compressedDataURL = canvas.toDataURL('image/jpeg', quality / 100)
      resolve(compressedDataURL)
    }
    
    img.onerror = reject
    img.src = URL.createObjectURL(file)
  })
}

const getPhotoStatus = () => {
  if (!formData.value.photo) return 'Ambil foto'
  return `Foto tersimpan (${getPhotoSize()})`
}

const getPhotoSize = () => {
  if (!formData.value.photo) return ''
  
  // Calculate approximate size from base64 data
  const base64Length = formData.value.photo.length
  const sizeInBytes = (base64Length * 3) / 4
  
  if (sizeInBytes < 1024) {
    return `${Math.round(sizeInBytes)} B`
  } else if (sizeInBytes < 1024 * 1024) {
    return `${Math.round(sizeInBytes / 1024)} KB`
  } else {
    return `${Math.round(sizeInBytes / (1024 * 1024))} MB`
  }
}

const removePhoto = () => {
  formData.value.photo = null
  showToast('Foto dihapus')
}

const openSignaturePad = async () => {
  showSignatureDialog.value = true
  await nextTick()
  initSignaturePad()
}

const initSignaturePad = () => {
  if (!signatureCanvas.value) return
  
  const canvas = signatureCanvas.value
  const container = canvas.parentElement
  
  // Set canvas size to container size with high DPI support
  const rect = container.getBoundingClientRect()
  const dpr = window.devicePixelRatio || 1
  
  canvas.width = rect.width * dpr
  canvas.height = 200 * dpr
  canvas.style.width = rect.width + 'px'
  canvas.style.height = '200px'
  
  signatureContext.value = canvas.getContext('2d')
  const ctx = signatureContext.value
  
  // Scale context for high DPI
  ctx.scale(dpr, dpr)
  
  // Set drawing properties for smooth lines
  ctx.lineCap = 'round'
  ctx.lineJoin = 'round'
  ctx.strokeStyle = '#000000'
  ctx.lineWidth = 2
  ctx.imageSmoothingEnabled = true
  
  // Clear canvas with white background
  ctx.fillStyle = '#ffffff'
  ctx.fillRect(0, 0, rect.width, 200)
  
  // Reset signature state
  hasSignature.value = false
  signatureStrokes.value = []
  signatureQuality.value = 0
  
  // Prevent scrolling when drawing on mobile
  canvas.addEventListener('touchstart', preventScroll, { passive: false })
  canvas.addEventListener('touchmove', preventScroll, { passive: false })
  canvas.addEventListener('touchend', preventScroll, { passive: false })
}

const preventScroll = (event) => {
  event.preventDefault()
}

const getEventPosition = (event) => {
  const canvas = signatureCanvas.value
  const rect = canvas.getBoundingClientRect()
  
  let clientX, clientY
  
  if (event.touches && event.touches.length > 0) {
    // Touch event
    clientX = event.touches[0].clientX
    clientY = event.touches[0].clientY
  } else {
    // Mouse event
    clientX = event.clientX
    clientY = event.clientY
  }
  
  return {
    x: clientX - rect.left,
    y: clientY - rect.top
  }
}

const startDrawing = (event) => {
  event.preventDefault()
  
  if (!signatureContext.value) return
  
  isDrawing.value = true
  const pos = getEventPosition(event)
  
  lastPoint.value = pos
  
  // Start new stroke
  const newStroke = [pos]
  signatureStrokes.value.push(newStroke)
  
  signatureContext.value.beginPath()
  signatureContext.value.moveTo(pos.x, pos.y)
}

const draw = (event) => {
  if (!isDrawing.value || !signatureContext.value) return
  
  event.preventDefault()
  const pos = getEventPosition(event)
  
  // Add point to current stroke
  const currentStroke = signatureStrokes.value[signatureStrokes.value.length - 1]
  currentStroke.push(pos)
  
  // Draw smooth line using quadratic curves
  const ctx = signatureContext.value
  const midPoint = {
    x: (lastPoint.value.x + pos.x) / 2,
    y: (lastPoint.value.y + pos.y) / 2
  }
  
  ctx.quadraticCurveTo(lastPoint.value.x, lastPoint.value.y, midPoint.x, midPoint.y)
  ctx.stroke()
  
  lastPoint.value = pos
  hasSignature.value = true
  
  // Update signature quality based on stroke complexity
  updateSignatureQuality()
}

const stopDrawing = (event) => {
  if (!isDrawing.value) return
  
  event.preventDefault()
  isDrawing.value = false
  
  if (signatureContext.value) {
    signatureContext.value.closePath()
  }
}

const updateSignatureQuality = () => {
  const validation = validateSignatureQuality(signatureStrokes.value)
  signatureQuality.value = validation.quality
  hasSignature.value = validation.isValid || signatureStrokes.value.length > 0
}

const clearSignature = () => {
  if (signatureContext.value && signatureCanvas.value) {
    const canvas = signatureCanvas.value
    const ctx = signatureContext.value
    const rect = canvas.getBoundingClientRect()
    
    // Clear canvas and fill with white background
    ctx.clearRect(0, 0, rect.width, 200)
    ctx.fillStyle = '#ffffff'
    ctx.fillRect(0, 0, rect.width, 200)
    
    // Reset state
    hasSignature.value = false
    signatureStrokes.value = []
    signatureQuality.value = 0
  }
}

const validateSignature = () => {
  const validation = validateSignatureQuality(signatureStrokes.value)
  
  if (!validation.isValid) {
    showToast(validation.feedback)
    return false
  }
  
  return true
}

const previewSignatureBeforeSave = () => {
  if (!validateSignature()) return
  
  // Create preview image
  previewSignature.value = compressCurrentSignature()
  showSignaturePreview.value = true
}

const compressCurrentSignature = () => {
  if (!hasSignature.value || !signatureCanvas.value) return null
  
  return compressSignature(signatureCanvas.value, 400, 200)
}

const confirmSignature = async () => {
  if (previewSignature.value) {
    formData.value.signature = previewSignature.value
    
    // Store signature offline
    await storeSignatureOffline(previewSignature.value)
    
    showSignaturePreview.value = false
    showSignatureDialog.value = false
    showSuccessToast('Tanda tangan berhasil disimpan')
  }
}

const cancelSignaturePreview = () => {
  showSignaturePreview.value = false
  previewSignature.value = null
}

const storeSignatureOffline = async (signatureData) => {
  try {
    // Store in IndexedDB for offline access
    const signatureRecord = {
      taskId: deliveryTask.value.id,
      signatureData: signatureData,
      quality: signatureQuality.value,
      timestamp: new Date().toISOString(),
      synced: false
    }
    
    // Add to signatures table (we'll need to update the DB schema)
    await db.signatures.add(signatureRecord)
    console.log('Signature stored offline successfully')
  } catch (error) {
    console.error('Error storing signature offline:', error)
    // Continue without offline storage if it fails
  }
}

const saveSignature = () => {
  // Use the new preview flow instead of direct save
  previewSignatureBeforeSave()
}

const getSignatureStatus = () => {
  if (!formData.value.signature) return 'Buat tanda tangan'
  return `Tanda tangan tersimpan (${getQualityText()})`
}

const removeSignature = async () => {
  try {
    const confirmed = await showConfirmDialog({
      title: 'Hapus Tanda Tangan',
      message: 'Apakah Anda yakin ingin menghapus tanda tangan ini?',
      confirmButtonText: 'Ya, Hapus',
      cancelButtonText: 'Batal'
    })
    
    if (confirmed) {
      formData.value.signature = null
      signatureQuality.value = 0
      hasSignature.value = false
      signatureStrokes.value = []
      showToast('Tanda tangan dihapus')
    }
  } catch (error) {
    // User cancelled
  }
}

const closeSignaturePad = () => {
  // Clean up event listeners
  if (signatureCanvas.value) {
    const canvas = signatureCanvas.value
    canvas.removeEventListener('touchstart', preventScroll)
    canvas.removeEventListener('touchmove', preventScroll)
    canvas.removeEventListener('touchend', preventScroll)
  }
  
  showSignatureDialog.value = false
}

const validateForm = () => {
  errors.value = {
    omprengDropOff: '',
    omprengPickUp: '',
    recipientName: ''
  }
  
  let isValid = true
  
  if (!formData.value.omprengDropOff || formData.value.omprengDropOff < 0) {
    errors.value.omprengDropOff = 'Jumlah ompreng diantar harus diisi dan tidak boleh negatif'
    isValid = false
  }
  
  if (!formData.value.omprengPickUp || formData.value.omprengPickUp < 0) {
    errors.value.omprengPickUp = 'Jumlah ompreng diambil harus diisi dan tidak boleh negatif'
    isValid = false
  }
  
  if (!formData.value.recipientName.trim()) {
    errors.value.recipientName = 'Nama penerima wajib diisi'
    isValid = false
  }
  
  // Additional signature validation
  if (!formData.value.signature) {
    showToast('Tanda tangan digital wajib dibuat')
    isValid = false
  }
  
  // GPS validation
  if (!currentLocation.value.latitude || !currentLocation.value.longitude) {
    showToast('Lokasi GPS wajib diambil')
    isValid = false
  }
  
  if (currentLocation.value.accuracy > 50) {
    showToast('Akurasi GPS terlalu rendah. Silakan ambil ulang lokasi GPS')
    isValid = false
  }
  
  // Photo validation
  if (!formData.value.photo) {
    showToast('Foto bukti pengiriman wajib diambil')
    isValid = false
  }
  
  return isValid
}

const submitePOD = async () => {
  if (!validateForm()) {
    showToast('Mohon lengkapi semua data yang diperlukan')
    return
  }
  
  if (!canSubmit.value) {
    showToast('Pastikan GPS akurat dan semua data sudah lengkap')
    return
  }
  
  try {
    const confirmed = await showConfirmDialog({
      title: 'Kirim Bukti Pengiriman',
      message: `Apakah Anda yakin ingin mengirim bukti pengiriman untuk ${deliveryTask.value.school?.name}?`,
      confirmButtonText: 'Ya, Kirim',
      cancelButtonText: 'Batal'
    })
    
    if (confirmed) {
      isSubmitting.value = true
      
      // Prepare comprehensive e-POD data
      const ePODData = {
        delivery_task_id: deliveryTask.value.id,
        latitude: currentLocation.value.latitude,
        longitude: currentLocation.value.longitude,
        accuracy: currentLocation.value.accuracy,
        recipient_name: formData.value.recipientName.trim(),
        ompreng_drop_off: parseInt(formData.value.omprengDropOff),
        ompreng_pick_up: parseInt(formData.value.omprengPickUp),
        photo_url: formData.value.photo,
        signature_url: formData.value.signature,
        completed_at: new Date().toISOString(),
        device_info: {
          userAgent: navigator.userAgent,
          timestamp: new Date().toISOString(),
          timezone: Intl.DateTimeFormat().resolvedOptions().timeZone
        }
      }
      
      // Submit e-POD with enhanced error handling and sync status
      const result = await deliveryTasksStore.submitePOD(ePODData)
      
      // Handle different submission results
      if (result.success) {
        if (result.synced) {
          showSuccessToast('✅ Bukti pengiriman berhasil dikirim dan status pengiriman diperbarui')
        } else if (result.offline) {
          showSuccessToast('📱 Bukti pengiriman disimpan offline dan akan disinkronkan saat online')
        } else if (result.queued) {
          showSuccessToast(`⏳ ${result.message || 'Bukti pengiriman dalam antrian sinkronisasi'}`)
        }
        
        // Show sync status if not immediately synced
        if (!result.synced) {
          showSyncStatusToast()
        }
        
        // Navigate back to task detail with success state
        router.push({
          path: `/tasks/${deliveryTask.value.id}`,
          query: { epodSubmitted: 'true' }
        })
      } else {
        throw new Error(result.message || 'Gagal menyimpan bukti pengiriman')
      }
    }
  } catch (error) {
    console.error('Error submitting e-POD:', error)
    
    // Enhanced error handling with specific messages
    if (error.offline || error.message?.includes('offline')) {
      showSuccessToast('📱 Bukti pengiriman disimpan offline dan akan disinkronkan saat online')
      router.push(`/tasks/${deliveryTask.value.id}`)
    } else if (error.message?.includes('GPS')) {
      showToast('❌ Error GPS: ' + error.message)
    } else if (error.message?.includes('photo')) {
      showToast('❌ Error foto: ' + error.message)
    } else if (error.message?.includes('signature')) {
      showToast('❌ Error tanda tangan: ' + error.message)
    } else {
      showToast('❌ ' + (error.message || 'Gagal mengirim bukti pengiriman. Coba lagi.'))
    }
  } finally {
    isSubmitting.value = false
  }
}

// Show sync status toast with progress info
const showSyncStatusToast = async () => {
  try {
    const pendingCount = await deliveryTasksStore.getPendingSyncCount()
    if (pendingCount > 0) {
      showToast(`📊 ${pendingCount} item menunggu sinkronisasi`)
    }
  } catch (error) {
    console.error('Error getting sync status:', error)
  }
}

const goBack = () => {
  router.go(-1)
}

const showHelp = () => {
  showHelpDialog.value = true
}

// Sync status methods
const updateSyncStatus = async () => {
  try {
    const pendingCount = await deliveryTasksStore.getPendingSyncCount()
    const progress = deliveryTasksStore.getSyncProgress()
    
    syncStatus.value = {
      pending: pendingCount,
      syncing: progress.status === 'syncing',
      progress: progress.total > 0 ? Math.round((progress.completed / progress.total) * 100) : 0,
      lastSync: new Date().toISOString()
    }
  } catch (error) {
    console.error('Error updating sync status:', error)
  }
}

const showSyncDetails = () => {
  showSyncStatusModal.value = true
}

// Sync progress listener
const onSyncProgress = (progress) => {
  syncStatus.value = {
    ...syncStatus.value,
    syncing: progress.status === 'syncing',
    progress: progress.total > 0 ? Math.round((progress.completed / progress.total) * 100) : 0
  }
  
  if (progress.status === 'completed') {
    updateSyncStatus() // Refresh pending count
    showSuccessToast('✅ Sinkronisasi selesai')
  } else if (progress.status === 'completed_with_errors') {
    updateSyncStatus()
    showToast('⚠️ Sinkronisasi selesai dengan beberapa error')
  }
}

// Network status handlers
const handleOnline = async () => {
  isOnline.value = true
  showSuccessToast('🌐 Koneksi internet tersambung')
  
  // Update sync status and start sync
  await updateSyncStatus()
  deliveryTasksStore.syncAllOfflineData()
}

const handleOffline = () => {
  isOnline.value = false
  showToast('📱 Mode offline - Data akan disimpan lokal')
}

// Lifecycle
onMounted(async () => {
  await loadDeliveryTask()
  
  // Initialize sync status
  await updateSyncStatus()
  
  // Add sync progress listener
  deliveryTasksStore.addSyncProgressListener(onSyncProgress)
  
  // Listen for network status changes
  window.addEventListener('online', handleOnline)
  window.addEventListener('offline', handleOffline)
  
  // Periodic sync status update
  const syncStatusInterval = setInterval(updateSyncStatus, 30000) // Every 30 seconds
  
  // Store interval for cleanup
  window.syncStatusInterval = syncStatusInterval
})

onUnmounted(() => {
  // Clean up camera stream
  closeCameraPreview()
  
  // Remove sync progress listener
  deliveryTasksStore.removeSyncProgressListener(onSyncProgress)
  
  // Remove event listeners
  window.removeEventListener('online', handleOnline)
  window.removeEventListener('offline', handleOffline)
  
  // Clear sync status interval
  if (window.syncStatusInterval) {
    clearInterval(window.syncStatusInterval)
    delete window.syncStatusInterval
  }
})
</script>

<style scoped>
.epod-form-container {
  min-height: 100vh;
  background-color: #f7f8fa;
  padding-top: 46px; /* Nav bar height */
  padding-bottom: 16px;
}

.form-content {
  padding: 16px;
}

.task-summary {
  margin-bottom: 16px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

.task-info {
  margin-top: 12px;
}

.info-item {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
  color: #646566;
  font-size: 14px;
}

.info-item .van-icon {
  margin-right: 8px;
  color: #1989fa;
}

.form-section {
  margin-bottom: 16px;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.gps-warning {
  color: #ff976a;
}

.gps-success {
  color: #07c160;
}

.accuracy-good {
  color: #07c160;
}

.accuracy-fair {
  color: #ff976a;
}

.accuracy-poor {
  color: #ee0a24;
}

.gps-button {
  margin: 12px 16px;
  border-radius: 6px;
}

.photo-preview,
.signature-preview {
  padding: 16px;
  text-align: center;
}

.photo-preview img,
.signature-preview img {
  max-width: 100%;
  max-height: 200px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.photo-info {
  display: flex;
  justify-content: space-between;
  margin: 8px 0;
  font-size: 12px;
  color: #969799;
}

.photo-actions {
  display: flex;
  gap: 8px;
  justify-content: center;
  margin-top: 12px;
}

.retake-photo-btn,
.remove-photo-btn,
.remove-signature-btn {
  border-radius: 6px;
}

/* Camera Dialog Styles */
.camera-dialog {
  width: 95%;
  max-width: none;
  height: 90vh;
  max-height: none;
}

.camera-container {
  position: relative;
  width: 100%;
  height: 70vh;
  background-color: #000;
  border-radius: 8px;
  overflow: hidden;
}

.camera-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.photo-canvas {
  position: absolute;
  top: 0;
  left: 0;
}

.camera-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
}

.camera-frame {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 80%;
  height: 60%;
  border: 2px solid rgba(255, 255, 255, 0.8);
  border-radius: 8px;
  box-shadow: 0 0 0 9999px rgba(0, 0, 0, 0.3);
}

.camera-controls {
  position: absolute;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: 20px;
  align-items: center;
}

.camera-control-btn {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.3);
}

.capture-btn {
  width: 80px;
  height: 80px;
  background: linear-gradient(135deg, #1989fa, #1c7cd6);
}

.camera-info {
  position: absolute;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(0, 0, 0, 0.6);
  color: white;
  padding: 8px 16px;
  border-radius: 20px;
  font-size: 14px;
  backdrop-filter: blur(4px);
}

.submit-section {
  margin-top: 24px;
  padding: 0 4px;
}

.submit-button {
  height: 48px;
  font-size: 16px;
  font-weight: 600;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(7, 193, 96, 0.3);
}

.submit-button .van-icon {
  margin-right: 8px;
  font-size: 18px;
}

.validation-info {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 12px;
  color: #969799;
  font-size: 14px;
}

.validation-info .van-icon {
  margin-right: 6px;
}

.signature-dialog {
  width: 90%;
  max-width: 400px;
}

.signature-pad-container {
  padding: 16px;
}

.signature-instructions {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 16px;
  color: #646566;
  font-size: 14px;
}

.signature-instructions .van-icon {
  margin-right: 8px;
  color: #1989fa;
}

.signature-canvas-wrapper {
  position: relative;
  margin-bottom: 16px;
}

.signature-canvas {
  width: 100%;
  height: 200px;
  border: 2px solid #dcdee0;
  border-radius: 8px;
  background-color: #fff;
  touch-action: none;
  cursor: crosshair;
}

.signature-canvas:active {
  border-color: #1989fa;
}

.signature-placeholder {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  color: #c8c9cc;
  pointer-events: none;
  font-size: 14px;
}

.signature-placeholder .van-icon {
  margin-bottom: 8px;
}

.signature-quality {
  margin-bottom: 16px;
  padding: 12px;
  background-color: #f7f8fa;
  border-radius: 6px;
}

.quality-label {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-size: 14px;
}

.quality-excellent { color: #07c160; font-weight: 600; }
.quality-good { color: #1989fa; font-weight: 600; }
.quality-fair { color: #ff976a; font-weight: 600; }
.quality-poor { color: #ee0a24; font-weight: 600; }

.signature-actions {
  display: flex;
  justify-content: space-between;
  gap: 8px;
}

.signature-actions .van-button {
  flex: 1;
  border-radius: 6px;
}

.signature-preview-dialog {
  width: 85%;
  max-width: 350px;
}

.signature-preview-container {
  padding: 16px;
}

.preview-instructions {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 16px;
  color: #07c160;
  font-size: 14px;
  font-weight: 500;
}

.preview-instructions .van-icon {
  margin-right: 8px;
}

.signature-preview-image {
  text-align: center;
  margin-bottom: 16px;
  padding: 16px;
  background-color: #f7f8fa;
  border-radius: 8px;
}

.signature-preview-image img {
  max-width: 100%;
  height: auto;
  border-radius: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.signature-preview-info {
  margin-bottom: 20px;
}

.signature-preview-info .info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid #ebedf0;
  font-size: 14px;
}

.signature-preview-info .info-item:last-child {
  border-bottom: none;
}

.signature-preview-actions {
  display: flex;
  gap: 12px;
}

.signature-preview-actions .van-button {
  flex: 1;
  height: 44px;
  border-radius: 6px;
  font-weight: 500;
}

.signature-preview {
  padding: 16px;
  text-align: center;
}

.signature-image-container {
  position: relative;
  display: inline-block;
  margin-bottom: 12px;
}

.signature-image-container img {
  max-width: 100%;
  max-height: 120px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  background-color: #fff;
}

.signature-overlay {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 28px;
  height: 28px;
  background-color: #07c160;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.signature-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-size: 12px;
}

.signature-quality-badge {
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
}

.signature-quality-badge.quality-excellent {
  background-color: rgba(7, 193, 96, 0.1);
  color: #07c160;
}

.signature-quality-badge.quality-good {
  background-color: rgba(25, 137, 250, 0.1);
  color: #1989fa;
}

.signature-quality-badge.quality-fair {
  background-color: rgba(255, 151, 106, 0.1);
  color: #ff976a;
}

.signature-quality-badge.quality-poor {
  background-color: rgba(238, 10, 36, 0.1);
  color: #ee0a24;
}

.signature-timestamp {
  color: #969799;
}

.signature-actions-inline {
  display: flex;
  gap: 8px;
  justify-content: center;
}

.re-sign-btn,
.remove-signature-btn {
  border-radius: 6px;
}

.help-content {
  padding: 16px;
  text-align: left;
}

.help-content h4 {
  margin: 16px 0 8px 0;
  color: #323233;
  font-size: 16px;
}

.help-content ol,
.help-content ul {
  margin: 0;
  padding-left: 20px;
}

.help-content li {
  margin-bottom: 8px;
  line-height: 1.5;
  color: #646566;
}

/* Responsive adjustments */
@media (max-width: 375px) {
  .form-content {
    padding: 12px;
  }
  
  .task-header {
    font-size: 14px;
  }
  
  .submit-button {
    height: 44px;
    font-size: 15px;
  }
  
  .camera-control-btn {
    width: 50px;
    height: 50px;
  }
  
  .capture-btn {
    width: 70px;
    height: 70px;
  }
  
  .photo-actions {
    flex-direction: column;
    gap: 8px;
  }
  
  .signature-dialog {
    width: 95%;
  }
  
  .signature-canvas {
    height: 160px;
  }
  
  .signature-actions {
    flex-direction: column;
    gap: 8px;
  }
  
  .signature-actions .van-button {
    width: 100%;
  }
  
  .signature-preview-actions {
    flex-direction: column;
    gap: 8px;
  }
  
  .signature-actions-inline {
    flex-direction: column;
    gap: 8px;
  }
}

/* Landscape orientation adjustments */
@media (orientation: landscape) and (max-height: 500px) {
  .signature-dialog {
    width: 95%;
    height: 90vh;
  }
  
  .signature-canvas {
    height: 120px;
  }
  
  .signature-pad-container {
    padding: 12px;
  }
  
  .signature-quality {
    margin-bottom: 12px;
    padding: 8px;
  }
}

/* Touch-specific improvements */
@media (pointer: coarse) {
  .signature-canvas {
    border-width: 3px;
  }
  
  .signature-actions .van-button {
    height: 44px;
    font-size: 16px;
  }
  
  .signature-preview-actions .van-button {
    height: 48px;
    font-size: 16px;
  }
}

/* Landscape orientation for camera */
@media (orientation: landscape) {
  .camera-container {
    height: 80vh;
  }
  
  .camera-frame {
    width: 60%;
    height: 80%;
  }
}

/* Dark mode support */
@media (prefers-color-scheme: dark) {
  .epod-form-container {
    background-color: #1a1a1a;
  }
  
  .task-summary,
  .form-section {
    background-color: #2a2a2a;
  }
  
  .signature-canvas {
    background-color: #2a2a2a;
    border-color: #4a4a4a;
  }
  
  .camera-container {
    background-color: #000;
  }
  
  .photo-info {
    color: #969799;
  }
}
</style>