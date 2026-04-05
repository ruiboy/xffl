import { test, expect } from '@playwright/test'

test.describe('FFL Home', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
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
})
