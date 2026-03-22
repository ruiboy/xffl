import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/features/afl/views/SeasonsView.vue'),
    },
    {
      path: '/afl/seasons/:seasonId',
      name: 'season',
      component: () => import('@/features/afl/views/SeasonView.vue'),
      props: true,
    },
    {
      path: '/afl/seasons/:seasonId/matches/:matchId',
      name: 'match',
      component: () => import('@/features/afl/views/MatchView.vue'),
      props: true,
    },
  ],
})

export default router
