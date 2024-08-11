<template>
  <div
    v-loading="loading"
    :element-loading-text="loadingOptions.text"
    :element-loading-spinner="loadingOptions.svg"
    :element-loading-svg-view-box="loadingOptions.svgViewBox"
    :element-loading-background="loadingOptions.background">
    <el-alert title="自定义请求在线模型时使用的网络代理信息" :closable="false" center style="border-radius: 0;margin-bottom: 10px;"/>
    <el-form ref="proxyFormRef" :model="proxyFormData" :rules="proxyFormRule" label-width="100px" label-position="left" @submit.prevent>
      <el-form-item label="协议" prop="scheme">
        <el-select v-model="proxyFormData.scheme" placeholder="请选择协议" style="width: 100%">
          <el-option v-for="(scheme, index) in schemes" :key="index" :label="scheme" :value="scheme"/>
        </el-select>
      </el-form-item>
      <el-form-item label="主机地址" prop="host">
        <el-input v-model.trim="proxyFormData.host" placeholder="请输入主机地址"/>
      </el-form-item>
      <el-form-item label="端口" prop="port">
        <el-input v-model.trim="proxyFormData.port" placeholder="请输入端口"/>
      </el-form-item>
      <el-form-item label="用户名" prop="username">
        <template #label>
          <span style="margin-left: 10.38px;">用户名</span>
        </template>
        <el-input v-model.trim="proxyFormData.username" placeholder="请输入用户名"/>
      </el-form-item>
      <el-form-item label="密码" prop="password">
        <template #label>
          <span style="margin-left: 10.38px;">密码</span>
        </template>
        <el-input v-model.trim="proxyFormData.password" placeholder="请输入密码"/>
      </el-form-item>
      <el-form-item label-width="0">
        <div style="text-align: center;width: 100%;">
          <el-button type="primary" @click="handleSubmitProxyConfig">保存</el-button>
          <el-button @click="$refs.proxyFormRef.resetFields()">重置</el-button>
        </div>
      </el-form-item>
    </el-form>
  </div>
  <!-- <div>Proxy</div> -->
</template>

<script setup>
import { ElMessage } from 'element-plus'
import { runQuietly } from '~/utils/wrapper.js'
import { ProxyConfigs, SaveProxyConfigs } from '@/go/app/Config.js'
import loadingOptions from '~/utils/loading.js'

const loading = ref(false)

const emptyData = {
  scheme: 'http',
  host: '',
  port: '80',
  username: '',
  password: ''
}

const schemes = ['http', 'https', 'socks4', 'socks5']

const proxyFormRef = ref(null)
const proxyFormData = ref({ ...emptyData })
const proxyFormRule = ref({
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

function handleSubmitProxyConfig() {
  proxyFormRef.value?.validate().then(_ => {
    loading.value = true
    runQuietly(() => SaveProxyConfigs({
      scheme: proxyFormData.value.scheme,
      host: proxyFormData.value.host,
      port: proxyFormData.value.port
    }), _ => ElMessage.success('保存代理配置成功'), _ => ElMessage.error('保存代理配置失败'), _ => { loading.value = false })
  })
}

onMounted(() => {
  loading.value = true
  runQuietly(ProxyConfigs, data => {
    proxyFormData.value = { ...emptyData, ...data }
  }, _ => ElMessage.error('获取代理配置失败'), _ => {
    nextTick(_ => proxyFormRef.value?.clearValidate())
    loading.value = false
  })
})
</script>

<style lang="scss" scoped>
</style>
