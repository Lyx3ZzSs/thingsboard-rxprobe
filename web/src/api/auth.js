import request from './request'

/**
 * 登录
 */
export function login(data) {
  return request.post('/auth/login', data)
}

/**
 * 检查系统是否已初始化
 */
export function checkInit() {
  return request.get('/auth/check-init')
}

/**
 * 初始化系统
 */
export function initSystem(data) {
  return request.post('/auth/init', data)
}

/**
 * 获取当前用户信息
 */
export function getCurrentUser() {
  return request.get('/auth/me')
}

/**
 * 修改密码
 */
export function changePassword(data) {
  return request.post('/auth/change-password', data)
}

/**
 * 登出
 */
export function logout() {
  return request.post('/auth/logout')
}
