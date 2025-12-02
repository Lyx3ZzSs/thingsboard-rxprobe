<template>
  <div class="target-detail" v-loading="loading">
    <el-page-header @back="$router.back()">
      <template #content>
        <div class="page-title">
          <el-tag :type="getStatusType(target.status)" effect="dark" round>
            {{ getStatusText(target.status) }}
          </el-tag>
          <span>{{ target.name }}</span>
        </div>
      </template>
    </el-page-header>
    
    <el-row :gutter="20" class="content-row">
      <!-- 基本信息 -->
      <el-col :span="8">
        <el-card class="info-card">
          <template #header>
            <span>基本信息</span>
          </template>
          
          <el-descriptions :column="1" border>
            <el-descriptions-item label="目标名称">{{ target.name }}</el-descriptions-item>
            <el-descriptions-item label="组件类型">{{ getTypeLabel(target.type) }}</el-descriptions-item>
            <el-descriptions-item label="探测间隔">{{ target.interval_seconds }}s</el-descriptions-item>
            <el-descriptions-item label="超时时间">{{ target.timeout_seconds }}s</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="target.enabled ? 'success' : 'info'" size="small">
                {{ target.enabled ? '已启用' : '已禁用' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="最后检查">{{ formatTime(target.last_check_at) }}</el-descriptions-item>
            <el-descriptions-item label="响应时间">
              <span :class="getLatencyClass(target.last_latency_ms)">
                {{ target.last_latency_ms ? `${target.last_latency_ms}ms` : '-' }}
              </span>
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
        
        <!-- 统计信息 -->
        <el-card class="stats-card">
          <template #header>
            <span>24小时统计</span>
          </template>
          
          <div class="stats-item">
            <span class="stats-label">成功率</span>
            <el-progress
              :percentage="stats.success_rate_24h || 0"
              :color="getSuccessRateColor(stats.success_rate_24h)"
              :stroke-width="12"
              :format="(val) => val.toFixed(2) + '%'"
            />
          </div>
          
          <div class="stats-item">
            <span class="stats-label">平均延迟</span>
            <span class="stats-value">{{ (stats.avg_latency_ms_24h || 0).toFixed(1) }}ms</span>
          </div>
        </el-card>
      </el-col>
      
      <!-- 探测结果列表 -->
      <el-col :span="16">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>探测记录</span>
              <el-button type="primary" link @click="loadResults">
                <el-icon><Refresh /></el-icon>
                刷新
              </el-button>
            </div>
          </template>
          
          <el-table :data="results" v-loading="resultsLoading" stripe max-height="500">
            <el-table-column label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.success ? 'success' : 'danger'" size="small">
                  {{ row.success ? '成功' : '失败' }}
                </el-tag>
              </template>
            </el-table-column>
            
            <el-table-column label="响应时间" width="100">
              <template #default="{ row }">
                {{ row.latency_ms }}ms
              </template>
            </el-table-column>
            
            <el-table-column label="消息" prop="message" min-width="200" show-overflow-tooltip />
            
            <el-table-column label="检查时间" width="180">
              <template #default="{ row }">
                {{ formatTime(row.checked_at) }}
              </template>
            </el-table-column>
          </el-table>
          
          <div class="pagination-wrapper">
            <el-pagination
              v-model:current-page="resultsPagination.page"
              v-model:page-size="resultsPagination.size"
              :total="resultsPagination.total"
              :page-sizes="[20, 50, 100]"
              layout="total, sizes, prev, pager, next"
              @change="loadResults"
            />
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getTarget, getTargetResults, getTargetStats } from '@/api/probe'

const route = useRoute()
const targetId = route.params.id

const loading = ref(false)
const resultsLoading = ref(false)
const target = ref({})
const stats = ref({})
const results = ref([])

const resultsPagination = reactive({
  page: 1,
  size: 20,
  total: 0
})

const typeLabels = {
  postgresql: 'PostgreSQL',
  cassandra: 'Cassandra',
  redis: 'Redis',
  kafka: 'Kafka',
  http: 'HTTP',
  tcp: 'TCP'
}

onMounted(() => {
  loadTarget()
  loadResults()
  loadStats()
})

async function loadTarget() {
  loading.value = true
  try {
    const res = await getTarget(targetId)
    target.value = res.data
  } finally {
    loading.value = false
  }
}

async function loadResults() {
  resultsLoading.value = true
  try {
    const res = await getTargetResults(targetId, {
      page: resultsPagination.page,
      size: resultsPagination.size
    })
    results.value = res.data.items || []
    resultsPagination.total = res.data.total
  } finally {
    resultsLoading.value = false
  }
}

async function loadStats() {
  try {
    const res = await getTargetStats(targetId)
    stats.value = res.data
  } catch (e) {
    console.error(e)
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
</script>

<style lang="scss" scoped>
.target-detail {
  .page-title {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  
  .content-row {
    margin-top: 20px;
  }
  
  .info-card {
    margin-bottom: 20px;
  }
  
  .stats-card {
    .stats-item {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 12px 0;
      
      &:not(:last-child) {
        border-bottom: 1px solid #ebeef5;
      }
      
      .stats-label {
        color: #606266;
      }
      
      .stats-value {
        font-size: 18px;
        font-weight: 600;
        color: #303133;
      }
      
      .el-progress {
        width: 60%;
      }
    }
  }
  
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  
  .pagination-wrapper {
    display: flex;
    justify-content: flex-end;
    margin-top: 20px;
  }
  
  .latency-good { color: #67c23a; }
  .latency-warn { color: #e6a23c; }
  .latency-bad { color: #f56c6c; }
}
</style>

