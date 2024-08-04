import { defineStore } from 'pinia'

export const useDownloaderStore = defineStore('downloader', () => {
  const list = ref([{}, {}])

  return { list }
})
