<template>
  <div class="header" style="--wails-draggable:drag;cursor: move;">
    <div style="margin-left: 10px;">
      <img width="45" height="45" src="/ollama.png" />
      <el-text size="large" style="font-weight: 700;margin-left: 10px;">Ollama Desktop</el-text>
    </div>
    <div style="flex:1"></div>
    <el-menu style="--wails-draggable:no-drag;cursor: default;" :default-active="activeIndex" mode="horizontal" :ellipsis="false" @select="handleSelect">
      <el-menu-item index="/home">主页</el-menu-item>
      <!-- <el-menu-item index="/chat">聊天</el-menu-item> -->
      <!-- <el-menu-item index="/setting">设置</el-menu-item> -->
      <el-menu-item index="/about">关于</el-menu-item>
    </el-menu>
    <div style="--wails-draggable:no-drag;cursor: default;">
      <div class="icon-wrapper" style="width: var(--app-layout-header);">
        <el-switch v-model="isDark" style="--el-switch-on-color: #303133; --el-switch-off-color: #606266" inline-prompt :active-icon="Moon" :inactive-icon="Sunny" @change="toggleDark" />
      </div>
      <!-- <svg-icon icon-class="fullscreen" /> -->
      <div class="icon-wrapper" @click="hanldeMinimise">
        <i-ep-minus />
      </div>
      <div class="icon-wrapper">
        <i-ep-copy-document v-if="isMax" @click="hanldeUnmaximise"/>
        <i-ep-full-screen v-else @click="hanldeMaximise"/>
      </div>
      <div class="icon-wrapper danger" @click="handleQuit">
        <i-ep-close />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ElMessageBox } from 'element-plus'
import { isDark, toggleDark } from '~/composables'
import { Sunny, Moon } from '@element-plus/icons-vue'
import { useRoute, useRouter } from 'vue-router'
import { runQuietly } from '~/utils/wrapper.js'
import { WindowIsMaximised, WindowMinimise, WindowMaximise, WindowUnmaximise, Quit } from '@/runtime/runtime.js'

const route = useRoute()
const router = useRouter()
const isMax = ref(false)

onMounted(() => {
  runQuietly(() => {
    WindowIsMaximised().then(data => {
      isMax.value = data
    })
  })
})

const activeIndex = computed(() => {
  const matched = route.matched || []
  if (matched.length >= 2) {
    return matched[1].path
  }
  return '/home'
})
const handleSelect = (key, keyPath) => { router.replace(key) }

function hanldeMinimise() {
  runQuietly(WindowMinimise)
}

function hanldeMaximise() {
  runQuietly(() => {
    WindowMaximise()
    isMax.value = true
  })
}

function hanldeUnmaximise() {
  runQuietly(() => {
    WindowUnmaximise()
    isMax.value = false
  })
}

function handleQuit() {
  ElMessageBox.confirm('确认要退出Ollama Desktop', '退出', {
    confirmButtonText: '确认',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => { runQuietly(Quit) })
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
    .icon-wrapper {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 40px;
      height: var(--app-layout-header);
      background-color: var(--el-menu-bg-color);
      &:hover {
        cursor: pointer;
        color: var(--el-menu-active-color) !important;
        outline: none;
        background-color: var(--el-menu-hover-bg-color);
      }
      &.danger:hover {
        color: white !important;
        background-color: var(--el-color-danger);
      }
    }
  }

  .el-menu {
    border-bottom: none !important;

    .el-menu-item:hover {
      border-bottom: 1px solid var(--el-menu-border-color);
    }
  }
</style>
