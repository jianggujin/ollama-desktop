<template>
  <el-container
     v-loading="loading"
    :element-loading-text="loadingOptions.text"
    :element-loading-spinner="loadingOptions.svg"
    :element-loading-svg-view-box="loadingOptions.svgViewBox"
    :element-loading-background="loadingOptions.background"
  >
    <el-aside width="250px">
      <session-panel @change="handleSessionChange"/>
    </el-aside>
    <el-main style="display: flex;flex-direction: column;">
      <el-scrollbar style="flex: 1;" ref="chatScrollbar" @scroll="handleChatScroll">
        <div ref="chatContent" style="margin: 10px auto 0 auto;width: 80%;position: relative;display: flex;flex-direction: column;gap: 16px;">
          <div v-for="(message, index) in messages" :key="index" :class="{question: message.role == 'user', answer: message.role == 'assistant' || message.id == 'thinking'}">
            <template v-if="message.role == 'user'">
              <pre class="message">{{ message.content }}</pre>
              <i-ep-avatar class="avatar"/>
            </template>
            <template v-if="message.role == 'assistant'">
              <svg-icon icon-class="ollama" class-name="avatar"/>
              <div class="message" v-html="marked.parse(message.content || '')"></div>
            </template>
            <template v-if="message.id == 'thinking'">
              <el-button disabled class="avatar" :icon="Loading" circle plain type="primary" loading/>
              <pre class="message">{{ message.content }}</pre>
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
            @click="sendQuestion"
            :loading="answering"
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
import SessionPanel from './session-panel.vue'
import { Promotion, Loading } from '@element-plus/icons-vue'
import { throttle } from 'lodash'
import marked from '~/utils/markdown.js'
import { ElMessage } from 'element-plus'
import { BrowserOpenURL, EventsOn, EventsOff } from '@/runtime/runtime.js'
import { SessionHistoryMessages, Conversation } from '@/go/app/Chat.js'
import { runQuietly } from '~/utils/wrapper.js'
import loadingOptions from '~/utils/loading.js'

const loading = ref(false)

const sessionId = ref('')

const messages = ref([])

const question = ref('')
const answering = ref(false)
const canSendQuestion = computed(() => !answering.value && !isAllWhitespace(question.value))
const hasHistory = ref(false)

const chatScrollbar = ref(null)
const chatContent = ref(null)

function handleSessionChange(value) {
  sessionId.value = value
}

watch(() => sessionId.value, newValue => {
  messages.value = []
  if (!newValue) {
    return
  }
  hasHistory.value = true
  loadSessionMessages(true)
})

let chatContainer

const scrollToBottom = throttle(() => {
  nextTick(() => {
    const chatScrollbarHeight = (chatScrollbar.value?.$el || chatScrollbar.value)?.clientHeight || 0
    const chatContentHeight = (chatContent.value?.$el || chatContent.value)?.clientHeight || 0
    chatScrollbar.value?.setScrollTop(chatContentHeight - chatScrollbarHeight + 20)
  })
}, 500)

function isAllWhitespace(str) {
  return /^\s*$/.test(str)
}

function handleQuestionKeydown(event) {
  if (event.altKey && event.key === 'Enter') {
    event.preventDefault()
    sendQuestion()
  }
}

function loadSessionMessages(first) {
  loading.value = true
  runQuietly(() => SessionHistoryMessages({ sessionId: sessionId.value, nextMarker: messages.value[0]?.id || '' }), data => {
    data = data || []
    if (data.length === 0) {
      hasHistory.value = false
      return
    }
    messages.value.unshift(...data)
    if (first) {
      scrollToBottom()
    }
    // chatScrollbar.value?.setScrollTop(0)
  }, _ => { ElMessage.error('获取会话历史消息失败') }, () => { loading.value = false })
}

let timeout
function lazyloadSessionMessages() {
  if (timeout) {
    clearTimeout(timeout)
    timeout = null
  }

  timeout = setTimeout(loadSessionMessages, 300)
}

let currentAnswerId = ''

function sendQuestion() {
  if (!canSendQuestion.value) {
    return
  }
  answering.value = true

  loading.value = true
  runQuietly(() => Conversation({ sessionId: sessionId.value, content: question.value }), ({ id, sessionId, content, createdAt }) => {
    question.value = ''
    messages.value.push({ id, sessionId, role: 'user', content, success: true, createdAt })
    messages.value.push({ id: 'thinking', sessionId, role: '', content: '思考中...', success: false, createdAt })
    scrollToBottom()

    currentAnswerId = id
    runQuietly(() => {
      EventsOn(id, (answerContent, answerDone, answerSuccess) => {
        const lastMessage = messages.value[messages.value.length - 1]
        if (lastMessage.id === 'thinking') {
          messages.value[messages.value.length - 1] = { id: currentAnswerId, sessionId, role: 'assistant', content: answerContent, answerSuccess, createdAt }
          scrollToBottom()
          return
        }
        // 回答失败或完成
        if (!answerSuccess || answerDone) {
          lastMessage.content = answerContent
          runQuietly(() => EventsOff(currentAnswerId))
          currentAnswerId = ''

          answering.value = false
          return
        }
        // 回答中
        lastMessage.content = answerContent + '_'
        scrollToBottom()
      })
    })
  }, _ => { ElMessage.error('发送消息失败') }, () => { loading.value = false })
}

function handleChatScroll({ scrollTop }) {
  if (scrollTop < 50 && hasHistory.value) {
    lazyloadSessionMessages()
  }
}

onMounted(() => {
  chatContainer = (chatContent.value?.$el || chatContent.value)
  chatContainer.addEventListener('click', handleChatClick)
})

onUnmounted(() => {
  chatContainer.removeEventListener('click', handleChatClick)
  if (currentAnswerId) {
    runQuietly(() => EventsOff(currentAnswerId))
    currentAnswerId = ''
  }
})

const showViewer = ref(false)
const previewSrcList = ref([])

let prevOverflow = ''

function closeViewer() {
  document.body.style.overflow = prevOverflow
  showViewer.value = false
}

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
</script>

<style lang="scss" scoped>
// .el-aside {
  // background-color: var(--el-bg-color-page);
// }
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
  :deep(p) {
    margin-block: unset;
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
    margin: 0 0 0 auto;
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
    margin: 0;
    background-color: var(--el-bg-color);
  }
}
</style>
