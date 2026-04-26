import { test, expect } from './fixtures'
import { setupFflSession } from './helpers'

test.describe('FFL Squad', () => {
  test.beforeEach(async ({ page }) => {
    await setupFflSession(page)
    await page.getByRole('link', { name: 'Squad' }).click()
    await page.waitForURL(/\/ffl\/seasons\/.*\/clubs\/.*\/squad/)
    await page.waitForLoadState('networkidle')
  })

  test('displays club name as heading', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 1 })).toContainText('The Howling Cows')
  })

  test('shows breadcrumb with FFL, season and club name', async ({ page }) => {
    await expect(page.locator('main').getByRole('link', { name: 'FFL 2026' })).toBeVisible()
    await expect(page.locator('main').getByText('The Howling Cows', { exact: true }).first()).toBeVisible()
  })

  test('displays player list', async ({ page }) => {
    await expect(page.getByRole('columnheader', { name: 'Player' })).toBeVisible()
    await expect(page.getByRole('cell', { name: 'Henry Smith' })).toBeVisible()
  })

  test('shows Manage button initially', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Manage' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Done' })).not.toBeVisible()
  })

  test('clicking Manage shows search panel and Done button', async ({ page }) => {
    await page.getByRole('button', { name: 'Manage' }).click()
    await expect(page.getByRole('button', { name: 'Done' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Add Player' })).toBeVisible()
    await expect(page.getByPlaceholder('Search AFL players by name...')).toBeVisible()
  })

  test('clicking Manage shows Remove buttons on player rows', async ({ page }) => {
    await page.getByRole('button', { name: 'Manage' }).click()
    await expect(page.getByRole('button', { name: 'Remove' }).first()).toBeVisible()
  })

  test('shows Saved message after removing a player', async ({ page }) => {
    await page.getByRole('button', { name: 'Manage' }).click()
    await page.getByRole('button', { name: 'Remove' }).first().click()
    await expect(page.getByText('Saved')).toBeVisible()
  })

  test('clicking Done exits manage mode', async ({ page }) => {
    await page.getByRole('button', { name: 'Manage' }).click()
    await page.getByRole('button', { name: 'Done' }).click()
    await expect(page.getByRole('button', { name: 'Manage' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Add Player' })).not.toBeVisible()
  })

  test('player search returns results', async ({ page }) => {
    await page.getByRole('button', { name: 'Manage' }).click()
    await page.getByPlaceholder('Search AFL players by name...').fill('Jordan')
    await expect(page.getByText('Jordan Dawson')).toBeVisible()
  })

  test('switching to a different club hides Manage button', async ({ page }) => {
    await page.getByRole('button', { name: 'Manage' }).click()
    await expect(page.getByRole('button', { name: 'Done' })).toBeVisible()

    // Open club selector and pick a different club (Ruiboys)
    await page.getByRole('button', { name: /The Howling Cows|Ruiboys/ }).click()
    await page.getByRole('button', { name: /Ruiboys/ }).click()

    // Page still shows The Howling Cows squad, but Manage is hidden (not the selected club)
    await expect(page.getByRole('button', { name: 'Manage' })).not.toBeVisible()
    await expect(page.getByRole('button', { name: 'Done' })).not.toBeVisible()
  })

  test('Manage button hidden when viewing another club squad', async ({ page }) => {
    // Navigate to Ruiboys squad while The Howling Cows is selected
    await page.goto('/ffl')
    await page.locator('main').getByRole('link', { name: 'Ruiboys' }).first().click()
    await page.waitForURL(/\/ffl\/seasons\/.*\/clubs\/.*\/squad/)
    await page.waitForLoadState('networkidle')
    await expect(page.getByRole('button', { name: 'Manage' })).not.toBeVisible()
  })
})
