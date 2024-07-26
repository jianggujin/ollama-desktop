<template>
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
        <el-text style="margin-left: 5px;cursor: pointer;" type="primary" @click="openHomePage">Jianggujin</el-text>
      </div>
    </div>
  </el-scrollbar>
</template>

<script setup>
import { ElMessage } from 'element-plus'
import { AppInfo } from '@/go/app/App.js'
import { BrowserOpenURL } from '@/runtime/runtime.js'
import { runAsync, runQuietly } from '~/utils/wrapper.js'

const appInfo = ref({})
const subTitle = computed(() => {
  return (appInfo.value.Version || '') + ' ' + (appInfo.value.BuildTime || '')
})
onMounted(() => {
  runAsync(AppInfo,
    data => { appInfo.value = data },
    _ => { ElMessage.error('获取应用信息失败') })
})

function openHomePage() {
  runQuietly(() => { BrowserOpenURL('https://www.jianggujin.com') })
}
</script>

<style lang="scss" scoped>
</style>
