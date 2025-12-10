import { clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs) {
  return twMerge(clsx(inputs))
}

export function formatTime(time) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

export function formatDuration(ms) {
  if (!ms || ms < 0) return '-'
  if (ms < 1000) return `${ms}ms`
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`
  if (ms < 3600000) return `${Math.floor(ms / 60000)}m ${Math.floor((ms % 60000) / 1000)}s`
  return `${Math.floor(ms / 3600000)}h ${Math.floor((ms % 3600000) / 60000)}m`
}

export function formatBytes(bytes) {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
}

export function formatPercent(value) {
  if (value == null) return '-'
  return `${value.toFixed(1)}%`
}

export const serviceTypeLabels = {
  postgresql: 'PostgreSQL',
  cassandra: 'Cassandra',
  redis: 'Redis',
  kafka: 'Kafka',
  http: 'HTTP',
  tcp: 'TCP',
  mysql: 'MySQL',
  mongodb: 'MongoDB',
  elasticsearch: 'Elasticsearch'
}

export function getServiceTypeLabel(type) {
  return serviceTypeLabels[type] || type.toUpperCase()
}

