<template>
  <router-view :key="key" />
</template>
<script>
export default {
  name: 'Nested',
  computed: {
    key() {
      return this.$route.path
    }
  }
}
</script>

