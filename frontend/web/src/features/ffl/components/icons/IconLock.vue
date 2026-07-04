<template>
  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
    <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
    <path v-if="!open" d="M7 11V7a5 5 0 0 1 10 0v4"/>
    <path v-else d="M7 11V7a5 5 0 0 1 9.9-1"/>
  </svg>
</template>

<script setup lang="ts">
defineProps<{ open?: boolean }>()
</script>
