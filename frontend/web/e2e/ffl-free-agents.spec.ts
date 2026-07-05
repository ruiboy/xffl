import { test, expect } from './fixtures'
import { setupFflSession } from './helpers'

test.describe('Free Agents', () => {
  test.beforeEach(async ({ page }) => {
    await setupFflSession(page)
    await page.goto('/ffl/free-agents')
    await page.waitForLoadState('networkidle')
  })

  test('shows Free Agents heading', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Free Agents')
  })

  test('shows all 7 stat tabs', async ({ page }) => {
    for (const label of ['Kicks', 'Handballs', 'Marks', 'Tackles', 'Hitouts', 'Goals', 'Star']) {
      await expect(page.getByRole('button', { name: label, exact: true })).toBeVisible()
    }
  })

  test('Kicks tab is active by default', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Kicks', exact: true })).toHaveClass(/border-active/)
  })

  test('shows player table with Player, Club and Gms columns', async ({ page }) => {
    await expect(page.getByRole('columnheader', { name: 'Player' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Club' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Gms' })).toBeVisible()
  })

  test('switching to Goals tab makes Goals tab active', async ({ page }) => {
    await page.getByRole('button', { name: 'Goals', exact: true }).click()
    await expect(page.getByRole('button', { name: 'Goals', exact: true })).toHaveClass(/border-active/)
    await expect(page.getByRole('button', { name: 'Kicks', exact: true })).not.toHaveClass(/border-active/)
  })
})
