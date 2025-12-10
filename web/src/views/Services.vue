<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { 
  Plus, 
  Search, 
  RefreshCw,
  Filter,
  LayoutGrid,
  List,
  MoreHorizontal,
  Edit,
  Trash2,
  Play,
  Pause
} from 'lucide-vue-next'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Dialog } from '@/components/ui/dialog'
import { 
  Table, 
  TableHeader, 
  TableBody, 
  TableRow, 
  TableHead, 
  TableCell 
} from '@/components/ui/table'
import { ServiceCard, StatusBadge } from '@/components/monitor'
import ServiceForm from '@/components/ServiceForm.vue'
import { getTargets, deleteTarget, updateTarget, getProbeTypes } from '@/api/probe'
import { formatTime, getServiceTypeLabel } from '@/lib/utils'

const router = useRouter()

// 状态
const loading = ref(false)
const services = ref([])
const probeTypes = ref([])
const viewMode = ref('grid') // grid | list
const formVisible = ref(false)
const editingService = ref(null)
const deleteDialogVisible = ref(false)
const deletingService = ref(null)

// 筛选
const filters = reactive({
  keyword: '',
  type: '',
  status: ''
})

// 分页
const pagination = reactive({
  page: 1,
  size: 20,
  total: 0
})

// 选项
const statusOptions = [
  { value: '', label: '全部状态' },
  { value: 'healthy', label: '正常' },
  { value: 'unhealthy', label: '异常' },
  { value: 'unknown', label: '未知' }
]

const typeOptions = computed(() => [
  { value: '', label: '全部类型' },
  ...probeTypes.value.map(t => ({ value: t.value, label: t.label }))
])

// 加载探针类型
async function loadProbeTypes() {
  try {
    const res = await getProbeTypes()
    probeTypes.value = res.data
  } catch (e) {
    // Mock 数据
    probeTypes.value = [
      { value: 'postgresql', label: 'PostgreSQL' },
      { value: 'redis', label: 'Redis' },
      { value: 'kafka', label: 'Kafka' },
      { value: 'cassandra', label: 'Cassandra' },
      { value: 'http', label: 'HTTP' },
      { value: 'tcp', label: 'TCP' }
    ]
  }
}

// 加载服务列表
async function loadServices() {
  loading.value = true
  try {
    const res = await getTargets({
      ...filters,
      page: pagination.page,
      size: pagination.size
    })
    services.value = res.data.items || []
    pagination.total = res.data.total
  } catch (e) {
    // Mock 数据
    services.value = [
      { id: 1, name: 'PostgreSQL 主库', type: 'postgresql', status: 'healthy', enabled: true, last_latency_ms: 12, interval_seconds: 30, timeout_seconds: 5, stats: { success_rate_24h: 99.8 }, last_check_at: new Date().toISOString() },
      { id: 2, name: 'PostgreSQL 从库', type: 'postgresql', status: 'healthy', enabled: true, last_latency_ms: 15, interval_seconds: 30, timeout_seconds: 5, stats: { success_rate_24h: 99.5 }, last_check_at: new Date().toISOString() },
      { id: 3, name: 'Redis 集群', type: 'redis', status: 'unhealthy', enabled: true, last_latency_ms: 250, interval_seconds: 10, timeout_seconds: 3, stats: { success_rate_24h: 95.2 }, last_check_at: new Date().toISOString() },
      { id: 4, name: 'Kafka Broker 1', type: 'kafka', status: 'healthy', enabled: true, last_latency_ms: 45, interval_seconds: 30, timeout_seconds: 10, stats: { success_rate_24h: 99.9 }, last_check_at: new Date().toISOString() },
      { id: 5, name: 'Cassandra 节点', type: 'cassandra', status: 'healthy', enabled: true, last_latency_ms: 88, interval_seconds: 60, timeout_seconds: 15, stats: { success_rate_24h: 99.5 }, last_check_at: new Date().toISOString() },
      { id: 6, name: 'API Gateway', type: 'http', status: 'warning', enabled: true, last_latency_ms: 520, interval_seconds: 15, timeout_seconds: 5, stats: { success_rate_24h: 97.8 }, last_check_at: new Date().toISOString() },
      { id: 7, name: 'MQTT Broker', type: 'tcp', status: 'unknown', enabled: false, last_latency_ms: null, interval_seconds: 30, timeout_seconds: 5, stats: { success_rate_24h: 0 }, last_check_at: null }
    ]
    pagination.total = services.value.length
  } finally {
    loading.value = false
  }
}

// 切换启用状态
async function toggleEnabled(service) {
  try {
    await updateTarget(service.id, { enabled: !service.enabled })
    service.enabled = !service.enabled
  } catch (e) {
    // 恢复状态
    service.enabled = !service.enabled
  }
}

// 删除服务
async function confirmDelete() {
  if (!deletingService.value) return
  try {
    await deleteTarget(deletingService.value.id)
    deleteDialogVisible.value = false
    loadServices()
  } catch (e) {
    console.error('删除失败:', e)
  }
}

// 打开编辑
function editService(service) {
  editingService.value = { ...service }
  formVisible.value = true
}

// 查看详情
function goToDetail(service) {
  router.push(`/services/${service.id}`)
}

// 表单成功回调
function onFormSuccess() {
  loadServices()
}

onMounted(() => {
  loadProbeTypes()
  loadServices()
})
</script>

<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold font-display">监控服务管理</h1>
        <p class="text-muted-foreground text-sm mt-1">管理所有监控目标和探针配置</p>
      </div>
      <Button @click="editingService = null; formVisible = true">
        <Plus class="w-4 h-4 mr-2" />
        新增服务
      </Button>
    </div>

    <!-- 筛选栏 -->
    <Card>
      <CardContent class="p-4">
        <div class="flex flex-wrap items-center gap-4">
          <div class="flex-1 min-w-[200px] max-w-xs">
            <div class="relative">
              <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input
                v-model="filters.keyword"
                placeholder="搜索服务名称..."
                class="pl-9"
                @keyup.enter="loadServices"
              />
            </div>
          </div>
          
          <Select
            v-model="filters.type"
            :options="typeOptions"
            placeholder="组件类型"
            class="w-40"
            @update:modelValue="loadServices"
          />
          
          <Select
            v-model="filters.status"
            :options="statusOptions"
            placeholder="状态"
            class="w-32"
            @update:modelValue="loadServices"
          />
          
          <div class="flex items-center gap-2 ml-auto">
            <Button variant="outline" size="icon" @click="loadServices" :disabled="loading">
              <RefreshCw :class="['w-4 h-4', loading && 'animate-spin']" />
            </Button>
            
            <div class="flex items-center border rounded-md">
              <Button 
                variant="ghost" 
                size="icon"
                :class="viewMode === 'grid' && 'bg-muted'"
                @click="viewMode = 'grid'"
              >
                <LayoutGrid class="w-4 h-4" />
              </Button>
              <Button 
                variant="ghost" 
                size="icon"
                :class="viewMode === 'list' && 'bg-muted'"
                @click="viewMode = 'list'"
              >
                <List class="w-4 h-4" />
              </Button>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- 网格视图 -->
    <div v-if="viewMode === 'grid'" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <ServiceCard
        v-for="service in services"
        :key="service.id"
        :service="service"
        @click="goToDetail"
      />
    </div>

    <!-- 列表视图 -->
    <Card v-else>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead class="w-[80px]">状态</TableHead>
            <TableHead>名称</TableHead>
            <TableHead class="w-[120px]">类型</TableHead>
            <TableHead class="w-[100px]">响应时间</TableHead>
            <TableHead class="w-[100px]">成功率</TableHead>
            <TableHead class="w-[100px]">探测间隔</TableHead>
            <TableHead class="w-[80px]">启用</TableHead>
            <TableHead class="w-[180px]">最后检查</TableHead>
            <TableHead class="w-[100px] text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="service in services" :key="service.id" class="cursor-pointer" @click="goToDetail(service)">
            <TableCell>
              <StatusBadge :status="service.status" size="sm" :show-label="false" />
            </TableCell>
            <TableCell class="font-medium">{{ service.name }}</TableCell>
            <TableCell>
              <Badge variant="secondary">{{ getServiceTypeLabel(service.type) }}</Badge>
            </TableCell>
            <TableCell class="font-mono">
              <span :class="{
                'text-success': service.last_latency_ms < 100,
                'text-warning': service.last_latency_ms >= 100 && service.last_latency_ms < 500,
                'text-destructive': service.last_latency_ms >= 500
              }">
                {{ service.last_latency_ms ? `${service.last_latency_ms}ms` : '-' }}
              </span>
            </TableCell>
            <TableCell class="font-mono">
              {{ service.stats?.success_rate_24h?.toFixed(1) || '0.0' }}%
            </TableCell>
            <TableCell>{{ service.interval_seconds }}s</TableCell>
            <TableCell @click.stop>
              <Switch :model-value="service.enabled" @update:model-value="toggleEnabled(service)" />
            </TableCell>
            <TableCell class="text-muted-foreground text-sm">
              {{ formatTime(service.last_check_at) }}
            </TableCell>
            <TableCell class="text-right" @click.stop>
              <div class="flex items-center justify-end gap-1">
                <Button variant="ghost" size="icon" @click="editService(service)">
                  <Edit class="w-4 h-4" />
                </Button>
                <Button variant="ghost" size="icon" @click="deletingService = service; deleteDialogVisible = true">
                  <Trash2 class="w-4 h-4 text-destructive" />
                </Button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </Card>

    <!-- 空状态 -->
    <div v-if="!loading && services.length === 0" class="text-center py-12">
      <div class="w-16 h-16 mx-auto mb-4 rounded-full bg-muted flex items-center justify-center">
        <Filter class="w-8 h-8 text-muted-foreground" />
      </div>
      <h3 class="text-lg font-medium mb-2">暂无监控服务</h3>
      <p class="text-muted-foreground mb-4">点击右上角按钮添加第一个监控目标</p>
      <Button @click="editingService = null; formVisible = true">
        <Plus class="w-4 h-4 mr-2" />
        新增服务
      </Button>
    </div>

    <!-- 新增/编辑表单 -->
    <ServiceForm
      v-model:open="formVisible"
      :edit-data="editingService"
      :probe-types="probeTypes"
      @success="onFormSuccess"
    />

    <!-- 删除确认 -->
    <Dialog
      v-model:open="deleteDialogVisible"
      title="删除确认"
      description="确定要删除这个监控服务吗？此操作不可撤销。"
    >
      <div class="flex justify-end gap-3 mt-4">
        <Button variant="outline" @click="deleteDialogVisible = false">取消</Button>
        <Button variant="destructive" @click="confirmDelete">确认删除</Button>
      </div>
    </Dialog>
  </div>
</template>

