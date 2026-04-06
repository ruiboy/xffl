import { test, expect } from '@playwright/test'

test.describe('FFL Match', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate: FFL Home → Round 1 → match
    await page.goto('/ffl')
    await page.locator('main nav').getByRole('link', { name: '1', exact: true }).click()
    await page.getByRole('link', { name: /Ruiboys.+v.+The Howling Cows/ }).click()
  })

  test('displays match header with teams', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Ruiboys')
    await expect(page.getByRole('heading', { level: 1 })).toContainText('The Howling Cows')
  })

  test('displays venue', async ({ page }) => {
    await expect(page.getByText('MCG')).toBeVisible()
  })

  test('displays fantasy scores', async ({ page }) => {
    await expect(page.getByText('Fantasy score:')).toHaveCount(2)
  })

  test('displays squad table with player columns', async ({ page }) => {
    await expect(page.getByRole('columnheader', { name: 'Player' }).first()).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Position' }).first()).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Status' }).first()).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Score' }).first()).toBeVisible()
  })

  test('displays Ruiboys players', async ({ page }) => {
    await expect(page.getByText('Jordan Dawson')).toBeVisible()
  })

  test('displays status badges', async ({ page }) => {
    await expect(page.getByText('Played').first()).toBeVisible()
  })

  test('displays total row', async ({ page }) => {
    await expect(page.getByText('Total').first()).toBeVisible()
  })

  test('shows Build Team link in selected club column only', async ({ page }) => {
    // Selected club is The Howling Cows — link should appear once (in their column)
    await expect(page.getByRole('link', { name: 'Build Team →' })).toHaveCount(1)
  })

  test('Build Team link navigates to team builder', async ({ page }) => {
    await page.getByRole('link', { name: 'Build Team →' }).click()
    await expect(page).toHaveURL(/\/ffl\/.*\/team-builder/)
  })
})
