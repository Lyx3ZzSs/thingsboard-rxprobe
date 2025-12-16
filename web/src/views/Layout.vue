<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { useAuth } from '@/store/auth'
import { 
  Zap,
  LayoutDashboard, 
  Server, 
  Bell, 
  MessageSquare,
  ChevronLeft,
  Menu,
  LogOut,
  User,
  ChevronDown
} from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const { user, loadCurrentUser, clearAuth } = useAuth()

const collapsed = ref(false)
const mobileMenuOpen = ref(false)
const userMenuOpen = ref(false)
const userMenuRef = ref(null)

// 加载用户信息
onMounted(async () => {
  if (!user.value) {
    await loadCurrentUser()
  }
  
  // 添加点击外部关闭菜单
  document.addEventListener('click', handleClickOutside)
  
  // 监听窗口滚动和调整大小，关闭菜单
  window.addEventListener('scroll', closeUserMenu, true)
  window.addEventListener('resize', closeUserMenu)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  window.removeEventListener('scroll', closeUserMenu, true)
  window.removeEventListener('resize', closeUserMenu)
})

// 关闭用户菜单
function closeUserMenu() {
  userMenuOpen.value = false
}

// 点击外部关闭菜单
function handleClickOutside(event) {
  if (!userMenuOpen.value) return
  
  const target = event.target
  const button = userMenuRef.value?.querySelector('button')
  const menu = document.querySelector('[data-user-menu]')
  
  // 检查点击是否在按钮或菜单内
  const isClickInButton = button?.contains(target)
  const isClickInMenu = menu?.contains(target)
  
  if (!isClickInButton && !isClickInMenu) {
    userMenuOpen.value = false
  }
}

// 切换用户菜单
function toggleUserMenu(event) {
  event.stopPropagation()
  userMenuOpen.value = !userMenuOpen.value
  if (userMenuOpen.value) {
    updateMenuPosition()
  }
}

// 更新菜单位置
function updateMenuPosition() {
  if (!userMenuRef.value) return
  
  setTimeout(() => {
    const button = userMenuRef.value.querySelector('button')
    if (button) {
      const rect = button.getBoundingClientRect()
      const menu = document.querySelector('[data-user-menu]')
      if (menu) {
        menu.style.top = `${rect.bottom + 8}px`
        menu.style.right = `${window.innerWidth - rect.right}px`
      }
    }
  }, 0)
}

// 登出
function handleLogout() {
  userMenuOpen.value = false
  clearAuth()
  router.push('/login')
}

const menuItems = [
  { path: '/dashboard', label: '监控总览', icon: LayoutDashboard },
  { path: '/services', label: '服务管理', icon: Server },
  { path: '/alerts', label: '告警中心', icon: Bell },
  { path: '/notifiers', label: '通知中心', icon: MessageSquare }
]

const currentPath = computed(() => route.path)

function toggleSidebar() {
  collapsed.value = !collapsed.value
}
</script>

<template>
  <div class="flex h-screen bg-background bg-grid">
    <!-- 侧边栏 -->
    <aside 
      :class="cn(
        'fixed inset-y-0 left-0 z-50 flex flex-col border-r border-primary/20 bg-card/95 backdrop-blur-xl transition-all duration-300 lg:relative',
        collapsed ? 'w-16' : 'w-60',
        mobileMenuOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'
      )"
    >
      <!-- Logo -->
      <div class="flex items-center h-16 px-4 border-b border-primary/20">
        <div class="flex items-center gap-3">
          <div class="w-9 h-9 rounded-lg bg-primary/20 flex items-center justify-center shrink-0 glow">
            <Zap class="w-5 h-5 text-primary" />
          </div>
          <div v-if="!collapsed" class="flex flex-col">
            <span class="font-bold text-base font-display text-gradient whitespace-nowrap">
              RxIOTProbe
            </span>
            <span class="text-[10px] text-muted-foreground -mt-0.5">生产环境通用监控工具</span>
          </div>
        </div>
      </div>
      
      <!-- 导航菜单 -->
      <nav class="flex-1 p-3 space-y-1 overflow-y-auto">
        <router-link
          v-for="item in menuItems"
          :key="item.path"
          :to="item.path"
          :class="cn(
            'flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-all',
            currentPath === item.path || currentPath.startsWith(item.path + '/')
              ? 'bg-primary/20 text-primary border border-primary/30 glow'
              : 'text-muted-foreground hover:text-foreground hover:bg-muted/50'
          )"
        >
          <component :is="item.icon" class="w-5 h-5 shrink-0" />
          <span v-if="!collapsed" class="whitespace-nowrap">{{ item.label }}</span>
        </router-link>
      </nav>
      
      <!-- 底部系统状态 -->
      <div v-if="!collapsed" class="p-3 border-t border-primary/20">
        <div class="flex items-center gap-2 px-3 py-2 rounded-lg bg-muted/30">
          <div class="w-2 h-2 rounded-full indicator-online" />
          <span class="text-xs text-muted-foreground">系统运行正常</span>
        </div>
      </div>
      
      <!-- 折叠按钮 -->
      <button
        @click="toggleSidebar"
        class="absolute -right-3 top-20 w-6 h-6 rounded-full bg-card border border-primary/30 flex items-center justify-center hover:bg-primary/20 hover:border-primary/50 transition-colors hidden lg:flex"
      >
        <ChevronLeft :class="cn('w-4 h-4 text-primary transition-transform', collapsed && 'rotate-180')" />
      </button>
    </aside>
    
    <!-- 遮罩层 (移动端) -->
    <div 
      v-if="mobileMenuOpen"
      class="fixed inset-0 z-40 bg-black/60 backdrop-blur-sm lg:hidden"
      @click="mobileMenuOpen = false"
    />
    
    <!-- 主内容区 -->
    <div class="flex-1 flex flex-col overflow-hidden">
      <!-- 顶部导航 -->
      <header class="h-14 border-b border-primary/20 bg-card/80 backdrop-blur-xl flex items-center px-4 lg:px-6 gap-4">
        <!-- 移动端菜单按钮 -->
        <Button
          variant="ghost"
          size="icon"
          class="lg:hidden"
          @click="mobileMenuOpen = !mobileMenuOpen"
        >
          <Menu class="w-5 h-5" />
        </Button>
        
        <!-- 面包屑 -->
        <div class="flex items-center gap-2 text-sm flex-1">
          <Zap class="w-4 h-4 text-primary" />
          <span class="text-muted-foreground">生产环境通用监控工具</span>
          <span class="text-primary/50">/</span>
          <span class="font-medium text-foreground">{{ route.meta.title || '监控总览' }}</span>
        </div>
        
        <!-- 用户菜单 -->
        <div class="relative" ref="userMenuRef">
          <Button
            variant="ghost"
            size="sm"
            :class="cn('gap-2 transition-all', userMenuOpen && 'bg-muted/50')"
            @click="toggleUserMenu"
          >
            <User class="w-4 h-4" />
            <span class="hidden sm:inline">{{ user?.nickname || user?.username || '用户' }}</span>
            <ChevronDown :class="cn('w-3 h-3 transition-transform duration-200', userMenuOpen && 'rotate-180')" />
          </Button>
        </div>
      </header>
      
      <!-- 用户菜单下拉（使用 Teleport 挂载到 body） -->
      <Teleport to="body">
        <Transition
          enter-active-class="transition ease-out duration-200"
          enter-from-class="transform opacity-0 scale-95 -translate-y-2"
          enter-to-class="transform opacity-100 scale-100 translate-y-0"
          leave-active-class="transition ease-in duration-150"
          leave-from-class="transform opacity-100 scale-100 translate-y-0"
          leave-to-class="transform opacity-0 scale-95 -translate-y-2"
        >
          <div
            v-if="userMenuOpen"
            data-user-menu
            class="fixed min-w-[220px] bg-card/95 border-2 border-primary/30 rounded-lg shadow-2xl backdrop-blur-xl"
            style="z-index: 99999;"
            @click.stop
          >
            <!-- 用户信息 -->
            <div class="p-4 border-b border-primary/20 bg-gradient-to-br from-primary/10 to-transparent">
              <div class="flex items-center gap-3 mb-2">
                <div class="w-10 h-10 rounded-full bg-primary/20 border-2 border-primary/30 flex items-center justify-center">
                  <User class="w-5 h-5 text-primary" />
                </div>
                <div class="flex-1 min-w-0">
                  <div class="text-sm font-bold text-foreground truncate">
                    {{ user?.nickname || user?.username || '用户' }}
                  </div>
                  <div class="text-xs text-muted-foreground truncate">
                    {{ user?.email || user?.username }}
                  </div>
                </div>
              </div>
              <div class="flex items-center gap-2">
                <span class="inline-flex items-center text-xs px-2.5 py-1 rounded-md bg-primary/20 text-primary border border-primary/30 font-medium">
                  <span class="w-1.5 h-1.5 rounded-full bg-primary mr-1.5"></span>
                  {{ user?.role === 'admin' ? '管理员' : '普通用户' }}
                </span>
              </div>
            </div>
            
            <!-- 操作按钮 -->
            <div class="p-2">
              <button
                @click="handleLogout"
                class="w-full flex items-center gap-3 px-4 py-3 text-sm font-medium text-foreground hover:bg-destructive/20 hover:text-destructive transition-all rounded-md border border-transparent hover:border-destructive/30 group"
              >
                <LogOut class="w-4 h-4 transition-transform group-hover:translate-x-0.5" />
                <span>退出登录</span>
              </button>
            </div>
          </div>
        </Transition>
      </Teleport>
      
      <!-- 页面内容 -->
      <main class="flex-1 overflow-y-auto p-4 lg:p-6 bg-gradient-radial" @click="userMenuOpen = false">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </main>
    </div>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
