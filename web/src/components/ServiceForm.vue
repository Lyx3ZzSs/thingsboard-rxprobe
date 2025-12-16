<script setup>
import { ref, reactive, watch, computed, onMounted } from 'vue'
import { Dialog } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Card, CardContent } from '@/components/ui/card'
import { createTarget, updateTarget, testProbe, getTargets } from '@/api/probe'
import { getNotifiers } from '@/api/notifier'
import { formatLatency } from '@/lib/utils'
import { 
  Database, 
  Server, 
  HardDrive, 
  Radio, 
  Globe, 
  Network,
  Loader2,
  CheckCircle2,
  XCircle,
  Cpu
} from 'lucide-vue-next'

const props = defineProps({
  open: Boolean,
  editData: Object,
  probeTypes: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['update:open', 'success'])

const loading = ref(false)
const testing = ref(false)
const testResult = ref(null)
const notifiers = ref([])
const groups = ref([])

const form = reactive({
  name: '',
  type: 'http',
  enabled: true,
  interval_seconds: 30,
  timeout_seconds: 5,
  config: {},
  group: '',
  notify_channel_ids: []
})

// 根据类型显示不同的配置字段
const configFields = computed(() => {
  const fields = {
    postgresql: [
      { key: 'host', label: '主机地址', type: 'text', placeholder: 'localhost' },
      { key: 'port', label: '端口', type: 'number', placeholder: '5432' },
      { key: 'username', label: '用户名', type: 'text', placeholder: 'postgres' },
      { key: 'password', label: '密码', type: 'password', placeholder: '' },
      { key: 'database', label: '数据库', type: 'text', placeholder: 'postgres' }
    ],
    redis: [
      { key: 'host', label: '主机地址', type: 'text', placeholder: 'localhost' },
      { key: 'port', label: '端口', type: 'number', placeholder: '6379' },
      { key: 'password', label: '密码', type: 'password', placeholder: '' },
      { key: 'db', label: 'DB索引', type: 'number', placeholder: '0' }
    ],
    kafka: [
      { key: 'brokers', label: 'Broker地址', type: 'text', placeholder: 'kafka1:9092,kafka2:9092', hint: '多个地址用逗号分隔' }
    ],
    cassandra: [
      { key: 'hosts', label: '节点地址', type: 'text', placeholder: 'localhost' },
      { key: 'port', label: '端口', type: 'number', placeholder: '9042' },
      { key: 'username', label: '用户名', type: 'text', placeholder: '' },
      { key: 'password', label: '密码', type: 'password', placeholder: '' },
      { key: 'keyspace', label: 'Keyspace', type: 'text', placeholder: '' }
    ],
    http: [
      { key: 'url', label: 'URL地址', type: 'text', placeholder: 'https://example.com/health' },
      { key: 'method', label: '请求方法', type: 'select', options: ['GET', 'POST', 'HEAD'] },
      { key: 'expected_status', label: '期望状态码', type: 'number', placeholder: '200' }
    ],
    tcp: [
      { key: 'host', label: '主机地址', type: 'text', placeholder: 'localhost' },
      { key: 'port', label: '端口', type: 'number', placeholder: '80' }
    ],
    ping: [
      { key: 'host', label: '主机地址', type: 'text', placeholder: 'example.com 或 192.168.1.1', hint: '支持域名或 IP 地址' }
    ],
    cpu: [
      { key: 'threshold', label: 'CPU告警阈值 (%)', type: 'number', placeholder: '80', hint: '当 CPU 占用率超过此值时触发告警 (0-100)' },
      { key: 'sample_duration', label: '采样时长 (秒)', type: 'number', placeholder: '3', hint: 'CPU 使用率采样时长，建议 1-10 秒' }
    ]
  }
  return fields[form.type] || []
})

const typeOptions = computed(() => 
  props.probeTypes.map(t => ({ value: t.value, label: t.label }))
)

const isEdit = computed(() => !!props.editData)

// 通知渠道选项
const notifierOptions = computed(() => 
  notifiers.value
    .filter(n => n.enabled)
    .map(n => ({ 
      value: n.id, 
      label: `${n.name} (${n.type})` 
    }))
)

// 分组选项
const groupOptions = computed(() => 
  groups.value.map(g => ({ value: g, label: g }))
)

// 加载通知渠道
async function loadNotifiers() {
  try {
    const res = await getNotifiers()
    notifiers.value = res.data || []
  } catch (e) {
    console.error('加载通知渠道失败:', e)
  }
}

// 加载已有分组
async function loadGroups() {
  try {
    const res = await getTargets({ page: 1, size: 1000 })
    const items = res.data.items || []
    // 提取所有不为空的分组
    const uniqueGroups = [...new Set(items.map(item => item.group).filter(Boolean))]
    groups.value = uniqueGroups.sort()
  } catch (e) {
    console.error('加载分组失败:', e)
  }
}

// 监听 editData 变化
watch(() => props.editData, (val) => {
  if (val) {
    form.name = val.name
    form.type = val.type
    form.enabled = val.enabled
    form.interval_seconds = val.interval_seconds
    form.timeout_seconds = val.timeout_seconds
    form.config = { ...val.config }
    form.group = val.group || ''
    form.notify_channel_ids = val.notify_channel_ids || []
  } else {
    resetForm()
  }
}, { immediate: true })

// 监听类型变化，重置配置
watch(() => form.type, () => {
  if (!props.editData) {
    form.config = {}
  }
  testResult.value = null
})

function resetForm() {
  form.name = ''
  form.type = 'http'
  form.enabled = true
  form.interval_seconds = 30
  form.timeout_seconds = 5
  form.config = {}
  form.group = ''
  form.notify_channel_ids = []
  testResult.value = null
}

// 测试连接
async function handleTest() {
  testing.value = true
  testResult.value = null
  
  try {
    const res = await testProbe({
      type: form.type,
      config: form.config,
      timeout_seconds: form.timeout_seconds
    })
    testResult.value = {
      success: res.data.success,
      message: res.data.message,
      latency: res.data.latency_ms
    }
  } catch (e) {
    testResult.value = {
      success: false,
      message: e.response?.data?.message || '测试失败'
    }
  } finally {
    testing.value = false
  }
}

// 提交表单
async function handleSubmit() {
  if (!form.name) return
  
  loading.value = true
  try {
    // 根据类型设置合理的间隔和超时
    let intervalSeconds = form.interval_seconds
    let timeoutSeconds = form.timeout_seconds
    
    if (form.type === 'cpu') {
      // CPU 监控的检测间隔等于采样时长
      const sampleDuration = parseInt(form.config.sample_duration) || 3
      intervalSeconds = sampleDuration
      timeoutSeconds = sampleDuration + 5  // 超时设置宽松一些
    } else if (form.type === 'kafka') {
      // Kafka 集群连接检查：默认30秒检查一次，超时10秒
      intervalSeconds = 30
      timeoutSeconds = 10
    } else if (form.type === 'ping') {
      // Ping 探测：默认30秒检查一次，超时15秒（4次ping，每次3秒，加上网络延迟）
      intervalSeconds = 30
      timeoutSeconds = 15
    }
    
    const payload = {
      name: form.name,
      type: form.type,
      enabled: form.enabled,
      interval_seconds: intervalSeconds,
      timeout_seconds: timeoutSeconds,
      config: form.config,
      group: form.group || '',
      notify_channel_ids: form.notify_channel_ids || []
    }
    
    if (isEdit.value) {
      await updateTarget(props.editData.id, payload)
    } else {
      await createTarget(payload)
    }
    
    emit('update:open', false)
    emit('success')
    resetForm()
  } catch (e) {
    console.error('保存失败:', e)
  } finally {
    loading.value = false
  }
}

const typeIcons = {
  postgresql: Database,
  redis: Server,
  kafka: Radio,
  cassandra: HardDrive,
  http: Globe,
  tcp: Network,
  ping: Network,
  cpu: Cpu
}

// CPU 类型不需要手动配置探测间隔和超时时间
const shouldShowProbeParams = computed(() => {
  return !['cpu'].includes(form.type)
})

// Ping 类型不需要手动配置超时时间（固定为15秒，因为后台固定ping 4次，每次3秒）
const shouldShowTimeout = computed(() => {
  return !['cpu', 'ping'].includes(form.type)
})

// 组件挂载时加载数据
onMounted(() => {
  loadNotifiers()
  loadGroups()
})
</script>

<template>
  <Dialog
    :open="open"
    @update:open="emit('update:open', $event)"
    :title="isEdit ? '编辑监控服务' : '新增监控服务'"
    description="配置监控目标的连接信息和探测参数"
  >
    <form @submit.prevent="handleSubmit" class="space-y-4">
      <!-- 基本信息 -->
      <div class="grid grid-cols-2 gap-4">
        <div class="col-span-2">
          <label class="text-sm font-medium mb-2 block">服务名称</label>
          <Input v-model="form.name" placeholder="例如：PostgreSQL 主库" />
        </div>
        
        <div>
          <label class="text-sm font-medium mb-2 block">探针类型</label>
          <Select
            v-model="form.type"
            :options="typeOptions"
            :disabled="isEdit"
            class="w-full"
          />
        </div>
        
        <div class="flex items-center justify-between pt-6">
          <label class="text-sm font-medium">启用监控</label>
          <Switch v-model="form.enabled" />
        </div>
        
        <div class="col-span-2">
          <label class="text-sm font-medium mb-2 block">
            分组
            <span class="text-xs text-muted-foreground font-normal ml-1">（选填，用于组织和管理监控服务）</span>
          </label>
          <Input 
            v-model="form.group" 
            placeholder="例如：生产环境、测试环境" 
            list="groups-list"
          />
          <datalist id="groups-list">
            <option v-for="group in groups" :key="group" :value="group" />
          </datalist>
        </div>
        
        <div class="col-span-2">
          <label class="text-sm font-medium mb-2 block">
            通知渠道
            <span class="text-xs text-muted-foreground font-normal ml-1">（选填，发生告警时通过选定的渠道发送通知）</span>
          </label>
          <div class="space-y-2">
            <div v-if="notifierOptions.length === 0" class="text-sm text-muted-foreground p-3 bg-muted/50 rounded-md">
              暂无可用的通知渠道，请先在
              <router-link to="/notifiers" class="text-primary underline">通知渠道管理</router-link>
              中添加
            </div>
            <div v-else class="flex flex-wrap gap-2">
              <label 
                v-for="notifier in notifierOptions" 
                :key="notifier.value"
                class="flex items-center gap-2 px-3 py-2 border rounded-md cursor-pointer hover:bg-muted/50 transition-colors"
                :class="form.notify_channel_ids.includes(notifier.value) ? 'bg-primary/10 border-primary' : ''"
              >
                <input
                  type="checkbox"
                  :value="notifier.value"
                  v-model="form.notify_channel_ids"
                  class="rounded border-gray-300"
                />
                <span class="text-sm">{{ notifier.label }}</span>
              </label>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 配置字段 -->
      <Card class="bg-muted/50">
        <CardContent class="p-4 space-y-3">
          <h4 class="text-sm font-medium flex items-center gap-2">
            <component :is="typeIcons[form.type]" v-if="typeIcons[form.type]" class="w-4 h-4" />
            {{ form.type === 'cpu' ? '监控配置' : '连接配置' }}
          </h4>
          
          <div class="grid grid-cols-2 gap-3">
            <template v-for="field in configFields" :key="field.key">
              <div 
                v-if="!field.showWhen || form.config[field.showWhen]"
                :class="field.key === 'url' ? 'col-span-2' : ''"
              >
                <label class="text-xs text-muted-foreground mb-1 block">
                  {{ field.label }}
                  <span v-if="field.hint" class="text-xs text-muted-foreground/70 block mt-0.5">{{ field.hint }}</span>
                </label>
                <!-- Switch 类型 -->
                <div v-if="field.type === 'switch'" class="flex items-center h-9">
                  <Switch v-model="form.config[field.key]" />
                </div>
                <!-- Select 类型 -->
                <Select
                  v-else-if="field.type === 'select'"
                  v-model="form.config[field.key]"
                  :options="field.options.map(o => ({ value: o, label: o }))"
                  class="w-full"
                />
                <!-- Input 类型 -->
                <Input
                  v-else
                  v-model="form.config[field.key]"
                  :type="field.type"
                  :placeholder="field.placeholder"
                />
              </div>
            </template>
          </div>
        </CardContent>
      </Card>
      
      <!-- 探测参数 (CPU监控不显示) -->
      <div v-if="shouldShowProbeParams" class="grid grid-cols-2 gap-4">
        <div>
          <label class="text-sm font-medium mb-2 block">探测间隔 (秒)</label>
          <Input v-model.number="form.interval_seconds" type="number" min="5" />
        </div>
        <div v-if="shouldShowTimeout">
          <label class="text-sm font-medium mb-2 block">超时时间 (秒)</label>
          <Input v-model.number="form.timeout_seconds" type="number" min="1" />
        </div>
      </div>
      
      <!-- CPU 监控的特殊说明 -->
      <div v-if="form.type === 'cpu'" class="p-3 rounded-lg bg-primary/10 border border-primary/20">
        <p class="text-sm text-foreground">
          <span class="font-medium">💡 提示：</span>
          CPU 监控会实时采样当前服务器的 CPU 占用率。采样时长即为检测耗时，无需额外设置探测间隔。
        </p>
      </div>
      
      <!-- Ping 监控的特殊说明 -->
      <div v-if="form.type === 'ping'" class="p-3 rounded-lg bg-primary/10 border border-primary/20">
        <p class="text-sm text-foreground">
          <span class="font-medium">💡 提示：</span>
          Ping 探测固定发送 4 个数据包，每个数据包超时时间为 3 秒，总超时时间固定为 15 秒，无需手动设置。
        </p>
      </div>
      
      <!-- Kafka 监控的特殊说明 -->
      <!-- <div v-if="form.type === 'kafka'" class="p-3 rounded-lg bg-primary/10 border border-primary/20">
        <p class="text-sm text-foreground">
          <span class="font-medium">💡 提示：</span>
          Kafka 监控将检查集群连接可用性，验证 Broker 是否在线。
        </p>
      </div> -->
      
      <!-- 测试结果 -->
      <div v-if="testResult" :class="[
        'p-3 rounded-lg flex items-center gap-2 text-sm',
        testResult.success ? 'bg-success/10 text-success' : 'bg-destructive/10 text-destructive'
      ]">
        <CheckCircle2 v-if="testResult.success" class="w-4 h-4" />
        <XCircle v-else class="w-4 h-4" />
        <span>{{ testResult.message }}</span>
        <span v-if="testResult.latency" class="ml-auto font-mono">{{ formatLatency(testResult.latency) }}</span>
      </div>
      
      <!-- 操作按钮 -->
      <div class="flex justify-between pt-2">
        <Button 
          type="button" 
          variant="outline" 
          @click="handleTest"
          :disabled="testing"
        >
          <Loader2 v-if="testing" class="w-4 h-4 mr-2 animate-spin" />
          测试连接
        </Button>
        
        <div class="flex gap-2">
          <Button type="button" variant="outline" @click="emit('update:open', false)">
            取消
          </Button>
          <Button type="submit" :loading="loading">
            {{ isEdit ? '保存' : '创建' }}
          </Button>
        </div>
      </div>
    </form>
  </Dialog>
</template>

