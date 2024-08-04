import { showLoading, closeLoading } from '~/utils/loading.js'

export function runQuietly(fn, successCallback, errorCallback) {
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
    fn()
    successCallback()
  } catch (e) {
    console.error(e)
    try {
      errorCallback(e)
    } catch {
      //
    }
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
    console.error(e)
    closeLoading()
    try {
      errorCallback(e)
    } catch {
      //
    }
  }
}
