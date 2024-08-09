<template>
  <el-dialog v-model="visible" :title="isUpdate ? '修改会话' : '新建会话'" width="800">
    <el-form ref="sessionFormRef"
      :model="sessionFormData"
      :rules="sessionFormRule"
      label-width="120px"
      label-position="left"
      @submit.prevent
      v-loading.body.fullscreen.lock="loading"
      :element-loading-text="loadingOptions.text"
      :element-loading-spinner="loadingOptions.svg"
      :element-loading-svg-view-box="loadingOptions.svgViewBox"
      :element-loading-background="loadingOptions.background">
      <div style="display: flex;gap: 10px;">
        <el-form-item label="会话名称" prop="sessionName" style="flex: 1;">
          <el-input v-model.trim="sessionFormData.sessionName" placeholder="请输入会话名称"/>
        </el-form-item>
        <el-form-item prop="modelName" style="flex: 1;">
          <template #label>
            <div style="display:flex; align-items: center;gap:5px;">
              <span>模型名称</span>
              <el-tooltip effect="dark" content="会话聊天中使用的模型" placement="bottom">
                <i-ep-question-filled style="cursor: pointer;"/>
              </el-tooltip>
            </div>
          </template>
          <el-select v-model="sessionFormData.modelName" placeholder="请选择会话模型" style="width: 100%">
            <el-option v-for="(item, index) in models" :key="index" :label="item.name" :value="item.name"/>
          </el-select>
        </el-form-item>
      </div>
      <div style="display: flex;gap: 10px;">
        <el-form-item prop="messageHistoryCount" style="flex: 1;">
          <template #label>
            <div style="display:flex; align-items: center;gap:5px;">
              <span>历史轮次</span>
              <el-tooltip effect="dark" content="会话聊天中使用的历史会话最大轮次，用于构建聊天消息的上下文信息" placement="bottom">
                <i-ep-question-filled style="cursor: pointer;"/>
              </el-tooltip>
            </div>
          </template>
          <el-input v-model.trim="sessionFormData.messageHistoryCount" placeholder="请输入最大历史轮次"/>
        </el-form-item>
        <el-form-item prop="keepAlive" style="flex: 1;">
          <template #label>
            <div style="display:flex; align-items: center;gap:5px;">
              <span style="margin-left: 10.38px;">存活时间</span>
              <el-tooltip effect="dark" content="模型被加载到内存中后，在内存中保留的时间，例如：5m、2h45m，可用单位有：ns、us、ms、s、m、h" placement="bottom">
                <i-ep-question-filled style="cursor: pointer;"/>
              </el-tooltip>
            </div>
          </template>
          <el-input v-model.trim="sessionFormData.keepAlive" placeholder="请输入存活时间"/>
        </el-form-item>
      </div>
      <el-form-item prop="systemMessage">
        <template #label>
          <div style="display:flex; align-items: center;gap:5px;">
            <span style="margin-left: 10.38px;">系统消息</span>
            <el-tooltip effect="dark" content="设置系统消息后，在与大模型聊天时会自动将其作为system角色的消息插入到第一条聊天信息前" placement="bottom">
              <i-ep-question-filled style="cursor: pointer;"/>
            </el-tooltip>
          </div>
        </template>
        <el-input v-model.trim="sessionFormData.systemMessage" type="textarea" resize="none" placeholder="请输入系统消息" :autosize="{ minRows: 2, maxRows: 6 }"/>
      </el-form-item>
      <div style="display: flex;gap: 10px;">
        <el-form-item prop="optionsSeed" style="flex: 1;">
          <template #label>
            <div style="display:flex; align-items: center;gap:5px;">
              <span style="margin-left: 10.38px;">随机种子</span>
              <el-tooltip effect="dark" content="用于控制生成结果的随机性。如果设定相同的种子，可以得到一致的输出" placement="bottom">
                <i-ep-question-filled style="cursor: pointer;"/>
              </el-tooltip>
            </div>
          </template>
          <el-input v-model.trim="sessionFormData.optionsSeed" placeholder="请输入随机种子"/>
        </el-form-item>
        <el-form-item prop="optionsNumPredict" style="flex: 1;">
          <template #label>
            <div style="display:flex; align-items: center;gap:5px;">
              <span style="margin-left: 10.38px;">令牌数量</span>
              <el-tooltip effect="dark" content="要生成的令牌数量，即模型应预测的最大令牌数" placement="bottom">
                <i-ep-question-filled style="cursor: pointer;"/>
              </el-tooltip>
            </div>
          </template>
          <el-input v-model.trim="sessionFormData.optionsNumPredict" placeholder="请输入令牌数量"/>
        </el-form-item>
      </div>
      <div style="display: flex;gap: 10px;">
        <el-form-item prop="optionsTopK" style="flex: 1;">
          <template #label>
            <div style="display:flex; align-items: center;gap:5px;">
              <span style="margin-left: 10.38px;">TopK</span>
              <el-tooltip effect="dark" content="在预测时，从最高概率的K个令牌中选择下一个令牌。较高的K值增加生成的多样性" placement="bottom">
                <i-ep-question-filled style="cursor: pointer;"/>
              </el-tooltip>
            </div>
          </template>
          <el-input v-model.trim="sessionFormData.optionsTopK" placeholder="请输入TopK"/>
        </el-form-item>
        <el-form-item prop="optionsTopP" style="flex: 1;">
          <template #label>
            <div style="display:flex; align-items: center;gap:5px;">
              <span style="margin-left: 10.38px;">TopP</span>
              <el-tooltip effect="dark" content="使用nucleus sampling（核采样）进行生成，从概率累积超过P的令牌中选择" placement="bottom">
                <i-ep-question-filled style="cursor: pointer;"/>
              </el-tooltip>
            </div>
          </template>
          <el-input v-model.trim="sessionFormData.optionsTopP" placeholder="请输入TopP"/>
        </el-form-item>
      </div>
      <el-form-item prop="optionsNumCtx">
        <template #label>
          <div style="display:flex; align-items: center;gap:5px;">
            <span style="margin-left: 10.38px;">上下文长度</span>
            <el-tooltip effect="dark" content="模型使用的上下文长度，影响模型可以记住的前文长度" placement="bottom">
              <i-ep-question-filled style="cursor: pointer;"/>
            </el-tooltip>
          </div>
        </template>
        <el-input v-model.trim="sessionFormData.optionsNumCtx" placeholder="请输入上下文长度"/>
      </el-form-item>
      <div style="display: flex;gap: 10px;">
        <el-form-item prop="optionsTemperature" style="flex: 1;">
          <template #label>
            <div style="display:flex; align-items: center;gap:5px;">
              <span style="margin-left: 10.38px;">温度</span>
              <el-tooltip effect="dark" content="控制生成文本的随机性。较高的温度值会使生成的输出更加随机，而较低的值则使输出更加确定" placement="bottom">
                <i-ep-question-filled style="cursor: pointer;"/>
              </el-tooltip>
            </div>
          </template>
          <el-input v-model.trim="sessionFormData.optionsTemperature" placeholder="请输入温度"/>
        </el-form-item>
        <el-form-item prop="optionsRepeatPenalty" style="flex: 1;">
          <template #label>
            <div style="display:flex; align-items: center;gap:5px;">
              <span style="margin-left: 10.38px;">惩罚</span>
              <el-tooltip effect="dark" content="为重复的令牌施加惩罚，以减少生成过程中的重复内容" placement="bottom">
                <i-ep-question-filled style="cursor: pointer;"/>
              </el-tooltip>
            </div>
          </template>
          <el-input v-model.trim="sessionFormData.optionsRepeatPenalty" placeholder="请输入惩罚"/>
        </el-form-item>
      </div>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitSession">确认</el-button>
      </div>
    </template>
  </el-dialog>
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
