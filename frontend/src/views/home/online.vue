<template>
  <div id="loading-wrapper" style="height: 100%;">
    <el-scrollbar ref="scrollbar">
      <div style="display: flex;align-items: center;justify-content: center;margin-top: 50px;">
        <el-input v-model="searchForm.q" style="width: 80%" size="large" placeholder="输入模型名称" :suffix-icon="Search" maxlength="100"/>
      </div>
      <div style="margin: 20px auto 0 auto;width: 80%;display: flex;align-items: center;justify-content: space-between;">
        <el-radio-group v-model="searchForm.searchType">
          <el-radio-button label="功能" value="function" />
          <el-radio-button label="排序" value="sort" />
        </el-radio-group>
        <el-radio-group v-show="searchForm.searchType == 'sort'" v-model="searchForm.sort" class="split-radio">
          <el-radio-button label="特色" value="featured" />
          <el-radio-button label="最受欢迎" value="Most popular" />
          <el-radio-button label="最新" value="Newest" />
        </el-radio-group>
        <el-radio-group v-show="searchForm.searchType == 'function'" v-model="searchForm.c" class="split-radio">
          <el-radio-button label="All" value="" />
          <el-radio-button label="Embedding" value="embedding" />
          <el-radio-button label="Vision" value="vision" />
          <el-radio-button label="Tools" value="tools" />
          <el-radio-button label="Code" value="code" />
        </el-radio-group>
      </div>
      <div style="margin: 20px auto 0 auto;width: 80%;">
        <el-empty v-if="!list.length" />
        <div class="model-item" v-for="(item, index) in list" :key="index">
          <div><el-link style="font-weight: 500;font-size: 1.5rem;" @click="openLibrary(item.name)">{{ item.name }}</el-link></div>
          <div style="margin-top: 10px;"><el-text style="">{{ item.description }}</el-text></div>
          <div style="margin-top: 10px;" v-if="item.tags?.length">
            <el-tag v-for="(tag, ti) in item.tags" :key="ti" :type="tag == 'Embedding' || tag == 'Vision' || tag == 'Tools' || tag == 'Code' ? 'success' : 'primary'">{{ tag }}</el-tag>
          </div>
          <div style="margin-top: 10px;display: flex;align-items: center;">
            <i-ep-download style="color:var(--el-color-info);"/>
            <el-text type="info" style="margin-left: 5px;">{{ item.pullCount }} Pulls</el-text>
            <i-ep-price-tag style="color:var(--el-color-info);margin-left: 10px;"/>
            <el-text type="info" style="margin-left: 5px;">{{ item.tagCount}} Tags</el-text>
            <i-ep-clock style="color:var(--el-color-info);margin-left: 10px;"/>
            <el-text type="info" style="margin-left: 5px;">Updated {{ item.updateTime }}</el-text>
          </div>
        </div>
      </div>
      <el-pagination v-show="searchForm.searchType == 'function'"
        style="margin-top: 40px;display: flex;justify-content: center;margin-bottom: 40px;"
        hide-on-single-page
        background
        :current-page="pagination.page"
        :page-count="pagination.pageCount"
        layout="prev, pager, next"
        @current-change="changeCurrentPage"
        @change="$refs.scrollbar.setScrollTop(0)"
      />
    </el-scrollbar>
  </div>
</template>

<script setup>
import { Search } from '@element-plus/icons-vue'
import { SearchOnline, LibraryOnline } from '@/go/app/Ollama.js'
import { ElMessage } from 'element-plus'
import { runAsync } from '~/utils/wrapper.js'
import { useRouter } from 'vue-router'

const router = useRouter()

function openLibrary(name) {
  router.push('/home/library/' + name)
}

const pagination = ref({
  page: 1,
  pageCount: 0
})

const searchForm = ref({
  q: '',
  searchType: 'function',
  sort: 'featured',
  c: ''
})

const list = ref([])

function changeCurrentPage(page) {
  handleSearch(page)
}

function handleSearch(page) {
  if (searchForm.value.searchType === 'sort') {
    runAsync(() => LibraryOnline(JSON.stringify({ q: searchForm.value.q, sort: searchForm.value.sort })), data => { list.value = data }, _ => {
      list.value = []
      ElMessage.error('查询模型失败')
    })
  } else if (searchForm.value.searchType === 'function') {
    runAsync(() => SearchOnline(JSON.stringify({ q: searchForm.value.q, p: page || 1, c: searchForm.value.c })), data => {
      pagination.value = { page: data.page, pageCount: data.pageCount }
      list.value = data.items
    }, _ => {
      list.value = []
      ElMessage.error('查询模型失败')
    })
  }
}

let timeout
function lazySearch() {
  if (timeout) {
    clearTimeout(timeout)
    timeout = null
  }

  timeout = setTimeout(handleSearch, 300)
}

watch(searchForm, lazySearch, {
  deep: true,
  immediate: true
})

// onMounted(() => {
//   handleSearch()
// })
</script>

<style lang="scss" scoped>
.split-radio {
  .el-radio-button {
    & + .el-radio-button {
      margin-left: 20px;
    }
  }
  // 可以用，但是已过时
  // ::v-deep .el-radio-button__inner {
  :deep(.el-radio-button__inner) {
    border: var(--el-border)!important;
    border-radius: var(--el-border-radius-base)!important;
  }
}
.model-item {
  .el-tag + .el-tag {
    margin-left: 10px;
  }
  & + .model-item {
    border-top: 1px solid var(--el-border-color);
    margin-top: 20px;
    padding-top: 20px;
  }
}
</style>
