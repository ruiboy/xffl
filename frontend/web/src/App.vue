<template>
  <div class="min-h-screen bg-gray-50 text-gray-900">
    <nav class="border-b border-gray-200 bg-white px-6 py-4">
      <div class="max-w-5xl mx-auto flex items-center gap-6">
        <router-link to="/" class="text-lg font-semibold tracking-tight text-gray-900 hover:text-gray-600 transition-colors">
          xFFL
        </router-link>
        <router-link to="/" class="text-sm text-gray-500 hover:text-gray-900 transition-colors">
          Home
        </router-link>
      </div>
    </nav>
    <main class="max-w-5xl mx-auto px-6 py-8">
      <router-view />
    </main>
  </div>
</template>
