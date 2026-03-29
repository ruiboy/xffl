import { test, expect } from '@playwright/test'

test.describe('Match view', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate: Home → first match
    await page.goto('/')
    await page.getByRole('link', { name: /Adelaide Crows.+v.+Brisbane Lions/ }).first().click()
  })

  test('displays match header with teams and venue', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Adelaide Crows')
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Brisbane Lions')
    await expect(page.getByText('Adelaide Oval')).toBeVisible()
  })

  test('displays home team player stats', async ({ page }) => {
    await expect(page.getByText('Jordan Dawson')).toBeVisible()
    await expect(page.getByText('Rory Laird')).toBeVisible()
    await expect(page.getByText('Ben Keays')).toBeVisible()
  })

  test('displays away team player stats', async ({ page }) => {
    await expect(page.getByText('Lachie Neale')).toBeVisible()
    await expect(page.getByText('Hugh McCluggage')).toBeVisible()
    await expect(page.getByText('Dayne Zorko')).toBeVisible()
  })

  test('displays stat column headers', async ({ page }) => {
    for (const label of ['K', 'HB', 'M', 'HO', 'T', 'G', 'B', 'D', 'SC']) {
      await expect(page.getByRole('columnheader', { name: label }).first()).toBeVisible()
    }
  })

  test('displays totals row', async ({ page }) => {
    await expect(page.getByText('Totals').first()).toBeVisible()
  })

  test('stats are read-only (no input fields)', async ({ page }) => {
    const dawsonRow = page.getByRole('row').filter({ hasText: 'Jordan Dawson' })
    await expect(dawsonRow.getByRole('spinbutton')).toHaveCount(0)
  })
})
