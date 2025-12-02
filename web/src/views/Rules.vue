<template>
  <div class="rules-page">
    <el-card>
      <div class="page-header">
        <h3>告警规则</h3>
        <el-button type="primary" @click="showAddDialog">
          <el-icon><Plus /></el-icon>
          新增规则
        </el-button>
      </div>
      
      <el-table :data="rules" v-loading="loading" stripe>
        <el-table-column label="规则名称" prop="name" min-width="150" />
        
        <el-table-column label="告警阈值" width="120">
          <template #default="{ row }">
            连续失败 {{ row.threshold }} 次
          </template>
        </el-table-column>
        
        <el-table-column label="静默时间" width="120">
          <template #default="{ row }">
            {{ row.silence_minutes }} 分钟
          </template>
        </el-table-column>
        
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.enabled ? 'success' : 'info'" size="small">
              {{ row.enabled ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="editRule(row)">编辑</el-button>
            <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    
    <!-- 新增/编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑规则' : '新增规则'"
      width="500px"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="formData.name" placeholder="如: 默认告警规则" />
        </el-form-item>
        
        <el-form-item label="告警阈值" prop="threshold">
          <el-input-number v-model="formData.threshold" :min="1" :max="10" />
          <span class="form-tip">连续失败次数</span>
        </el-form-item>
        
        <el-form-item label="静默时间" prop="silence_minutes">
          <el-input-number v-model="formData.silence_minutes" :min="1" :max="1440" />
          <span class="form-tip">分钟</span>
        </el-form-item>
        
        <el-form-item label="启用" prop="enabled">
          <el-switch v-model="formData.enabled" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSubmit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getRules, createRule, updateRule, deleteRule } from '@/api/alert'

const loading = ref(false)
const saving = ref(false)
const rules = ref([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const editingId = ref(null)

const formRef = ref()
const formData = ref({
  name: '',
  threshold: 3,
  silence_minutes: 30,
  enabled: true
})

const formRules = {
  name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }]
}

onMounted(() => {
  loadRules()
})

async function loadRules() {
  loading.value = true
  try {
    const res = await getRules()
    rules.value = res.data || []
  } finally {
    loading.value = false
  }
}

function showAddDialog() {
  isEdit.value = false
  editingId.value = null
  formData.value = {
    name: '',
    threshold: 3,
    silence_minutes: 30,
    enabled: true
  }
  dialogVisible.value = true
}

function editRule(row) {
  isEdit.value = true
  editingId.value = row.id
  formData.value = {
    name: row.name,
    threshold: row.threshold,
    silence_minutes: row.silence_minutes,
    enabled: row.enabled
  }
  dialogVisible.value = true
}

async function handleSubmit() {
  await formRef.value.validate()
  
  saving.value = true
  try {
    if (isEdit.value) {
      await updateRule(editingId.value, formData.value)
    } else {
      await createRule(formData.value)
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    loadRules()
  } finally {
    saving.value = false
  }
}

async function handleDelete(row) {
  try {
    await ElMessageBox.confirm(`确定要删除规则 "${row.name}" 吗？`, '提示', {
      type: 'warning'
    })
    await deleteRule(row.id)
    ElMessage.success('删除成功')
    loadRules()
  } catch (e) {
    if (e !== 'cancel') {
      console.error(e)
    }
  }
}

function formatTime(time) {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}
</script>

<style lang="scss" scoped>
.rules-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    
    h3 {
      margin: 0;
    }
  }
  
  .form-tip {
    margin-left: 10px;
    color: #909399;
    font-size: 12px;
  }
}
</style>

