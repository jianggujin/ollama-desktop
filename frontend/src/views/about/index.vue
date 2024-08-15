<template>
  <div
    style="height: 100%;"
    v-loading.lock="loading"
    :element-loading-text="loadingOptions.text"
    :element-loading-spinner="loadingOptions.svg"
    :element-loading-svg-view-box="loadingOptions.svgViewBox"
    :element-loading-background="loadingOptions.background"
  >
    <el-scrollbar>
      <div style="display: flex;align-items: center;justify-content: center;margin-top: 15px;flex-direction: column;">
        <el-result title="Ollama DeskTop" :sub-title="subTitle" style="--el-result-extra-margin-top: 10px;">
          <template #icon>
            <img src="/ollama.png" />
          </template>
          <template #extra>
            <el-text>{{appInfo.Platform || ""}} {{appInfo.Arch || ""}}</el-text>
          </template>
        </el-result>
        <div>
          <el-text>Pwered By</el-text>
          <el-text style="margin-left: 5px;cursor: pointer;" type="primary" @click="openBrowser('https://www.jianggujin.com')">Jianggujin</el-text>
        </div>
        <div style="margin-top: 20px;display: flex;gap: 10px;align-items: center;justify-content: center;">
          <el-text style="cursor: pointer;" type="primary" @click="openBrowser('https://github.com/jianggujin/ollama-desktop')">项目主页</el-text>
          <el-text>|</el-text>
          <el-text style="cursor: pointer;" type="primary" @click="openBrowser('https://www.jianggujin.com')">个人主页</el-text>
        </div>
      </div>
    </el-scrollbar>
  </div>
</template>

<script setup>
import { ElMessage } from 'element-plus'
import { AppInfo } from '@/go/app/App.js'
import { BrowserOpenURL } from '@/runtime/runtime.js'
import { runQuietly } from '~/utils/wrapper.js'
import loadingOptions from '~/utils/loading.js'

const loading = ref(false)

const appInfo = ref({})
const subTitle = computed(() => {
  return (appInfo.value.Version || '') + ' ' + (appInfo.value.BuildShortHash || '')
})

onMounted(() => {
  loading.value = true
  runQuietly(AppInfo,
    data => { appInfo.value = data },
    _ => { ElMessage.error('获取应用信息失败') }, () => { loading.value = false })
})

function openBrowser(url) {
  runQuietly(() => BrowserOpenURL(url))
}
</script>

<style lang="scss" scoped>
</style>
