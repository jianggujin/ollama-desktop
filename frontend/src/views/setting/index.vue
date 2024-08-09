<template>
  <div
    style="height: 100%;"
    v-loading="loading"
    :element-loading-text="loadingOptions.text"
    :element-loading-spinner="loadingOptions.svg"
    :element-loading-svg-view-box="loadingOptions.svgViewBox"
    :element-loading-background="loadingOptions.background"
    >
    <el-scrollbar>
      <div style="text-align: center;margin-top: 50px;">
        <el-segmented v-model="segmentedValue" :options="segmentedOptions" />
      </div>
      <div style="width: 800px;margin: 20px auto;">
        <component :is="componentValue"/>
      </div>
    </el-scrollbar>
  </div>
</template>

<script setup>
import OllamaPanel from './ollama-panel.vue'
import ProxyPanel from './proxy-panel.vue'
import { ElMessage } from 'element-plus'
import { AppInfo } from '@/go/app/App.js'
import { BrowserOpenURL } from '@/runtime/runtime.js'
import { runQuietly } from '~/utils/wrapper.js'
import loadingOptions from '~/utils/loading.js'

const loading = ref(false)

const segmentedValue = ref('ollama')

const segmentedOptions = [{ label: 'Ollama', value: 'ollama' }, { label: '代理', value: 'proxy' }]

const componentValue = computed(() => {
  if (segmentedValue.value === 'ollama') {
    return OllamaPanel
  }
  if (segmentedValue.value === 'proxy') {
    return ProxyPanel
  }
  return 'el-empty'
})

</script>

<style lang="scss" scoped>
</style>
