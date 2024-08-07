<template>
    <el-dialog v-model="visible" title="新建会话" width="500" >
        <el-form ref="sessionFormRef" :model="sessionFormData" :rules="sessionFormRule" label-width="auto" status-icon>
        <el-form-item label="会话名称" prop="sessionName">
            <el-input v-model="sessionFormData.sessionName" />
        </el-form-item>
        <el-form-item label="模型名称" prop="modelName">
            <el-select v-model="sessionFormData.modelName" placeholder="请选择模型" style="width: 100%">
            <el-option v-for="(item, index) in models" :key="index" :label="item.name" :value="item.name"/>
            </el-select>
        </el-form-item>
        <el-form-item label="历史轮次" prop="messageHistoryCount">
            <el-input-number v-model="sessionFormData.messageHistoryCount" :min="0" controls-position="right" :precision="0" style="width: 100%;"/>
        </el-form-item>
        </el-form>
        <template #footer>
        <div class="dialog-footer">
            <el-button @click="visible = false">取消</el-button>
            <el-button type="primary" @click="handleCreateSession">确认</el-button>
        </div>
        </template>
    </el-dialog>
</template>

<script setup>
import { throttle } from 'lodash'
import marked from '~/utils/markdown.js'
import { ElMessage } from 'element-plus'
import { CreateSession } from '@/go/app/Chat.js'
import { runAsync, runQuietly } from '~/utils/wrapper.js'
import { List as listModels } from '@/go/app/Ollama.js'
import { humanize } from '~/utils/humanize.js'

const emptyData = {
  sessionName: '',
  modelName: '',
  messageHistoryCount: 5
}

const emits = defineEmits(['create'])

const models = ref([])
const visible = ref(false)

const sessionFormRef = ref(null)
const sessionFormData = ref({ ...emptyData })
const sessionFormRule = ref({
  sessionName: [
    { required: true, message: '请输入会话名称', trigger: 'blur' },
    { max: 50, message: '会话名称长度不能大于50', trigger: 'blur' }
  ],
  modelName: [{ required: true, message: '请选择会话模型', trigger: 'change' }],
  messageHistoryCount: [{ required: true, message: '请输入历史会话轮次', trigger: 'change' }]
})

function loadModels() {
  // 获取模型信息
  runAsync(listModels, ({ models }) => {
    models.value = (models || []).map(item => {
      item.formatModifiedAt = humanize.date('Y-m-d H:i:s',
        new Date(item.modified_at))
      item.formatSize = humanize.filesize(item.size)
      item.parameterSize = item.details?.parameter_size
      item.quantizationLevel = item.details?.quantization_level
      return item
    })
  }, _ => { ElMessage.error('获取本地模型列表失败') })
}

function handleCreateSession() {
  sessionFormRef.value?.validate().then(_ => {
    runAsync(() => CreateSession(JSON.stringify({
      sessionName: sessionFormData.value.sessionName,
      modelName: sessionFormData.value.modelName,
      prompts: '',
      messageHistoryCount: sessionFormData.value.messageHistoryCount
    })), data => {
      visible.value = false
      emits('create', data)
    }, _ => { ElMessage.error('添加会话失败') })
  })
}

function showDialog() {
  loadModels()
  sessionFormData.value = { ...emptyData }
  visible.value = true
  sessionFormRef.value?.clearValidate()
}

defineExpose({
  showDialog
})
</script>

<style lang="scss" scoped>
</style>
