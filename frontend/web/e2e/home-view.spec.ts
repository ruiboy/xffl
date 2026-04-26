import { test, expect } from '@playwright/test'
import { setupAflSession } from './helpers'

test.describe('AFL Home view', () => {
  test.beforeEach(async ({ page }) => {
    await setupAflSession(page)
  })

  test('displays season name in heading', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 1 })).toContainText('AFL 2026')
  })

  test('displays ladder', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Ladder' })).toBeVisible()
    await expect(page.getByRole('cell', { name: 'Adelaide Crows' })).toBeVisible()
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

  test('round 3 has the open live-round ring indicator', async ({ page }) => {
    const round3 = page.locator('main nav').last().getByRole('link', { name: '3', exact: true })
    await expect(round3).toHaveClass(/ring-active/)
  })
})
