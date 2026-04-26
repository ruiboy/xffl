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

  test('shows FFL breadcrumb on home page', async ({ page }) => {
    await expect(page.locator('main').getByText('FFL', { exact: true })).toBeVisible()
  })

  test('club name in ladder links to squad page', async ({ page }) => {
    await page.getByRole('link', { name: 'Ruiboys' }).first().click()
    await page.waitForURL(/\/ffl\/seasons\/.*\/clubs\/.*\/squad/)
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Ruiboys')
  })

  test('displays ladder with percentage column', async ({ page }) => {
    await expect(page.getByRole('columnheader', { name: '%' })).toBeVisible()
  })

  test('does not display matches section', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Matches' })).not.toBeVisible()
  })

  test('displays round selector with round circles', async ({ page }) => {
    const roundNav = page.locator('main nav').last()
    await expect(roundNav).toBeVisible()
    await expect(roundNav.getByRole('link', { name: '1', exact: true })).toBeVisible()
  })

  test('round circle navigates to round page', async ({ page }) => {
    await page.locator('main nav').last().getByRole('link', { name: '1', exact: true }).click()
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

  test('global nav shows Team Builder and Squad links', async ({ page }) => {
    const topNav = page.getByRole('navigation').first()
    await expect(topNav.getByRole('link', { name: 'Team Builder' })).toBeVisible()
    await expect(topNav.getByRole('link', { name: 'Squad' })).toBeVisible()
  })

  test('global Team Builder link navigates to team builder for live round', async ({ page }) => {
    await page.getByRole('navigation').first().getByRole('link', { name: 'Team Builder' }).click()
    await expect(page).toHaveURL(/\/ffl\/.*\/team-builder/)
  })

  test('round 3 has the open live-round ring indicator', async ({ page }) => {
    const round3 = page.locator('main nav').last().getByRole('link', { name: '3', exact: true })
    await expect(round3).toHaveClass(/ring-active/)
  })
})
