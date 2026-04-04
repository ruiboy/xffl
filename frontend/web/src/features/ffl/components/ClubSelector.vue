<template>
  <div class="relative" ref="container">
    <button
      @click="open = !open"
      class="flex items-center gap-2 rounded-lg border border-border bg-surface px-3 py-1 text-sm text-text hover:bg-surface-hover transition-colors focus:outline-none focus:border-active"
    >
      <img
        v-if="selectedLogoUrl"
        :src="selectedLogoUrl"
        :alt="selectedClubName"
        class="w-5 h-5 object-contain"
      />
      <span>{{ selectedClubName || 'Select club' }}</span>
      <svg class="w-3 h-3 text-text-muted" viewBox="0 0 12 12" fill="currentColor">
        <path d="M2 4l4 4 4-4" stroke="currentColor" stroke-width="1.5" fill="none" stroke-linecap="round"/>
      </svg>
    </button>

    <div
      v-if="open"
      class="absolute right-0 top-full mt-1 z-50 min-w-[160px] rounded-lg border border-border bg-surface-raised shadow-lg py-1"
    >
      <button
        v-for="cs in clubs"
        :key="cs.club.id"
        @click="select(cs.club.id, cs.club.name)"
        class="flex items-center gap-2 w-full px-3 py-2 text-sm text-left hover:bg-surface-hover transition-colors"
        :class="cs.club.id === modelValue ? 'text-text font-medium' : 'text-text-muted'"
      >
        <img
          v-if="logoUrl(cs.club.name)"
          :src="logoUrl(cs.club.name)"
          :alt="cs.club.name"
          class="w-5 h-5 object-contain"
        />
        {{ cs.club.name }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { clubLogoUrl } from '../utils/clubLogos'

interface ClubSeasonEntry {
  club: { id: string; name: string }
}

const props = defineProps<{
  modelValue: string
  clubs: ClubSeasonEntry[]
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', id: string): void
}>()

const open = ref(false)
const container = ref<HTMLElement | null>(null)

const selectedClub = computed(() => props.clubs.find(cs => cs.club.id === props.modelValue)?.club ?? null)
const selectedClubName = computed(() => selectedClub.value?.name ?? '')
const selectedLogoUrl = computed(() => selectedClub.value ? clubLogoUrl(selectedClub.value.name) : '')

function logoUrl(name: string) {
  return clubLogoUrl(name)
}

function select(id: string, _name: string) {
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
