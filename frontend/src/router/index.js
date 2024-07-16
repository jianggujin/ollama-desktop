import { createWebHistory, createRouter } from "vue-router";

const routes = [
  {
    path: '/',
    component: () => import('~/components/HelloWorld.vue')  // 首页组件
  }, {
    path: '/about',
    component: () => import('~/components/About.vue')  // 关于我们组件
  }
]

const router = createRouter({
  history: createWebHistory(), // 路由类型
  routes // short for `routes: routes`
})


export default router