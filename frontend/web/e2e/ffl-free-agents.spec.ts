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

  test('shows all 7 stat column headers', async ({ page }) => {
    for (const label of ['Kicks', 'Handballs', 'Marks', 'Tackles', 'Hitouts', 'Goals']) {
      await expect(page.getByRole('button', { name: label, exact: true })).toBeVisible()
    }
    await expect(page.getByRole('button', { name: /^Star/ })).toBeVisible()
  })

  test('Kicks header is active by default', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Kicks', exact: true })).toHaveClass(/text-sky-400/)
  })

  test('rank basis toggle switches between last N and season', async ({ page }) => {
    await expect(page.getByRole('button', { name: /^Last \d+$/ })).toHaveClass(/bg-control/)
    await page.getByRole('button', { name: 'Season', exact: true }).click()
    await expect(page.getByRole('button', { name: 'Season', exact: true })).toHaveClass(/bg-control/)
  })

  test('shows player table with Player, Club and Gms columns', async ({ page }) => {
    await expect(page.getByRole('columnheader', { name: 'Player' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Club' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Gms' })).toBeVisible()
  })

  test('clicking the Goals header re-ranks by Goals', async ({ page }) => {
    await page.getByRole('button', { name: 'Goals', exact: true }).click()
    await expect(page.getByRole('button', { name: 'Goals', exact: true })).toHaveClass(/text-sky-400/)
    await expect(page.getByRole('button', { name: 'Kicks', exact: true })).not.toHaveClass(/text-sky-400/)
  })
})
