<template>
  <el-dialog
    v-model="visible"
    :title="isEdit ? '编辑监控目标' : '新增监控目标'"
    width="650px"
    destroy-on-close
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      label-width="120px"
    >
      <!-- 基础信息 -->
      <el-form-item label="目标名称" prop="name">
        <el-input v-model="formData.name" placeholder="如: TB PostgreSQL Master" />
      </el-form-item>

      <el-form-item label="组件类型" prop="type">
        <el-select
          v-model="formData.type"
          placeholder="选择组件类型"
          style="width: 100%"
          @change="onTypeChange"
          :disabled="isEdit"
        >
          <el-option
            v-for="item in probeTypes"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-form-item>

      <!-- 动态配置字段 -->
      <el-divider v-if="Object.keys(configSchema).length">连接配置</el-divider>
      
      <template v-for="(field, key) in configSchema" :key="key">
        <el-form-item
          v-if="shouldShowField(field)"
          :label="field.label"
          :prop="'config.' + key"
        >
          <!-- 文本输入 -->
          <el-input
            v-if="field.type === 'string'"
            v-model="formData.config[key]"
            :placeholder="field.placeholder"
          />

          <!-- 密码输入 -->
          <el-input
            v-else-if="field.type === 'password'"
            v-model="formData.config[key]"
            type="password"
            show-password
            :placeholder="field.placeholder"
          />

          <!-- 数字输入 -->
          <el-input-number
            v-else-if="field.type === 'number'"
            v-model="formData.config[key]"
            :min="0"
            style="width: 100%"
          />

          <!-- 开关 -->
          <el-switch
            v-else-if="field.type === 'boolean'"
            v-model="formData.config[key]"
          />

          <!-- 下拉选择 -->
          <el-select
            v-else-if="field.type === 'select'"
            v-model="formData.config[key]"
            style="width: 100%"
          >
            <el-option
              v-for="opt in field.options"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>

          <!-- 字段提示 -->
          <template v-if="field.hint">
            <div class="field-hint">{{ field.hint }}</div>
          </template>
        </el-form-item>
      </template>

      <!-- 探测配置 -->
      <el-divider>探测配置</el-divider>

      <el-form-item label="探测间隔(秒)" prop="interval_seconds">
        <el-input-number
          v-model="formData.interval_seconds"
          :min="10"
          :max="3600"
          style="width: 200px"
        />
      </el-form-item>

      <el-form-item label="超时时间(秒)" prop="timeout_seconds">
        <el-input-number
          v-model="formData.timeout_seconds"
          :min="1"
          :max="60"
          style="width: 200px"
        />
      </el-form-item>

      <el-form-item label="启用监控" prop="enabled">
        <el-switch v-model="formData.enabled" />
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="visible = false">取消</el-button>
        <el-button
          type="warning"
          :loading="testing"
          @click="testConnection"
        >
          测试连接
        </el-button>
        <el-button
          type="primary"
          :loading="saving"
          @click="handleSubmit"
        >
          保存
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { getProbeSchema, createTarget, updateTarget, testTarget } from '@/api/probe'

const props = defineProps({
  modelValue: Boolean,
  editData: Object,
  probeTypes: Array
})

const emit = defineEmits(['update:modelValue', 'success'])

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const isEdit = computed(() => !!props.editData)

const formRef = ref()
const formData = ref({
  name: '',
  type: '',
  config: {},
  interval_seconds: 30,
  timeout_seconds: 5,
  enabled: true
})

const configSchema = ref({})
const testing = ref(false)
const saving = ref(false)

const formRules = {
  name: [{ required: true, message: '请输入目标名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择组件类型', trigger: 'change' }]
}

// 监听类型变化，加载对应的配置 schema
watch(() => formData.value.type, async (newType) => {
  if (newType && !isEdit.value) {
    try {
      const res = await getProbeSchema(newType)
      configSchema.value = res.data || {}
      
      // 设置默认值
      formData.value.config = {}
      for (const [key, field] of Object.entries(configSchema.value)) {
        if (field.default_value !== undefined && field.default_value !== null) {
          formData.value.config[key] = field.default_value
        }
      }
    } catch (e) {
      console.error(e)
    }
  }
})

// 编辑模式初始化
watch(() => props.editData, async (data) => {
  if (data) {
    // 先加载 schema
    try {
      const res = await getProbeSchema(data.type)
      configSchema.value = res.data || {}
    } catch (e) {
      console.error(e)
    }
    
    // 解析配置
    let config = data.config
    if (typeof config === 'string') {
      try {
        config = JSON.parse(config)
      } catch (e) {
        config = {}
      }
    }
    
    formData.value = {
      name: data.name,
      type: data.type,
      config: config || {},
      interval_seconds: data.interval_seconds || 30,
      timeout_seconds: data.timeout_seconds || 5,
      enabled: data.enabled !== false
    }
  } else {
    formData.value = {
      name: '',
      type: '',
      config: {},
      interval_seconds: 30,
      timeout_seconds: 5,
      enabled: true
    }
    configSchema.value = {}
  }
}, { immediate: true })

function onTypeChange() {
  // 类型变化时清空配置
  formData.value.config = {}
}

// 判断字段是否应该显示
function shouldShowField(field) {
  if (!field.show_when) return true
  
  for (const [key, expectedValue] of Object.entries(field.show_when)) {
    const actualValue = formData.value.config[key]
    if (Array.isArray(expectedValue)) {
      if (!expectedValue.includes(actualValue)) return false
    } else {
      if (actualValue !== expectedValue) return false
    }
  }
  return true
}

// 测试连接
async function testConnection() {
  testing.value = true
  try {
    const result = await testTarget({
      type: formData.value.type,
      config: formData.value.config,
      timeout_seconds: formData.value.timeout_seconds
    })
    
    if (result.data.success) {
      ElMessage.success(`连接成功! 响应时间: ${result.data.latency_ms}ms`)
    } else {
      ElMessage.error(`连接失败: ${result.data.message}`)
    }
  } catch (err) {
    ElMessage.error(`测试失败: ${err.message}`)
  } finally {
    testing.value = false
  }
}

// 提交保存
async function handleSubmit() {
  await formRef.value.validate()
  
  saving.value = true
  try {
    if (isEdit.value) {
      await updateTarget(props.editData.id, formData.value)
    } else {
      await createTarget(formData.value)
    }
    ElMessage.success('保存成功')
    visible.value = false
    emit('success')
  } catch (err) {
    // error handled by interceptor
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.field-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>

