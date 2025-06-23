<template>
  <div class="min-h-screen bg-surface-0 dark:bg-surface-900">
    <nav class="bg-surface-50 dark:bg-surface-800 border-b border-surface-200 dark:border-surface-700 p-4">
      <div class="max-w-7xl mx-auto">
        <div class="flex justify-center space-x-6">
          <Button 
            :label="'Home'"
            @click="$router.push('/')"
            :outlined="$route.path !== '/'"
            :severity="$route.path === '/' ? 'primary' : 'secondary'"
            class="px-6 py-2"
          />
          <Button 
            :label="'Players'"
            @click="$router.push('/players')"
            :outlined="$route.path !== '/players'"
            :severity="$route.path === '/players' ? 'primary' : 'secondary'"
            class="px-6 py-2"
          />
          <Button 
            :label="'AFL Clubs'"
            @click="$router.push('/afl-clubs')"
            :outlined="$route.path !== '/afl-clubs'"
            :severity="$route.path === '/afl-clubs' ? 'primary' : 'secondary'"
            class="px-6 py-2"
          />
        </div>
      </div>
    </nav>
    <main class="max-w-7xl mx-auto p-6">
      <router-view></router-view>
    </main>
  </div>
</template>

<script setup lang="ts">
// Component logic will go here
</script> 
