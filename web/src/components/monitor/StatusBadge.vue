<script setup>
import { computed } from 'vue'
import { cn } from '@/lib/utils'
import { CheckCircle, XCircle, AlertCircle, HelpCircle } from 'lucide-vue-next'

const props = defineProps({
  status: {
    type: String,
    required: true,
    validator: (v) => ['healthy', 'unhealthy', 'warning', 'unknown'].includes(v)
  },
  size: {
    type: String,
    default: 'default',
    validator: (v) => ['sm', 'default', 'lg'].includes(v)
  },
  pulse: Boolean,
  showLabel: {
    type: Boolean,
    default: true
  }
})

const statusConfig = {
  healthy: {
    icon: CheckCircle,
    label: '正常',
    color: 'text-success',
    bg: 'bg-success/10',
    border: 'border-success/30',
    glow: 'shadow-success/20'
  },
  unhealthy: {
    icon: XCircle,
    label: '异常',
    color: 'text-destructive',
    bg: 'bg-destructive/10',
    border: 'border-destructive/30',
    glow: 'shadow-destructive/20'
  },
  warning: {
    icon: AlertCircle,
    label: '警告',
    color: 'text-warning',
    bg: 'bg-warning/10',
    border: 'border-warning/30',
    glow: 'shadow-warning/20'
  },
  unknown: {
    icon: HelpCircle,
    label: '未知',
    color: 'text-muted-foreground',
    bg: 'bg-muted',
    border: 'border-muted-foreground/30',
    glow: ''
  }
}

const config = computed(() => statusConfig[props.status] || statusConfig.unknown)

const sizeClasses = {
  sm: 'h-4 w-4',
  default: 'h-5 w-5',
  lg: 'h-6 w-6'
}

const iconSize = computed(() => sizeClasses[props.size])

const badgeSizeClasses = {
  sm: 'px-2 py-0.5 text-xs gap-1',
  default: 'px-2.5 py-1 text-sm gap-1.5',
  lg: 'px-3 py-1.5 text-base gap-2'
}
</script>

<template>
  <div
    :class="cn(
      'inline-flex items-center rounded-full border font-medium transition-all',
      config.bg,
      config.border,
      config.color,
      badgeSizeClasses[size],
      pulse && 'animate-pulse-glow',
      config.glow && `shadow-lg ${config.glow}`
    )"
  >
    <component :is="config.icon" :class="iconSize" />
    <span v-if="showLabel">{{ config.label }}</span>
  </div>
</template>

