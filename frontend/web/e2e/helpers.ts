import type { Page } from '@playwright/test'

/**
 * Navigates to the AFL home page and waits until the AFL session cookie is set.
 * The cookie is written once the aflLiveRound query resolves.
 */
export async function setupAflSession(page: Page) {
  await page.goto('/afl')
  await page.waitForFunction(
    () => document.cookie.includes('xffl_afl='),
    { timeout: 15000 },
  )
}

/**
 * Navigates to the FFL home page and waits until all session cookies are set.
 *
 * Cookie chain:
 *   1. HomeView queries GET_AFL_LIVE_ROUND → GET_FFL_ROUND_BY_AFL_ROUND → writes xffl_ffl JSON cookie
 *   2. App.vue's GET_FFL_SEASON_CLUBS becomes enabled (needs liveSeasonId from xffl_ffl) →
 *      clubs load → first club auto-selected → sets xffl_club_id
 */
export async function setupFflSession(page: Page) {
  await page.goto('/ffl')
  await page.waitForFunction(
    () =>
      document.cookie.includes('xffl_ffl=') &&
      document.cookie.includes('xffl_club_id='),
    { timeout: 15000 },
  )
}
