<script setup>
import { ref, reactive, onMounted } from 'vue'
import { 
  RefreshCw, 
  BellOff, 
  Filter,
  AlertTriangle,
  CheckCircle2
} from 'lucide-vue-next'

import { Button } from '@/components/ui/button'
import { Select } from '@/components/ui/select'
import { Card, CardContent } from '@/components/ui/card'
import { Dialog } from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { 
  Table, 
  TableHeader, 
  TableBody, 
  TableRow, 
  TableHead, 
  TableCell 
} from '@/components/ui/table'
import { AlertItem, HealthCard } from '@/components/monitor'
import { getAlerts, silenceAlert } from '@/api/alert'
import { formatTime, getServiceTypeLabel } from '@/lib/utils'

// 状态
const loading = ref(false)
const alerts = ref([])
const silenceDialogVisible = ref(false)
const silencingAlert = ref(null)
const silenceDuration = ref(30)

// 统计
const stats = reactive({
  firing: 0,
  resolved: 0,
  total: 0
})

// 筛选
const filters = reactive({
  status: ''
})

// 分页
const pagination = reactive({
  page: 1,
  size: 50,
  total: 0
})

// 选项
const statusOptions = [
  { value: '', label: '全部状态' },
  { value: 'firing', label: '告警中' },
  { value: 'resolved', label: '已恢复' }
]

const durationOptions = [
  { value: 30, label: '30 分钟' },
  { value: 60, label: '1 小时' },
  { value: 120, label: '2 小时' },
  { value: 360, label: '6 小时' },
  { value: 720, label: '12 小时' },
  { value: 1440, label: '24 小时' }
]

// 加载告警列表
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
    
    // 统计
    stats.total = pagination.total
    stats.firing = alerts.value.filter(a => a.status === 'firing').length
    stats.resolved = alerts.value.filter(a => a.status === 'resolved').length
  } catch (e) {
    // Mock 数据
    alerts.value = [
      { 
        id: 1, 
        target_id: 2,
        target_name: 'Redis 集群', 
        target_type: 'redis',
        message: '连接超时，无法获取服务状态', 
        status: 'firing', 
        latency_ms: 3500,
        fired_at: new Date().toISOString() 
      },
      { 
        id: 2, 
        target_id: 5,
        target_name: 'API Gateway', 
        target_type: 'http',
        message: '响应时间超过阈值 (520ms > 500ms)', 
        status: 'firing', 
        latency_ms: 520,
        fired_at: new Date(Date.now() - 1800000).toISOString() 
      },
      { 
        id: 3, 
        target_id: 1,
        target_name: 'PostgreSQL 主库', 
        target_type: 'postgresql',
        message: '连接数超过警告阈值', 
        status: 'resolved', 
        latency_ms: 45,
        fired_at: new Date(Date.now() - 7200000).toISOString(),
        resolved_at: new Date(Date.now() - 3600000).toISOString()
      },
      { 
        id: 4, 
        target_id: 4,
        target_name: 'Kafka Broker 1', 
        target_type: 'kafka',
        message: '消费延迟过高', 
        status: 'resolved', 
        latency_ms: 88,
        fired_at: new Date(Date.now() - 86400000).toISOString(),
        resolved_at: new Date(Date.now() - 82800000).toISOString()
      }
    ]
    pagination.total = alerts.value.length
    stats.total = pagination.total
    stats.firing = alerts.value.filter(a => a.status === 'firing').length
    stats.resolved = alerts.value.filter(a => a.status === 'resolved').length
  } finally {
    loading.value = false
  }
}

// 静默告警
function openSilenceDialog(alert) {
  silencingAlert.value = alert
  silenceDialogVisible.value = true
}

async function confirmSilence() {
  if (!silencingAlert.value) return
  try {
    await silenceAlert(silencingAlert.value.target_id, silenceDuration.value)
    silenceDialogVisible.value = false
    loadAlerts()
  } catch (e) {
    console.error('静默失败:', e)
  }
}

onMounted(() => {
  loadAlerts()
})
</script>

<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold font-display">告警记录</h1>
        <p class="text-muted-foreground text-sm mt-1">查看和管理所有告警事件</p>
      </div>
      <Button variant="outline" size="sm" @click="loadAlerts" :disabled="loading">
        <RefreshCw :class="['w-4 h-4 mr-2', loading && 'animate-spin']" />
        刷新
      </Button>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <HealthCard
        title="告警中"
        :value="stats.firing"
        :icon="AlertTriangle"
        variant="destructive"
        class="animate-fade-in"
      />
      <HealthCard
        title="已恢复"
        :value="stats.resolved"
        :icon="CheckCircle2"
        variant="success"
        class="animate-fade-in delay-100"
      />
      <HealthCard
        title="总计"
        :value="stats.total"
        variant="primary"
        class="animate-fade-in delay-200"
      />
    </div>

    <!-- 筛选和列表 -->
    <Card>
      <CardContent class="p-4">
        <div class="flex items-center justify-between mb-4">
          <Select
            v-model="filters.status"
            :options="statusOptions"
            placeholder="状态筛选"
            class="w-40"
            @update:modelValue="loadAlerts"
          />
          
          <div class="text-sm text-muted-foreground">
            共 {{ pagination.total }} 条告警记录
          </div>
        </div>

        <!-- 告警表格 -->
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead class="w-[100px]">状态</TableHead>
              <TableHead>目标名称</TableHead>
              <TableHead class="w-[120px]">类型</TableHead>
              <TableHead>告警信息</TableHead>
              <TableHead class="w-[100px]">响应时间</TableHead>
              <TableHead class="w-[180px]">触发时间</TableHead>
              <TableHead class="w-[180px]">恢复时间</TableHead>
              <TableHead class="w-[100px] text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="alert in alerts" :key="alert.id">
              <TableCell>
                <Badge :variant="alert.status === 'firing' ? 'destructive' : 'success'">
                  {{ alert.status === 'firing' ? '告警中' : '已恢复' }}
                </Badge>
              </TableCell>
              <TableCell class="font-medium">{{ alert.target_name }}</TableCell>
              <TableCell>
                <Badge variant="secondary">{{ getServiceTypeLabel(alert.target_type) }}</Badge>
              </TableCell>
              <TableCell class="max-w-[300px] truncate text-muted-foreground">
                {{ alert.message }}
              </TableCell>
              <TableCell class="font-mono">
                {{ alert.latency_ms }}ms
              </TableCell>
              <TableCell class="text-muted-foreground text-sm">
                {{ formatTime(alert.fired_at) }}
              </TableCell>
              <TableCell class="text-muted-foreground text-sm">
                {{ alert.resolved_at ? formatTime(alert.resolved_at) : '-' }}
              </TableCell>
              <TableCell class="text-right">
                <Button
                  v-if="alert.status === 'firing'"
                  variant="ghost"
                  size="sm"
                  @click="openSilenceDialog(alert)"
                >
                  <BellOff class="w-4 h-4 mr-1" />
                  静默
                </Button>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>

        <!-- 空状态 -->
        <div v-if="!loading && alerts.length === 0" class="text-center py-12">
          <div class="w-16 h-16 mx-auto mb-4 rounded-full bg-success/10 flex items-center justify-center">
            <CheckCircle2 class="w-8 h-8 text-success" />
          </div>
          <h3 class="text-lg font-medium mb-2">暂无告警</h3>
          <p class="text-muted-foreground">所有服务运行正常，没有告警记录</p>
        </div>
      </CardContent>
    </Card>

    <!-- 静默对话框 -->
    <Dialog
      v-model:open="silenceDialogVisible"
      title="静默告警"
      description="静默期间将不会收到该目标的告警通知"
    >
      <div class="space-y-4 mt-4">
        <div>
          <label class="text-sm font-medium mb-2 block">静默时长</label>
          <Select
            v-model="silenceDuration"
            :options="durationOptions"
            class="w-full"
          />
        </div>
        
        <div class="flex justify-end gap-3">
          <Button variant="outline" @click="silenceDialogVisible = false">取消</Button>
          <Button @click="confirmSilence">确认静默</Button>
        </div>
      </div>
    </Dialog>
  </div>
</template>
