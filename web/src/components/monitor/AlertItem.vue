<script setup>
import { computed } from 'vue'
import { cn, formatTime, getServiceTypeLabel } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { 
  AlertTriangle, 
  CheckCircle2, 
  Clock,
  BellOff
} from 'lucide-vue-next'

const props = defineProps({
  alert: {
    type: Object,
    required: true
  },
  compact: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['silence', 'view'])

const isFiring = computed(() => props.alert.status === 'firing')
</script>

<template>
  <div 
    :class="cn(
      'group rounded-lg border transition-all duration-200',
      isFiring 
        ? 'bg-destructive/5 border-destructive/30 hover:border-destructive/50' 
        : 'bg-success/5 border-success/30 hover:border-success/50',
      compact ? 'p-3' : 'p-4'
    )"
  >
    <div class="flex items-start gap-3">
      <!-- 图标 -->
      <div 
        :class="cn(
          'flex items-center justify-center w-8 h-8 rounded-full shrink-0',
          isFiring ? 'bg-destructive/10 text-destructive' : 'bg-success/10 text-success'
        )"
      >
        <AlertTriangle v-if="isFiring" class="w-4 h-4" />
        <CheckCircle2 v-else class="w-4 h-4" />
      </div>
      
      <!-- 内容 -->
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2 mb-1">
          <h4 class="font-medium truncate">{{ alert.target_name }}</h4>
          <Badge 
            :variant="isFiring ? 'destructive' : 'success'" 
            class="shrink-0"
          >
            {{ isFiring ? '告警中' : '已恢复' }}
          </Badge>
        </div>
        
        <p class="text-sm text-muted-foreground truncate mb-2">
          {{ alert.message }}
        </p>
        
        <div class="flex items-center gap-4 text-xs text-muted-foreground">
          <span class="flex items-center gap-1">
            <Clock class="w-3 h-3" />
            {{ formatTime(alert.fired_at) }}
          </span>
          <span v-if="alert.resolved_at">
            恢复于 {{ formatTime(alert.resolved_at) }}
          </span>
          <span v-if="alert.latency_ms">
            {{ alert.latency_ms }}ms
          </span>
        </div>
      </div>
      
      <!-- 操作 -->
      <div v-if="isFiring && !compact" class="shrink-0">
        <Button 
          variant="ghost" 
          size="sm"
          @click.stop="emit('silence', alert)"
          class="opacity-0 group-hover:opacity-100 transition-opacity"
        >
          <BellOff class="w-4 h-4 mr-1" />
          静默
        </Button>
      </div>
    </div>
  </div>
</template>

