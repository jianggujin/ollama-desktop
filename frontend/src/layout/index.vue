<template>
  <el-container>
    <el-header height="50px">
      <nav-header />
    </el-header>
    <el-main>
      <!-- <el-scrollbar> -->
      <router-view :key="key" />
      <!-- </el-scrollbar> -->
    </el-main>
  </el-container>
</template>
<script setup>
  import NavHeader from './NavHeader.vue'
  import {
    useRoute
  } from 'vue-router'

  const route = useRoute()
  const key = computed(() => {
    return route.path
  })
</script>
<style lang="scss" scoped>
  .el-container {
    height: 100%;

    .el-header {
      padding: 0;
    }

    .el-main {
      padding: 0;
      height: calc(100vh - 50px);
    }
  }
</style>