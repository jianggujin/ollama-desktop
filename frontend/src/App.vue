<template>
<el-config-provider>
  <img id="logo" alt="Wails logo" src="./assets/images/logo-universal.png"/>
  <el-button type="primary">Primary</el-button>
    <el-button type="success">Success</el-button>
</el-config-provider>
</template>

<script>
import { defineComponent } from 'vue'
import { ElConfigProvider } from 'element-plus'
export default defineComponent({
  components: {
    ElConfigProvider,
  },
  setup() {
    return {
    }
  },
})
</script>



<style>
#logo {
  display: block;
  width: 50%;
  height: 50%;
  margin: auto;
  padding: 10% 0 0;
  background-position: center;
  background-repeat: no-repeat;
  background-size: 100% 100%;
  background-origin: content-box;
}
</style>
