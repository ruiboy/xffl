import { test, expect } from '@playwright/test'

test.describe('Home view', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  test('displays season and round name', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 1 })).toContainText('AFL 2025')
    await expect(page.getByRole('paragraph')).toContainText('Round 13')
  })

  test('displays ladder', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Ladder' })).toBeVisible()
    await expect(page.getByRole('cell', { name: 'Adelaide Crows' })).toBeVisible()
  })

  test('displays match summary with teams', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Matches' })).toBeVisible()
    const matchLink = page.getByRole('link', { name: /Adelaide Crows.+v.+Brisbane Lions/ })
    await expect(matchLink).toBeVisible()
  })

  test('match summary links to match page', async ({ page }) => {
    const matchLink = page.getByRole('link', { name: /Adelaide Crows.+v.+Brisbane Lions/ })
    await matchLink.click()
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Adelaide Crows')
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Brisbane Lions')
  })

  test('displays round navigation', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Rounds' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Round 13' })).toBeVisible()
  })

  test('navbar has Home link', async ({ page }) => {
    await expect(page.getByRole('navigation').getByRole('link', { name: 'Home' })).toBeVisible()
  })
})
