<template>
  <div class="relative" ref="container">
    <button
      @click="open = !open"
      class="flex items-center gap-2 rounded-lg border border-border bg-surface px-3 py-1 text-sm text-text hover:bg-surface-hover transition-colors focus:outline-none focus:border-active"
      title="Season"
    >
      <span>{{ selectedName || 'Season' }}</span>
      <svg class="w-3 h-3 text-text-muted" viewBox="0 0 12 12" fill="currentColor">
        <path d="M2 4l4 4 4-4" stroke="currentColor" stroke-width="1.5" fill="none" stroke-linecap="round"/>
      </svg>
    </button>

    <div
      v-if="open"
      class="absolute right-0 top-full mt-1 z-50 min-w-[140px] max-h-80 overflow-y-auto rounded-lg border border-border bg-surface-raised shadow-lg py-1"
    >
      <button
        v-for="s in seasons"
        :key="s.id"
        @click="select(s.id)"
        class="block w-full px-3 py-2 text-sm text-left hover:bg-surface-hover transition-colors"
        :class="s.id === modelValue ? 'text-text font-medium' : 'text-text-muted'"
      >
        {{ s.name }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

interface Season {
  id: string
  name: string
}

const props = defineProps<{
  modelValue: string
  seasons: Season[]
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', id: string): void
}>()

const open = ref(false)
const container = ref<HTMLElement | null>(null)

const selectedName = computed(() => props.seasons.find(s => s.id === props.modelValue)?.name ?? '')

function select(id: string) {
  emit('update:modelValue', id)
  open.value = false
}

function onClickOutside(e: MouseEvent) {
  if (container.value && !container.value.contains(e.target as Node)) {
    open.value = false
  }
}

onMounted(() => document.addEventListener('mousedown', onClickOutside))
onUnmounted(() => document.removeEventListener('mousedown', onClickOutside))
</script>
