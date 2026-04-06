import { test, expect } from '@playwright/test'
import { setupFflSession } from './helpers'

test.describe('FFL Squad', () => {
  test.beforeEach(async ({ page }) => {
    await setupFflSession(page)
    await page.getByRole('link', { name: 'Squad' }).click()
    await page.waitForURL(/\/ffl\/seasons\/.*\/squad/)
    await page.waitForLoadState('networkidle')
  })

  test('displays club name as heading', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 1 })).toContainText('The Howling Cows')
  })

  test('displays season name in subheading', async ({ page }) => {
    await expect(page.getByText(/FFL 2026 Squad/)).toBeVisible()
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

  test('switching club exits manage mode', async ({ page }) => {
    await page.getByRole('button', { name: 'Manage' }).click()
    await expect(page.getByRole('button', { name: 'Done' })).toBeVisible()

    // Open club selector and pick a different club (Ruiboys)
    await page.getByRole('button', { name: /The Howling Cows|Ruiboys/ }).click()
    await page.getByRole('button', { name: /Ruiboys/ }).click()

    await expect(page.getByRole('button', { name: 'Manage' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Done' })).not.toBeVisible()
  })
})
