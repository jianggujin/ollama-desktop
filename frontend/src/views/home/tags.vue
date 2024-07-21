<template>
  <el-scrollbar>
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
          <el-popconfirm :title="`确定要删除模型(${scope.row.name})?`" @confirm="handleDelete(scope.row)">
            <template #reference>
              <el-button :icon="Delete" size="small" link type="danger"></el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
  </el-scrollbar>
</template>

<script setup>
  import { Refresh, Delete } from '@element-plus/icons-vue'
  import { ElMessage } from 'element-plus'
  import { OllamaList, OllamaDelete } from "@/go/app/App.js"
  import { EventsOn } from "@/runtime/runtime.js"
  import { useOllamaStore } from '~/store/ollama.js'
  import { runAsync } from "~/utils/wrapper.js"
  import { humanize } from "~/utils/humanize.js"

  const ollamaStore = useOllamaStore()
  const list = ref([])

  function handleRefresh() {
    runAsync(OllamaList, ({ models }) => {
      list.value = (models || []).map(item => {
        item.formatModifiedAt = humanize.date('Y-m-d H:i:s',
          new Date(item.modified_at));
        item.formatSize = humanize.filesize(item.size)
        item.parameterSize = item.details?.parameter_size
        item.quantizationLevel = item.details?.quantization_level
        return item
      })
    }, _ => { ElMessage.error('获取本地模型列表失败') })
  }

  function handleDelete(row) {
    runAsync(() => OllamaDelete(row.name), handleRefresh, _ => { ElMessage.error(`删除模型(${row.name})失败`) })
  }

  onMounted(() => {
    handleRefresh()
  })
</script>

<style lang="scss" scoped>
</style>