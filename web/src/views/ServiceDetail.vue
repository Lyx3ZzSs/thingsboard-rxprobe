<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { 
  ArrowLeft, 
  RefreshCw, 
  Edit, 
  Trash2,
  Clock,
  Activity,
  Zap,
  CheckCircle2,
  AlertTriangle,
  Cpu,
  HardDrive,
  Gauge
} from 'lucide-vue-next'

import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Dialog } from '@/components/ui/dialog'
import { Progress } from '@/components/ui/progress'
import { StatusBadge, MetricChart, HealthCard, AlertItem } from '@/components/monitor'
import { 
  Table, 
  TableHeader, 
  TableBody, 
  TableRow, 
  TableHead, 
  TableCell 
} from '@/components/ui/table'
import ServiceForm from '@/components/ServiceForm.vue'
import { getTarget, getTargetResults, deleteTarget, getProbeTypes } from '@/api/probe'
import { getNotifiers } from '@/api/notifier'
import { formatTime, getServiceTypeLabel, formatDuration, formatLatency } from '@/lib/utils'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const service = ref(null)
const results = ref([])
const probeTypes = ref([])
const notifiers = ref([])
const formVisible = ref(false)
const deleteDialogVisible = ref(false)
let refreshTimer = null

// 图表数据
const chartLabels = ref([])
const latencyData = ref([])

// 统计数据
const stats = computed(() => {
  if (!results.value.length) {
    return {
      avgLatency: 0,
      maxLatency: 0,
      minLatency: 0,
      successRate: 0
    }
  }
  
  const latencies = results.value.map(r => r.latency_ms).filter(l => l > 0)
  const successCount = results.value.filter(r => r.success).length
  
  return {
    avgLatency: latencies.length ? Math.round(latencies.reduce((a, b) => a + b, 0) / latencies.length) : 0,
    maxLatency: latencies.length ? Math.max(...latencies) : 0,
    minLatency: latencies.length ? Math.min(...latencies) : 0,
    successRate: results.value.length ? (successCount / results.value.length * 100).toFixed(1) : 0
  }
})

// CPU 监控专用统计
const cpuStats = computed(() => {
  if (!results.value.length || service.value?.type !== 'cpu') {
    return {
      avgCpu: 0,
      maxCpu: 0,
      minCpu: 0,
      currentCpu: 0,
      cpuCores: 0,
      threshold: 0
    }
  }
  
  const cpuValues = results.value
    .filter(r => r.metrics && r.metrics.cpu_percent)
    .map(r => parseFloat(r.metrics.cpu_percent))
  
  const latestResult = results.value[0]
  
  return {
    avgCpu: cpuValues.length ? (cpuValues.reduce((a, b) => a + b, 0) / cpuValues.length).toFixed(2) : 0,
    maxCpu: cpuValues.length ? Math.max(...cpuValues).toFixed(2) : 0,
    minCpu: cpuValues.length ? Math.min(...cpuValues).toFixed(2) : 0,
    currentCpu: latestResult?.metrics?.cpu_percent || 0,
    cpuCores: latestResult?.metrics?.cpu_cores || 0,
    threshold: latestResult?.metrics?.threshold || 0
  }
})

// 是否为 CPU 监控
const isCpuMonitor = computed(() => service.value?.type === 'cpu')

// 获取服务绑定的通知渠道
const serviceNotifiers = computed(() => {
  if (!service.value || !service.value.notify_channel_ids || !notifiers.value.length) {
    return []
  }
  const ids = service.value.notify_channel_ids || []
  return notifiers.value.filter(n => ids.includes(n.id))
})

// 加载服务详情
async function loadService() {
  loading.value = true
  try {
    const [serviceRes, resultsRes, typesRes, notifiersRes] = await Promise.all([
      getTarget(route.params.id),
      getTargetResults(route.params.id, { page: 1, size: 100 }),
      getProbeTypes(),
      getNotifiers()
    ])
    
    service.value = serviceRes.data
    results.value = resultsRes.data?.items || []
    probeTypes.value = typesRes.data
    notifiers.value = notifiersRes.data || []
    
    // 处理图表数据
    updateChartData()
  } catch (e) {
    console.error('加载失败:', e)
    // Mock 数据
    service.value = {
      id: route.params.id,
      name: 'PostgreSQL 主库',
      type: 'postgresql',
      status: 'healthy',
      enabled: true,
      last_latency_ms: 12,
      interval_seconds: 30,
      timeout_seconds: 5,
      last_check_at: new Date().toISOString(),
      config: {
        host: '192.168.1.100',
        port: 5432,
        database: 'thingsboard',
        username: 'postgres'
      },
      stats: {
        success_rate_24h: 99.8,
        avg_latency_24h: 15,
        total_checks_24h: 2880
      }
    }
    
    // Mock 历史数据
    const mockResults = []
    const now = Date.now()
    for (let i = 0; i < 50; i++) {
      mockResults.push({
        id: i,
        success: Math.random() > 0.05,
        latency_ms: Math.floor(Math.random() * 50) + 5,
        message: Math.random() > 0.05 ? 'OK' : '连接超时',
        checked_at: new Date(now - i * 30000).toISOString()
      })
    }
    results.value = mockResults
    
    probeTypes.value = [
      { value: 'postgresql', label: 'PostgreSQL' },
      { value: 'redis', label: 'Redis' },
      { value: 'kafka', label: 'Kafka' }
    ]
    
    updateChartData()
  } finally {
    loading.value = false
  }
}

function updateChartData() {
  const labels = []
  const data = []
  
  // 取最近 24 个点
  const recent = results.value.slice(0, 24).reverse()
  
  if (isCpuMonitor.value) {
    // CPU 监控：提取 CPU 占用率数据
    recent.forEach(r => {
      const time = new Date(r.checked_at)
      labels.push(`${time.getHours()}:${String(time.getMinutes()).padStart(2, '0')}`)
      const cpuPercent = r.metrics?.cpu_percent ? parseFloat(r.metrics.cpu_percent) : 0
      data.push(cpuPercent)
    })
  } else {
    // 其他监控：提取延迟数据
    recent.forEach(r => {
      const time = new Date(r.checked_at)
      labels.push(`${time.getHours()}:${String(time.getMinutes()).padStart(2, '0')}`)
      data.push(r.latency_ms)
    })
  }
  
  chartLabels.value = labels
  latencyData.value = data
}

// 删除服务
async function confirmDelete() {
  try {
    await deleteTarget(service.value.id)
    router.push('/services')
  } catch (e) {
    console.error('删除失败:', e)
  }
}

// 编辑成功回调
function onFormSuccess() {
  loadService()
}

onMounted(() => {
  loadService()
  refreshTimer = setInterval(loadService, 30000)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<template>
  <div class="space-y-6">
    <!-- 返回和操作栏 -->
    <div class="flex items-center justify-between">
      <Button variant="ghost" @click="router.push('/services')">
        <ArrowLeft class="w-4 h-4 mr-2" />
        返回列表
      </Button>
      
      <div class="flex items-center gap-2">
        <Button variant="outline" size="sm" @click="loadService" :disabled="loading">
          <RefreshCw :class="['w-4 h-4 mr-2', loading && 'animate-spin']" />
          刷新
        </Button>
        <Button variant="outline" size="sm" @click="formVisible = true">
          <Edit class="w-4 h-4 mr-2" />
          编辑
        </Button>
        <Button variant="destructive" size="sm" @click="deleteDialogVisible = true">
          <Trash2 class="w-4 h-4 mr-2" />
          删除
        </Button>
      </div>
    </div>

    <!-- 服务信息卡片 -->
    <Card v-if="service">
      <CardContent class="p-6">
        <div class="flex items-start justify-between">
          <div class="flex items-center gap-4">
            <div :class="[
              'w-16 h-16 rounded-xl flex items-center justify-center',
              service.status === 'healthy' ? 'bg-success/10' : 'bg-destructive/10'
            ]">
              <Activity :class="[
                'w-8 h-8',
                service.status === 'healthy' ? 'text-success' : 'text-destructive'
              ]" />
            </div>
            <div>
              <h1 class="text-2xl font-bold font-display">{{ service.name }}</h1>
              <div class="flex items-center gap-3 mt-2">
                <Badge variant="secondary">{{ getServiceTypeLabel(service.type) }}</Badge>
                <StatusBadge :status="service.status" />
                <Badge v-if="service.group" variant="outline">{{ service.group }}</Badge>
                <span class="text-sm text-muted-foreground">
                  间隔: {{ service.interval_seconds }}s
                </span>
              </div>
              <div v-if="serviceNotifiers.length > 0" class="flex items-center gap-2 mt-2">
                <span class="text-xs text-muted-foreground">通知渠道:</span>
                <div class="flex flex-wrap gap-1">
                  <Badge 
                    v-for="notifier in serviceNotifiers" 
                    :key="notifier.id" 
                    variant="outline" 
                    class="text-xs"
                  >
                    {{ notifier.name }}
                  </Badge>
                </div>
              </div>
            </div>
          </div>
          
          <div class="text-right">
            <p v-if="isCpuMonitor" class="text-3xl font-bold font-mono">
              {{ cpuStats.currentCpu }}
              <span class="text-lg text-muted-foreground">%</span>
            </p>
            <p v-else class="text-3xl font-bold font-mono">
              {{ service.last_latency_ms || '-' }}
              <span class="text-lg text-muted-foreground">ms</span>
            </p>
            <p class="text-sm text-muted-foreground mt-1">
              最后检查: {{ formatTime(service.last_check_at) }}
            </p>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- 统计卡片 - CPU 监控 -->
    <div v-if="isCpuMonitor" class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <HealthCard
        title="当前 CPU 占用率"
        :value="`${cpuStats.currentCpu}%`"
        :icon="Cpu"
        :variant="parseFloat(cpuStats.currentCpu) > cpuStats.threshold ? 'destructive' : 'success'"
      />
      <HealthCard
        title="平均 CPU (24h)"
        :value="`${cpuStats.avgCpu}%`"
        :icon="Activity"
        variant="primary"
      />
      <HealthCard
        title="峰值 CPU (24h)"
        :value="`${cpuStats.maxCpu}%`"
        :icon="AlertTriangle"
        :variant="parseFloat(cpuStats.maxCpu) > cpuStats.threshold ? 'warning' : 'primary'"
      />
      <HealthCard
        title="CPU 核心数"
        :value="`${cpuStats.cpuCores}`"
        :icon="HardDrive"
        variant="secondary"
      />
    </div>
    
    <!-- 统计卡片 - 其他监控类型 -->
    <div v-else class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <HealthCard
        title="成功率 (24h)"
        :value="`${stats.successRate}%`"
        :icon="CheckCircle2"
        variant="success"
      />
      <HealthCard
        title="平均响应"
        :value="`${stats.avgLatency}ms`"
        :icon="Zap"
        variant="primary"
      />
      <HealthCard
        title="最大响应"
        :value="`${stats.maxLatency}ms`"
        :icon="AlertTriangle"
        variant="warning"
      />
      <HealthCard
        title="最小响应"
        :value="`${stats.minLatency}ms`"
        :icon="CheckCircle2"
        variant="success"
      />
    </div>

    <!-- CPU 阈值信息卡片 -->
    <Card v-if="isCpuMonitor && cpuStats.threshold > 0" class="bg-primary/5 border-primary/20">
      <CardContent class="p-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-lg bg-primary/20 flex items-center justify-center">
              <Gauge class="w-5 h-5 text-primary" />
            </div>
            <div>
              <p class="text-sm font-medium text-foreground">告警阈值设置</p>
              <p class="text-xs text-muted-foreground mt-0.5">当 CPU 占用率超过 {{ cpuStats.threshold }}% 时触发告警</p>
            </div>
          </div>
          <div class="text-right">
            <p class="text-2xl font-bold text-primary">{{ cpuStats.threshold }}%</p>
            <p class="text-xs text-muted-foreground">阈值</p>
          </div>
        </div>
        
        <!-- 进度条显示当前 CPU 相对阈值的位置 -->
        <div class="mt-4">
          <div class="flex items-center justify-between mb-2">
            <span class="text-xs text-muted-foreground">当前占用率</span>
            <span class="text-xs font-medium">{{ cpuStats.currentCpu }}% / {{ cpuStats.threshold }}%</span>
          </div>
          <Progress 
            :value="(parseFloat(cpuStats.currentCpu) / cpuStats.threshold) * 100" 
            :class="parseFloat(cpuStats.currentCpu) > cpuStats.threshold ? 'bg-destructive/20' : 'bg-success/20'"
          />
        </div>
      </CardContent>
    </Card>

    <!-- 趋势图表 -->
    <MetricChart
      :title="isCpuMonitor ? 'CPU 占用率趋势' : '响应时间趋势'"
      :labels="chartLabels"
      :data="latencyData"
      :unit="isCpuMonitor ? '%' : 'ms'"
      color="#8b5cf6"
      :height="250"
    />

    <!-- 配置信息 -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- 配置详情 -->
      <Card>
        <CardHeader>
          <CardTitle class="text-base">
            {{ isCpuMonitor ? '监控配置' : '连接配置' }}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <dl class="space-y-3">
            <template v-for="(value, key) in service?.config" :key="key">
              <div class="flex justify-between py-2 border-b border-border/50 last:border-0">
                <dt class="text-muted-foreground capitalize">
                  {{ key === 'threshold' ? 'CPU告警阈值' : 
                     key === 'sample_duration' ? '采样时长' :
                     key }}
                </dt>
                <dd class="font-mono text-sm">
                  {{ key.includes('password') ? '••••••••' : 
                     key === 'threshold' ? value + '%' :
                     key === 'sample_duration' ? value + '秒' :
                     value }}
                </dd>
              </div>
            </template>
          </dl>
        </CardContent>
      </Card>

      <!-- 探测历史 -->
      <Card>
        <CardHeader>
          <CardTitle class="text-base">最近检测记录</CardTitle>
        </CardHeader>
        <CardContent class="max-h-[300px] overflow-y-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead class="w-[80px]">状态</TableHead>
                <TableHead class="w-[80px]">
                  {{ isCpuMonitor ? 'CPU' : '延迟' }}
                </TableHead>
                <TableHead>时间</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="result in results.slice(0, 10)" :key="result.id">
                <TableCell>
                  <StatusBadge :status="result.success ? 'healthy' : 'unhealthy'" size="sm" :show-label="false" />
                </TableCell>
                <TableCell class="font-mono text-sm">
                  <span v-if="isCpuMonitor">{{ result.metrics?.cpu_percent || '0' }}%</span>
                  <span v-else>{{ formatLatency(result.latency_ms) }}</span>
                </TableCell>
                <TableCell class="text-muted-foreground text-sm">
                  {{ formatTime(result.checked_at) }}
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>

    <!-- 编辑表单 -->
    <ServiceForm
      v-model:open="formVisible"
      :edit-data="service"
      :probe-types="probeTypes"
      @success="onFormSuccess"
    />

    <!-- 删除确认 -->
    <Dialog
      v-model:open="deleteDialogVisible"
      title="删除确认"
      description="确定要删除这个监控服务吗？此操作不可撤销，相关的告警记录也会被删除。"
    >
      <div class="flex justify-end gap-3 mt-4">
        <Button variant="outline" @click="deleteDialogVisible = false">取消</Button>
        <Button variant="destructive" @click="confirmDelete">确认删除</Button>
      </div>
    </Dialog>
  </div>
</template>

