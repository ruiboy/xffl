import { test, expect } from '@playwright/test'

test.describe('Admin match view', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate: AFL Home → match, then switch to admin URL
    await page.goto('/afl')
    const matchLink = page.getByRole('link', { name: /Adelaide Crows.+v.+Brisbane Lions/ }).first()
    const href = await matchLink.getAttribute('href')
    // Replace /afl/ with /admin/afl/ to get admin route
    const adminHref = href!.replace('/afl/', '/admin/afl/')
    await page.goto(adminHref)
  })

  test('displays admin label', async ({ page }) => {
    await expect(page.getByText('Admin')).toBeVisible()
  })

  test('displays match header with teams', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Adelaide Crows')
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Brisbane Lions')
  })

  test('stats are editable (has input fields)', async ({ page }) => {
    const dawsonRow = page.getByRole('row').filter({ hasText: 'Jordan Dawson' })
    await expect(dawsonRow.getByRole('spinbutton').first()).toBeVisible()
  })

  test('edits a player stat', async ({ page }) => {
    const dawsonRow = page.getByRole('row').filter({ hasText: 'Jordan Dawson' })
    const kicksInput = dawsonRow.getByRole('spinbutton').first()

    const currentValue = await kicksInput.inputValue()
    const newValue = String(Number(currentValue) + 1)

    await kicksInput.fill(newValue)
    await kicksInput.press('Tab')

    await expect(kicksInput).toHaveValue(newValue)
  })
})
