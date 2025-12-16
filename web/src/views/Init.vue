<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { initSystem } from '@/api/auth'
import { setToken, setUser, checkSystemInit } from '@/store/auth'
import { Settings, Loader2, CheckCircle2 } from 'lucide-vue-next'

const router = useRouter()

const loading = ref(false)
const success = ref(false)
const form = ref({
  username: '',
  password: '',
  confirmPassword: '',
  nickname: '',
  email: ''
})
const error = ref('')

// 检查系统是否已初始化
onMounted(async () => {
  const initialized = await checkSystemInit()
  if (initialized) {
    // 如果已初始化，跳转到登录页
    router.push('/login')
  }
})

async function handleInit() {
  // 验证表单
  if (!form.value.username || !form.value.password) {
    error.value = '请输入用户名和密码'
    return
  }

  if (form.value.username.length < 3) {
    error.value = '用户名至少需要3个字符'
    return
  }

  if (form.value.password.length < 6) {
    error.value = '密码至少需要6个字符'
    return
  }

  if (form.value.password !== form.value.confirmPassword) {
    error.value = '两次输入的密码不一致'
    return
  }

  loading.value = true
  error.value = ''

  try {
    const res = await initSystem({
      username: form.value.username,
      password: form.value.password,
      nickname: form.value.nickname || form.value.username,
      email: form.value.email
    })
    
    // 初始化成功后自动登录
    const loginRes = await import('@/api/auth').then(m => m.login({
      username: form.value.username,
      password: form.value.password
    }))
    
    setToken(loginRes.data.token)
    setUser(loginRes.data.user)
    
    success.value = true
    
    // 2秒后跳转到首页
    setTimeout(() => {
      router.push('/dashboard')
    }, 2000)
  } catch (e) {
    error.value = e.response?.data?.message || e.message || '初始化失败，请重试'
  } finally {
    loading.value = false
  }
}

function handleKeyPress(e) {
  if (e.key === 'Enter') {
    handleInit()
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-purple-50 to-pink-100 p-4">
    <Card class="w-full max-w-lg">
      <CardHeader class="text-center space-y-2">
        <div class="mx-auto w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center mb-4">
          <Settings class="w-8 h-8 text-primary" />
        </div>
        <CardTitle class="text-2xl font-bold">系统初始化</CardTitle>
        <CardDescription>
          欢迎使用生产环境监控工具！请创建管理员账户以开始使用
        </CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        <div v-if="success" class="p-4 rounded-lg bg-success/10 text-success flex items-center gap-2">
          <CheckCircle2 class="w-5 h-5" />
          <div>
            <p class="font-medium">初始化成功！</p>
            <p class="text-sm mt-1">正在跳转到首页...</p>
          </div>
        </div>

        <div v-else>
          <div v-if="error" class="p-3 rounded-lg bg-destructive/10 text-destructive text-sm">
            {{ error }}
          </div>

          <form @submit.prevent="handleInit" class="space-y-4">
            <div>
              <label class="text-sm font-medium mb-2 block">
                用户名 <span class="text-destructive">*</span>
              </label>
              <Input
                v-model="form.username"
                placeholder="至少3个字符"
                @keypress="handleKeyPress"
                :disabled="loading"
              />
            </div>

            <div>
              <label class="text-sm font-medium mb-2 block">
                昵称
              </label>
              <Input
                v-model="form.nickname"
                placeholder="可选，默认为用户名"
                @keypress="handleKeyPress"
                :disabled="loading"
              />
            </div>

            <div>
              <label class="text-sm font-medium mb-2 block">
                邮箱
              </label>
              <Input
                v-model="form.email"
                type="email"
                placeholder="可选"
                @keypress="handleKeyPress"
                :disabled="loading"
              />
            </div>

            <div>
              <label class="text-sm font-medium mb-2 block">
                密码 <span class="text-destructive">*</span>
              </label>
              <Input
                v-model="form.password"
                type="password"
                placeholder="至少6个字符"
                @keypress="handleKeyPress"
                :disabled="loading"
              />
            </div>

            <div>
              <label class="text-sm font-medium mb-2 block">
                确认密码 <span class="text-destructive">*</span>
              </label>
              <Input
                v-model="form.confirmPassword"
                type="password"
                placeholder="请再次输入密码"
                @keypress="handleKeyPress"
                :disabled="loading"
              />
            </div>

            <Button
              type="submit"
              class="w-full"
              :disabled="loading"
            >
              <Loader2 v-if="loading" class="w-4 h-4 mr-2 animate-spin" />
              {{ loading ? '初始化中...' : '完成初始化' }}
            </Button>
          </form>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
