<template>
  <div id="loading-wrapper" style="height: 100%;">
    <el-scrollbar ref="scrollbar">
      <div style="display: flex;align-items: center;justify-content: center;margin-top: 50px;">
        <el-input v-model="searchForm.q" style="width: 80%" size="large" placeholder="输入模型名称" :suffix-icon="Search" />
      </div>
      <div style="margin-top: 15px;margin: 20px auto 0 auto;width: 80%;display: flex;align-items: center;justify-content: space-between;">
        <el-radio-group v-model="searchForm.searchType">
          <el-radio-button label="排序" value="sort" />
          <el-radio-button label="功能" value="function" />
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
          <!--模型名称-->
          <!-- <div><el-text size="large" style="font-weight: 500;font-size: 1.5rem;">llama3.1</el-text></div> -->
          <div><el-link style="font-weight: 500;font-size: 1.5rem;">{{ item.model }}</el-link></div>
          <div style="margin-top: 10px;"><el-text style="">{{ item.description }}</el-text></div>
          <div style="margin-top: 10px;" v-if="item.tags?.length">
            <el-tag v-for="(tag, ti) in item.tags" :key="ti" :type="tas == 'Embedding' || tas == 'Vision' || tag == 'Tools' || tag == 'Code' ? 'success' : 'primary'">{{ tag }}</el-tag>
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
        :page-size="20"
        :pager-count="11"
        layout="prev, pager, next"
        :total="1000"
        @change="$refs.scrollbar.setScrollTop(0)"
      />
    </el-scrollbar>
  </div>
</template>

<script setup>
import { Search } from '@element-plus/icons-vue'

import { ElMessage } from 'element-plus'
import { runAsync } from '~/utils/wrapper.js'

const searchForm = ref({
  searchType: 'sort',
  sort: 'featured',
  c: ''
})

const list = ref([{
  model: 'llama3.1',
  description: 'Llama 3.1 is a new state-of-the-art model from Meta available in 8B, 70B and 405B parameter sizes.',
  tags: ['Tools', '8B', '70B'],
  pullCount: '577.1K',
  tagCount: 36,
  updateTime: '6 days ago'
}])

watch(searchForm, newValue => {
  console.log(newValue.searchType)
}, {
  deep: true,
  immediate: true
})

onMounted(() => {
  // runAsync(Envs, data => { envs.value = data }, _ => { ElMessage.error('获取Ollama环境信息失败') })
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
