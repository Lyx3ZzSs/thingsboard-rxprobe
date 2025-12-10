<script setup>
import { computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps({
  modelValue: Boolean,
  disabled: Boolean,
  class: String
})

const emit = defineEmits(['update:modelValue'])

const isChecked = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

function toggle() {
  if (!props.disabled) {
    isChecked.value = !isChecked.value
  }
}
</script>

<template>
  <button
    type="button"
    role="switch"
    :aria-checked="isChecked"
    @click="toggle"
    :disabled="disabled"
    :class="cn(
      'peer inline-flex h-6 w-11 shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:cursor-not-allowed disabled:opacity-50',
      isChecked ? 'bg-primary' : 'bg-input',
      props.class
    )"
  >
    <span
      :class="cn(
        'pointer-events-none block h-5 w-5 rounded-full bg-background shadow-lg ring-0 transition-transform',
        isChecked ? 'translate-x-5' : 'translate-x-0'
      )"
    />
  </button>
</template>

