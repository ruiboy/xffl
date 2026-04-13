import { test, expect } from '@playwright/test'
import { setupFflSession } from './helpers'

test.describe('FFL Home', () => {
  test.beforeEach(async ({ page }) => {
    await setupFflSession(page)
  })

  test('displays season name in heading', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 1 })).toContainText('FFL 2026')
  })

  test('displays ladder with clubs', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Ladder' })).toBeVisible()
    await expect(page.getByRole('cell', { name: 'Ruiboys' })).toBeVisible()
    await expect(page.getByRole('cell', { name: 'The Howling Cows' })).toBeVisible()
  })

  test('displays ladder with percentage column', async ({ page }) => {
    await expect(page.getByRole('columnheader', { name: '%' })).toBeVisible()
  })

  test('does not display matches section', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Matches' })).not.toBeVisible()
  })

  test('displays round selector with ladder icon and round circles', async ({ page }) => {
    const roundNav = page.locator('main nav')
    await expect(roundNav).toBeVisible()
    await expect(roundNav.getByTitle('Ladder')).toBeVisible()
    await expect(roundNav.getByRole('link', { name: '1', exact: true })).toBeVisible()
  })

  test('round circle navigates to round page', async ({ page }) => {
    await page.locator('main nav').getByRole('link', { name: '1', exact: true }).click()
    await page.waitForURL(/\/ffl\/seasons\/.*\/rounds\//)
    await page.waitForLoadState('networkidle')
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Round 1')
  })

  test('navbar has FFL and AFL links', async ({ page }) => {
    const topNav = page.getByRole('navigation').first()
    await expect(topNav.getByRole('link', { name: 'FFL', exact: true }).filter({ hasText: 'FFL' })).toBeVisible()
    await expect(topNav.getByRole('link', { name: 'AFL', exact: true })).toBeVisible()
  })

  test('navbar has settings cog', async ({ page }) => {
    await expect(page.getByTitle('Settings')).toBeVisible()
  })

  test('navbar does not have a Team Builder link', async ({ page }) => {
    await expect(page.getByRole('navigation').first().getByRole('link', { name: 'Team Builder' })).not.toBeVisible()
  })

  test('round 3 has the open live-round ring indicator', async ({ page }) => {
    const round3 = page.locator('main nav').getByRole('link', { name: '3', exact: true })
    await expect(round3).toHaveClass(/ring-active/)
  })
})
