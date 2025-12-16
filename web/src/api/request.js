import axios from 'axios'
import { useRouter } from 'vue-router'
import { clearAuth } from '@/store/auth'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 30000
})

// 请求拦截器
request.interceptors.request.use(
  config => {
    // 添加token到请求头
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  response => {
    const res = response.data
    // 兼容不同的响应格式
    if (res.code !== undefined && res.code !== 0) {
      console.error('API Error:', res.message)
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    return res
  },
  error => {
    // 处理401未授权错误
    if (error.response?.status === 401) {
      clearAuth()
      // 如果不在登录页，跳转到登录页
      if (window.location.pathname !== '/login' && window.location.pathname !== '/init') {
        window.location.href = '/login'
      }
    }
    console.error('Request Error:', error.message)
    return Promise.reject(error)
  }
)

export default request
