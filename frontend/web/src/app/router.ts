import { createRouter, createWebHistory } from 'vue-router'
import { useFflState } from '@/features/ffl/composables/useFflState'
import { useAflState } from '@/features/afl/composables/useAflState'

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
      redirect: () => {
        const { liveRoundId } = useFflState()
        return liveRoundId.value
          ? { name: 'ffl-round', params: { roundId: liveRoundId.value } }
          : { name: 'ffl-ladder' }
      },
    },
    {
      path: '/ffl/ladder',
      name: 'ffl-ladder',
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
      path: '/ffl/data-ops',
      name: 'ffl-data-ops',
      component: () => import('@/features/data-ops/views/DataOpsView.vue'),
    },

    // AFL routes
    {
      path: '/afl',
      name: 'afl-home',
      redirect: () => {
        const { liveRoundId } = useAflState()
        return liveRoundId.value
          ? { name: 'afl-round', params: { roundId: liveRoundId.value } }
          : { name: 'afl-ladder' }
      },
    },
    {
      path: '/afl/ladder',
      name: 'afl-ladder',
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
  ],
})

export default router
