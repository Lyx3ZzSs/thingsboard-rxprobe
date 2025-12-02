import request from './request'

// 告警规则
export function getRules() {
  return request.get('/rules')
}

export function createRule(data) {
  return request.post('/rules', data)
}

export function updateRule(id, data) {
  return request.put(`/rules/${id}`, data)
}

export function deleteRule(id) {
  return request.delete(`/rules/${id}`)
}

// 告警记录
export function getAlerts(params) {
  return request.get('/alerts', { params })
}

export function getAlert(id) {
  return request.get(`/alerts/${id}`)
}

export function silenceAlert(id, durationMinutes) {
  return request.put(`/alerts/${id}/silence`, { duration_minutes: durationMinutes })
}

