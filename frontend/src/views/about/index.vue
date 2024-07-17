<template>
  <div class="wrapper">
    <img src="/ollama.png" />
    <div>{{ setting.count }}</div>
  </div>
</template>

<script setup>
  import {
    useSettingStore
  } from "~/store/setting.js"

  const setting = useSettingStore()
</script>

<style lang="scss" scoped>
  .wrapper {
    width: 100%;
    height: 100%;
    display: flex;
    justify-content: center;
    align-items: center;
  }
</style>