<template>
  <div
    style="height: 100%;"
    v-loading="loading"
    :element-loading-text="loadingOptions.text"
    :element-loading-spinner="loadingOptions.svg"
    :element-loading-svg-view-box="loadingOptions.svgViewBox"
    :element-loading-background="loadingOptions.background"
  >
    <el-scrollbar ref="scrollbar">
      <div style="display: flex;align-items: center;justify-content: center;margin-top: 50px;">
        <el-input v-model.trim="searchForm.q" style="width: 80%" size="large" placeholder="输入模型名称" :suffix-icon="Search" maxlength="100"/>
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
          <div style="display: flex;align-items: center;">
            <el-link style="font-weight: 500;font-size: 1.5rem;" @click="openLibrary(item.name)">{{ item.name }}</el-link>
            <el-tag v-show="item.archive" type="warning" round style="margin-left: 5px;">Archive</el-tag>
          </div>
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
import { runQuietly } from '~/utils/wrapper.js'
import { useRouter } from 'vue-router'
import loadingOptions from '~/utils/loading.js'

const loading = ref(false)

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

const cacheKey = '/home/library'

function saveCache() {
  sessionStorage.setItem(cacheKey, JSON.stringify({
    pagination: pagination.value,
    searchForm: searchForm.value,
    list: list.value
  }))
}

let inited = false

onMounted(() => {
  let cacheValue = sessionStorage.getItem(cacheKey)
  if (cacheValue) {
    cacheValue = JSON.parse(cacheValue)
    pagination.value = cacheValue.pagination
    searchForm.value = cacheValue.searchForm
    list.value = cacheValue.list
    nextTick(() => { inited = true })
  } else {
    inited = true
    handleSearch()
  }
})

function handleSearch(page) {
  if (searchForm.value.searchType === 'sort') {
    loading.value = true
    runQuietly(() => LibraryOnline({ q: searchForm.value.q, sort: searchForm.value.sort }), data => { list.value = data || [] }, _ => {
      list.value = []
      ElMessage.error('查询模型失败')
    }, _ => {
      saveCache()
      loading.value = false
    })
  } else if (searchForm.value.searchType === 'function') {
    loading.value = true
    runQuietly(() => SearchOnline({ q: searchForm.value.q, p: page || 1, c: searchForm.value.c }), data => {
      pagination.value = { page: data.page, pageCount: data.pageCount }
      list.value = data.items || []
    }, _ => {
      list.value = []
      ElMessage.error('查询模型失败')
    }, _ => {
      saveCache()
      loading.value = false
    })
  }
}

let timeout
function lazySearch() {
  if (!inited) {
    return
  }
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
    box-shadow: none;
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
