<template>
  <el-container>
    <el-aside width="200px">
      <el-scrollbar>
        <el-menu :default-openeds="['1']" :default-active="activeIndex" @select="handleSelect">
          <el-sub-menu index="1">
            <template #title>
              <i-ep-grid />模型列表
            </template>
            <el-menu-item index="/home/ps">运行中</el-menu-item>
            <el-menu-item index="/home/local">本地模型</el-menu-item>
            <el-menu-item index="/home/online">模型仓库</el-menu-item>
          </el-sub-menu>
          <el-sub-menu index="2">
            <template #title>
              <i-ep-grid />聊天
            </template>
            <el-menu-item index="/home/ps">运行中</el-menu-item>
            <el-menu-item index="/home/local">本地模型</el-menu-item>
            <el-menu-item index="/home/online">模型仓库</el-menu-item>
          </el-sub-menu>
        </el-menu>
      </el-scrollbar>
    </el-aside>
    <el-main>
      <router-view :key="key" />
    </el-main>
  </el-container>
</template>
<script setup>
  import NavHeader from './NavHeader.vue'
  import {
    useRoute,
    useRouter
  } from 'vue-router'

  const route = useRoute()
  const router = useRouter()
  const key = computed(() => {
    console.log(route.path)
    return route.path
  })
  const activeIndex = computed(() => {
    const matched = route.matched || []
    if (matched.length >= 3) {
      return matched[2].path
    }
    return "/home"
  })
  const handleSelect = (key, keyPath) => {
    router.replace(key)
  }
</script>
<style lang="scss" scoped>
  .el-container {
    height: 100%;

    .el-aside,
    .el-main {
      padding: 0;
      height: calc(100vh - 50px);
    }

    .el-menu {
      min-height: calc(100vh - 50px);
    }
  }
</style>