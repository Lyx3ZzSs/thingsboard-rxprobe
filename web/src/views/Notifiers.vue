<script setup>
import { ref, reactive, onMounted } from 'vue'
import { 
  Plus, 
  RefreshCw,
  Edit,
  Trash2,
  Send,
  Bell,
  MessageSquare,
  CheckCircle,
  XCircle
} from 'lucide-vue-next'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Dialog } from '@/components/ui/dialog'
import { Select } from '@/components/ui/select'
import { 
  Table, 
  TableHeader, 
  TableBody, 
  TableRow, 
  TableHead, 
  TableCell 
} from '@/components/ui/table'
import { 
  getNotifiers, 
  createNotifier, 
  updateNotifier, 
  deleteNotifier, 
  testNotifier,
  getNotifierTypes
} from '@/api/notifier'
import { formatTime } from '@/lib/utils'

// 状态
const loading = ref(false)
const notifiers = ref([])
const notifierTypes = ref([])

// 表单状态
const formVisible = ref(false)
const editingNotifier = ref(null)
const formLoading = ref(false)
const testLoading = ref(false)
const testResult = ref(null)

// 删除确认
const deleteDialogVisible = ref(false)
const deletingNotifier = ref(null)

// 错误提示
const errorDialogVisible = ref(false)
const errorMessage = ref('')

// 格式化错误消息
function formatErrorMessage(error) {
  // 如果有响应数据中的消息
  if (error.response?.data?.message) {
    return error.response.data.message
  }
  
  // 根据状态码返回友好提示
  if (error.response?.status) {
    const statusMessages = {
      400: '请求参数错误，请检查输入信息',
      401: '未授权，请重新登录',
      403: '没有权限执行此操作',
      404: '请求的资源不存在',
      409: '数据冲突，可能已存在相同的记录',
      500: '服务器内部错误，请稍后重试',
      502: '网关错误，请稍后重试',
      503: '服务暂时不可用，请稍后重试'
    }
    return statusMessages[error.response.status] || `请求失败 (状态码: ${error.response.status})`
  }
  
  // 网络错误
  if (error.message === 'Network Error') {
    return '网络连接失败，请检查网络设置'
  }
  
  // 其他错误
  return error.message || '操作失败，请重试'
}

// 表单数据
const formData = reactive({
  name: '',
  type: 'wecom',
  webhook_url: '',
  mention_all: true,
  enabled: true,
  description: ''
})

// 类型选项
const typeOptions = [
  { value: 'wecom', label: '企业微信' }
]

// 重置表单
function resetForm() {
  formData.name = ''
  formData.type = 'wecom'
  formData.webhook_url = ''
  formData.mention_all = true
  formData.enabled = true
  formData.description = ''
  testResult.value = null
}

// 加载通知类型
async function loadNotifierTypes() {
  try {
    const res = await getNotifierTypes()
    notifierTypes.value = res.data
  } catch (e) {
    console.error('加载通知类型失败:', e)
  }
}

// 加载通知渠道列表
async function loadNotifiers() {
  loading.value = true
  try {
    const res = await getNotifiers()
    notifiers.value = res.data || []
  } catch (e) {
    console.error('加载通知渠道失败:', e)
    notifiers.value = []
  } finally {
    loading.value = false
  }
}

// 打开新建表单
function openCreateForm() {
  editingNotifier.value = null
  resetForm()
  formVisible.value = true
}

// 打开编辑表单
function openEditForm(notifier) {
  editingNotifier.value = notifier
  formData.name = notifier.name
  formData.type = notifier.type
  formData.webhook_url = notifier.webhook_url
  formData.mention_all = notifier.mention_all !== undefined ? notifier.mention_all : true
  formData.enabled = notifier.enabled
  formData.description = notifier.description || ''
  testResult.value = null
  formVisible.value = true
}

// 关闭表单
function closeForm() {
  formVisible.value = false
  resetForm()
}

// 提交表单
async function submitForm() {
  if (!formData.name || !formData.webhook_url) {
    return
  }

  formLoading.value = true
  try {
    if (editingNotifier.value) {
      await updateNotifier(editingNotifier.value.id, formData)
    } else {
      await createNotifier(formData)
    }
    closeForm()
    loadNotifiers()
  } catch (e) {
    console.error('保存失败:', e)
    errorMessage.value = formatErrorMessage(e)
    errorDialogVisible.value = true
  } finally {
    formLoading.value = false
  }
}

// 测试通知
async function handleTestNotifier() {
  if (!formData.webhook_url) {
    testResult.value = { success: false, message: '请先填写 Webhook URL' }
    return
  }

  testLoading.value = true
  testResult.value = null
  try {
    await testNotifier({
      type: formData.type,
      webhook_url: formData.webhook_url,
      mention_all: formData.mention_all
    })
    testResult.value = { success: true, message: '测试消息发送成功！' }
  } catch (e) {
    testResult.value = { success: false, message: formatErrorMessage(e) }
  } finally {
    testLoading.value = false
  }
}

// 切换启用状态
async function toggleEnabled(notifier) {
  try {
    await updateNotifier(notifier.id, { enabled: !notifier.enabled })
    notifier.enabled = !notifier.enabled
  } catch (e) {
    console.error('更新失败:', e)
    errorMessage.value = formatErrorMessage(e)
    errorDialogVisible.value = true
  }
}

// 删除通知渠道
async function confirmDelete() {
  if (!deletingNotifier.value) return
  try {
    await deleteNotifier(deletingNotifier.value.id)
    deleteDialogVisible.value = false
    deletingNotifier.value = null
    loadNotifiers()
  } catch (e) {
    console.error('删除失败:', e)
    deleteDialogVisible.value = false
    errorMessage.value = formatErrorMessage(e)
    errorDialogVisible.value = true
  }
}

// 获取类型标签
function getTypeLabel(type) {
  const typeMap = {
    'wecom': '企业微信'
  }
  return typeMap[type] || type
}

onMounted(() => {
  loadNotifierTypes()
  loadNotifiers()
})
</script>

<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold font-display">通知中心</h1>
        <p class="text-muted-foreground text-sm mt-1">管理告警通知渠道和消息配置</p>
      </div>
      <Button @click="openCreateForm">
        <Plus class="w-4 h-4 mr-2" />
        新增通知渠道
      </Button>
    </div>

    <!-- 通知渠道列表 -->
    <Card>
      <CardHeader class="pb-3">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-lg bg-primary/20 flex items-center justify-center">
              <Bell class="w-5 h-5 text-primary" />
            </div>
            <div>
              <CardTitle class="text-lg">通知渠道</CardTitle>
              <CardDescription>配置告警消息的推送目标</CardDescription>
            </div>
          </div>
          <Button variant="outline" size="icon" @click="loadNotifiers" :disabled="loading">
            <RefreshCw :class="['w-4 h-4', loading && 'animate-spin']" />
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <Table v-if="notifiers.length > 0">
          <TableHeader>
            <TableRow>
              <TableHead class="w-[80px]">状态</TableHead>
              <TableHead>名称</TableHead>
              <TableHead class="w-[120px]">类型</TableHead>
              <TableHead>Webhook URL</TableHead>
              <TableHead class="w-[180px]">创建时间</TableHead>
              <TableHead class="w-[80px]">启用</TableHead>
              <TableHead class="w-[120px] text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="notifier in notifiers" :key="notifier.id">
              <TableCell>
                <div 
                  :class="[
                    'w-2.5 h-2.5 rounded-full',
                    notifier.enabled ? 'bg-success indicator-online' : 'bg-muted-foreground'
                  ]"
                />
              </TableCell>
              <TableCell class="font-medium">
                <div>
                  <div>{{ notifier.name }}</div>
                  <div v-if="notifier.description" class="text-xs text-muted-foreground">
                    {{ notifier.description }}
                  </div>
                </div>
              </TableCell>
              <TableCell>
                <Badge variant="secondary">{{ getTypeLabel(notifier.type) }}</Badge>
              </TableCell>
              <TableCell class="max-w-[300px]">
                <div class="truncate text-sm text-muted-foreground font-mono">
                  {{ notifier.webhook_url }}
                </div>
              </TableCell>
              <TableCell class="text-muted-foreground text-sm">
                {{ formatTime(notifier.created_at) }}
              </TableCell>
              <TableCell>
                <Switch 
                  :model-value="notifier.enabled" 
                  @update:model-value="toggleEnabled(notifier)" 
                />
              </TableCell>
              <TableCell class="text-right">
                <div class="flex items-center justify-end gap-1">
                  <Button variant="ghost" size="icon" @click="openEditForm(notifier)">
                    <Edit class="w-4 h-4" />
                  </Button>
                  <Button 
                    variant="ghost" 
                    size="icon" 
                    @click="deletingNotifier = notifier; deleteDialogVisible = true"
                  >
                    <Trash2 class="w-4 h-4 text-destructive" />
                  </Button>
                </div>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>

        <!-- 空状态 -->
        <div v-else class="text-center py-12">
          <div class="w-16 h-16 mx-auto mb-4 rounded-full bg-muted flex items-center justify-center">
            <MessageSquare class="w-8 h-8 text-muted-foreground" />
          </div>
          <h3 class="text-lg font-medium mb-2">暂无通知渠道</h3>
          <p class="text-muted-foreground mb-4">添加通知渠道以接收告警消息</p>
          <Button @click="openCreateForm">
            <Plus class="w-4 h-4 mr-2" />
            新增通知渠道
          </Button>
        </div>
      </CardContent>
    </Card>

    <!-- 新增/编辑表单弹窗 -->
    <Dialog
      v-model:open="formVisible"
      :title="editingNotifier ? '编辑通知渠道' : '新增通知渠道'"
      :description="editingNotifier ? '修改通知渠道配置' : '添加一个新的告警通知渠道'"
    >
      <form @submit.prevent="submitForm" class="space-y-4 mt-4">
        <!-- 名称 -->
        <div class="space-y-2">
          <label class="text-sm font-medium">渠道名称 <span class="text-destructive">*</span></label>
          <Input 
            v-model="formData.name" 
            placeholder="例如：运维告警群" 
            required
          />
        </div>

        <!-- 类型 -->
        <div class="space-y-2">
          <label class="text-sm font-medium">通知类型</label>
          <Select
            v-model="formData.type"
            :options="typeOptions"
            :disabled="!!editingNotifier"
          />
        </div>

        <!-- Webhook URL -->
        <div class="space-y-2">
          <label class="text-sm font-medium">Webhook URL <span class="text-destructive">*</span></label>
          <Input 
            v-model="formData.webhook_url" 
            placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx"
            required
          />
          <p class="text-xs text-muted-foreground">
            企业微信群机器人的 Webhook 地址
          </p>
        </div>

        <!-- 描述 -->
        <div class="space-y-2">
          <label class="text-sm font-medium">描述（可选）</label>
          <Input 
            v-model="formData.description" 
            placeholder="渠道用途说明"
          />
        </div>

        <!-- @所有人 -->
        <div class="flex items-center justify-between py-2">
          <div>
            <label class="text-sm font-medium">@所有人</label>
            <p class="text-xs text-muted-foreground">开启后消息将@群内所有成员</p>
          </div>
          <Switch v-model="formData.mention_all" />
        </div>

        <!-- 启用 -->
        <div class="flex items-center justify-between py-2">
          <div>
            <label class="text-sm font-medium">启用通知</label>
            <p class="text-xs text-muted-foreground">关闭后将不会发送告警到此渠道</p>
          </div>
          <Switch v-model="formData.enabled" />
        </div>

        <!-- 测试结果 -->
        <div 
          v-if="testResult" 
          :class="[
            'flex items-center gap-2 p-3 rounded-lg text-sm',
            testResult.success ? 'bg-success/10 text-success' : 'bg-destructive/10 text-destructive'
          ]"
        >
          <CheckCircle v-if="testResult.success" class="w-4 h-4" />
          <XCircle v-else class="w-4 h-4" />
          {{ testResult.message }}
        </div>

        <!-- 操作按钮 -->
        <div class="flex justify-between pt-4">
          <Button 
            type="button" 
            variant="outline" 
            @click="handleTestNotifier"
            :disabled="testLoading"
          >
            <Send :class="['w-4 h-4 mr-2', testLoading && 'animate-pulse']" />
            {{ testLoading ? '发送中...' : '测试发送' }}
          </Button>
          <div class="flex gap-3">
            <Button type="button" variant="outline" @click="closeForm">取消</Button>
            <Button type="submit" :disabled="formLoading">
              {{ formLoading ? '保存中...' : '保存' }}
            </Button>
          </div>
        </div>
      </form>
    </Dialog>

    <!-- 删除确认 -->
    <Dialog
      v-model:open="deleteDialogVisible"
      title="删除确认"
      description="确定要删除这个通知渠道吗？删除后将无法接收告警通知。"
    >
      <div class="flex justify-end gap-3 mt-4">
        <Button variant="outline" @click="deleteDialogVisible = false">取消</Button>
        <Button variant="destructive" @click="confirmDelete">确认删除</Button>
      </div>
    </Dialog>

    <!-- 错误提示弹窗 -->
    <Dialog v-model:open="errorDialogVisible">
      <div class="flex items-start gap-4">
        <div class="w-10 h-10 rounded-full bg-destructive/10 flex items-center justify-center flex-shrink-0">
          <XCircle class="w-5 h-5 text-destructive" />
        </div>
        <div class="flex-1 space-y-2">
          <h3 class="text-lg font-semibold">保存失败</h3>
          <p class="text-sm text-muted-foreground">{{ errorMessage }}</p>
        </div>
      </div>
      <div class="flex justify-end gap-3 mt-6">
        <Button @click="errorDialogVisible = false">确定</Button>
      </div>
    </Dialog>
  </div>
</template>


