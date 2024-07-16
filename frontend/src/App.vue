<template>
  <el-config-provider>
    <router-view />
  </el-config-provider>
</template>

<style>
#app {
  text-align: center;
  color: var(--el-text-color-primary);
}

.main-container {
  height: calc(100vh - var(--el-menu-item-height) - 3px);
}
</style>
