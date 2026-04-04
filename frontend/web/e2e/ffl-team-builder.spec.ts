import { test, expect } from '@playwright/test'

test.describe('FFL Team Builder', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate: Home → Build Team
    await page.goto('/')
    await page.getByRole('link', { name: 'Build Team' }).click()
  })

  test('displays Team Builder heading', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Team Builder')
  })

  test('displays club selector', async ({ page }) => {
    await expect(page.getByRole('combobox')).toBeVisible()
  })

  test('displays position groups', async ({ page }) => {
    for (const position of ['Goals', 'Kicks', 'Handballs', 'Marks', 'Tackles', 'Hitouts', 'Star']) {
      await expect(page.getByRole('heading', { name: position })).toBeVisible()
    }
  })

  test('displays bench section', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /Bench/ })).toBeVisible()
  })

  test('displays squad panel with players', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /Squad/ })).toBeVisible()
    // Ruiboys has 30 players, some already assigned in seed data
    await expect(page.getByText('Marcus Bontempelli')).toBeVisible()
  })

  test('displays Save Lineup button', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Save Lineup' })).toBeVisible()
  })

  test('loads existing lineup from seed data', async ({ page }) => {
    // Seed data has 7 starters + 2 bench for Ruiboys
    // Starters should appear in position slots (not in squad panel as available)
    await expect(page.getByText('Christian Petracca')).toBeVisible()
  })
})
