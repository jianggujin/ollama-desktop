<template>
    <el-config-provider :locale="locale">
        <router-view />
    </el-config-provider>
</template>

<script setup>
    import {
        ElMessageBox
    } from 'element-plus'
    import {
        EventsOn
    } from "@/runtime/runtime.js"
    import {
        Quit
    } from "@/go/app/App.js"
    import zhCn from 'element-plus/es/locale/lang/zh-cn'
    const locale = ref(zhCn)

    onMounted(() => {
        try {
            EventsOn("beforeClose", function() {
                ElMessageBox.confirm("确认要退出Ollama Desktop", '退出', {
                    confirmButtonText: '确认',
                    cancelButtonText: '取消',
                    type: 'warning',
                }).then(() => {
                    Quit()
                })
            })
        } catch (e) {
            console.error(e)
        }
    })
</script>

<style>
</style>