import { test, expect } from './fixtures'

test.describe('FFL Match', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate: FFL Home → Round 1 → match (match card is a div, not a link)
    await page.goto('/ffl')
    await page.locator('main nav').getByRole('link', { name: '1', exact: true }).click()
    await page.locator('.cursor-pointer').filter({ hasText: 'Ruiboys' }).filter({ hasText: 'The Howling Cows' }).click()
    await page.waitForURL(/\/ffl\/seasons\/.*\/matches\//)
  })

  test('displays match header with teams', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Ruiboys')
    await expect(page.getByRole('heading', { level: 1 })).toContainText('The Howling Cows')
  })

  test('displays venue', async ({ page }) => {
    await expect(page.getByText('MCG')).toBeVisible()
  })

  test('displays fantasy scores', async ({ page }) => {
    await expect(page.getByText('Score:')).toHaveCount(2)
  })

  test('displays squad table with player columns', async ({ page }) => {
    await expect(page.getByRole('columnheader', { name: 'Player' }).first()).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Status' }).first()).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Score' }).first()).toBeVisible()
  })

  test('squad table groups players by position with subtotals', async ({ page }) => {
    // SquadTable now groups starters by position (Goals, Kicks, etc.) with subtotals
    await expect(page.getByText('Goals').first()).toBeVisible()
  })

  test('displays Ruiboys players', async ({ page }) => {
    await expect(page.getByText('Jordan Dawson')).toBeVisible()
  })

  test('displays status badges', async ({ page }) => {
    // Status derives from the AFL match's statsImportStatus, not from the FFL
    // player_match.status column (commit 94033fe — "Gate team builder scores on
    // AFL match stats import status"). The seed leaves Round 1's AFL match at
    // stats_import_status='no_data', so badges read "Named" rather than
    // "Played" until the AFL stats are imported.
    await expect(page.getByText('Named').first()).toBeVisible()
  })

  test('displays total row', async ({ page }) => {
    await expect(page.getByText('Total').first()).toBeVisible()
  })

  test('shows Team Builder button in selected club column only', async ({ page }) => {
    // Selected club is The Howling Cows — button should appear once (in their column)
    await expect(page.getByTitle('Team Builder')).toHaveCount(1)
  })

  test('Team Builder button navigates to team builder', async ({ page }) => {
    await page.getByTitle('Team Builder').click()
    await expect(page).toHaveURL(/\/ffl\/.*\/team-builder/)
  })

  test('shows breadcrumb with FFL, season and round', async ({ page }) => {
    await expect(page.locator('main').getByRole('link', { name: 'FFL 2026' })).toBeVisible()
    await expect(page.locator('main').getByRole('link', { name: 'Round 1' })).toBeVisible()
  })

  test('club name links to squad page', async ({ page }) => {
    await page.locator('main').getByRole('link', { name: 'Ruiboys' }).first().click()
    await page.waitForURL(/\/ffl\/seasons\/.*\/clubs\/.*\/squad/)
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Ruiboys')
  })
})
