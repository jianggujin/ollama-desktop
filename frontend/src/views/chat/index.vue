<template>
  <el-container id="loading-wrapper">
    <el-aside width="250px" style="display: flex;flex-direction: column;">
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
        <el-button :icon="DocumentAdd" @click="handleCreateSession">添加会话</el-button>
      </div>
    </el-aside>
    <el-main style="display: flex;flex-direction: column;">
      <el-scrollbar style="flex: 1;" ref="chatScrollbar">
        <div ref="chatContent" style="margin: 10px auto 0 auto;width: 80%;position: relative;display: flex;flex-direction: column;gap: 16px;">
          <div v-for="(message, index) in messages" :key="index" :class="{question: message.role == 'user', answer: message.role == 'assistant'}">
            <template v-if="message.role == 'user'">
              <pre class="message">{{ content }}</pre>
              <i-ep-avatar class="avatar"/>
            </template>
            <template v-if="message.role == 'assistant'">
              <svg-icon icon-class="ollama" class-name="avatar"/>
              <div class="message" v-html="marked.parse(content)"></div>
            </template>
          </div>
        </div>
        <el-image-viewer @close="closeViewer" v-if="showViewer" :url-list="previewSrcList" />
      </el-scrollbar>
      <div style="margin: 10px auto 10px auto;width: 80%;position: relative;">
        <template v-if="sessionId">
          <el-input
            v-model="question"
            show-word-limit
            type="textarea"
            resize="none"
            placeholder="请输入问题"
            :autosize="{ minRows: 2, maxRows: 6 }"
            @keydown="handleQuestionKeydown"
          />
          <el-button :disabled="!canSendQuestion"
            :icon="answering ? Loading : Promotion"
            circle
            plain
            style="position: absolute;z-index: 1;right: 10px;bottom: 10px;"
            type="primary"/>
        </template>
      </div>
    </el-main>
  </el-container>
</template>

<script setup>
import { Promotion, Loading, DocumentAdd, Delete } from '@element-plus/icons-vue'
import { throttle } from 'lodash'
import marked from '~/utils/markdown.js'
import { ElMessage } from 'element-plus'
import { BrowserOpenURL } from '@/runtime/runtime.js'
import { runAsync, runQuietly } from '~/utils/wrapper.js'

// type SessionModel struct {
// 	Id                  string        `json:"id"`
// 	SessionName         string        `json:"sessionName"`
// 	ModelName           string        `json:"modelName"`
// 	Prompts             string        `json:"prompts,omitempty"`
// 	MessageHistoryCount int           `json:"messageHistoryCount"`
// 	UseStream           bool          `json:"stream,omitempty"`
// 	ResponseFormat      string        `json:"responseFormat,omitempty"`
// 	KeepAlive           time.Duration `json:"keepAlive,omitempty"`
// 	Options             string        `json:"options,omitempty"`
// 	SessionType         string        `json:"sessionType"`
// 	CreatedAt           time.Time     `json:"createdAt"`
// 	UpdatedAt           time.Time     `json:"updatedAt"`
// }
const sessions = ref([])
const sessionId = ref('')
const messages = ref([])

const question = ref('')
const answering = ref(false)
const chatScrollbar = ref(null)
const chatContent = ref(null)
const showViewer = ref(false)
const previewSrcList = ref([])
let prevOverflow = ''

let chatContentMacContent = -1

let chatContainer
function handleChatClick(event) {
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

const scrollToBottom = throttle(() => {
  nextTick(() => {
    const chatScrollbarHeight = (chatScrollbar.value?.$el || chatScrollbar.value)?.clientHeight || 0
    const chatContentHeight = (chatContent.value?.$el || chatContent.value)?.clientHeight || 0
    if (chatContentHeight > chatContentMacContent) {
      chatContentMacContent = chatContentHeight
      chatScrollbar.value?.setScrollTop(chatContentHeight - chatScrollbarHeight)
    }
  })
}, 500)

function forceScrollToBottom() {
  chatContentMacContent = -1
  scrollToBottom()
}

function isAllWhitespace(str) {
  return /^\s*$/.test(str)
}

const answer = ref('')
const canSendQuestion = computed(() => !answering.value && !isAllWhitespace(question.value))

function handleQuestionKeydown(event) {
  if (event.altKey && event.key === 'Enter') {
    event.preventDefault()
    sendQuestion()
  }
}

function sendQuestion() {
  if (!canSendQuestion.value) {
    return
  }
  answering.value = true
  question.value = ''
  // let pos = 0
  // let body = ''
  // const id = setInterval(() => {
  //   if (pos < length) {
  //     body += readme.charAt(pos)
  //     answer.value = marked.parse(body + '_')
  //     scrollToBottom()
  //     pos++
  //     return
  //   }
  //   answer.value = marked.parse(body)
  //   scrollToBottom()
  //   answering.value = false
  //   chatContentMacContent = -1
  //   clearInterval(id)
  // }, 50)
}

onMounted(() => {
  chatContainer = (chatContent.value?.$el || chatContent.value)
  chatContainer.addEventListener('click', handleChatClick)
})

onUnmounted(() => {
  chatContainer.removeEventListener('click', handleChatClick)
})

function closeViewer() {
  document.body.style.overflow = prevOverflow
  showViewer.value = false
}

watch(() => sessionId.value, newValue => {
  messages.value = []
})

let index = 1
function handleCreateSession() {
  const id = String(index++)
  sessions.value.push({ id, sessionName: 'dsdsds' + id, modelName: 'qwen:0.5B' })
  sessionId.value = id
}

function handleDeleteSesson(session, index) {
  sessions.value.splice(index, 1)
}

</script>

<style lang="scss" scoped>
.el-aside {
  background-color: var(--el-bg-color-page);
}
.el-main {
  background-color: var(--el-fill-color-extra-light);
  padding: 0;
  :deep(.el-textarea__inner) {
    padding-right: 50px;
    &:focus {
      box-shadow: 0 0 0 1px var(--el-input-focus-border-color);
    }
    &::-webkit-scrollbar {
      width: 7px;
    }
    &::-webkit-scrollbar-thumb {
      border-radius: inherit;
      background-color: var(--el-scrollbar-bg-color, var(--el-text-color-secondary));
      opacity: var(--el-scrollbar-opacity, .3);

      // background-color: var(--el-scrollbar-hover-bg-color, var(--el-text-color-secondary));
      // opacity: var(--el-scrollbar-hover-opacity, .5);
    }
  }
}

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

.avatar {
  width: 40px;
  height: 40px;
  border-radius: 20px;
}

.message {
  padding: 10px 16px;
  border-radius: 12px;
  font-size: 14px;
  letter-spacing: .25px;
  line-height: 24px;
  max-width: calc(100% - 82px);
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
.question {
  display: flex;
  gap: 8px;
  color: var(--el-text-color);
  // color: var(--el-color-primary);
  // & > .avatar {
  // }
  & > .message {
    margin-left: auto;
    // color: white;
    background-color: var(--el-bg-color);
    // background-color: var(--el-color-primary);
    white-space: pre-wrap;
    word-break: break-all;
  }
}
.answer {
  display: flex;
  gap: 8px;
  color: var(--el-text-color);
  // & > .avatar {
  // }
  & > .message {
    // color: var(--el-text-color);
    background-color: var(--el-bg-color);
  }
}
</style>
