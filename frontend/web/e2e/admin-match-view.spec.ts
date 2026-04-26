import { test, expect } from './fixtures'
import { setupAflSession } from './helpers'

test.describe('Admin match view', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate: AFL Home → Round 1 → get match href → navigate to admin URL
    await setupAflSession(page)
    await page.locator('main nav').last().getByRole('link', { name: '1', exact: true }).click()
    const matchLink = page.getByRole('link', { name: /Adelaide Crows.+v.+Brisbane Lions/ }).first()
    const href = await matchLink.getAttribute('href')
    // Replace /afl/ with /admin/afl/ to get admin route
    const adminHref = href!.replace('/afl/', '/admin/afl/')
    await page.goto(adminHref)
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible({ timeout: 15000 })
  })

  test('displays match header with teams', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Adelaide Crows')
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Brisbane Lions')
  })

  test('shows breadcrumb with AFL, season and round', async ({ page }) => {
    await expect(page.locator('main').getByRole('link', { name: 'AFL 2026' })).toBeVisible()
    await expect(page.locator('main').getByRole('link', { name: 'Round 1' })).toBeVisible()
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
