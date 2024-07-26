import path from 'path'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

import Icons from 'unplugin-icons/vite'
import IconsResolver from 'unplugin-icons/resolver'

import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

import ElementPlus from 'unplugin-element-plus/vite'

import { createSvgIconsPlugin } from 'vite-plugin-svg-icons'

const pathSrc = path.resolve(__dirname, 'src')
const pathWailsjs = path.resolve(__dirname, 'wailsjs')

// https://vitejs.dev/config/
export default defineConfig({
  resolve: {
    alias: {
      '~/': `${pathSrc}/`,
      '@/': `${pathWailsjs}/`
    }
  },
  plugins: [
    vue(),
    ElementPlus(),
    createSvgIconsPlugin({
      iconDirs: [path.resolve(process.cwd(), 'src/assets/icons')],
      symbolId: 'icon-[dir]-[name]'
    }),
    AutoImport({
      // 自动导入 Vue 相关函数，如：ref, reactive, toRef 等
      imports: ['vue'],
      resolvers: [
        ElementPlusResolver(),
        IconsResolver({
          prefix: 'Icon'
        })
      ],
      dts: path.resolve(pathSrc, 'auto-imports.d.ts'),
      eslintrc: {
        // enabled: true
      }
    }),
    Components({
      resolvers: [
        IconsResolver({
          enabledCollections: ['ep']
        }),
        ElementPlusResolver()
      ],
      dts: path.resolve(pathSrc, 'components.d.ts')
    }),
    Icons({
      autoInstall: true
    })
  ]
})
