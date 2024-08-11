<template>
  <el-dialog v-model="visible" top="50px" :title="title" width="800">
    <div style="text-align: center;">
      <el-segmented v-model="segmentedValue" :options="segmentedOptions" />
    </div>
    <el-scrollbar ref="scrollbarRef" style="margin-top: 20px;height: calc(100% - 62px);">
      <div v-show="segmentedValue == 'basic'">
        <el-descriptions title="基本信息" :column="2" border>
          <el-descriptions-item label="名称">{{ modelBasic.name }}</el-descriptions-item>
          <el-descriptions-item label="大小">{{ modelBasic.formatSize }}</el-descriptions-item>
          <el-descriptions-item label="参数大小">{{ modelBasic.parameterSize }}</el-descriptions-item>
          <el-descriptions-item label="量化水平">{{ modelBasic.quantizationLevel }}</el-descriptions-item>
          <el-descriptions-item label="格式">{{ modelBasic.format }}</el-descriptions-item>
          <el-descriptions-item label="修改时间">{{ modelBasic.formatModifiedAt }}</el-descriptions-item>
        </el-descriptions>
        <el-descriptions title="模型信息" :column="1" border style="margin-top: 20px;" class="model-info">
          <el-descriptions-item v-for="(value, name) in modelInfo.model_info" :key="name" :label="name">{{ value }}</el-descriptions-item>
        </el-descriptions>
      </div>
      <pre v-show="segmentedValue == 'parameters'" style="margin: 0;white-space: pre-wrap;word-break: break-all;">{{ modelInfo.parameters }}</pre>
      <pre v-show="segmentedValue == 'template'" style="margin: 0;white-space: pre-wrap;word-break: break-all;">{{ modelInfo.template }}</pre>
      <pre v-show="segmentedValue == 'modelfile'" style="margin: 0;white-space: pre-wrap;word-break: break-all;">{{ modelInfo.modelfile }}</pre>
      <pre v-show="segmentedValue == 'license'" style="margin: 0;white-space: pre-wrap;word-break: break-all;">{{ modelInfo.license }}</pre>
    </el-scrollbar>
  </el-dialog>
</template>

<script setup>
import { ElMessage } from 'element-plus'
import { runQuietly } from '~/utils/wrapper.js'
import { Show } from '@/go/app/Ollama.js'

const modelBasic = ref({})
const modelInfo = ref({})
const visible = ref(false)
const scrollbarRef = ref(null)

const segmentedValue = ref('basic')
const segmentedOptions = [{ label: '信息', value: 'basic' },
  { label: '参数', value: 'parameters' },
  { label: '模板', value: 'template' },
  { label: '模型文件', value: 'modelfile' },
  { label: 'License', value: 'license' }]

const title = computed(() => { return `模型(${modelBasic.value.name})信息` })

function showDialog(model) {
  modelBasic.value = { ...model }
  segmentedValue.value = 'basic'
  visible.value = true
  runQuietly(() => Show({ model: model.name }), data => { modelInfo.value = data }, _ => ElMessage.error('获取模型信息失败'))
}

watch(() => segmentedValue.value, _ => nextTick(() => scrollbarRef.value?.setScrollTop(0)))

defineExpose({
  showDialog
})
</script>

<style lang="scss" scoped>
.model-info {
  :deep(.el-descriptions__label) {
    width: 200px;
  }
}
</style>
