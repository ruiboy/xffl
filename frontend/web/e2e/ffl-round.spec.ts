import { test, expect } from '@playwright/test'

test.describe('FFL Round', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate: Home → Round 1 via round nav
    await page.goto('/')
    await page.getByRole('link', { name: 'Round 1' }).click()
  })

  test('displays round and season name', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Round 1')
    await expect(page.getByText('2024 Season')).toBeVisible()
  })

  test('displays match summaries', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Matches' })).toBeVisible()
    const matchLink = page.getByRole('link', { name: /Ruiboys.+v.+The Howling Cows/ })
    await expect(matchLink).toBeVisible()
  })

  test('displays top fantasy scorers', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Top Fantasy Scorers' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Player' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Score' })).toBeVisible()
  })

  test('displays round navigation', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Rounds' })).toBeVisible()
  })

  test('has Build Team button', async ({ page }) => {
    await expect(page.getByRole('link', { name: 'Build Team' })).toBeVisible()
  })
})
