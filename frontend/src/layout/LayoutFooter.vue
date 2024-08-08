<template>
  <div
    v-loading.body.fullscreen.lock="loading"
    :element-loading-text="loadingOptions.text"
    :element-loading-spinner="loadingOptions.svg"
    :element-loading-svg-view-box="loadingOptions.svgViewBox"
    :element-loading-background="loadingOptions.background"
  >
    <div class="footer">
      <el-text style="margin-left: 10px;margin-right: 5px;">Ollama</el-text>
      <i-ep-circle-check-filled v-if="ollamaStore.started" style="color: var(--el-color-success);font-size: var(--el-font-size-base);" />
      <i-ep-circle-close-filled v-else style="color: var(--el-color-warning);font-size: var(--el-font-size-base);" />
      <el-text v-if="ollamaStore.canStart" style="margin-left: 5px;cursor: pointer;" type="primary" @click="startOllamaApp">启动服务</el-text>
      <template v-if="downloaderStore.list?.length">
        <el-text style="margin-left: 10px;">当前有</el-text>
        <el-text style="margin-left: 3px;margin-right: 3px;cursor: pointer;" type="primary" @click="drawer = true">{{ downloaderStore.list?.length || 0 }}</el-text>
        <el-text>个模型正在下载</el-text>
      </template>
      <el-text style="margin-left: auto;">Ollama Desktop Pwered By</el-text>
      <el-text style="margin-left: 5px;margin-right: 10px;cursor: pointer;" type="primary" @click="openHomePage">Jianggujin</el-text>
    </div>
    <el-drawer v-model="drawer" title="下载进度" :size="500">
      <el-scrollbar>
        <template v-if="downloaderStore.list?.length">
          <div class="download-item" v-for="(item, index) in downloaderStore.list" :key="index">
            <div style="display: flex;align-items: center;">
              <div class="line-1" style="font-size: 1.1rem;width: calc(100% - 30px);">{{ item.model }}</div>
              <div style="display: flex;align-items: center;justify-content: center;width: 30px;">
                <el-popconfirm :title="`确定要取消下载?`" @confirm="handleDeleteDownload(item)">
                  <template #reference>
                    <el-button :icon="Delete" size="large" link type="danger"></el-button>
                  </template>
                </el-popconfirm>
              </div>
            </div>
            <el-progress v-for="(bar, bi) in item.bars"
              :key="bi"
              :percentage="bar.percentage"
              :status="bar.status"
              text-inside
              :stroke-width="20"
              :striped="bar.status != 'success'"
              :striped-flow="bar.status != 'success'"
              :duration="20">
              <template #default="{ percentage }">
                <div style="width: 100%;display: flex;align-items: center;">
                  <span style="margin-left: 10px;">{{ bar.name }}</span>
                  <span style="margin-left: auto;margin-right: 10px;">{{ percentage.toFixed(2) }}%</span>
                </div>
              </template>
            </el-progress>
          </div>
        </template>
        <el-empty v-else />
      </el-scrollbar>
    </el-drawer>
  </div>
</template>

<script setup>
import { Delete } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { ElNotification } from 'element-plus'
import { onUnmounted, ref } from 'vue'
import { BrowserOpenURL, EventsOn, EventsOff } from '@/runtime/runtime.js'
import { Heartbeat, Start } from '@/go/app/Ollama.js'
import { runQuietly } from '~/utils/wrapper.js'
import loadingOptions from '~/utils/loading.js'
import { useOllamaStore } from '~/store/ollama.js'
import { useDownloaderStore } from '~/store/downloader.js'
import { Cancel } from '@/go/app/DownLoader.js'

const loading = ref(false)

const ollamaStore = useOllamaStore()
const downloaderStore = useDownloaderStore()
const drawer = ref(false)

let autoStarted = false

onMounted(() => {
  runQuietly(Heartbeat)
  runQuietly(() => {
    EventsOn('ollamaHeartbeat', (installed, started, canStart) => {
      ollamaStore.installed = installed
      ollamaStore.started = started
      ollamaStore.canStart = canStart
      if (canStart && !autoStarted) {
        autoStarted = true
        startOllamaApp()
      }
    })
  })
  runQuietly(() => { EventsOn('pull_list', list => { downloaderStore.list = list }) })
  runQuietly(() => {
    EventsOn('pull_success', item => {
      ElNotification({
        title: '成功',
        message: `模型${item.model}下载成功`,
        type: 'success'
      })
    })
  })
  runQuietly(() => {
    EventsOn('pull_error', item => {
      ElNotification({
        title: '错误',
        message: `模型${item.model}下载失败`,
        type: 'error'
      })
    })
  })
})

onUnmounted(() => {
  runQuietly(() => { EventsOff('ollamaHeartbeat') })
  runQuietly(() => { EventsOff('pull_list') })
  runQuietly(() => { EventsOff('pull_success') })
  runQuietly(() => { EventsOff('pull_error') })
})

function startOllamaApp() {
  loading.value = true
  runQuietly(Start, _ => ElMessage.success('启动Ollama服务成功'),
    _ => ElMessage.error('启动Ollama服务失败'),
    _ => { loading.value = false })
}

function openHomePage() {
  runQuietly(() => { BrowserOpenURL('https://www.jianggujin.com') })
}

function handleDeleteDownload(item) {
  loading.value = true
  runQuietly(() => Cancel(item.model), _ => ElMessage.success(`取消模型${item.model}下载成功`),
    _ => ElMessage.error(`取消模型${item.model}下载失败`), _ => { loading.value = false })
}
</script>

<style lang="scss" scoped>
.footer {
  height: var(--app-layout-footer);
  display: flex;
  align-items: center;
}
:deep(.el-drawer__header) {
  margin-bottom: 20px;
}
:deep(.el-drawer__body) {
  border-top: var(--el-border);
  padding: 0;
}

.download-item {
  width: calc(100% - 20px);
  padding: 10px;
  // display: flex;
  cursor: pointer;
  // align-items: center;
  & + .download-item {
    border-top: 1px solid var(--el-border-color);
  }
  &:hover {
    background-color: var(--el-menu-hover-bg-color);
  }
  .el-progress {
    margin-top: 10px;
  }
  :deep(.el-progress-bar__innerText) {
    width: 100%;
  }
}
</style>
