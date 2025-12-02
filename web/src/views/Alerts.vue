<template>
  <div class="alerts-page">
    <el-card>
      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-select v-model="filters.status" placeholder="状态" clearable style="width: 120px" @change="loadAlerts">
          <el-option label="告警中" value="firing" />
          <el-option label="已恢复" value="resolved" />
        </el-select>
        
        <el-button type="primary" @click="loadAlerts">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
      
      <!-- 告警列表 -->
      <el-table :data="alerts" v-loading="loading" stripe>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'firing' ? 'danger' : 'success'" effect="dark">
              {{ row.status === 'firing' ? '告警中' : '已恢复' }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="目标名称" prop="target_name" min-width="150" />
        
        <el-table-column label="类型" width="120">
          <template #default="{ row }">
            {{ getTypeLabel(row.target_type) }}
          </template>
        </el-table-column>
        
        <el-table-column label="告警信息" prop="message" min-width="200" show-overflow-tooltip />
        
        <el-table-column label="响应时间" width="100">
          <template #default="{ row }">
            {{ row.latency_ms }}ms
          </template>
        </el-table-column>
        
        <el-table-column label="触发时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.fired_at) }}
          </template>
        </el-table-column>
        
        <el-table-column label="恢复时间" width="180">
          <template #default="{ row }">
            {{ row.resolved_at ? formatTime(row.resolved_at) : '-' }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 'firing'"
              type="warning"
              link
              @click="handleSilence(row)"
            >
              静默
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.size"
          :total="pagination.total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @change="loadAlerts"
        />
      </div>
    </el-card>
    
    <!-- 静默弹窗 -->
    <el-dialog v-model="silenceVisible" title="静默告警" width="400px">
      <el-form label-width="100px">
        <el-form-item label="静默时长">
          <el-select v-model="silenceDuration" style="width: 100%">
            <el-option label="30 分钟" :value="30" />
            <el-option label="1 小时" :value="60" />
            <el-option label="2 小时" :value="120" />
            <el-option label="6 小时" :value="360" />
            <el-option label="12 小时" :value="720" />
            <el-option label="24 小时" :value="1440" />
          </el-select>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="silenceVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmSilence">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getAlerts, silenceAlert } from '@/api/alert'

const loading = ref(false)
const alerts = ref([])
const silenceVisible = ref(false)
const silenceDuration = ref(30)
const silencingAlert = ref(null)

const filters = reactive({
  status: ''
})

const pagination = reactive({
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
  loadAlerts()
})

async function loadAlerts() {
  loading.value = true
  try {
    const res = await getAlerts({
      ...filters,
      page: pagination.page,
      size: pagination.size
    })
    alerts.value = res.data.items || []
    pagination.total = res.data.total
  } finally {
    loading.value = false
  }
}

function handleSilence(row) {
  silencingAlert.value = row
  silenceVisible.value = true
}

async function confirmSilence() {
  try {
    await silenceAlert(silencingAlert.value.target_id, silenceDuration.value)
    ElMessage.success('已静默该告警')
    silenceVisible.value = false
    loadAlerts()
  } catch (e) {
    // error handled by interceptor
  }
}

function getTypeLabel(type) {
  return typeLabels[type] || type
}

function formatTime(time) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}
</script>

<style lang="scss" scoped>
.alerts-page {
  .filter-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
  }
  
  .pagination-wrapper {
    display: flex;
    justify-content: flex-end;
    margin-top: 20px;
  }
}
</style>

