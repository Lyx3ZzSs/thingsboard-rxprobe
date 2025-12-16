import { ref, computed } from 'vue'
import { checkInit, getCurrentUser } from '@/api/auth'

// 状态
const token = ref(localStorage.getItem('token') || '')
const user = ref(null)
const initialized = ref(null) // null表示未检查，true/false表示已检查

// 计算属性
const isAuthenticated = computed(() => !!token.value)
const isAdmin = computed(() => user.value?.role === 'admin')

// 设置token
export function setToken(newToken) {
  token.value = newToken
  if (newToken) {
    localStorage.setItem('token', newToken)
  } else {
    localStorage.removeItem('token')
  }
}

// 设置用户信息
export function setUser(userData) {
  user.value = userData
}

// 清除认证信息
export function clearAuth() {
  token.value = ''
  user.value = null
  localStorage.removeItem('token')
}

// 检查系统初始化状态
export async function checkSystemInit() {
  try {
    const res = await checkInit()
    initialized.value = res.data.initialized
    return res.data.initialized
  } catch (e) {
    console.error('检查初始化状态失败:', e)
    initialized.value = true // 假设已初始化，避免阻塞
    return true
  }
}

// 加载当前用户信息
export async function loadCurrentUser() {
  if (!token.value) {
    return null
  }
  
  try {
    const res = await getCurrentUser()
    user.value = res.data
    return res.data
  } catch (e) {
    // token可能已过期
    clearAuth()
    return null
  }
}

// 导出
export function useAuth() {
  return {
    token,
    user,
    initialized,
    isAuthenticated,
    isAdmin,
    setToken,
    setUser,
    clearAuth,
    checkSystemInit,
    loadCurrentUser
  }
}
