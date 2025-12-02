<template>
  <div class="targets-page">
    <el-card>
      <!-- 头部操作栏 -->
      <div class="page-header">
        <div class="header-left">
          <el-input
            v-model="filters.keyword"
            placeholder="搜索目标名称"
            clearable
            style="width: 200px"
            @clear="loadTargets"
            @keyup.enter="loadTargets"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          
          <el-select v-model="filters.type" placeholder="组件类型" clearable style="width: 150px" @change="loadTargets">
            <el-option v-for="t in probeTypes" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
          
          <el-select v-model="filters.status" placeholder="状态" clearable style="width: 120px" @change="loadTargets">
            <el-option label="正常" value="healthy" />
            <el-option label="异常" value="unhealthy" />
            <el-option label="未知" value="unknown" />
          </el-select>
        </div>
        
        <el-button type="primary" @click="showAddDialog">
          <el-icon><Plus /></el-icon>
          新增目标
        </el-button>
      </div>
      
      <!-- 目标列表 -->
      <el-table :data="targets" v-loading="loading" stripe>
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
        
        <el-table-column label="探测间隔" width="100">
          <template #default="{ row }">
            {{ row.interval_seconds }}s
          </template>
        </el-table-column>
        
        <el-table-column label="最后检查" width="180">
          <template #default="{ row }">
            {{ formatTime(row.last_check_at) }}
          </template>
        </el-table-column>
        
        <el-table-column label="启用" width="80">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" @change="toggleEnabled(row)" />
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="goToDetail(row)">详情</el-button>
            <el-button type="primary" link @click="editTarget(row)">编辑</el-button>
            <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.size"
          :total="pagination.total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @change="loadTargets"
        />
      </div>
    </el-card>
    
    <!-- 新增/编辑弹窗 -->
    <TargetForm
      v-model="formVisible"
      :edit-data="editingTarget"
      :probe-types="probeTypes"
      @success="onFormSuccess"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getTargets, deleteTarget, updateTarget, getProbeTypes } from '@/api/probe'
import TargetForm from '@/components/TargetForm.vue'

const router = useRouter()

const loading = ref(false)
const targets = ref([])
const probeTypes = ref([])
const formVisible = ref(false)
const editingTarget = ref(null)

const filters = reactive({
  keyword: '',
  type: '',
  status: ''
})

const pagination = reactive({
  page: 1,
  size: 10,
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
  loadProbeTypes()
  loadTargets()
})

async function loadProbeTypes() {
  try {
    const res = await getProbeTypes()
    probeTypes.value = res.data
  } catch (e) {
    console.error(e)
  }
}

async function loadTargets() {
  loading.value = true
  try {
    const res = await getTargets({
      ...filters,
      page: pagination.page,
      size: pagination.size
    })
    targets.value = res.data.items || []
    pagination.total = res.data.total
  } finally {
    loading.value = false
  }
}

function showAddDialog() {
  editingTarget.value = null
  formVisible.value = true
}

function editTarget(row) {
  editingTarget.value = { ...row }
  formVisible.value = true
}

function goToDetail(row) {
  router.push(`/targets/${row.id}`)
}

async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(`确定要删除目标 "${row.name}" 吗？`, '提示', {
      type: 'warning'
    })
    await deleteTarget(row.id)
    ElMessage.success('删除成功')
    loadTargets()
  } catch (e) {
    if (e !== 'cancel') {
      console.error(e)
    }
  }
}

async function toggleEnabled(row) {
  try {
    await updateTarget(row.id, { enabled: row.enabled })
    ElMessage.success(row.enabled ? '已启用' : '已禁用')
  } catch (e) {
    row.enabled = !row.enabled
  }
}

function onFormSuccess() {
  loadTargets()
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

function formatTime(time) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}
</script>

<style lang="scss" scoped>
.targets-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    
    .header-left {
      display: flex;
      gap: 12px;
    }
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

