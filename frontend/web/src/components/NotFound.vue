<template>
  <div class="py-16 text-center">
    <p class="text-5xl font-bold text-text-faint mb-3">404</p>
    <h1 class="text-xl font-semibold text-text-heading mb-2">{{ entity }} not found</h1>
    <p class="text-text-muted mb-6">
      Looks like this {{ entity.toLowerCase() }} has been delisted. It can't be found anywhere.
    </p>
    <router-link
      :to="{ name: 'ffl-ladder' }"
      class="inline-block rounded-md bg-control px-4 py-2 text-sm font-medium hover:bg-control-hover transition-colors"
    >
      Back to the ladder
    </router-link>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ entity?: string }>(), { entity: 'Page' })
</script>
