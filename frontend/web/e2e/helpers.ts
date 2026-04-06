import type { Page } from '@playwright/test'

/**
 * Navigates to the FFL home page and waits until all three session cookies are set.
 *
 * Cookie chain:
 *   1. HomeView runs GET_FFL_LATEST_ROUND → sets xffl_season_id + xffl_round_id
 *   2. App.vue's GET_FFL_SEASON_CLUBS becomes enabled (needs xffl_season_id) →
 *      clubs load → first club auto-selected → sets xffl_club_id
 *
 * Polling document.cookie directly is the most reliable signal — it fires only
 * once all three watches have completed, regardless of render timing.
 */
export async function setupFflSession(page: Page) {
  await page.goto('/ffl')
  await page.waitForFunction(
    () =>
      document.cookie.includes('xffl_season_id=') &&
      document.cookie.includes('xffl_round_id=') &&
      document.cookie.includes('xffl_club_id='),
    { timeout: 15000 },
  )
}
