<template>
  <el-container>
    <el-aside width="200px">
      <el-scrollbar>
        <ul class="el-menu el-menu--vertical" style="--el-menu-level: 0;">
          <li v-for="(menu, mIndex) in menus" :key="mIndex" class="el-sub-menu is-active is-opened">
            <div class="el-sub-menu__title"> {{menu.group}} </div>
            <ul class="el-menu el-menu--inline" style="--el-menu-level: 1;">
              <li v-for="item in menu.items" :key="item.path" @click="handleSelect(item.path)" :class="{'el-menu-item':true, 'is-active': item.path == key}">
                {{item.name}}
              </li>
            </ul>
          </li>
        </ul>
      </el-scrollbar>
    </el-aside>
    <el-main>
      <router-view :key="key" />
    </el-main>
  </el-container>
</template>
<script setup>
import NavHeader from './NavHeader.vue'
import { useRoute, useRouter } from 'vue-router'

const menus = ref([{
  group: 'Ollama',
  items: [{
    name: '环境信息',
    path: '/home/ollama'
  }, {
    name: '本地模型',
    path: '/home/tags'
  }, {
    name: '在线模型',
    path: '/home/online'
  }]
}])

const route = useRoute()
const router = useRouter()
const key = computed(() => { return route.path })
const handleSelect = (key) => { router.replace(key) }
</script>
<style lang="scss" scoped>
  .el-container {
    height: 100%;

    .el-sub-menu__title:hover {
      background-color: transparent !important;
      cursor: default !important;
    }

    .el-menu .el-menu {
      min-height: unset !important;
    }

    .el-aside,
    .el-main {
      padding: 0;
      height: var(--app-layout-main);
    }

    .el-menu {
      min-height: var(--app-layout-main);
    }
  }
</style>
