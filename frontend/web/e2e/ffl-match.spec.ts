import { test, expect } from '@playwright/test'

test.describe('FFL Match', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate: Home → match
    await page.goto('/')
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

  test('displays roster table with player columns', async ({ page }) => {
    await expect(page.getByRole('columnheader', { name: 'Player' }).first()).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Position' }).first()).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Status' }).first()).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Score' }).first()).toBeVisible()
  })

  test('displays Ruiboys players', async ({ page }) => {
    await expect(page.getByText('Marcus Bontempelli')).toBeVisible()
  })

  test('displays status badges', async ({ page }) => {
    await expect(page.getByText('Played').first()).toBeVisible()
  })

  test('displays total row', async ({ page }) => {
    await expect(page.getByText('Total').first()).toBeVisible()
  })
})
