<template>
  <div
    v-loading="loading"
    :element-loading-text="loadingOptions.text"
    :element-loading-spinner="loadingOptions.svg"
    :element-loading-svg-view-box="loadingOptions.svgViewBox"
    :element-loading-background="loadingOptions.background">
    <el-alert title="自定义Ollama服务端信息，默认为本机" :closable="false" center style="border-radius: 0;margin-bottom: 10px;"/>
    <el-form ref="ollamaFormRef" :model="ollamaFormData" :rules="ollamaFormRule" label-width="100px" label-position="left" @submit.prevent>
      <el-form-item label="协议" prop="scheme">
        <el-select v-model="ollamaFormData.scheme" placeholder="请选择协议" style="width: 100%">
          <el-option v-for="(scheme, index) in schemes" :key="index" :label="scheme" :value="scheme"/>
        </el-select>
      </el-form-item>
      <el-form-item label="主机地址" prop="host">
        <el-input v-model.trim="ollamaFormData.host" placeholder="请输入主机地址"/>
      </el-form-item>
      <el-form-item label="端口" prop="port">
        <el-input v-model.trim="ollamaFormData.port" placeholder="请输入端口"/>
      </el-form-item>
      <el-form-item label-width="0">
        <div style="text-align: center;width: 100%;">
          <el-button type="primary" @click="handleSubmitOllamaConfig">保存</el-button>
          <el-button @click="$refs.ollamaFormRef.resetFields()">重置</el-button>
        </div>
      </el-form-item>
    </el-form>
  </div>
  <!-- <div>Proxy</div> -->
</template>

<script setup>
import { ElMessage } from 'element-plus'
import { runQuietly } from '~/utils/wrapper.js'
import { OllamaConfigs, SaveOllamaConfigs } from '@/go/app/Config.js'
import loadingOptions from '~/utils/loading.js'

const loading = ref(false)

const emptyData = {
  scheme: 'http',
  host: '127.0.0.1',
  port: '11434'
}

const schemes = ['http', 'https']

const ollamaFormRef = ref(null)
const ollamaFormData = ref({ ...emptyData })
const ollamaFormRule = ref({
  scheme: [{ required: true, message: '请输入协议', trigger: 'change' }],
  host: [{ required: true, message: '请选择主机地址', trigger: 'blur' }],
  port: [{ required: true, message: '请输入主机端口', trigger: 'blur' },
    { validator: (rule, value, callback) => {
      value = parseInt(value)
      if (isNaN(value) || value < 0) {
        callback(new Error('主机端口不合法，必须为正整数'))
      } else {
        callback()
      }
    }, trigger: 'blur' }]
})

function handleSubmitOllamaConfig() {
  ollamaFormRef.value?.validate().then(_ => {
    loading.value = true
    runQuietly(() => SaveOllamaConfigs({
      scheme: ollamaFormData.value.scheme,
      host: ollamaFormData.value.host,
      port: ollamaFormData.value.port
    }), _ => ElMessage.success('保存Ollama配置成功'), _ => ElMessage.error('保存Ollama配置失败'), _ => { loading.value = false })
  })
}

onMounted(() => {
  loading.value = true
  runQuietly(OllamaConfigs, data => {
    ollamaFormData.value = { ...emptyData, ...data }
  }, _ => ElMessage.error('获取Ollama配置失败'), _ => {
    nextTick(_ => ollamaFormRef.value?.clearValidate())
    loading.value = false
  })
})
</script>

<style lang="scss" scoped>
</style>
