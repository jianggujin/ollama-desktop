<template>
  <el-scrollbar>
    <el-table :data="list" style="width: 100%">
      <el-empty #empty />
      <el-table-column prop="name" align="center" label="名称" />
      <el-table-column prop="formatSize" align="center" label="大小" width="100" />
      <el-table-column prop="parameterSize" align="center" label="参数大小" width="100" />
      <el-table-column prop="quantizationLevel" align="center" label="量化水平" width="100" />
      <el-table-column prop="formatModifiedAt" align="center" label="修改时间" width="180" />
    </el-table>
  </el-scrollbar>
</template>

<script setup>
  import { ElMessage } from 'element-plus'
  import { OllamaList } from "@/go/app/App.js"
  import { useOllamaStore } from '~/store/ollama.js'
  import { runAsync } from "~/utils/wrapper.js"
  import { humanize } from "~/utils/humanize.js"

  const ollamaStore = useOllamaStore()
  const list = ref([])

  onMounted(() => {
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
  })
</script>

<style lang="scss" scoped>
</style>