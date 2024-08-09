<template>
  <div>Ollama</div>
</template>

<script setup>
import { ElMessage } from 'element-plus'
import { CreateSession, UpdateSession } from '@/go/app/Chat.js'
import { runQuietly } from '~/utils/wrapper.js'
import { List as listModels } from '@/go/app/Ollama.js'
import { humanize } from '~/utils/humanize.js'
import loadingOptions from '~/utils/loading.js'

const emptyData = {
  sessionName: '',
  modelName: '',
  messageHistoryCount: 5,
  keepAlive: '',
  systemMessage: '',

  optionsSeed: '',
  optionsNumPredict: '',
  optionsTopK: '',
  optionsTopP: '',
  optionsNumCtx: '',
  optionsTemperature: '',
  optionsRepeatPenalty: ''
}

const loading = ref(false)

const emits = defineEmits(['create', 'update'])

const models = ref([])
const visible = ref(false)

const sessionFormRef = ref(null)
const sessionFormData = ref({ ...emptyData })
const sessionFormRule = ref({
  sessionName: [{ required: true, message: '请输入会话名称', trigger: 'blur' },
    { max: 50, message: '会话名称长度不能大于50', trigger: 'blur' }],
  modelName: [{ required: true, message: '请选择会话模型', trigger: 'change' }],
  messageHistoryCount: [{ required: true, message: '请输入历史会话轮次', trigger: 'change' },
    { validator: (rule, value, callback) => {
      value = parseInt(value)
      if (isNaN(value) || value < 0) {
        callback(new Error('历史会话轮次不合法，必须为正整数或0'))
      } else {
        callback()
      }
    }, trigger: 'blur' }],
  keepAlive: [{ pattern: /^(([1-9][0-9]*)(ns|us|ms|s|m|h))+$/, message: '存活时间不合法', trigger: 'blur' }],
  optionsSeed: [{ validator: (rule, value, callback) => {
    if (value) {
      value = parseInt(value)
      if (isNaN(value) || value <= 0) {
        callback(new Error('随机种子不合法，必须为正整数'))
        return
      }
    }
    callback()
  }, trigger: 'blur' }],
  optionsNumPredict: [{ validator: (rule, value, callback) => {
    if (value) {
      value = parseInt(value)
      if (isNaN(value) || value <= 0) {
        callback(new Error('令牌数量不合法，必须为正整数'))
        return
      }
    }
    callback()
  }, trigger: 'blur' }],
  optionsTopK: [{ validator: (rule, value, callback) => {
    if (value) {
      value = parseInt(value)
      if (isNaN(value) || value <= 0) {
        callback(new Error('TopK不合法，必须为正整数'))
        return
      }
    }
    callback()
  }, trigger: 'blur' }],
  optionsTopP: [{ validator: (rule, value, callback) => {
    if (value) {
      value = parseFloat(value)
      if (isNaN(value) || value <= 0) {
        callback(new Error('TopP不合法，必须为正数'))
        return
      }
    }
    callback()
  }, trigger: 'blur' }],
  optionsNumCtx: [{ validator: (rule, value, callback) => {
    if (value) {
      value = parseInt(value)
      if (isNaN(value) || value <= 0) {
        callback(new Error('上下文长度不合法，必须为正整数'))
        return
      }
    }
    callback()
  }, trigger: 'blur' }],
  optionsTemperature: [{ validator: (rule, value, callback) => {
    if (value) {
      value = parseFloat(value)
      if (isNaN(value) || value <= 0) {
        callback(new Error('温度不合法，必须为正数'))
        return
      }
    }
    callback()
  }, trigger: 'blur' }],
  optionsRepeatPenalty: [{ validator: (rule, value, callback) => {
    if (value) {
      value = parseFloat(value)
      if (isNaN(value) || value <= 0) {
        callback(new Error('惩罚不合法，必须为正数'))
        return
      }
    }
    callback()
  }, trigger: 'blur' }]
})

const isUpdate = computed(() => !!sessionFormData.value.id)

function loadModels() {
  loading.value = true
  // 获取模型信息
  runQuietly(listModels, data => {
    models.value = (data.models || []).map(item => {
      item.formatModifiedAt = humanize.date('Y-m-d H:i:s',
        new Date(item.modified_at))
      item.formatSize = humanize.filesize(item.size)
      item.parameterSize = item.details?.parameter_size
      item.quantizationLevel = item.details?.quantization_level
      return item
    })

    sessionFormRef.value?.clearValidate()
  }, _ => { ElMessage.error('获取本地模型列表失败') }, () => { loading.value = false })
}

function handleSubmitSession() {
  sessionFormRef.value?.validate().then(_ => {
    loading.value = true
    const fn = isUpdate.value ? UpdateSession : CreateSession

    const formData = {
      id: sessionFormData.value.id || '',
      sessionName: sessionFormData.value.sessionName,
      modelName: sessionFormData.value.modelName,
      messageHistoryCount: parseInt(sessionFormData.value.messageHistoryCount),
      keepAlive: sessionFormData.value.keepAlive,
      systemMessage: sessionFormData.value.systemMessage,
      options: JSON.stringify({
        seed: sessionFormData.value.optionsSeed,
        numPredict: sessionFormData.value.optionsNumPredict,
        topK: sessionFormData.value.optionsTopK,
        topP: sessionFormData.value.optionsTopP,
        numCtx: sessionFormData.value.optionsNumCtx,
        temperature: sessionFormData.value.optionsTemperature,
        repeatPenalty: sessionFormData.value.optionsRepeatPenalty
      })
    }
    runQuietly(() => fn(formData), data => {
      visible.value = false
      emits(isUpdate.value ? 'update' : 'create', data)
    }, _ => { ElMessage.error((isUpdate.value ? '修改' : '新建') + '会话失败') }, _ => { loading.value = false })
  })
}

function showDialog(session) {
  loadModels()
  sessionFormData.value = { ...emptyData, ...session }
  visible.value = true
  sessionFormRef.value?.clearValidate()
}

defineExpose({
  showDialog
})
</script>

<style lang="scss" scoped>
</style>
