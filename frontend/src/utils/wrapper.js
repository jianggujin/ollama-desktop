import { showLoading, closeLoading } from '~/utils/loading.js'

export function runQuietly(fn) {
  if (typeof fn !== 'function') {
    return
  }
  try {
    fn()
  } catch (e) {
    console.error(e)
  }
}

export function runAsync(fn, successCallback, errorCallback) {
  if (typeof fn !== 'function') {
    return
  }
  if (!successCallback) {
    successCallback = function() {}
  }
  if (!errorCallback) {
    errorCallback = function() {}
  }
  try {
    showLoading()
    fn().then(successCallback).catch(errorCallback).finally(closeLoading)
  } catch (e) {
    closeLoading()
    try {
      errorCallback(e)
    } catch {
      //
    }
  }
}
