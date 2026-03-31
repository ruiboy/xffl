import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    // FFL routes
    {
      path: '/',
      name: 'home',
      component: () => import('@/features/ffl/views/HomeView.vue'),
    },

    // AFL routes
    {
      path: '/afl',
      name: 'afl-home',
      component: () => import('@/features/afl/views/HomeView.vue'),
    },
    {
      path: '/afl/seasons/:seasonId/rounds/:roundId',
      name: 'afl-round',
      component: () => import('@/features/afl/views/RoundView.vue'),
      props: true,
    },
    {
      path: '/afl/seasons/:seasonId/matches/:matchId',
      name: 'afl-match',
      component: () => import('@/features/afl/views/MatchView.vue'),
      props: true,
    },
    {
      path: '/admin/afl/seasons/:seasonId/matches/:matchId',
      name: 'afl-admin-match',
      component: () => import('@/features/afl/views/AdminMatchView.vue'),
      props: true,
    },
  ],
})

export default router
