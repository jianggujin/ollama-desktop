/**
 * 判断传入参数是否为 Promise (最小实现)
 * @param { any } value
 * @return { boolean } 是否为 Promise
 * @description 代码源于第三方库 is-promise
 * @tutorial https://www.npmjs.com/package/is-promise
 * */
function isPromise(value) {
  return !!value &&
  (typeof value === 'object' || typeof value === 'function') && typeof value.then === 'function'
}

export function runQuietly(fn, successCallback, errorCallback, finallyCallback) {
  if (typeof fn !== 'function') {
    return
  }
  if (typeof successCallback !== 'function') {
    successCallback = function() {}
  }
  if (typeof errorCallback !== 'function') {
    errorCallback = function() {}
  }
  if (typeof finallyCallback !== 'function') {
    finallyCallback = function() {}
  }
  try {
    const promise = fn()
    if (isPromise(promise)) {
      promise.then(successCallback).catch(errorCallback).finally(finallyCallback)
    } else {
      successCallback()
    }
  } catch (e) {
    console.error(e)
    try {
      errorCallback(e)
    } catch (ex) {
      console.error(ex)
    }
    try {
      finallyCallback()
    } catch (ex) {
      console.error(ex)
    }
  }
}
