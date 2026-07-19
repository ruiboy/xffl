import { test, expect } from './fixtures'
import { setupFflSession } from './helpers'
import type { Page } from '@playwright/test'

/**
 * The fixture builder's round body: matches lead and own the only button, with
 * byes and the superbye as quiet leftovers below.
 *
 * The seeded FFL 2026 season has two clubs and its own rounds, so each test
 * appends a fresh round and works on that. Nothing here saves, so the seeded
 * fixtures stay untouched.
 */

/** The round card added by a test — always the last one on the page. */
function lastRound(page: Page) {
  return page.getByTestId('round').last()
}

test.describe('Admin fixture builder', () => {
  test.beforeEach(async ({ page }) => {
    await setupFflSession(page)
    await page.goto('/ffl/admin')
    // Wait for the view to mount before clicking — clicking a pre-hydration
    // node lands on a element Vue is about to replace, and the click is lost.
    await page.getByRole('heading', { name: 'Admin', level: 1 }).waitFor()
    // A season row's Edit opens the builder on that season, which is the path
    // an admin actually takes.
    await page.getByRole('row', { name: /FFL 2026/ })
      .getByRole('button', { name: 'Edit' }).click()
    // "+ Add round" renders as soon as a season is picked, but the seeded rounds
    // arrive with a later query. Wait for those too, or a round added here can
    // be counted before them and stop being the last one on the page.
    await page.getByRole('button', { name: '+ Add round' }).waitFor({ timeout: 15000 })
    await page.getByTestId('round').first().waitFor({ timeout: 15000 })
  })

  test("Edit loads the season's existing rounds", async ({ page }) => {
    const rounds = page.getByTestId('round')
    await expect(rounds.first()).toBeVisible()
    expect(await rounds.count()).toBeGreaterThan(0)
  })

  test('Add round appends a round to the end', async ({ page }) => {
    const before = await page.getByTestId('round').count()
    await page.getByRole('button', { name: '+ Add round' }).click()
    await expect(page.getByTestId('round')).toHaveCount(before + 1)
  })

  test('a new round leads with Add match', async ({ page }) => {
    await page.getByRole('button', { name: '+ Add round' }).click()
    await expect(lastRound(page).getByRole('button', { name: '+ Add match' })).toBeVisible()
  })

  test('a new round lists its clubs as unplaced', async ({ page }) => {
    await page.getByRole('button', { name: '+ Add round' }).click()
    // Both seeded clubs start with nowhere to be.
    await expect(lastRound(page).getByTestId('unplaced-club')).toHaveCount(2)
  })

  test('Add match seeds a real pair rather than empty dropdowns', async ({ page }) => {
    await page.getByRole('button', { name: '+ Add round' }).click()
    await lastRound(page).getByRole('button', { name: '+ Add match' }).click()

    const match = lastRound(page).getByTestId('match')
    await expect(match).toHaveCount(1)
    await expect(match.locator('select').first()).not.toHaveValue('')
    await expect(match.locator('select').last()).not.toHaveValue('')
  })

  test('clubs placed in a match leave the unplaced line', async ({ page }) => {
    await page.getByRole('button', { name: '+ Add round' }).click()
    await expect(lastRound(page).getByTestId('unplaced-club')).toHaveCount(2)

    await lastRound(page).getByRole('button', { name: '+ Add match' }).click()

    // Both clubs are now in the match, so nothing is left over at all.
    await expect(lastRound(page).getByTestId('unplaced')).toHaveCount(0)
  })

  test('clicking an unplaced club gives it a bye', async ({ page }) => {
    await page.getByRole('button', { name: '+ Add round' }).click()
    const club = lastRound(page).getByTestId('unplaced-club').first()
    const name = (await club.textContent())!.trim()

    await club.click()

    await expect(lastRound(page).getByTestId('byes')).toContainText(name)
    await expect(lastRound(page).getByTestId('unplaced-club')).toHaveCount(1)
  })

  test('superbye all sweeps up every remaining club and explains itself', async ({ page }) => {
    await page.getByRole('button', { name: '+ Add round' }).click()
    await lastRound(page).getByTestId('superbye-all').click()

    await expect(lastRound(page).getByTestId('superbye-club')).toHaveCount(2)
    await expect(lastRound(page).getByTestId('superbye'))
      .toContainText('each submit a team; the highest scorer earns 1 point')
    await expect(lastRound(page).getByTestId('unplaced')).toHaveCount(0)
  })

  test('bye all gives every remaining club a bye', async ({ page }) => {
    await page.getByRole('button', { name: '+ Add round' }).click()
    await lastRound(page).getByTestId('bye-all').click()

    await expect(lastRound(page).getByTestId('bye-club')).toHaveCount(2)
    await expect(lastRound(page).getByTestId('unplaced')).toHaveCount(0)
  })

  test('a one-club superbye is rejected as a bye in disguise', async ({ page }) => {
    await page.getByRole('button', { name: '+ Add round' }).click()
    await lastRound(page).getByTestId('superbye-all').click()

    // Drop one club, leaving a superbye of one — which has no field to top.
    await lastRound(page).getByTitle('Remove from superbye').first().click()

    await expect(page.getByText(/a superbye needs at least two clubs/)).toBeVisible()
    await expect(page.getByRole('button', { name: 'Save fixtures' })).toBeDisabled()
  })

  test('removing a club from the superbye returns it to unplaced', async ({ page }) => {
    await page.getByRole('button', { name: '+ Add round' }).click()
    await lastRound(page).getByTestId('superbye-all').click()
    await expect(lastRound(page).getByTestId('unplaced')).toHaveCount(0)

    await lastRound(page).getByTitle('Remove from superbye').first().click()

    await expect(lastRound(page).getByTestId('unplaced-club')).toHaveCount(1)
  })
})
