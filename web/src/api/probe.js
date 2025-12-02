import request from './request'

// 探针类型
export function getProbeTypes() {
  return request.get('/probe/types')
}

// 获取探针配置 Schema
export function getProbeSchema(type) {
  return request.get(`/probe/schema/${type}`)
}

// 测试探测目标
export function testTarget(data) {
  return request.post('/probe/test', data)
}

// 目标管理
export function getTargets(params) {
  return request.get('/targets', { params })
}

export function getTarget(id) {
  return request.get(`/targets/${id}`)
}

export function createTarget(data) {
  return request.post('/targets', data)
}

export function updateTarget(id, data) {
  return request.put(`/targets/${id}`, data)
}

export function deleteTarget(id) {
  return request.delete(`/targets/${id}`)
}

export function getTargetResults(id, params) {
  return request.get(`/targets/${id}/results`, { params })
}

export function getTargetStats(id) {
  return request.get(`/targets/${id}/stats`)
}

// 仪表盘
export function getDashboardSummary() {
  return request.get('/dashboard/summary')
}

export function getDashboardMetrics() {
  return request.get('/dashboard/metrics')
}

