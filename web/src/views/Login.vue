<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { login } from '@/api/auth'
import { setToken, setUser, checkSystemInit } from '@/store/auth'
import { Activity, Loader2 } from 'lucide-vue-next'

const router = useRouter()

const loading = ref(false)
const form = ref({
  username: '',
  password: ''
})
const error = ref('')

// 检查系统是否已初始化
onMounted(async () => {
  const initialized = await checkSystemInit()
  if (!initialized) {
    // 如果未初始化，跳转到初始化页面
    router.push('/init')
  }
})

async function handleLogin() {
  if (!form.value.username || !form.value.password) {
    error.value = '请输入用户名和密码'
    return
  }

  loading.value = true
  error.value = ''

  try {
    const res = await login(form.value)
    setToken(res.data.token)
    setUser(res.data.user)
    
    // 跳转到首页
    router.push('/dashboard')
  } catch (e) {
    error.value = e.response?.data?.message || e.message || '登录失败，请检查用户名和密码'
  } finally {
    loading.value = false
  }
}

function handleKeyPress(e) {
  if (e.key === 'Enter') {
    handleLogin()
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 p-4">
    <Card class="w-full max-w-md">
      <CardHeader class="text-center space-y-2">
        <div class="mx-auto w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center mb-4">
          <Activity class="w-8 h-8 text-primary" />
        </div>
        <CardTitle class="text-2xl font-bold">生产环境监控工具</CardTitle>
        <p class="text-muted-foreground text-sm">请登录您的账户</p>
      </CardHeader>
      <CardContent class="space-y-4">
        <div v-if="error" class="p-3 rounded-lg bg-destructive/10 text-destructive text-sm">
          {{ error }}
        </div>

        <form @submit.prevent="handleLogin" class="space-y-4">
          <div>
            <label class="text-sm font-medium mb-2 block">用户名</label>
            <Input
              v-model="form.username"
              placeholder="请输入用户名"
              @keypress="handleKeyPress"
              :disabled="loading"
            />
          </div>

          <div>
            <label class="text-sm font-medium mb-2 block">密码</label>
            <Input
              v-model="form.password"
              type="password"
              placeholder="请输入密码"
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
            {{ loading ? '登录中...' : '登录' }}
          </Button>
        </form>
      </CardContent>
    </Card>
  </div>
</template>
