<template>
  <el-scrollbar
    v-loading="loading"
    :element-loading-text="loadingOptions.text"
    :element-loading-spinner="loadingOptions.svg"
    :element-loading-svg-view-box="loadingOptions.svgViewBox"
    :element-loading-background="loadingOptions.background">
    <div style="margin-top: 15px;">
      <el-button :icon="Refresh" style="margin-left: 15px;" @click="handleRefresh" />
    </div>
    <el-table :data="list" style="width: 100%;margin-top: 15px;">
      <template #empty><el-empty /></template>
      <el-table-column fixed="left" prop="name" align="center" label="名称" min-width="200" />
      <el-table-column prop="formatSize" align="center" label="大小" width="100" />
      <el-table-column prop="parameterSize" align="center" label="参数大小" width="100" />
      <el-table-column prop="quantizationLevel" align="center" label="量化水平" width="100" />
      <el-table-column prop="formatModifiedAt" align="center" label="修改时间" width="180" />
      <el-table-column fixed="right" label="操作" align="center" min-width="80">
        <template #default="scope">
          <el-button :icon="View" size="small" link type="primary" @click="$refs.showModelDialog.showDialog(scope.row)"></el-button>
          <el-popconfirm :title="`确定要删除模型(${scope.row.name})?`" @confirm="handleDelete(scope.row)">
            <template #reference>
              <el-button :icon="Delete" size="small" link type="danger"></el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <show-model-dialog ref="showModelDialog" />
  </el-scrollbar>
</template>

<script setup>
import ShowModelDialog from './show-model-dialog.vue'
import { Refresh, Delete, View } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { List, Delete as deleteOllamaModel } from '@/go/app/Ollama.js'
import { useOllamaStore } from '~/store/ollama.js'
import { runQuietly } from '~/utils/wrapper.js'
import { humanize } from '~/utils/humanize.js'
import { EventsOn, EventsOff } from '@/runtime/runtime.js'
import loadingOptions from '~/utils/loading.js'

const loading = ref(false)

const ollamaStore = useOllamaStore()
const list = ref([])

function handleRefresh() {
  if (!ollamaStore.started) {
    ElMessage.warning('Ollama服务尚未启动')
    return
  }
  loading.value = true
  runQuietly(List, ({ models }) => {
    list.value = (models || []).map(item => {
      item.formatModifiedAt = humanize.date('Y-m-d H:i:s',
        new Date(item.modified_at))
      item.formatSize = humanize.filesize(item.size)
      item.parameterSize = item.details?.parameter_size
      item.quantizationLevel = item.details?.quantization_level
      item.format = item.details?.format
      return item
    })
  }, _ => ElMessage.error('获取本地模型列表失败'), _ => { loading.value = false })
}

function handleDelete(row) {
  if (!ollamaStore.started) {
    ElMessage.warning('Ollama服务尚未启动')
    return
  }
  loading.value = true
  runQuietly(() => deleteOllamaModel({ model: row.name }), handleRefresh, _ => ElMessage.error(`删除模型(${row.name})失败`), _ => { loading.value = false })
}

onMounted(() => {
  handleRefresh()
  runQuietly(() => { EventsOn('model_refresh', handleRefresh) })
})

onUnmounted(() => {
  runQuietly(() => { EventsOff('model_refresh') })
})
</script>

<style lang="scss" scoped>
:deep(.el-dialog) {
  height: calc(100vh - 100px);
  .el-dialog__body {
    height: calc(100% - 32px);
  }
}
</style>
