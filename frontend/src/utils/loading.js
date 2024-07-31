import { ElLoading } from 'element-plus'
window._loading_count = 0
const svg = '<path fill="var(--el-color-primary)" d="M512 64a32 32 0 0 1 32 32v192a32 32 0 0 1-64 0V96a32 32 0 0 1 32-32m0 640a32 32 0 0 1 32 32v192a32 32 0 1 1-64 0V736a32 32 0 0 1 32-32m448-192a32 32 0 0 1-32 32H736a32 32 0 1 1 0-64h192a32 32 0 0 1 32 32m-640 0a32 32 0 0 1-32 32H96a32 32 0 0 1 0-64h192a32 32 0 0 1 32 32M195.2 195.2a32 32 0 0 1 45.248 0L376.32 331.008a32 32 0 0 1-45.248 45.248L195.2 240.448a32 32 0 0 1 0-45.248zm452.544 452.544a32 32 0 0 1 45.248 0L828.8 783.552a32 32 0 0 1-45.248 45.248L647.744 692.992a32 32 0 0 1 0-45.248zM828.8 195.264a32 32 0 0 1 0 45.184L692.992 376.32a32 32 0 0 1-45.248-45.248l135.808-135.808a32 32 0 0 1 45.248 0m-452.544 452.48a32 32 0 0 1 0 45.248L240.448 828.8a32 32 0 0 1-45.248-45.248l135.808-135.808a32 32 0 0 1 45.248 0z"></path>'
export function showLoading() {
  window._loading_count++
  if (!window._loading_instance) {
    const target = document.querySelector('#loading-wrapper') || document.body
    window._loading_instance = ElLoading.service({
      text: '加载中...',
      target,
      svg,
      svgViewBox: '0, 0, 1024, 1024',
      background: 'rgba(0, 0, 0, 0.3)'
    })
  }
}

export function closeLoading() {
  window._loading_count--
  if (window._loading_count < 1) {
    window._loading_count = 0
    window._loading_instance && window._loading_instance.close()
    window._loading_instance = null
  }
}

export function loading(fn) {
  if (typeof fn !== 'function') {
    return
  }
  showLoading()
  fn(closeLoading)
}
