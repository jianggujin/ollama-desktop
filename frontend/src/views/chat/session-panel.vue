<template>
  <div style="display: flex;flex-direction: column;height: 100%;"
    v-loading="loading"
    :element-loading-text="loadingOptions.text"
    :element-loading-spinner="loadingOptions.svg"
    :element-loading-svg-view-box="loadingOptions.svgViewBox"
    :element-loading-background="loadingOptions.background"
  >
    <el-scrollbar>
      <div v-for="(session, index) in sessions" :key="index" :class="{'session-item':true, 'is-active': session.id == sessionId}" @click="sessionId = session.id">
        <div style="display: flex;flex-direction: column;width: calc(100% - 30px);gap: 5px;">
          <div class="line-1" style="font-size: 14px;">{{ session.sessionName }}</div>
          <div class="line-1" style="font-size: 12px;">{{ session.modelName }}</div>
        </div>
        <div style="display: flex;align-items: center;justify-content: center;width: 30px;" @click.stop>
          <!-- <el-popconfirm :title="`确定要删除会话?`" @confirm="handleDeleteSesson(session, index)">
            <template #reference>
              <el-button :icon="Delete" size="large" link type="danger"></el-button>
            </template>
          </el-popconfirm> -->
          <el-dropdown trigger="click" @command="handleMoreCommand(session, $event)">
            <!-- <i-ep-more/> -->
            <i-ep-more-filled class="more"/>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="delete" :icon="Delete">删除</el-dropdown-item>
                <el-dropdown-item command="edit" :icon="Edit">编辑</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>
    </el-scrollbar>
    <div style="display: flex;align-items: center;justify-content: center;margin: 10px 0;">
      <el-button :icon="DocumentAdd" @click="showCreateSession">添加会话</el-button>
    </div>
    <create-sesion-dialog ref="createSesionDialog" @create="handleCreated"/>
  </div>
</template>

<script setup>
import CreateSesionDialog from './create-sesion-dialog.vue'
import { DocumentAdd, Delete, Edit } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Sessions, DeleteSession } from '@/go/app/Chat.js'
import { runQuietly } from '~/utils/wrapper.js'
import loadingOptions from '~/utils/loading.js'

const emits = defineEmits(['change'])

const loading = ref(false)

const sessions = ref([{ id: 'dd', sessionName: '测试', modelName: 'qwen2:0.5b' }, { id: 'dsd', sessionName: '测试', modelName: 'qwen2:0.5b' }])
const sessionId = ref('dd')

const createSesionDialog = ref(null)

function loadSessions() {
  loading.value = true
  runQuietly(Sessions, data => {
    sessions.value = data
    if (data.length) {
      sessionId.value = data.find(item => item.id === sessionId.value)?.id || data[0].id
    } else {
      showCreateSession()
    }
  }, _ => { ElMessage.error('获取会话列表失败') }, () => { loading.value = false })
}

onMounted(loadSessions)

function showCreateSession() {
  createSesionDialog.value.showDialog()
}

function handleCreated(session) {
  sessions.value.push(session)
  sessionId.value = session.id
}

function handleMoreCommand(session, command) {
  console.log(command)
}

function handleDeleteSesson(session, index) {
  loading.value = true
  runQuietly(() => DeleteSession(session.id), _ => {
    // 删除选中的会话，需要切换会话
    if (sessionId.value === session.id) {
      // 下面有数据
      if (index < sessions.value.length - 1) {
        sessionId.value = sessions.value[index + 1].id
      } else if (index > 0) { // 上面有数据
        sessionId.value = sessions.value[index - 1].id
      } else {
        sessionId.value = ''
      }
      return
    }
    sessions.value.splice(index, 1)
  }, _ => { ElMessage.error('获取会话列表失败') }, () => { loading.value = false })
}

watch(() => sessionId.value, newValue => emits('change', newValue), { immediate: true })

</script>

<style lang="scss" scoped>
.session-item {
  width: calc(100% - 20px);
  padding: 10px;
  display: flex;
  cursor: pointer;
  align-items: center;
  & + .session-item {
    border-top: 1px solid var(--el-border-color);
  }
  .el-dropdown {
    display: none;
  }
  &:hover {
    background-color: var(--el-menu-hover-bg-color);
  }
  &.is-active {
    background-color: var(--el-menu-hover-bg-color);
    color: var(--el-menu-active-color);
    & .el-dropdown {
      display: inline-flex;
    }
  }
  .more:hover {
    color: var(--el-menu-active-color);
  }
}
</style>
