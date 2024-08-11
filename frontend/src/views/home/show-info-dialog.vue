<template>
  <el-dialog v-model="visible" top="50px" :title="title" width="800">
    <el-scrollbar ref="scrollbarRef" style="height: calc(100% - 12px);">
      <pre style="margin: 0;white-space: pre-wrap;word-break: break-all;">{{ content }}</pre>
    </el-scrollbar>
  </el-dialog>
</template>

<script setup>
const title = ref('')
const content = ref('')
const visible = ref(false)
const scrollbarRef = ref(null)

function showDialog(titleValue, contentValue) {
  title.value = titleValue
  content.value = contentValue
  visible.value = true
  nextTick(() => scrollbarRef.value?.setScrollTop(0))
}

defineExpose({
  showDialog
})
</script>

<style lang="scss" scoped>
</style>
