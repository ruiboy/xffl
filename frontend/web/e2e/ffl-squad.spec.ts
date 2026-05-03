import { test, expect } from './fixtures'
import { setupFflSession } from './helpers'

test.describe('FFL Squad', () => {
  test.beforeEach(async ({ page }) => {
    await setupFflSession(page)
    await page.getByRole('link', { name: 'Squad' }).click()
    await page.waitForURL(/\/ffl\/seasons\/.*\/clubs\/.*\/squad/)
    await page.waitForLoadState('networkidle')
  })

  // ── Read-only / layout ─────────────────────────────────────────────────────

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

  test('shows Club column header', async ({ page }) => {
    await expect(page.getByRole('columnheader', { name: 'Club' })).toBeVisible()
  })

  test('shows AFL club name for each player', async ({ page }) => {
    const row = page.getByRole('row', { name: /Henry Smith/ })
    await expect(row.getByText('Brisbane Lions')).toBeVisible()
  })

  // ── Manage toolbar ─────────────────────────────────────────────────────────

  test('shows Manage button initially', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Manage' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Done' })).not.toBeVisible()
  })

  test('clicking Manage reveals + Add Player and Done', async ({ page }) => {
    await page.getByRole('button', { name: 'Manage' }).click()
    await expect(page.getByRole('button', { name: 'Done' })).toBeVisible()
    await expect(page.getByRole('button', { name: '+ Add Player' })).toBeVisible()
  })

  test('Manage shows bin (Remove) button on each player row', async ({ page }) => {
    await page.getByRole('button', { name: 'Manage' }).click()
    await expect(page.getByRole('button', { name: 'Remove' }).first()).toBeVisible()
  })

  test('clicking Done exits manage mode and hides + Add Player', async ({ page }) => {
    await page.getByRole('button', { name: 'Manage' }).click()
    await page.getByRole('button', { name: 'Done' }).click()
    await expect(page.getByRole('button', { name: 'Manage' })).toBeVisible()
    await expect(page.getByRole('button', { name: '+ Add Player' })).not.toBeVisible()
  })

  test('switching to a different club hides Manage button', async ({ page }) => {
    await page.getByRole('button', { name: 'Manage' }).click()
    await expect(page.getByRole('button', { name: 'Done' })).toBeVisible()

    // Open club selector and pick a different club (Ruiboys)
    await page.getByRole('button', { name: /The Howling Cows|Ruiboys/ }).click()
    await page.getByRole('button', { name: /Ruiboys/ }).click()

    await expect(page.getByRole('button', { name: 'Manage' })).not.toBeVisible()
    await expect(page.getByRole('button', { name: 'Done' })).not.toBeVisible()
  })

  test('Manage button hidden when viewing another club squad', async ({ page }) => {
    await page.goto('/ffl')
    await page.locator('main').getByRole('link', { name: 'Ruiboys' }).first().click()
    await page.waitForURL(/\/ffl\/seasons\/.*\/clubs\/.*\/squad/)
    await page.waitForLoadState('networkidle')
    await expect(page.getByRole('button', { name: 'Manage' })).not.toBeVisible()
  })

  // ── Add player two-step flow (search → confirm with from-round) ────────────

  test.describe('add player flow', () => {
    test.beforeEach(async ({ page }) => {
      await page.getByRole('button', { name: 'Manage' }).click()
      await page.getByRole('button', { name: '+ Add Player' }).click()
    })

    test('search dialog opens with placeholder and Cancel', async ({ page }) => {
      await expect(page.getByPlaceholder('Search by name...')).toBeVisible()
      await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible()
    })

    test('typing a query shows AFL player results with their AFL club', async ({ page }) => {
      await page.getByPlaceholder('Search by name...').fill('Jordan')
      await expect(page.getByText('Jordan Dawson')).toBeVisible()
      // The result row shows the AFL club name underneath the player name
      const searchDialog = page.getByPlaceholder('Search by name...').locator('xpath=ancestor::div[contains(@class, "rounded-xl")]')
      await expect(searchDialog.getByText('Adelaide Crows')).toBeVisible()
    })

    test('selecting Add opens the confirm dialog with a From round selector', async ({ page }) => {
      await page.getByPlaceholder('Search by name...').fill('Jordan')
      await expect(page.getByText('Jordan Dawson')).toBeVisible()
      await page.getByRole('button', { name: 'Add', exact: true }).click()
      // Search input is gone — confirm dialog is now showing
      await expect(page.getByPlaceholder('Search by name...')).not.toBeVisible()
      await expect(page.getByText('From round')).toBeVisible()
      const select = page.locator('select').last()
      await expect(select).toBeVisible()
      // A round is preselected (defaultRoundId picks live round or last)
      expect(await select.inputValue()).not.toBe('')
    })

    test('Cancel closes the search dialog without adding', async ({ page }) => {
      await page.getByRole('button', { name: 'Cancel' }).click()
      await expect(page.getByPlaceholder('Search by name...')).not.toBeVisible()
      await expect(page.getByText('Saved')).not.toBeVisible()
    })

    test('Back from the confirm dialog returns to the search dialog', async ({ page }) => {
      await page.getByPlaceholder('Search by name...').fill('Jordan')
      await expect(page.getByText('Jordan Dawson')).toBeVisible()
      await page.getByRole('button', { name: 'Add', exact: true }).click()
      await page.getByRole('button', { name: 'Back' }).click()
      await expect(page.getByPlaceholder('Search by name...')).toBeVisible()
    })

    test('confirming Add adds the player and shows Saved', async ({ page }) => {
      await page.getByPlaceholder('Search by name...').fill('Jordan')
      await expect(page.getByText('Jordan Dawson')).toBeVisible()
      await page.getByRole('button', { name: 'Add', exact: true }).click()
      // After search → confirm, the only Add button on the page is the confirm one.
      await page.getByRole('button', { name: 'Add', exact: true }).click()
      await expect(page.getByText('Saved')).toBeVisible()
      await expect(page.getByRole('cell', { name: 'Jordan Dawson' })).toBeVisible()
    })
  })

  // ── Remove player + Traded section ─────────────────────────────────────────

  test.describe('remove player flow', () => {
    test.beforeEach(async ({ page }) => {
      await page.getByRole('button', { name: 'Manage' }).click()
    })

    test('clicking the bin opens the Remove Player modal with a round selector', async ({ page }) => {
      await page.getByRole('button', { name: 'Remove' }).first().click()
      await expect(page.getByRole('heading', { name: 'Remove Player' })).toBeVisible()
      await expect(page.getByText('Round', { exact: true })).toBeVisible()
      const dialog = page.getByRole('heading', { name: 'Remove Player' }).locator('..')
      await expect(dialog.getByRole('button', { name: 'Remove' })).toBeVisible()
      await expect(dialog.getByRole('button', { name: 'Cancel' })).toBeVisible()
    })

    test('Cancel closes the modal without removing', async ({ page }) => {
      await page.getByRole('button', { name: 'Remove' }).first().click()
      await page.getByRole('button', { name: 'Cancel' }).click()
      await expect(page.getByRole('heading', { name: 'Remove Player' })).not.toBeVisible()
      await expect(page.getByText('Saved')).not.toBeVisible()
    })

    test('confirming Remove moves the player to the Traded section', async ({ page }) => {
      // Pick Henry Smith specifically so we can assert his row moved.
      const henrysRow = page.getByRole('row', { name: /Henry Smith/ })
      await henrysRow.getByRole('button', { name: 'Remove' }).click()
      const dialog = page.getByRole('heading', { name: 'Remove Player' }).locator('..')
      await dialog.getByRole('button', { name: 'Remove' }).click()
      await expect(page.getByText('Saved')).toBeVisible()

      const tradedToggle = page.getByRole('button', { name: /Traded \(\d+\)/ })
      await expect(tradedToggle).toBeVisible()
      await tradedToggle.click()
      // Henry Smith now appears in the muted Traded section.
      await expect(page.getByRole('row', { name: /Henry Smith/ })).toBeVisible()
    })
  })

  // ── Inline notes ───────────────────────────────────────────────────────────

  test.describe('inline notes', () => {
    test.beforeEach(async ({ page }) => {
      await page.getByRole('button', { name: 'Manage' }).click()
    })

    test('clicking a player row in manage mode expands a notes textarea', async ({ page }) => {
      await page.getByRole('row', { name: /Henry Smith/ }).click()
      await expect(page.getByPlaceholder('Add notes...')).toBeVisible()
    })

    test('Save button only appears once notes have been edited', async ({ page }) => {
      await page.getByRole('row', { name: /Henry Smith/ }).click()
      const textarea = page.getByPlaceholder('Add notes...')
      await expect(textarea).toBeVisible()
      await expect(page.getByRole('button', { name: 'Save' })).not.toBeVisible()
      await textarea.fill('Strong start, watch backline')
      await expect(page.getByRole('button', { name: 'Save' })).toBeVisible()
    })

    test('saving notes flashes Saved and persists across reload', async ({ page }) => {
      await page.getByRole('row', { name: /Henry Smith/ }).click()
      await page.getByPlaceholder('Add notes...').fill('Strong start')
      await page.getByRole('button', { name: 'Save' }).click()
      await expect(page.getByText('Saved')).toBeVisible()

      await page.reload()
      await page.waitForLoadState('networkidle')
      await page.getByRole('button', { name: 'Manage' }).click()
      await page.getByRole('row', { name: /Henry Smith/ }).click()
      await expect(page.getByPlaceholder('Add notes...')).toHaveValue('Strong start')
    })
  })
})
