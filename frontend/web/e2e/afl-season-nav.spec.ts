import { test, expect } from './fixtures'
import { setupAflSession } from './helpers'

test.describe('AFL season navigation', () => {
  test.beforeEach(async ({ page }) => {
    await setupAflSession(page)
  })

  test('season selector shows the current season', async ({ page }) => {
    const selector = page.locator('div.relative', { has: page.getByTitle('Season') })
    await expect(selector).toBeVisible()
    await expect(selector).toContainText('AFL 2026')
  })

  test('selecting a season navigates to its ladder', async ({ page }) => {
    const selector = page.locator('div.relative', { has: page.getByTitle('Season') })
    await selector.getByTitle('Season').click()
    // second button in the selector is the season option (first is the trigger)
    await selector.getByRole('button').nth(1).click()

    await expect(page).toHaveURL(/\/afl\/seasons\/\d+$/)
    await expect(page.getByRole('heading', { level: 1 })).toContainText('AFL 2026')
    await expect(page.getByRole('heading', { name: 'Ladder' })).toBeVisible()
    await expect(page.getByRole('cell', { name: 'Adelaide Crows' })).toBeVisible()
  })
})
