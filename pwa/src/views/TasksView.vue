<template>
  <div class="tasks-container">
    <van-nav-bar title="Tugas Pengiriman" />
    <van-cell-group>
      <van-cell title="Implementasi DeliveryTasksView telah selesai" />
      <van-cell title="Silakan gunakan route /tasks untuk melihat daftar tugas pengiriman" />
    </van-cell-group>
    <van-tabbar v-model="active" route>
      <van-tabbar-item to="/tasks" icon="orders-o">Tugas</van-tabbar-item>
      <van-tabbar-item to="/attendance" icon="clock-o">Absensi</van-tabbar-item>
      <van-tabbar-item to="/profile" icon="user-o">Profil</van-tabbar-item>
    </van-tabbar>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const active = ref(0)
</script>

<style scoped>
.tasks-container {
  min-height: 100vh;
  background-color: #f7f8fa;
  padding-bottom: 50px;
}
</style>
