import request from './request'

// 获取告警列表
export function getAlerts(params) {
  return request.get('/alerts', { params })
}

// 获取单个告警
export function getAlert(id) {
  return request.get(`/alerts/${id}`)
}

// 静默告警
export function silenceAlert(targetId, durationMinutes) {
  return request.post(`/targets/${targetId}/silence`, { duration_minutes: durationMinutes })
}

// 取消静默
export function unsilenceAlert(targetId) {
  return request.delete(`/targets/${targetId}/silence`)
}
