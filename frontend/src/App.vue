<template>
  <div class="app">
    <nav>
      <div class="nav-links">
        <router-link to="/">Home</router-link>
        <router-link to="/players">Players</router-link>
      </div>
    </nav>
    <router-view></router-view>
  </div>
</template>

<script setup lang="ts">
// Component logic will go here
</script>

<style>
.app {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

nav {
  padding: 20px 0;
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

nav h1 {
  margin: 0;
  color: #2c3e50;
  font-size: 24px;
}

.nav-links {
  display: flex;
  justify-content: center;
  flex: 1;
}

nav a {
  color: #2c3e50;
  text-decoration: none;
  margin-right: 10px;
}

nav a.router-link-active {
  color: #42b983;
}
</style> 