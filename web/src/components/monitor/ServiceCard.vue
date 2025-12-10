<script setup>
import { computed } from 'vue'
import { Card, CardContent } from '@/components/ui/card'
import StatusBadge from './StatusBadge.vue'
import { cn, formatTime, getServiceTypeLabel } from '@/lib/utils'
import { 
  Database, 
  Server, 
  HardDrive, 
  Radio, 
  Globe, 
  Network,
  Activity
} from 'lucide-vue-next'

const props = defineProps({
  service: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['click', 'toggle'])

const typeIcons = {
  postgresql: Database,
  cassandra: HardDrive,
  redis: Server,
  kafka: Radio,
  http: Globe,
  tcp: Network,
  mysql: Database,
  mongodb: Database,
  elasticsearch: Server
}

const icon = computed(() => typeIcons[props.service.type] || Activity)

const latencyClass = computed(() => {
  const latency = props.service.last_latency_ms
  if (!latency) return 'text-muted-foreground'
  if (latency < 100) return 'text-success'
  if (latency < 500) return 'text-warning'
  return 'text-destructive'
})

const borderClass = computed(() => {
  const status = props.service.status
  if (status === 'healthy') return 'border-success/20 hover:border-success/40'
  if (status === 'unhealthy') return 'border-destructive/20 hover:border-destructive/40'
  if (status === 'warning') return 'border-warning/20 hover:border-warning/40'
  return 'border-muted-foreground/20 hover:border-muted-foreground/40'
})
</script>

<template>
  <Card 
    :class="cn(
      'cursor-pointer group relative overflow-hidden transition-all duration-300',
      borderClass
    )"
    @click="emit('click', service)"
  >
    <!-- 状态指示条 -->
    <div 
      :class="cn(
        'absolute top-0 left-0 right-0 h-0.5 transition-all',
        service.status === 'healthy' && 'bg-success',
        service.status === 'unhealthy' && 'bg-destructive',
        service.status === 'warning' && 'bg-warning',
        service.status === 'unknown' && 'bg-muted-foreground'
      )"
    />
    
    <CardContent class="p-4 pt-5">
      <div class="flex items-start justify-between mb-3">
        <div class="flex items-center gap-3">
          <div 
            :class="cn(
              'flex items-center justify-center w-9 h-9 rounded-lg transition-colors',
              service.status === 'healthy' && 'bg-success/10 text-success',
              service.status === 'unhealthy' && 'bg-destructive/10 text-destructive',
              service.status === 'warning' && 'bg-warning/10 text-warning',
              service.status === 'unknown' && 'bg-muted text-muted-foreground'
            )"
          >
            <component :is="icon" class="w-4 h-4" />
          </div>
          <div>
            <h3 class="font-medium text-sm text-foreground group-hover:text-primary transition-colors">
              {{ service.name }}
            </h3>
            <p class="text-xs text-muted-foreground">
              {{ getServiceTypeLabel(service.type) }}
            </p>
          </div>
        </div>
        <StatusBadge :status="service.status" size="sm" :show-label="false" />
      </div>
      
      <div class="grid grid-cols-2 gap-3 text-sm">
        <div>
          <p class="text-muted-foreground text-xs mb-0.5">延迟</p>
          <p :class="latencyClass" class="font-mono text-sm font-medium">
            {{ service.last_latency_ms ? `${service.last_latency_ms}ms` : '-' }}
          </p>
        </div>
        <div>
          <p class="text-muted-foreground text-xs mb-0.5">可用率</p>
          <p class="font-mono text-sm font-medium">
            {{ service.stats?.success_rate_24h?.toFixed(1) || '0.0' }}%
          </p>
        </div>
      </div>
      
      <div class="mt-3 pt-3 border-t border-border/50 flex items-center justify-between text-xs text-muted-foreground">
        <span>{{ service.interval_seconds }}s</span>
        <span>{{ formatTime(service.last_check_at) }}</span>
      </div>
    </CardContent>
  </Card>
</template>
