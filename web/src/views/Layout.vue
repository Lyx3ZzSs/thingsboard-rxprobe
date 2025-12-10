<script setup>
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { 
  Zap,
  LayoutDashboard, 
  Server, 
  Bell, 
  ChevronLeft,
  Menu
} from 'lucide-vue-next'

const route = useRoute()

const collapsed = ref(false)
const mobileMenuOpen = ref(false)

const menuItems = [
  { path: '/dashboard', label: '监控总览', icon: LayoutDashboard },
  { path: '/services', label: '服务管理', icon: Server },
  { path: '/alerts', label: '告警中心', icon: Bell }
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
        <div class="flex items-center gap-2 text-sm">
          <Zap class="w-4 h-4 text-primary" />
          <span class="text-muted-foreground">生产环境通用监控工具</span>
          <span class="text-primary/50">/</span>
          <span class="font-medium text-foreground">{{ route.meta.title || '监控总览' }}</span>
        </div>
        
      </header>
      
      <!-- 页面内容 -->
      <main class="flex-1 overflow-y-auto p-4 lg:p-6 bg-gradient-radial">
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
