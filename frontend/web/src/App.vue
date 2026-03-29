<template>
  <div class="min-h-screen bg-gray-950 text-gray-100">
    <nav class="border-b border-gray-800 px-6 py-4">
      <div class="max-w-5xl mx-auto flex items-center gap-6">
        <router-link to="/" class="text-lg font-semibold tracking-tight hover:text-white transition-colors">
          xFFL
        </router-link>
        <router-link to="/" class="text-sm text-gray-400 hover:text-gray-200 transition-colors">
          Home
        </router-link>
      </div>
    </nav>
    <main class="max-w-5xl mx-auto px-6 py-8">
      <router-view />
    </main>
  </div>
</template>
