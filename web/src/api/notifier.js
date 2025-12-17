import request from './request'

/**
 * 获取通知渠道列表
 */
export function getNotifiers() {
  return request.get('/notifiers')
}

/**
 * 创建通知渠道
 */
export function createNotifier(data) {
  return request.post('/notifiers', data)
}

/**
 * 更新通知渠道
 */
export function updateNotifier(id, data) {
  return request.put(`/notifiers/${id}`, data)
}

/**
 * 删除通知渠道
 */
export function deleteNotifier(id) {
  return request.delete(`/notifiers/${id}`)
}

/**
 * 获取通知渠道详情
 */
export function getNotifier(id) {
  return request.get(`/notifiers/${id}`)
}

/**
 * 获取支持的通知类型
 */
export function getNotifierTypes() {
  return request.get('/notifiers/types')
}

/**
 * 获取默认消息模板
 */
export function getDefaultTemplates() {
  return request.get('/notifiers/templates')
}

/**
 * 测试通知渠道
 */
export function testNotifier(data) {
  return request.post('/notifiers/test', data)
}





