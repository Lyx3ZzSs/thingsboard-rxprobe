import request from './request'

// 探针类型
export function getProbeTypes() {
  return request.get('/probe/types')
}

// 探针 Schema
export function getProbeSchema(type) {
  return request.get(`/probe/schema/${type}`)
}

// 测试探针
export function testProbe(data) {
  return request.post('/probe/test', data)
}

// 监控目标列表
export function getTargets(params) {
  return request.get('/targets', { params })
}

// 获取单个目标
export function getTarget(id) {
  return request.get(`/targets/${id}`)
}

// 创建目标
export function createTarget(data) {
  return request.post('/targets', data)
}

// 更新目标
export function updateTarget(id, data) {
  return request.put(`/targets/${id}`, data)
}

// 删除目标
export function deleteTarget(id) {
  return request.delete(`/targets/${id}`)
}

// 仪表盘概览
export function getDashboardSummary() {
  return request.get('/dashboard/summary')
}

// 仪表盘监控指标
export function getDashboardMetrics() {
  return request.get('/dashboard/metrics')
}

// 获取目标历史结果
export function getTargetResults(id, params) {
  return request.get(`/targets/${id}/results`, { params })
}
