<template>
  <div id="loading-wrapper" style="height: 100%;">
    <el-scrollbar>
      <div style="margin: 50px auto 0 auto;width: 80%;" v-if="model.name">
        <div class="model-item">
          <div><el-link style="font-weight: 500;font-size: 1.5rem;">{{ model.name }}</el-link></div>
          <div style="margin-top: 10px;"><el-text style="">{{ model.description }}</el-text></div>
          <div style="margin-top: 10px;" v-if="model.tags?.length">
            <el-tag v-for="(tag, ti) in model.tags" :key="ti" :type="tag == 'Embedding' || tag == 'Vision' || tag == 'Tools' || tag == 'Code' ? 'success' : 'primary'">{{ tag }}</el-tag>
          </div>
          <div style="margin-top: 10px;display: flex;align-items: center;">
          <i-ep-download style="color:var(--el-color-info);"/>
          <el-text type="info" style="margin-left: 5px;">{{ model.pullCount }} Pulls</el-text>
            <i-ep-price-tag style="color:var(--el-color-info);margin-left: 10px;"/>
            <el-text type="info" style="margin-left: 5px;">{{ model.tagCount}} Tags</el-text>
            <i-ep-clock style="color:var(--el-color-info);margin-left: 10px;"/>
            <el-text type="info" style="margin-left: 5px;">Updated {{ model.updateTime }}</el-text>
          </div>
        </div>
        <div style="display: flex;align-items: center;margin-top: 20px;">
          <el-select v-model="tag" placeholder="选择标签" style="width: 240px" size="large">
            <el-option v-for="(item, index) in tags" :key="index" :label="item.name" :value="item.name">
              <div style="display: flex;align-items: center;">
                <span>{{ item.name }}</span>
                <el-tag size="small" v-if="item.latest" type="success" style="margin-left: 5px;">latest</el-tag>
                <span style="color: var(--el-text-color-secondary);font-size: 13px;margin-left: auto;">{{ item.size }}</span>
              </div>
            </el-option>
          </el-select>
          <el-input v-show="copyCommand" v-model="copyCommand" style="width: 320px;margin-left: auto;" size="large" readonly>
            <template #append>
              <span class="copy-wrapper" @click="handleCopyCommand">
                <i-ep-copy-document/>
              </span>
            </template>
          </el-input>
          <el-text type="primary" v-show="ollamaStore.started" style="cursor: pointer;margin-left: 10px;" @click="handleDownload">下载</el-text>
        </div>
        <el-table :data="metas" style="width: 100%;margin-top: 20px;" size="small">
          <template #empty><el-empty /></template>
          <el-table-column fixed="left" prop="name" align="center" label="名称" width="90"/>
          <el-table-column prop="content" align="center" label="内容">
            <template #default="scope">
              <div v-if="(scope.row.content || '').replace(/[\s]+/g, ' ').length > 64">
                <span>{{ scope.row.content.replace(/[\s]+/g, ' ').substring(0, 64) }}</span>
              </div>
              <span v-else>{{ (scope.row.content || '').replace(/[\s]+/g, ' ') }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="unit" align="center" label="大小" width="80"/>
        </el-table>
      </div>
      <div style="margin: 20px auto 0 auto;width: 80%;" v-html="readme" class="readme" id="readme"></div>
      <el-image-viewer @close="closeViewer" v-if="showViewer" :url-list="previewSrcList" />
    </el-scrollbar>
  </div>
</template>

<script setup>
import { ModelInfoOnline } from '@/go/app/Ollama.js'
import { Pull } from '@/go/app/DownLoader.js'
import { BrowserOpenURL, ClipboardSetText } from '@/runtime/runtime.js'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { runAsync, runQuietly } from '~/utils/wrapper.js'
import marked from '~/utils/markdown.js'
import { useOllamaStore } from '~/store/ollama.js'

const ollamaStore = useOllamaStore()

const props = defineProps({
  modelTag: String
})

const modelInfo = ref({})

const name = ref('')
const tag = ref('')
const copyCommand = ref('')

const model = computed(() => { return modelInfo.value.model || {} })
const tags = computed(() => { return modelInfo.value.tags || [] })
const metas = computed(() => { return modelInfo.value.metas || {} })
const readme = computed(() => { return marked.parse(modelInfo.value.readme || '') })

const showViewer = ref(false)
const previewSrcList = ref([])
let prevOverflow = ''

const router = useRouter()

let readmeContainer
function handleReadmeClick(event) {
  if (event.target.tagName.toLowerCase() === 'a') {
    event.preventDefault()
    const href = event.target.getAttribute('href')
    runQuietly(() => { BrowserOpenURL(href) })
  } else if (event.target.tagName.toLowerCase() === 'img') {
    event.preventDefault()
    const src = event.target.getAttribute('src')
    previewSrcList.value = [src]
    prevOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    showViewer.value = true
  }
}

watch(() => tag.value, newValue => {
  if (newValue && (name.value + ':' + tag.value) !== props.modelTag) {
    router.replace('/home/library/' + name.value + ':' + tag.value)
  }
})

onMounted(() => {
  if (!props.modelTag) {
    router.replace('/home/library')
    return
  }
  copyCommand.value = 'ollama run ' + props.modelTag
  const index = props.modelTag.lastIndexOf(':')
  if (index > 0) {
    name.value = props.modelTag.substring(0, index)
    tag.value = props.modelTag.substring(index + 1)
  } else {
    name.value = props.modelTag
  }

  readmeContainer = document.getElementById('readme')
  readmeContainer.addEventListener('click', handleReadmeClick)
  runAsync(() => ModelInfoOnline(props.modelTag), data => {
    modelInfo.value = data
    if (!tag.value) {
      tag.value = data?.tags?.find(item => item.latest)?.name || ''
    }
  }, _ => {
    ElMessage.error('获取模型信息失败')
    router.replace('/home/library')
  })
})

onUnmounted(() => {
  readmeContainer.removeEventListener('click', handleReadmeClick)
})

function handleCopyCommand() {
  runQuietly(() => {
    ClipboardSetText(copyCommand.value).then(result => {
      if (result) {
        ElMessage.success('复制命令成功')
      } else {
        ElMessage.error('复制命令失败')
      }
    }).catch(_ => ElMessage.error('复制命令失败'))
  }, null, () => ElMessage.error('复制命令失败'))
}

function handleDownload() {
  runAsync(() => Pull(JSON.stringify({ model: props.modelTag })),
    _ => ElMessage.error('已加入下载队列'),
    _ => ElMessage.error('加入下载队列失败'))
}

function closeViewer() {
  document.body.style.overflow = prevOverflow
  showViewer.value = false
}
</script>

<style lang="scss" scoped>
.model-item {
  .el-tag + .el-tag {
    margin-left: 10px;
  }
}
:deep(.el-input-group__append) {
  padding: 0;
  background-color: transparent!important;
}
.copy-wrapper {
  width: 38px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: var(--el-text-color-secondary);
  &:hover {
    color: var(--el-text-color-primary);
  }
}
.readme {
  :deep(a) {
    color: var(--el-text-color-regular);
    &:hover {
      color: var(--el-text-color-primary);
    }
  }
  :deep(img) {
    max-width: 100%;
    cursor: pointer;
  }
}
</style>
