import { test, expect } from './fixtures'

test.describe('FFL Match', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate: FFL Home → Round 1 → match (match card is a div, not a link)
    await page.goto('/ffl')
    await page.locator('main nav').getByRole('link', { name: '1', exact: true }).click()
    await page.locator('.cursor-pointer').filter({ hasText: 'Ruiboys' }).filter({ hasText: 'The Howling Cows' }).click()
    await page.waitForURL(/\/ffl\/matches\//)
  })

  test('displays match header with teams', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Ruiboys')
    await expect(page.getByRole('heading', { level: 1 })).toContainText('The Howling Cows')
  })

  test('displays venue', async ({ page }) => {
    await expect(page.getByText('MCG')).toBeVisible()
  })

  test('displays fantasy scores', async ({ page }) => {
    // Each team header shows its total: Ruiboys 85, The Howling Cows 72.
    await expect(page.getByText('85', { exact: true })).toBeVisible()
    await expect(page.getByText('72', { exact: true })).toBeVisible()
  })

  test('displays on-field player count next to each score', async ({ page }) => {
    // Each side has one played starter (Jordan Dawson / Henry Smith) → (1/18),
    // shown while the team is not final, with the full text on hover.
    await expect(page.getByText('(1/18)')).toHaveCount(2)
    await expect(page.locator('[title="1 of 18 played"]')).toHaveCount(2)
  })

  test('displays squad table with player columns', async ({ page }) => {
    await expect(page.getByRole('columnheader', { name: 'Player' }).first()).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Status' }).first()).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Score' }).first()).toBeVisible()
  })

  test('squad table groups players by position with subtotals', async ({ page }) => {
    // SquadTable now groups starters by position (Goals, Kicks, etc.) with subtotals
    await expect(page.getByText('Goals').first()).toBeVisible()
  })

  test('displays Ruiboys players', async ({ page }) => {
    await expect(page.getByText('Jordan Dawson')).toBeVisible()
  })

  test('displays status badges', async ({ page }) => {
    // drv_afl_status='played' is seeded for Round 1 player matches.
    await expect(page.getByText('Played').first()).toBeVisible()
  })

  test('shows Team Builder link in selected club column only', async ({ page }) => {
    // The selected club's h2 link navigates to /edit; the other club's h2 link does not.
    // The Team Builder icon was merged into the club-name link in this branch (no separate title attr).
    await expect(page.locator('h2 a[href$="/edit"]')).toHaveCount(1)
  })

  test('Team Builder link navigates to team builder', async ({ page }) => {
    await page.locator('h2 a[href$="/edit"]').click()
    await expect(page).toHaveURL(/\/ffl\/club-matches\/.*\/edit/)
  })

  test('shows breadcrumb with FFL, season and round', async ({ page }) => {
    await expect(page.locator('main').getByRole('link', { name: 'FFL 2026' })).toBeVisible()
    await expect(page.locator('main').getByRole('link', { name: 'Round 1' })).toBeVisible()
  })

  test('club name links to squad page', async ({ page }) => {
    await page.locator('main').getByRole('link', { name: 'Ruiboys' }).first().click()
    await page.waitForURL(/\/ffl\/club-seasons\//)
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Ruiboys')
  })
})
