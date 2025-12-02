import request from './request'

export function login(username, password) {
  return request.post('/auth/login', { username, password })
}

export function logout() {
  return request.post('/auth/logout')
}

export function getCurrentUser() {
  return request.get('/auth/me')
}

export function changePassword(oldPassword, newPassword) {
  return request.put('/auth/password', { old_password: oldPassword, new_password: newPassword })
}

