<script setup>
import { cn } from '@/lib/utils'

const props = defineProps({
  class: String,
  hover: Boolean,
  glow: Boolean
})
</script>

<template>
  <div
    :class="cn(
      'rounded-xl border bg-card text-card-foreground shadow-sm transition-all duration-300',
      hover && 'hover:border-primary/50 hover:shadow-lg',
      glow && 'glow',
      props.class
    )"
  >
    <slot />
  </div>
</template>

