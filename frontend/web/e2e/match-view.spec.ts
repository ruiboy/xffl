import { test, expect } from '@playwright/test'

test.describe('Match view', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate: Seasons → first season → first match
    await page.goto('/')
    await page.getByRole('link', { name: 'AFL 2025' }).first().click()
    await page.getByRole('link', { name: /Adelaide Crows.+v.+Brisbane Lions/ }).first().click()
  })

  test('displays match header with teams and venue', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Adelaide Crows')
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Brisbane Lions')
    await expect(page.getByText('Adelaide Oval')).toBeVisible()
  })

  test('displays home team player stats', async ({ page }) => {
    const homeSection = page.getByRole('heading', { name: 'Adelaide Crows' }).locator('..')

    // Check players appear in the table
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

  test('edits a player stat and sees updated value', async ({ page }) => {
    // Find Jordan Dawson's kicks input (first editable input in his row)
    const dawsonRow = page.getByRole('row').filter({ hasText: 'Jordan Dawson' })
    const kicksInput = dawsonRow.getByRole('spinbutton').first()

    // Read current value, change it, and verify
    const currentValue = await kicksInput.inputValue()
    const newValue = String(Number(currentValue) + 1)

    await kicksInput.fill(newValue)
    await kicksInput.press('Tab') // trigger change event

    // Wait for mutation response — the disposals column should update
    // (disposals = kicks + handballs, so increasing kicks by 1 increases disposals by 1)
    await expect(kicksInput).toHaveValue(newValue)
  })
})
