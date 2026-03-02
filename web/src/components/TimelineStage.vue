<template>
  <div class="timeline-stage" :class="{ 'is-last': isLast }">
    <div class="stage-line" v-if="!isLast"></div>
    <div class="stage-indicator" :class="getIndicatorClass()">
      <check-circle-filled v-if="isCompleted" />
      <loading-outlined v-else-if="isInProgress" spin />
      <div v-else class="empty-circle"></div>
    </div>
    <div class="stage-content">
      <div class="stage-header">
        <h3 class="stage-title">{{ stage.title }}</h3>
        <div v-if="stage.completed_at" class="stage-timestamp">
          {{ formatTimeRange(stage.started_at, stage.completed_at) }}
        </div>
      </div>
      <p class="stage-description">{{ stage.description }}</p>
      
      <div v-if="stage.media" class="stage-media">
        <div v-if="stage.media.type === 'photo'" class="media-photo" @click="openMedia">
          <img :src="stage.media.thumbnail_url || stage.media.url" :alt="stage.title" />
        </div>
        <div v-else-if="stage.media.type === 'video'" class="media-video" @click="openMedia">
          <div class="video-thumbnail">
            <img :src="stage.media.thumbnail_url" :alt="stage.title" />
            <div class="play-button">
              <play-circle-outlined style="font-size: 48px; color: #fff;" />
            </div>
          </div>
        </div>
      </div>

      <div v-if="stage.transitioned_by" class="stage-user">
        <user-outlined />
        <span>{{ stage.transitioned_by.name }}</span>
      </div>
    </div>

    <a-modal
      v-model:visible="showMediaModal"
      :footer="null"
      :width="800"
      centered
    >
      <div v-if="stage.media" class="media-modal-content">
        <img
          v-if="stage.media.type === 'photo'"
          :src="stage.media.url"
          :alt="stage.title"
          style="width: 100%;"
        />
        <video
          v-else-if="stage.media.type === 'video'"
          :src="stage.media.url"
          controls
          style="width: 100%;"
        ></video>
      </div>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, defineProps } from 'vue';
import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import 'dayjs/locale/id';
import {
  CheckCircleFilled,
  LoadingOutlined,
  PlayCircleOutlined,
  UserOutlined,
} from '@ant-design/icons-vue';

dayjs.extend(utc);
dayjs.locale('id');

defineProps({
  stage: {
    type: Object,
    required: true,
  },
  isCompleted: {
    type: Boolean,
    default: false,
  },
  isInProgress: {
    type: Boolean,
    default: false,
  },
  isLast: {
    type: Boolean,
    default: false,
  },
});

const showMediaModal = ref(false);

const getIndicatorClass = () => {
  const props = defineProps({
    stage: Object,
    isCompleted: Boolean,
    isInProgress: Boolean,
    isLast: Boolean,
  });
  
  if (props.isCompleted) return 'completed';
  if (props.isInProgress) return 'in-progress';
  return 'pending';
};

const formatTimeRange = (startTime, endTime) => {
  if (!startTime || !endTime) return '';
  
  // Debug: log raw values
  console.log('=== formatTimeRange Debug ===');
  console.log('startTime:', startTime);
  console.log('endTime:', endTime);
  console.log('typeof startTime:', typeof startTime);
  
  // Extract date and time directly from ISO string
  // Format: "2026-03-02T21:26:04.483536+07:00"
  // We want: "Minggu, 02 Mar 2026, 21:26 - 21:32 WIB"
  
  // Extract date parts using regex
  const startMatch = startTime.match(/^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})/);
  const endMatch = endTime.match(/^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})/);
  
  console.log('startMatch:', startMatch);
  console.log('endMatch:', endMatch);
  
  if (!startMatch || !endMatch) return '';
  
  const [, startYear, startMonth, startDay, startHour, startMin] = startMatch;
  const [, endYear, endMonth, endDay, endHour, endMin] = endMatch;
  
  console.log('Extracted:', { startYear, startMonth, startDay, startHour, startMin });
  
  // Create date string for dayjs (without timezone)
  const startDateStr = `${startYear}-${startMonth}-${startDay} ${startHour}:${startMin}`;
  const endDateStr = `${endYear}-${endMonth}-${endDay} ${endHour}:${endMin}`;
  
  console.log('startDateStr:', startDateStr);
  console.log('endDateStr:', endDateStr);
  
  // Parse with dayjs
  const start = dayjs(startDateStr, 'YYYY-MM-DD HH:mm');
  const end = dayjs(endDateStr, 'YYYY-MM-DD HH:mm');
  
  console.log('Parsed start:', start.format('YYYY-MM-DD HH:mm'));
  console.log('Parsed end:', end.format('YYYY-MM-DD HH:mm'));
  
  // Format the values
  const startDayName = start.format('dddd');
  const startDate24 = start.format('DD MMM YYYY');
  const startTime24 = start.format('HH:mm');
  const endDayName = end.format('dddd');
  const endDate24 = end.format('DD MMM YYYY');
  const endTime24 = end.format('HH:mm');
  
  console.log('Final formatted:', startDate24);
  console.log('===========================');
  
  // Check if same day
  if (startDate24 === endDate24) {
    return `${startDayName}, ${startDate24}, ${startTime24} - ${endTime24} WIB`;
  }
  
  return `${startDayName}, ${startDate24}, ${startTime24} - ${endDayName}, ${endDate24}, ${endTime24} WIB`;
};

const openMedia = () => {
  showMediaModal.value = true;
};
</script>

<style scoped>
.timeline-stage {
  position: relative;
  display: flex;
  padding-bottom: 32px;
}

.timeline-stage.is-last {
  padding-bottom: 0;
}

.stage-line {
  position: absolute;
  left: 15px;
  top: 32px;
  bottom: 0;
  width: 2px;
  background: #e8e8e8;
}

.stage-indicator {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  z-index: 1;
  background: #fff;
}

.stage-indicator.completed {
  color: #52c41a;
  font-size: 32px;
}

.stage-indicator.in-progress {
  color: #1890ff;
  font-size: 24px;
  border: 2px solid #1890ff;
}

.stage-indicator.pending {
  border: 2px solid #d9d9d9;
}

.empty-circle {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #fff;
}

.stage-content {
  flex: 1;
  margin-left: 16px;
}

.stage-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8px;
}

.stage-title {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
  color: #262626;
}

.stage-timestamp {
  font-size: 13px;
  color: #8c8c8c;
  white-space: nowrap;
  margin-left: 16px;
}

.stage-description {
  font-size: 14px;
  color: #595959;
  margin-bottom: 12px;
  line-height: 1.6;
}

.stage-media {
  margin-bottom: 12px;
}

.media-photo,
.media-video {
  cursor: pointer;
  border-radius: 8px;
  overflow: hidden;
  max-width: 200px;
  transition: transform 0.2s;
}

.media-photo:hover,
.media-video:hover {
  transform: scale(1.05);
}

.media-photo img {
  width: 100%;
  height: auto;
  display: block;
}

.video-thumbnail {
  position: relative;
  width: 100%;
}

.video-thumbnail img {
  width: 100%;
  height: auto;
  display: block;
}

.play-button {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background: rgba(0, 0, 0, 0.5);
  border-radius: 50%;
  width: 64px;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stage-user {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #8c8c8c;
}

.media-modal-content {
  padding: 16px 0;
}
</style>
