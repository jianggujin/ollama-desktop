<template>
  <div>
    <div style="text-align: center;margin-top: 50px;">
      <el-segmented v-model="segmentedValue" :options="segmentedOptions" />
    </div>
    <el-scrollbar>
      <div style="width: 500px;margin: 10px auto;">
        <component :is="componentValue"/>
      </div>
    </el-scrollbar>
  </div>
</template>

<script setup>
import OllamaPanel from './ollama-panel.vue'
import ProxyPanel from './proxy-panel.vue'

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
