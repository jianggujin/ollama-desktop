<template>
  <el-container>
    <el-header>
      <nav-header />
    </el-header>
    <el-main>
      <router-view :key="key" />
    </el-main>
    <el-footer>
      <layout-footer />
    </el-footer>
  </el-container>
</template>
<script setup>
import NavHeader from './NavHeader.vue'
import LayoutFooter from './LayoutFooter.vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const key = computed(() => { return route.path })
</script>
<style lang="scss" scoped>
  .el-container {
    height: 100%;

    .el-header {
      padding: 0;
      height: var(--app-layout-header);
    }

    .el-main {
      padding: 0;
      height: var(--app-layout-main);
    }

    .el-footer {
      padding: 0;
      height: var(--app-layout-footer);
    }
  }
</style>
