<script setup>
import { computed } from 'vue'
import { Card, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils'

const props = defineProps({
  title: {
    type: String,
    required: true
  },
  value: {
    type: [Number, String],
    required: true
  },
  icon: Object,
  trend: {
    type: String,
    validator: (v) => ['up', 'down', 'stable'].includes(v)
  },
  trendValue: String,
  variant: {
    type: String,
    default: 'default',
    validator: (v) => ['default', 'primary', 'success', 'warning', 'destructive', 'secondary'].includes(v)
  },
  class: String
})

const variantStyles = {
  default: {
    icon: 'bg-muted text-muted-foreground',
    border: 'border-border',
    glow: ''
  },
  primary: {
    icon: 'bg-primary/15 text-primary',
    border: 'border-primary/30',
    glow: 'glow'
  },
  success: {
    icon: 'bg-success/15 text-success',
    border: 'border-success/30',
    glow: 'glow-success'
  },
  warning: {
    icon: 'bg-warning/15 text-warning',
    border: 'border-warning/30',
    glow: 'glow-warning'
  },
  destructive: {
    icon: 'bg-destructive/15 text-destructive',
    border: 'border-destructive/30',
    glow: 'glow-destructive'
  },
  secondary: {
    icon: 'bg-secondary/15 text-secondary-foreground',
    border: 'border-secondary/30',
    glow: ''
  }
}

const styles = computed(() => variantStyles[props.variant] || variantStyles.default)

const trendColors = {
  up: 'text-success',
  down: 'text-destructive',
  stable: 'text-muted-foreground'
}
</script>

<template>
  <Card 
    :class="cn(
      'overflow-hidden transition-all duration-300 hover:shadow-xl',
      styles.border,
      styles.glow,
      props.class
    )"
  >
    <CardContent class="p-5">
      <div class="flex items-center justify-between">
        <div class="flex-1">
          <p class="text-sm font-medium text-muted-foreground mb-1">{{ title }}</p>
          <p class="text-3xl font-bold tracking-tight font-mono">{{ value }}</p>
          <p v-if="trendValue" class="text-xs mt-1" :class="trendColors[trend]">
            <span v-if="trend === 'up'">↑</span>
            <span v-else-if="trend === 'down'">↓</span>
            <span v-else>→</span>
            {{ trendValue }}
          </p>
        </div>
        <div 
          v-if="icon" 
          :class="cn(
            'flex items-center justify-center w-12 h-12 rounded-lg',
            styles.icon
          )"
        >
          <component :is="icon" class="w-6 h-6" />
        </div>
      </div>
    </CardContent>
  </Card>
</template>
