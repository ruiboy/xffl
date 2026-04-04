import { test, expect } from '@playwright/test'

test.describe('FFL Squad', () => {
  test.beforeEach(async ({ page }) => {
    // Load home first so FFL state (seasonId) is populated, then navigate via nav
    await page.goto('/')
    await page.getByRole('link', { name: 'Squad' }).click()
  })

  test('displays club name as heading', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Ruiboys')
  })

  test('displays season name in subheading', async ({ page }) => {
    await expect(page.getByText(/2024 Season Squad/)).toBeVisible()
  })

  test('displays player list', async ({ page }) => {
    await expect(page.getByRole('columnheader', { name: 'Player' })).toBeVisible()
    await expect(page.getByRole('cell', { name: 'Marcus Bontempelli' })).toBeVisible()
  })

  test('shows Manage button initially', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Manage' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Done' })).not.toBeVisible()
  })

  test('clicking Manage shows search panel and Done button', async ({ page }) => {
    await page.getByRole('button', { name: 'Manage' }).click()
    await expect(page.getByRole('button', { name: 'Done' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Add Player' })).toBeVisible()
    await expect(page.getByPlaceholder('Search AFL players by name…')).toBeVisible()
  })

  test('clicking Manage shows Remove buttons on player rows', async ({ page }) => {
    await page.getByRole('button', { name: 'Manage' }).click()
    await expect(page.getByRole('button', { name: 'Remove' }).first()).toBeVisible()
  })

  test('clicking Done exits manage mode', async ({ page }) => {
    await page.getByRole('button', { name: 'Manage' }).click()
    await page.getByRole('button', { name: 'Done' }).click()
    await expect(page.getByRole('button', { name: 'Manage' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Add Player' })).not.toBeVisible()
  })

  test('player search returns results', async ({ page }) => {
    await page.getByRole('button', { name: 'Manage' }).click()
    await page.getByPlaceholder('Search AFL players by name…').fill('Patrick')
    await expect(page.getByText('Patrick Cripps')).toBeVisible()
  })
})
