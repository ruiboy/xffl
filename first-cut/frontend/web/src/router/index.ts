import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'
import Players from '../views/Players.vue'
import AFLClubs from '../views/AFLClubs.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home
    },
    {
      path: '/players',
      name: 'players',
      component: Players
    },
    {
      path: '/afl-clubs',
      name: 'afl-clubs',
      component: AFLClubs
    }
  ]
})

export default router 