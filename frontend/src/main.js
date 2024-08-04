import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
// import SvgIcon from '~/components/SvgIcon/index.vue'
import 'virtual:svg-icons-register'

import 'element-plus/theme-chalk/dark/css-vars.css'
// import "element-plus/theme-chalk/src/message.scss"
// import "element-plus/theme-chalk/src/message-box.scss"
import '~/styles/index.scss'

const app = createApp(App)
app.use(router)
app.use(store)
// app.component('svg-icon', SvgIcon)
app.mount('#app')
