<template>
  <div class="profile-container">
    <van-nav-bar title="Profil" />
    <van-cell-group>
      <van-cell title="Profil pengguna akan diimplementasikan pada task berikutnya" />
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

const active = ref(2)
</script>

<style scoped>
.profile-container {
  min-height: 100vh;
  background-color: #f7f8fa;
  padding-bottom: 50px;
}
</style>
