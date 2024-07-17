<template>
  <div class="header">
    <div style="margin-left: 10px;">
      <img width="45" height="45" src="/ollama.png" />
      <el-text size="large" style="font-weight: 700;margin-left: 10px;">Ollama Desktop</el-text>
    </div>
    <el-menu :default-active="activeIndex" mode="horizontal" style="margin-left: auto;" :ellipsis="false"
      @select="handleSelect">
      <el-menu-item index="/home">主页</el-menu-item>
      <el-menu-item index="/chat">聊天</el-menu-item>
      <el-menu-item index="/setting">设置</el-menu-item>
      <el-menu-item index="/about">关于</el-menu-item>
    </el-menu>
    <div style="margin-right: 10px;margin-left: 10px;">
      <!-- <el-tooltip effect="dark" :content="isDark? '暗黑模式':'明亮模式'" placement="bottom"> -->
      <el-switch v-model="isDark" style="--el-switch-on-color: #303133; --el-switch-off-color: #606266" inline-prompt
        :active-icon="Moon" :inactive-icon="Sunny" @change="toggleDark" />
      <!-- </el-tooltip> -->
    </div>
  </div>
</template>

<script setup>
  import {
    isDark,
    toggleDark
  } from "~/composables"
  import {
    Sunny,
    Moon
  } from '@element-plus/icons-vue'
  import {
    useRoute,
    useRouter
  } from 'vue-router'

  const route = useRoute()
  const router = useRouter()
  const activeIndex = computed(() => {
    const matched = route.matched || []
    if (matched.length >= 2) {
      return matched[1].path
    }
    return "/home"
  })
  const handleSelect = (key, keyPath) => {
    router.replace(key)
  }
</script>

<style lang="scss" scoped>
  .header {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;

    &>div {
      height: 100%;
      display: flex;
      align-items: center;
      justify-content: center;
    }
  }

  .el-menu {
    border-bottom: none !important;

    .el-menu-item:hover {
      border-bottom: 1px solid var(--el-menu-border-color);
    }
  }
</style>