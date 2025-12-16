import { createRouter, createWebHistory } from 'vue-router'
import { checkSystemInit, loadCurrentUser } from '@/store/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { title: '登录', requiresAuth: false }
  },
  {
    path: '/init',
    name: 'Init',
    component: () => import('@/views/Init.vue'),
    meta: { title: '系统初始化', requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('@/views/Layout.vue'),
    redirect: '/dashboard',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '仪表盘' }
      },
      {
        path: 'services',
        name: 'Services',
        component: () => import('@/views/Services.vue'),
        meta: { title: '监控服务' }
      },
      {
        path: 'services/:id',
        name: 'ServiceDetail',
        component: () => import('@/views/ServiceDetail.vue'),
        meta: { title: '服务详情' }
      },
      {
        path: 'alerts',
        name: 'Alerts',
        component: () => import('@/views/Alerts.vue'),
        meta: { title: '告警记录' }
      },
      {
        path: 'notifiers',
        name: 'Notifiers',
        component: () => import('@/views/Notifiers.vue'),
        meta: { title: '通知中心' }
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/dashboard'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach(async (to, from, next) => {
  // 检查系统初始化状态
  const initialized = await checkSystemInit()
  
  // 如果未初始化，且不在初始化页面，跳转到初始化页面
  if (!initialized && to.name !== 'Init') {
    next('/init')
    return
  }
  
  // 如果已初始化，且在初始化页面，跳转到登录页
  if (initialized && to.name === 'Init') {
    next('/login')
    return
  }
  
  // 检查是否需要认证
  if (to.meta.requiresAuth) {
    const token = localStorage.getItem('token')
    if (!token) {
      next('/login')
      return
    }
    
    // 尝试加载用户信息
    try {
      await loadCurrentUser()
    } catch (e) {
      // token无效，清除并跳转到登录页
      next('/login')
      return
    }
  }
  
  // 如果已登录，访问登录页则跳转到首页
  if (to.name === 'Login' && localStorage.getItem('token')) {
    next('/dashboard')
    return
  }
  
  next()
})

export default router
