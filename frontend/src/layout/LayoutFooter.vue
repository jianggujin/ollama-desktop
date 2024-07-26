<template>
  <div class="footer">
    <el-text style="margin-left: 10px;margin-right: 5px;">Ollama</el-text>
    <i-ep-circle-check-filled v-if="ollamaStore.started"
      style="color: var(--el-color-success);font-size: var(--el-font-size-base);" />
    <i-ep-circle-close-filled v-else style="color: var(--el-color-warning);font-size: var(--el-font-size-base);" />
    <el-text v-if="ollamaStore.canStart" style="margin-left: 5px;cursor: pointer;" type="primary" @click="startOllamaApp">启动服务</el-text>
    <el-text style="margin-left: auto;">Ollama Desktop Pwered By</el-text>
    <el-text style="margin-left: 5px;margin-right: 10px;cursor: pointer;" type="primary" @click="openHomePage">Jianggujin</el-text>
  </div>
</template>

<script setup>
import { ElMessage } from 'element-plus'
import { onUnmounted, ref } from 'vue'
import { BrowserOpenURL, EventsOn, EventsOff } from '@/runtime/runtime.js'
import { OllamaHeartbeat, StartOllama } from '@/go/app/App.js'
import { runAsync, runQuietly } from '~/utils/wrapper.js'
import { useOllamaStore } from '~/store/ollama.js'

const ollamaStore = useOllamaStore()
onMounted(() => {
  runQuietly(OllamaHeartbeat)
  runQuietly(() => {
    EventsOn('ollamaHeartbeat', (installed, started, canStart) => {
      ollamaStore.installed = installed
      ollamaStore.started = started
      ollamaStore.canStart = canStart
    })
  })
})

onUnmounted(() => {
  runQuietly(() => { EventsOff('ollamaHeartbeat') })
})

function startOllamaApp() {
  runAsync(StartOllama, () => { ElMessage.success('启动Ollama服务成功') },
    () => { ElMessage.error('启动Ollama服务失败') })
}

function openHomePage() {
  runQuietly(() => { BrowserOpenURL('https://www.jianggujin.com') })
}
</script>

<style lang="scss" scoped>
  .footer {
    height: var(--app-layout-footer);
    display: flex;
    align-items: center;
  }
</style>
