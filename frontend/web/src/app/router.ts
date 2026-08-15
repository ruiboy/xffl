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
      redirect: { name: 'ffl-ladder' },
    },
    {
      path: '/ffl/ladder',
      name: 'ffl-ladder',
      component: () => import('@/features/ffl/views/HomeView.vue'),
    },
    {
      path: '/ffl/seasons',
      name: 'ffl-seasons',
      component: () => import('@/features/ffl/views/SeasonsIndexView.vue'),
    },
    {
      path: '/ffl/seasons/:seasonId',
      name: 'ffl-season',
      component: () => import('@/features/ffl/views/SeasonView.vue'),
      props: true,
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
      path: '/ffl/clubs/:clubId',
      name: 'ffl-club',
      component: () => import('@/features/ffl/views/ClubView.vue'),
      props: true,
    },
    {
      path: '/ffl/club-seasons/:clubSeasonId',
      name: 'ffl-club-season',
      component: () => import('@/features/ffl/views/SquadView.vue'),
      props: true,
    },
    {
      path: '/ffl/club-matches/:clubMatchId',
      name: 'ffl-club-match',
      component: () => import('@/features/ffl/views/TeamBuilderView.vue'),
      props: (route) => ({ clubMatchId: route.params.clubMatchId as string, readonly: true }),
    },
    {
      path: '/ffl/club-matches/:clubMatchId/edit',
      name: 'ffl-club-match-edit',
      component: () => import('@/features/ffl/views/TeamBuilderView.vue'),
      props: true,
    },
    {
      path: '/ffl/afl/player-seasons/:aflPlayerSeasonId',
      name: 'ffl-afl-player-season',
      component: () => import('@/features/ffl/views/AFLPlayerSeasonView.vue'),
      props: true,
    },
    {
      path: '/ffl/afl/club-seasons/:clubSeasonId',
      name: 'ffl-afl-club-season',
      component: () => import('@/features/afl/views/ClubSeasonView.vue'),
      props: true,
    },
    {
      path: '/ffl/free-agents',
      name: 'ffl-free-agents',
      component: () => import('@/features/ffl/views/FreeAgentsView.vue'),
    },
    {
      path: '/ffl/data-ops',
      name: 'ffl-data-ops',
      component: () => import('@/features/data-ops/views/DataOpsView.vue'),
    },
    {
      path: '/ffl/admin',
      name: 'ffl-admin',
      component: () => import('@/features/admin/views/AdminView.vue'),
    },

    // AFL routes
    {
      path: '/afl',
      name: 'afl-home',
      redirect: { name: 'afl-ladder' },
    },
    {
      path: '/afl/ladder',
      name: 'afl-ladder',
      component: () => import('@/features/afl/views/HomeView.vue'),
    },
    {
      path: '/afl/seasons',
      name: 'afl-seasons',
      component: () => import('@/features/afl/views/SeasonsIndexView.vue'),
    },
    {
      path: '/afl/seasons/:seasonId',
      name: 'afl-season',
      component: () => import('@/features/afl/views/SeasonView.vue'),
      props: true,
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

    // Unmatched paths. Entity-level 404s are handled inside each view (the
    // route matches, the id just doesn't resolve); this catches the rest.
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/components/NotFound.vue'),
      props: { entity: 'Page' },
    },
  ],
})

export default router
