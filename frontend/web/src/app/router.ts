import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    // Redirect root to FFL
    {
      path: '/',
      redirect: '/ffl',
    },

    // FFL routes
    {
      path: '/ffl',
      name: 'home',
      component: () => import('@/features/ffl/views/HomeView.vue'),
    },
    {
      path: '/ffl/rounds/:roundId',
      name: 'ffl-round',
      component: () => import('@/features/ffl/views/RoundView.vue'),
      props: true,
    },
    {
      path: '/ffl/matches/:matchId',
      name: 'ffl-match',
      component: () => import('@/features/ffl/views/MatchView.vue'),
      props: true,
    },
    {
      path: '/ffl/club-seasons/:clubSeasonId',
      name: 'ffl-club-season',
      component: () => import('@/features/ffl/views/SquadView.vue'),
      props: true,
    },
    {
      path: '/ffl/seasons/:seasonId/rounds/:roundId/team-builder',
      name: 'ffl-team-builder',
      component: () => import('@/features/ffl/views/TeamBuilderView.vue'),
      props: true,
    },
    {
      path: '/ffl/data-ops',
      name: 'ffl-data-ops',
      component: () => import('@/features/data-ops/views/DataOpsView.vue'),
    },

    // AFL routes
    {
      path: '/afl',
      name: 'afl-home',
      component: () => import('@/features/afl/views/HomeView.vue'),
    },
    {
      path: '/afl/rounds/:roundId',
      name: 'afl-round',
      component: () => import('@/features/afl/views/RoundView.vue'),
      props: true,
    },
    {
      path: '/afl/matches/:matchId',
      name: 'afl-match',
      component: () => import('@/features/afl/views/MatchView.vue'),
      props: true,
    },
    {
      path: '/afl/matches/:matchId/edit',
      name: 'afl-match-edit',
      component: () => import('@/features/afl/views/AdminMatchView.vue'),
      props: true,
    },

    // Redirects from old hierarchical paths
    { path: '/ffl/seasons/:seasonId/rounds/:roundId', redirect: to => ({ name: 'ffl-round', params: { roundId: to.params.roundId } }) },
    { path: '/ffl/seasons/:seasonId/matches/:matchId', redirect: to => ({ name: 'ffl-match', params: { matchId: to.params.matchId } }) },
    { path: '/afl/seasons/:seasonId/rounds/:roundId', redirect: to => ({ name: 'afl-round', params: { roundId: to.params.roundId } }) },
    { path: '/afl/seasons/:seasonId/matches/:matchId', redirect: to => ({ name: 'afl-match', params: { matchId: to.params.matchId } }) },
    { path: '/admin/afl/seasons/:seasonId/matches/:matchId', redirect: to => ({ name: 'afl-match-edit', params: { matchId: to.params.matchId } }) },
  ],
})

export default router
