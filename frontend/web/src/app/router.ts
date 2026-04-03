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
    {
      path: '/ffl/seasons/:seasonId/rounds/:roundId',
      name: 'ffl-round',
      component: () => import('@/features/ffl/views/RoundView.vue'),
      props: true,
    },
    {
      path: '/ffl/seasons/:seasonId/matches/:matchId',
      name: 'ffl-match',
      component: () => import('@/features/ffl/views/MatchView.vue'),
      props: true,
    },
    {
      path: '/ffl/seasons/:seasonId/roster',
      name: 'ffl-roster',
      component: () => import('@/features/ffl/views/RosterView.vue'),
      props: true,
    },
    {
      path: '/ffl/seasons/:seasonId/rounds/:roundId/team-builder',
      name: 'ffl-team-builder',
      component: () => import('@/features/ffl/views/TeamBuilderView.vue'),
      props: true,
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
