<script setup>
import { computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps({
  value: {
    type: Number,
    default: 0
  },
  max: {
    type: Number,
    default: 100
  },
  variant: {
    type: String,
    default: 'default',
    validator: (v) => ['default', 'success', 'warning', 'destructive'].includes(v)
  },
  showValue: Boolean,
  class: String
})

const percentage = computed(() => Math.min(100, Math.max(0, (props.value / props.max) * 100)))

const indicatorClass = computed(() => {
  const variants = {
    default: 'bg-primary',
    success: 'bg-success',
    warning: 'bg-warning',
    destructive: 'bg-destructive'
  }
  return variants[props.variant]
})
</script>

<template>
  <div class="flex items-center gap-2">
    <div
      :class="cn('relative h-2 w-full overflow-hidden rounded-full bg-secondary', props.class)"
    >
      <div
        :class="cn('h-full transition-all duration-500 ease-out', indicatorClass)"
        :style="{ width: `${percentage}%` }"
      />
    </div>
    <span v-if="showValue" class="text-xs text-muted-foreground tabular-nums">
      {{ value.toFixed(1) }}%
    </span>
  </div>
</template>

