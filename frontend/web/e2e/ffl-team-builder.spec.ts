import { test, expect } from '@playwright/test'

test.describe('FFL Team Builder', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate via navbar Team Builder link (requires home to load first so state is set)
    await page.goto('/')
    await page.getByRole('link', { name: 'Team Builder' }).click()
  })

  test('displays club name as heading', async ({ page }) => {
    // Heading should be the selected club name, not "Team Builder"
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Ruiboys')
  })

  test('loads in read-only mode with Manage button', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Manage' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Done' })).not.toBeVisible()
  })

  test('read-only mode shows lineup without edit controls', async ({ page }) => {
    // Starters visible from seed data
    await expect(page.getByText('Christian Petracca')).toBeVisible()
    // No position action buttons or Remove buttons visible
    await expect(page.getByRole('button', { name: 'Remove' })).not.toBeVisible()
  })

  test('read-only mode does not show squad panel', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /Squad/ })).not.toBeVisible()
  })

  test('clicking Manage reveals edit controls and squad panel', async ({ page }) => {
    await page.getByRole('button', { name: 'Manage' }).click()
    await expect(page.getByRole('button', { name: 'Done' })).toBeVisible()
    await expect(page.getByRole('heading', { name: /Squad/ })).toBeVisible()
  })

  test('clicking Manage shows position groups with action buttons', async ({ page }) => {
    await page.getByRole('button', { name: 'Manage' }).click()
    for (const position of ['Goals', 'Kicks', 'Handballs', 'Marks', 'Tackles', 'Hitouts', 'Star']) {
      await expect(page.getByRole('heading', { name: position })).toBeVisible()
    }
    await expect(page.getByRole('heading', { name: /Bench/ })).toBeVisible()
  })

  test('clicking Done saves and returns to read-only mode', async ({ page }) => {
    await page.getByRole('button', { name: 'Manage' }).click()
    await expect(page.getByRole('button', { name: 'Done' })).toBeVisible()
    await page.getByRole('button', { name: 'Done' }).click()
    // Returns to Manage state (Done triggers save then exits)
    await expect(page.getByRole('button', { name: 'Manage' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Done' })).not.toBeVisible()
  })
})
