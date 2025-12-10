<script setup>
import { ref, computed, watch } from 'vue'
import { cn } from '@/lib/utils'
import { ChevronDown, Check } from 'lucide-vue-next'

const props = defineProps({
  modelValue: [String, Number],
  options: {
    type: Array,
    default: () => []
  },
  placeholder: {
    type: String,
    default: '请选择'
  },
  disabled: Boolean,
  class: String
})

const emit = defineEmits(['update:modelValue'])

const isOpen = ref(false)

const selectedLabel = computed(() => {
  const option = props.options.find(o => o.value === props.modelValue)
  return option?.label || props.placeholder
})

function select(value) {
  emit('update:modelValue', value)
  isOpen.value = false
}

function toggle() {
  if (!props.disabled) {
    isOpen.value = !isOpen.value
  }
}

// 点击外部关闭
function handleClickOutside(e) {
  if (!e.target.closest('.select-wrapper')) {
    isOpen.value = false
  }
}

watch(isOpen, (val) => {
  if (val) {
    document.addEventListener('click', handleClickOutside)
  } else {
    document.removeEventListener('click', handleClickOutside)
  }
})
</script>

<template>
  <div class="select-wrapper relative" :class="props.class">
    <button
      type="button"
      @click="toggle"
      :disabled="disabled"
      :class="cn(
        'flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50',
        isOpen && 'ring-2 ring-ring ring-offset-2'
      )"
    >
      <span :class="{ 'text-muted-foreground': !modelValue }">{{ selectedLabel }}</span>
      <ChevronDown class="h-4 w-4 opacity-50 transition-transform" :class="{ 'rotate-180': isOpen }" />
    </button>
    
    <Transition
      enter-active-class="transition duration-100 ease-out"
      enter-from-class="transform scale-95 opacity-0"
      enter-to-class="transform scale-100 opacity-100"
      leave-active-class="transition duration-75 ease-in"
      leave-from-class="transform scale-100 opacity-100"
      leave-to-class="transform scale-95 opacity-0"
    >
      <div
        v-if="isOpen"
        class="absolute z-50 mt-1 w-full rounded-md border bg-popover p-1 shadow-md"
      >
        <div
          v-for="option in options"
          :key="option.value"
          @click="select(option.value)"
          :class="cn(
            'relative flex cursor-pointer select-none items-center rounded-sm py-1.5 pl-8 pr-2 text-sm outline-none hover:bg-accent hover:text-accent-foreground',
            option.value === modelValue && 'bg-accent'
          )"
        >
          <span class="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
            <Check v-if="option.value === modelValue" class="h-4 w-4" />
          </span>
          {{ option.label }}
        </div>
      </div>
    </Transition>
  </div>
</template>

