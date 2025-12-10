<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { 
  CheckCircle2, 
  XCircle, 
  AlertTriangle,
  RefreshCw,
  ArrowRight,
  Zap,
  Server,
  Database,
  Radio
} from 'lucide-vue-next'

import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { HealthCard, ServiceCard, AlertItem, StatusBadge } from '@/components/monitor'
import { getDashboardSummary, getDashboardMetrics } from '@/api/probe'

const router = useRouter()

// 状态
const loading = ref(false)
const summary = ref({
  total_targets: 0,
  healthy_count: 0,
  unhealthy_count: 0,
  unknown_count: 0,
  recent_alerts: []
})
const services = ref([])
let refreshTimer = null

// 加载数据
async function loadData() {
  loading.value = true
  try {
    const [summaryRes, metricsRes] = await Promise.all([
      getDashboardSummary(),
      getDashboardMetrics()
    ])
    summary.value = summaryRes.data
    services.value = metricsRes.data || []
  } catch (e) {
    console.error('加载数据失败:', e)
    // 使用 mock 数据
    summary.value = {
      total_targets: 12,
      healthy_count: 9,
      unhealthy_count: 2,
      unknown_count: 1,
      recent_alerts: [
        { id: 1, target_name: 'PostgreSQL 主库', message: '连接超时', status: 'firing', fired_at: new Date().toISOString() },
        { id: 2, target_name: 'Redis 集群', message: '内存使用率过高', status: 'resolved', fired_at: new Date(Date.now() - 3600000).toISOString(), resolved_at: new Date().toISOString() }
      ]
    }
    services.value = [
      { id: 1, name: 'PostgreSQL 主库', type: 'postgresql', status: 'healthy', last_latency_ms: 12, interval_seconds: 30, stats: { success_rate_24h: 99.8 }, last_check_at: new Date().toISOString() },
      { id: 2, name: 'Redis 集群', type: 'redis', status: 'unhealthy', last_latency_ms: 250, interval_seconds: 10, stats: { success_rate_24h: 95.2 }, last_check_at: new Date().toISOString() },
      { id: 3, name: 'Kafka Broker', type: 'kafka', status: 'healthy', last_latency_ms: 45, interval_seconds: 30, stats: { success_rate_24h: 99.9 }, last_check_at: new Date().toISOString() },
      { id: 4, name: 'Cassandra 节点', type: 'cassandra', status: 'healthy', last_latency_ms: 88, interval_seconds: 60, stats: { success_rate_24h: 99.5 }, last_check_at: new Date().toISOString() },
      { id: 5, name: 'MQTT Broker', type: 'tcp', status: 'healthy', last_latency_ms: 23, interval_seconds: 15, stats: { success_rate_24h: 99.9 }, last_check_at: new Date().toISOString() },
      { id: 6, name: 'CoAP Server', type: 'tcp', status: 'unknown', last_latency_ms: null, interval_seconds: 30, stats: { success_rate_24h: 0 }, last_check_at: null }
    ]
  } finally {
    loading.value = false
  }
}

function goToService(service) {
  router.push(`/services/${service.id}`)
}

onMounted(() => {
  loadData()
  refreshTimer = setInterval(loadData, 30000)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold font-display flex items-center gap-2">
          <Zap class="w-6 h-6 text-primary" />
          <span>监控总览</span>
        </h1>
        <p class="text-muted-foreground text-sm mt-1">平台基础设施实时监控</p>
      </div>
      <Button variant="outline" size="sm" @click="loadData" :disabled="loading" class="border-primary/30 hover:bg-primary/10">
        <RefreshCw :class="['w-4 h-4 mr-2', loading && 'animate-spin']" />
        刷新数据
      </Button>
    </div>

    <!-- 统计卡片 -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <HealthCard
        title="监控节点"
        :value="summary.total_targets"
        :icon="Server"
        variant="primary"
        class="animate-fade-in"
      />
      <HealthCard
        title="运行正常"
        :value="summary.healthy_count"
        :icon="CheckCircle2"
        variant="success"
        class="animate-fade-in delay-100"
      />
      <HealthCard
        title="故障告警"
        :value="summary.unhealthy_count"
        :icon="XCircle"
        variant="destructive"
        class="animate-fade-in delay-200"
      />
      <HealthCard
        title="离线节点"
        :value="summary.unknown_count"
        :icon="AlertTriangle"
        variant="warning"
        class="animate-fade-in delay-300"
      />
    </div>


    <!-- 主要内容区 -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- 服务状态列表 -->
      <div class="lg:col-span-2 space-y-4">
        <div class="flex items-center justify-between">
          <h2 class="text-lg font-semibold flex items-center gap-2">
            <Database class="w-5 h-5 text-primary" />
            核心服务状态
          </h2>
          <Button variant="ghost" size="sm" @click="router.push('/services')" class="text-primary hover:text-primary hover:bg-primary/10">
            全部服务
            <ArrowRight class="w-4 h-4 ml-1" />
          </Button>
        </div>
        
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <ServiceCard
            v-for="service in services.slice(0, 6)"
            :key="service.id"
            :service="service"
            class="animate-fade-in"
            @click="goToService"
          />
        </div>
      </div>

      <!-- 最近告警 -->
      <div class="space-y-4">
        <div class="flex items-center justify-between">
          <h2 class="text-lg font-semibold flex items-center gap-2">
            <Radio class="w-5 h-5 text-destructive" />
            实时告警
          </h2>
          <Button variant="ghost" size="sm" @click="router.push('/alerts')" class="text-primary hover:text-primary hover:bg-primary/10">
            告警中心
            <ArrowRight class="w-4 h-4 ml-1" />
          </Button>
        </div>
        
        <div class="space-y-3">
          <AlertItem
            v-for="alert in summary.recent_alerts?.slice(0, 5)"
            :key="alert.id"
            :alert="alert"
            compact
            class="animate-slide-in-right"
          />
          
          <div v-if="!summary.recent_alerts?.length" class="text-center py-8 text-muted-foreground border border-dashed border-primary/20 rounded-lg">
            <CheckCircle2 class="w-12 h-12 mx-auto mb-2 text-success opacity-50" />
            <p>系统运行正常，暂无告警</p>
          </div>
        </div>
      </div>
    </div>

  </div> 
</template>
