import { test, expect } from '@playwright/test'

test.describe('FFL Round', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
    await page.locator('main nav').last().getByRole('link', { name: '1', exact: true }).click()
    // Wait for round data to load before running assertions
    await expect(page.getByRole('heading', { name: 'Matches' })).toBeVisible({ timeout: 15000 })
  })

  test('displays round name in heading', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Round 1')
  })

  test('displays season name in breadcrumb', async ({ page }) => {
    await expect(page.locator('main').getByRole('link', { name: 'FFL 2026' })).toBeVisible()
  })

  test('displays round selector with round circles', async ({ page }) => {
    const roundNav = page.locator('main nav').last()
    await expect(roundNav).toBeVisible()
    await expect(roundNav.getByRole('link', { name: '1', exact: true })).toBeVisible()
  })

  test('FFL breadcrumb link navigates back to home', async ({ page }) => {
    await page.locator('main').getByRole('link', { name: 'FFL 2026' }).click()
    await expect(page).toHaveURL('/ffl')
  })

  test('displays match summaries', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Matches' })).toBeVisible()
    const matchCard = page.locator('.cursor-pointer').filter({ hasText: 'Ruiboys' }).filter({ hasText: 'The Howling Cows' })
    await expect(matchCard).toBeVisible()
  })

  test('displays top scorers grouped by position', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Top Scorers' })).toBeVisible()
    // Grid layout by position — no table column headers
    await expect(page.getByRole('columnheader', { name: 'Player' })).not.toBeAttached()
    // Position group labels appear instead
    await expect(page.getByText('Kicks').first()).toBeVisible()
  })

  test('shows Team Builder button for selected club match', async ({ page }) => {
    await expect(page.locator('main').getByTitle('Team Builder')).toBeVisible()
  })

  test('Team Builder button navigates to team builder', async ({ page }) => {
    await page.locator('main').getByTitle('Team Builder').click()
    await expect(page).toHaveURL(/\/ffl\/.*\/team-builder/)
  })

  test('rounds display in numeric order', async ({ page }) => {
    const roundLinks = await page.locator('main nav').last().getByRole('link').allTextContents()
    const roundNumbers = roundLinks.map(t => parseInt(t.trim())).filter(n => !isNaN(n))
    expect(roundNumbers).toEqual([...roundNumbers].sort((a, b) => a - b))
  })
})
