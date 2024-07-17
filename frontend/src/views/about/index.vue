<template>
    <div class="wrapper">
        <img src="/ollama.png" />
    </div>
</template>

<script setup>
    import {
        searchPreview
    } from "~/api/ollama"
    onMounted(() => {
        searchPreview("qwen").then(res => {
            console.log(res)
        })
    })
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