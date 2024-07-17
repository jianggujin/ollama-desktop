import {defineStore} from 'pinia'

export const useSettingStore = defineStore('setting', () => {
    const count = ref(10)
    const doubleCount = computed(() => count.value * 2)

    function increment() {
        count.value++
    }

    return {count, doubleCount, increment}
})