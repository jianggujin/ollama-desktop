import {
  ElLoading
} from 'element-plus'
window._loading_count = 0

export function showLoading() {
  window._loading_count++
  if (!window._loading_instance) {
    window._loading_instance = ElLoading.service({
      text: '加载中...',
      spinner: 'el-icon-loading',
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