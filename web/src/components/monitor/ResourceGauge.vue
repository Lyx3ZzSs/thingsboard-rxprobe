<script setup>
import { computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps({
  label: {
    type: String,
    required: true
  },
  value: {
    type: Number,
    required: true
  },
  max: {
    type: Number,
    default: 100
  },
  unit: {
    type: String,
    default: '%'
  },
  thresholds: {
    type: Object,
    default: () => ({ warning: 70, danger: 90 })
  },
  size: {
    type: String,
    default: 'default',
    validator: (v) => ['sm', 'default', 'lg'].includes(v)
  }
})

const percentage = computed(() => Math.min(100, Math.max(0, (props.value / props.max) * 100)))

const status = computed(() => {
  if (percentage.value >= props.thresholds.danger) return 'danger'
  if (percentage.value >= props.thresholds.warning) return 'warning'
  return 'normal'
})

const colorClass = computed(() => {
  const colors = {
    normal: 'text-primary',
    warning: 'text-warning',
    danger: 'text-destructive'
  }
  return colors[status.value]
})

const strokeColor = computed(() => {
  const colors = {
    normal: '#0ea5e9', // 电力蓝
    warning: '#f59e0b',
    danger: '#ef4444'
  }
  return colors[status.value]
})

const trackColor = 'hsl(215, 25%, 15%)'

const sizes = {
  sm: { size: 70, stroke: 5 },
  default: { size: 90, stroke: 6 },
  lg: { size: 110, stroke: 8 }
}

const sizeConfig = computed(() => sizes[props.size])

const radius = computed(() => (sizeConfig.value.size - sizeConfig.value.stroke) / 2)
const circumference = computed(() => 2 * Math.PI * radius.value)
const offset = computed(() => circumference.value - (percentage.value / 100) * circumference.value)
</script>

<template>
  <div class="flex flex-col items-center gap-2">
    <div class="relative" :style="{ width: `${sizeConfig.size}px`, height: `${sizeConfig.size}px` }">
      <svg 
        :width="sizeConfig.size" 
        :height="sizeConfig.size"
        class="transform -rotate-90"
      >
        <!-- 背景圆环 -->
        <circle
          :cx="sizeConfig.size / 2"
          :cy="sizeConfig.size / 2"
          :r="radius"
          :stroke="trackColor"
          :stroke-width="sizeConfig.stroke"
          fill="none"
        />
        <!-- 进度圆环 -->
        <circle
          :cx="sizeConfig.size / 2"
          :cy="sizeConfig.size / 2"
          :r="radius"
          :stroke="strokeColor"
          :stroke-width="sizeConfig.stroke"
          fill="none"
          :stroke-dasharray="circumference"
          :stroke-dashoffset="offset"
          stroke-linecap="round"
          class="transition-all duration-500 ease-out"
          :style="{ filter: `drop-shadow(0 0 4px ${strokeColor})` }"
        />
      </svg>
      
      <!-- 中心文字 -->
      <div class="absolute inset-0 flex flex-col items-center justify-center">
        <span :class="cn('font-bold font-mono', colorClass)" :style="{ fontSize: size === 'sm' ? '14px' : size === 'lg' ? '22px' : '18px' }">
          {{ value.toFixed(1) }}
        </span>
        <span class="text-[10px] text-muted-foreground">{{ unit }}</span>
      </div>
    </div>
    
    <span class="text-xs text-muted-foreground">{{ label }}</span>
  </div>
</template>
