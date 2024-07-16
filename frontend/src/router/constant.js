import hallRouter from './modules/hall'
import Hall from '@/layout/hall'

const router = [
  hallRouter,
  { path: '/login', component: () => import('@/views/login/index'), hidden: true, meta: { title: 'login' }},
  // { path: '/form', component: () => import('@/views/form-design'), hidden: true, meta: { title: '表单设计' }},
  //  { path: '/account/active/:activeCode(\\w+)', component: () => import('@/views/login/active'), hidden: true, meta: { title: 'accountActive' }},
  {
    path: '/account/active/:activeCode(\\w+)',
    component: Hall,
    hidden: true,
    children: [{
      path: '',
      component: () => import('@/views/login/active'),
      hidden: true,
      meta: { title: 'accountActive' }
    }]
  },
  {
    path: '/redirect/:path(.*)',
    component: () => import('@/views/redirect/index'),
    hidden: true
  }
]

export default router
