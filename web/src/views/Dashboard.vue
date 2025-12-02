<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stat-cards">
      <el-col :span="6">
        <el-card class="stat-card total">
          <div class="stat-icon">
            <el-icon :size="40"><Monitor /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ summary.total_targets || 0 }}</div>
            <div class="stat-label">监控目标</div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card healthy">
          <div class="stat-icon">
            <el-icon :size="40"><CircleCheck /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ summary.healthy_count || 0 }}</div>
            <div class="stat-label">正常</div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card unhealthy">
          <div class="stat-icon">
            <el-icon :size="40"><CircleClose /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ summary.unhealthy_count || 0 }}</div>
            <div class="stat-label">异常</div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card class="stat-card unknown">
          <div class="stat-icon">
            <el-icon :size="40"><QuestionFilled /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ summary.unknown_count || 0 }}</div>
            <div class="stat-label">未知</div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    
    <el-row :gutter="20">
      <!-- 监控状态列表 -->
      <el-col :span="16">
        <el-card class="metrics-card">
          <template #header>
            <div class="card-header">
              <span>监控状态</span>
              <el-button type="primary" link @click="refreshMetrics">
                <el-icon><Refresh /></el-icon>
                刷新
              </el-button>
            </div>
          </template>
          
          <el-table :data="metrics" v-loading="metricsLoading" stripe>
            <el-table-column label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)" effect="dark" round>
                  {{ getStatusText(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            
            <el-table-column label="名称" prop="name" min-width="150" />
            
            <el-table-column label="类型" width="120">
              <template #default="{ row }">
                {{ getTypeLabel(row.type) }}
              </template>
            </el-table-column>
            
            <el-table-column label="响应时间" width="100">
              <template #default="{ row }">
                <span :class="getLatencyClass(row.last_latency_ms)">
                  {{ row.last_latency_ms ? `${row.last_latency_ms}ms` : '-' }}
                </span>
              </template>
            </el-table-column>
            
            <el-table-column label="成功率(24h)" width="120">
              <template #default="{ row }">
                <el-progress
                  :percentage="row.stats?.success_rate_24h || 0"
                  :color="getSuccessRateColor(row.stats?.success_rate_24h)"
                  :stroke-width="8"
                />
              </template>
            </el-table-column>
            
            <el-table-column label="最后检查" width="180">
              <template #default="{ row }">
                {{ formatTime(row.last_check_at) }}
              </template>
            </el-table-column>
            
            <el-table-column label="操作" width="100" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link @click="goToDetail(row)">
                  详情
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      
      <!-- 最近告警 -->
      <el-col :span="8">
        <el-card class="alerts-card">
          <template #header>
            <div class="card-header">
              <span>最近告警</span>
              <el-button type="primary" link @click="$router.push('/alerts')">
                查看全部
              </el-button>
            </div>
          </template>
          
          <div v-if="summary.recent_alerts?.length" class="alert-list">
            <div
              v-for="alert in summary.recent_alerts"
              :key="alert.id"
              class="alert-item"
              :class="alert.status"
            >
              <div class="alert-icon">
                <el-icon v-if="alert.status === 'firing'" color="#f56c6c"><Warning /></el-icon>
                <el-icon v-else color="#67c23a"><CircleCheck /></el-icon>
              </div>
              <div class="alert-content">
                <div class="alert-name">{{ alert.target_name }}</div>
                <div class="alert-message">{{ alert.message }}</div>
                <div class="alert-time">{{ formatTime(alert.fired_at) }}</div>
              </div>
            </div>
          </div>
          <el-empty v-else description="暂无告警" :image-size="80" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { getDashboardSummary, getDashboardMetrics } from '@/api/probe'

const router = useRouter()

const summary = ref({})
const metrics = ref([])
const metricsLoading = ref(false)
let refreshTimer = null

const typeLabels = {
  postgresql: 'PostgreSQL',
  cassandra: 'Cassandra',
  redis: 'Redis',
  kafka: 'Kafka',
  http: 'HTTP',
  tcp: 'TCP'
}

onMounted(() => {
  loadData()
  // 每30秒自动刷新
  refreshTimer = setInterval(loadData, 30000)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})

async function loadData() {
  try {
    const [summaryRes, metricsRes] = await Promise.all([
      getDashboardSummary(),
      getDashboardMetrics()
    ])
    summary.value = summaryRes.data
    metrics.value = metricsRes.data
  } catch (e) {
    console.error(e)
  }
}

async function refreshMetrics() {
  metricsLoading.value = true
  try {
    const res = await getDashboardMetrics()
    metrics.value = res.data
  } finally {
    metricsLoading.value = false
  }
}

function getStatusType(status) {
  const types = { healthy: 'success', unhealthy: 'danger', unknown: 'info' }
  return types[status] || 'info'
}

function getStatusText(status) {
  const texts = { healthy: '正常', unhealthy: '异常', unknown: '未知' }
  return texts[status] || '未知'
}

function getTypeLabel(type) {
  return typeLabels[type] || type
}

function getLatencyClass(latency) {
  if (!latency) return ''
  if (latency < 100) return 'latency-good'
  if (latency < 500) return 'latency-warn'
  return 'latency-bad'
}

function getSuccessRateColor(rate) {
  if (rate >= 99) return '#67c23a'
  if (rate >= 95) return '#e6a23c'
  return '#f56c6c'
}

function formatTime(time) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

function goToDetail(row) {
  router.push(`/targets/${row.id}`)
}
</script>

<style lang="scss" scoped>
.dashboard {
  .stat-cards {
    margin-bottom: 20px;
  }
  
  .stat-card {
    display: flex;
    align-items: center;
    padding: 20px;
    
    .stat-icon {
      width: 80px;
      height: 80px;
      border-radius: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
      margin-right: 20px;
    }
    
    .stat-content {
      .stat-value {
        font-size: 32px;
        font-weight: 700;
        color: #303133;
      }
      
      .stat-label {
        font-size: 14px;
        color: #909399;
        margin-top: 4px;
      }
    }
    
    &.total .stat-icon {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: #fff;
    }
    
    &.healthy .stat-icon {
      background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
      color: #fff;
    }
    
    &.unhealthy .stat-icon {
      background: linear-gradient(135deg, #eb3349 0%, #f45c43 100%);
      color: #fff;
    }
    
    &.unknown .stat-icon {
      background: linear-gradient(135deg, #bdc3c7 0%, #2c3e50 100%);
      color: #fff;
    }
  }
  
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  
  .latency-good { color: #67c23a; }
  .latency-warn { color: #e6a23c; }
  .latency-bad { color: #f56c6c; }
  
  .alerts-card {
    .alert-list {
      max-height: 400px;
      overflow-y: auto;
    }
    
    .alert-item {
      display: flex;
      padding: 12px;
      border-radius: 8px;
      margin-bottom: 8px;
      background: #f5f7fa;
      
      &.firing {
        background: #fef0f0;
      }
      
      &.resolved {
        background: #f0f9eb;
      }
      
      .alert-icon {
        margin-right: 12px;
        padding-top: 2px;
      }
      
      .alert-content {
        flex: 1;
        
        .alert-name {
          font-weight: 500;
          color: #303133;
        }
        
        .alert-message {
          font-size: 12px;
          color: #909399;
          margin-top: 4px;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
        
        .alert-time {
          font-size: 12px;
          color: #c0c4cc;
          margin-top: 4px;
        }
      }
    }
  }
}
</style>

