<script setup>
import { ref, reactive, watch, computed } from 'vue'
import { Dialog } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Card, CardContent } from '@/components/ui/card'
import { createTarget, updateTarget, testProbe } from '@/api/probe'
import { 
  Database, 
  Server, 
  HardDrive, 
  Radio, 
  Globe, 
  Network,
  Loader2,
  CheckCircle2,
  XCircle
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

const form = reactive({
  name: '',
  type: 'http',
  enabled: true,
  interval_seconds: 30,
  timeout_seconds: 5,
  config: {}
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
      { key: 'brokers', label: 'Broker地址', type: 'text', placeholder: 'localhost:9092' },
      { key: 'username', label: '用户名', type: 'text', placeholder: '' },
      { key: 'password', label: '密码', type: 'password', placeholder: '' }
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
    ]
  }
  return fields[form.type] || []
})

const typeOptions = computed(() => 
  props.probeTypes.map(t => ({ value: t.value, label: t.label }))
)

const isEdit = computed(() => !!props.editData)

// 监听 editData 变化
watch(() => props.editData, (val) => {
  if (val) {
    form.name = val.name
    form.type = val.type
    form.enabled = val.enabled
    form.interval_seconds = val.interval_seconds
    form.timeout_seconds = val.timeout_seconds
    form.config = { ...val.config }
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
    const payload = {
      name: form.name,
      type: form.type,
      enabled: form.enabled,
      interval_seconds: form.interval_seconds,
      timeout_seconds: form.timeout_seconds,
      config: form.config
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
  tcp: Network
}
</script>

<template>
  <Dialog
    :open="open"
    @update:open="emit('update:open', $event)"
    :title="isEdit ? '编辑监控服务' : '新增监控服务'"
    description="配置监控目标的连接信息和探测参数"
  >
    <form @submit.prevent="handleSubmit" class="space-y-4 mt-4">
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
      </div>
      
      <!-- 配置字段 -->
      <Card class="bg-muted/50">
        <CardContent class="p-4 space-y-3">
          <h4 class="text-sm font-medium flex items-center gap-2">
            <component :is="typeIcons[form.type]" class="w-4 h-4" />
            连接配置
          </h4>
          
          <div class="grid grid-cols-2 gap-3">
            <div v-for="field in configFields" :key="field.key" :class="field.key === 'url' ? 'col-span-2' : ''">
              <label class="text-xs text-muted-foreground mb-1 block">{{ field.label }}</label>
              <Select
                v-if="field.type === 'select'"
                v-model="form.config[field.key]"
                :options="field.options.map(o => ({ value: o, label: o }))"
                class="w-full"
              />
              <Input
                v-else
                v-model="form.config[field.key]"
                :type="field.type"
                :placeholder="field.placeholder"
              />
            </div>
          </div>
        </CardContent>
      </Card>
      
      <!-- 探测参数 -->
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="text-sm font-medium mb-2 block">探测间隔 (秒)</label>
          <Input v-model.number="form.interval_seconds" type="number" min="5" />
        </div>
        <div>
          <label class="text-sm font-medium mb-2 block">超时时间 (秒)</label>
          <Input v-model.number="form.timeout_seconds" type="number" min="1" />
        </div>
      </div>
      
      <!-- 测试结果 -->
      <div v-if="testResult" :class="[
        'p-3 rounded-lg flex items-center gap-2 text-sm',
        testResult.success ? 'bg-success/10 text-success' : 'bg-destructive/10 text-destructive'
      ]">
        <CheckCircle2 v-if="testResult.success" class="w-4 h-4" />
        <XCircle v-else class="w-4 h-4" />
        <span>{{ testResult.message }}</span>
        <span v-if="testResult.latency" class="ml-auto font-mono">{{ testResult.latency }}ms</span>
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

