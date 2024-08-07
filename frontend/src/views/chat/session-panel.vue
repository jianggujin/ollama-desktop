<template>
  <div style="display: flex;flex-direction: column;height: 100%;">
    <el-scrollbar>
      <div v-for="(session, index) in sessions" :key="index" :class="{'session-item':true, 'is-active': session.id == sessionId}" @click="sessionId = session.id">
        <div style="display: flex;flex-direction: column;width: calc(100% - 30px);">
          <div class="line-1" style="font-size: 1.1rem;">{{ session.sessionName }}</div>
          <div class="line-1" style="text-align: right;margin-top: 5px;">{{ session.modelName }}</div>
        </div>
        <div style="display: flex;align-items: center;justify-content: center;width: 30px;">
          <el-popconfirm :title="`确定要删除会话?`" @confirm="handleDeleteSesson(session, index)">
            <template #reference>
              <el-button :icon="Delete" size="large" link type="danger"></el-button>
            </template>
          </el-popconfirm>
        </div>
      </div>
    </el-scrollbar>
    <div style="display: flex;align-items: center;justify-content: center;margin: 10px 0;">
      <el-button :icon="DocumentAdd" @click="showCreateSession">添加会话</el-button>
    </div>
    <create-sesion-dialog ref="createSesionDialog"/>
  </div>
</template>

<script setup>
import CreateSesionDialog from './create-sesion-dialog.vue'
import { DocumentAdd, Delete } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { Sessions, DeleteSession } from '@/go/app/Chat.js'
import { runAsync } from '~/utils/wrapper.js'

const emits = defineEmits(['change'])

const sessions = ref([])
const sessionId = ref('')
const messages = ref([])

const createSesionDialog = ref(null)

function loadSessions() {
  runAsync(Sessions, data => {
    sessions.value = data
    if (data.length) {
      sessionId.value = data.find(item => item.id === sessionId.value)?.id || data[0].id
    } else {
      showCreateSession()
    }
  }, _ => { ElMessage.error('获取会话列表失败') })
}

onMounted(loadSessions)

function showCreateSession() {
  createSesionDialog.value.showDialog()
}

function handleDeleteSesson(session, index) {
  runAsync(() => DeleteSession(session.id), _ => {
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
  }, _ => { ElMessage.error('获取会话列表失败') })
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
  &:hover {
    background-color: var(--el-menu-hover-bg-color);
  }
  &.is-active {
    background-color: var(--el-menu-hover-bg-color);
    color: var(--el-menu-active-color);
  }
}
</style>
