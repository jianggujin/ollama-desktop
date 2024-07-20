<template>
  <el-scrollbar>
    <div style="display: flex;align-items: center;justify-content: center;margin-top: 15px;">
      <el-result :title="title" :sub-title="subTitle" style="--el-result-extra-margin-top: 10px;">
        <template #icon>
          <img src="/ollama.png" />
        </template>
        <template v-if="!ollamaStore.installed" #extra>
          <el-text style="cursor: pointer;" type="primary" @click="openDownload">点此下载安装</el-text>
        </template>
        <template v-else-if="ollamaStore.canStart" #extra>
          <el-text style="cursor: pointer;" type="primary" @click="startOllamaApp">启动服务</el-text>
        </template>
      </el-result>
    </div>
    <el-descriptions title="环境变量" :column="1" direction="vertical" style="width: 600px;margin: 15px auto;">
      <el-descriptions-item v-for="(item, index) in envs" :key="index">
        <template #label>
          <div style="display: flex;align-items: center;">
            <span style="margin-right: 5px;">{{item.Name}}</span>
            <el-tooltip effect="dark" :content="item.Description" placement="right">
              <i-ep-question-filled />
            </el-tooltip>
          </div>
        </template>
        {{item.Value}}
      </el-descriptions-item>
    </el-descriptions>
  </el-scrollbar>
</template>

<script setup>
  import { ElMessage } from 'element-plus'
  import { BrowserOpenURL } from "@/runtime/runtime.js"
  import { OllamaVersion, OllamaEnvs, StartOllama } from "@/go/app/App.js"
  import { useOllamaStore } from '~/store/ollama.js'
  import { runAsync, runQuietly } from "~/utils/wrapper.js"

  const ollamaStore = useOllamaStore()
  const version = ref("")
  const envs = ref([])

  const title = computed(() => {
    if (version.value) {
      return "Ollama " + version.value
    }
    return "Ollama"
  })

  const subTitle = computed(() => {
    return ollamaStore.installed ? "已安装" : "未安装"
  })

  onMounted(() => {
    runAsync(OllamaVersion, data => { version.value = data }, _ => { ElMessage.error('获取Ollama版本失败') })
    runAsync(OllamaEnvs, data => { envs.value = data }, _ => { ElMessage.error('获取Ollama环境信息失败') })
  })

  function startOllamaApp() {
    runAsync(StartOllama, () => {
        ElMessage.error('启动Ollama服务成功')
        runAsync(OllamaVersion, data => { version.value = data }, _ => { ElMessage.error('获取Ollama版本失败') })
      },
      () => { ElMessage.error('启动Ollama服务失败') })
  }

  function openDownload() {
    runQuietly(() => { BrowserOpenURL("https://ollama.com/download") })
  }
</script>

<style lang="scss" scoped>
</style>