import { test, expect } from '@playwright/test'

// Helpers
async function goToTeamBuilder(page: import('@playwright/test').Page) {
  await page.goto('/')
  await page.getByRole('link', { name: 'Team Builder' }).click()
  await page.waitForURL(/\/ffl\/.*\/team-builder/)
}

function positionSection(page: import('@playwright/test').Page, name: string) {
  return page.locator('div.mb-6').filter({ has: page.locator('h3').filter({ hasText: name }) })
}

function benchSection(page: import('@playwright/test').Page) {
  return page.locator('div.mb-6').filter({ has: page.locator('h3').filter({ hasText: 'Bench' }) })
}

// The squad panel h2 heading is "Squad (N)"; go up one level to get the panel div
function squadPanel(page: import('@playwright/test').Page) {
  return page.getByRole('heading', { name: /Squad \(/ }).locator('..')
}

test.describe('FFL Team Builder', () => {
  // ── Layout: read-only mode ────────────────────────────────────────────────

  test.describe('layout: read-only mode', () => {
    test.beforeEach(async ({ page }) => {
      await goToTeamBuilder(page)
    })

    test('shows club name as h1', async ({ page }) => {
      await expect(page.getByRole('heading', { level: 1 })).toContainText('The Howling Cows')
    })

    test('Manage button visible; Done not visible', async ({ page }) => {
      await expect(page.getByRole('button', { name: 'Manage' })).toBeVisible()
      await expect(page.getByRole('button', { name: 'Done' })).not.toBeVisible()
    })

    test('all position group headings present', async ({ page }) => {
      for (const name of ['Goals', 'Kicks', 'Handballs', 'Marks', 'Tackles', 'Hitouts', 'Star', 'Bench']) {
        await expect(page.getByRole('heading', { name })).toBeVisible()
      }
    })

    test('slot counts match composition rules (3/4/4/2/2/2/1)', async ({ page }) => {
      const expected: [string, number][] = [
        ['Goals', 3], ['Kicks', 4], ['Handballs', 4],
        ['Marks', 2], ['Tackles', 2], ['Hitouts', 2], ['Star', 1],
      ]
      for (const [name, count] of expected) {
        const section = positionSection(page, name)
        await expect(section.locator('.rounded-lg')).toHaveCount(count)
      }
    })

    test('bench has Backup Star row and 3 dual-position rows (B1/B2/B3)', async ({ page }) => {
      const bench = benchSection(page)
      // 1 backup star + 3 dual = 4 rows
      await expect(bench.locator('.rounded-lg')).toHaveCount(4)
      await expect(bench.getByText('Backup Star')).toBeVisible()
      await expect(bench.getByText('B1')).toBeVisible()
      await expect(bench.getByText('B2')).toBeVisible()
      await expect(bench.getByText('B3')).toBeVisible()
    })

    test('no Remove buttons visible', async ({ page }) => {
      await expect(page.getByRole('button', { name: 'Remove' })).not.toBeVisible()
    })

    test('no squad panel visible', async ({ page }) => {
      await expect(page.getByRole('heading', { name: /Squad \(/ })).not.toBeVisible()
    })

    test('no position selectors (selects) visible', async ({ page }) => {
      await expect(page.locator('select')).not.toBeVisible()
    })
  })

  // ── Layout: manage mode ───────────────────────────────────────────────────

  test.describe('layout: manage mode', () => {
    test.beforeEach(async ({ page }) => {
      await goToTeamBuilder(page)
      await page.getByRole('button', { name: 'Manage' }).click()
    })

    test('Done button visible; Manage button gone', async ({ page }) => {
      await expect(page.getByRole('button', { name: 'Done' })).toBeVisible()
      await expect(page.getByRole('button', { name: 'Manage' })).not.toBeVisible()
    })

    test('squad panel shows player count', async ({ page }) => {
      await expect(page.getByRole('heading', { name: /Squad \(/ })).toBeVisible()
    })

    test('Remove buttons visible on filled starter slots', async ({ page }) => {
      // Seed has 2 starters (Goals: Jordan Dawson, Kicks: Wayne Milera)
      await expect(page.getByRole('button', { name: 'Remove' }).first()).toBeVisible()
    })

    test('IC checkbox visible on Backup Star bench row', async ({ page }) => {
      const bench = benchSection(page)
      const starRow = bench.locator('.rounded-lg').first()
      await expect(starRow.getByText('IC')).toBeVisible()
    })

    test('dual-position selects appear on bench rows that have a player', async ({ page }) => {
      // All seed players are already assigned to starter slots, so remove one first
      // to make them available in the squad panel, then bench them
      await page.getByRole('button', { name: 'Remove' }).first().click()
      const panel = squadPanel(page)
      await panel.getByRole('button', { name: 'B' }).first().click()

      const bench = benchSection(page)
      // B1 is the 2nd rounded-lg (index 1, after backup star at index 0)
      const b1Row = bench.locator('.rounded-lg').nth(1)
      await expect(b1Row.locator('select')).toHaveCount(2)
    })

    test('★ and B buttons appear in squad panel', async ({ page }) => {
      // Remove a starter first so the squad panel has a player to show buttons for
      await page.getByRole('button', { name: 'Remove' }).first().click()
      const panel = squadPanel(page)
      await expect(panel.getByRole('button', { name: '★' }).first()).toBeVisible()
      await expect(panel.getByRole('button', { name: 'B' }).first()).toBeVisible()
    })

    test('Done saves and returns to read-only mode', async ({ page }) => {
      await page.getByRole('button', { name: 'Done' }).click()
      await expect(page.getByRole('button', { name: 'Manage' })).toBeVisible()
      await expect(page.getByRole('button', { name: 'Done' })).not.toBeVisible()
      await expect(page.getByRole('heading', { name: /Squad \(/ })).not.toBeVisible()
    })
  })

  // ── Partial edit → view → re-edit (state retention) ──────────────────────
  //
  // The bug being guarded against: Apollo's watch fires after a mutation
  // response and resets all local slot state. The initializedMatchId guard
  // prevents that reset.

  test.describe('state retention across Manage/Done cycles', () => {
    test.beforeEach(async ({ page }) => {
      await goToTeamBuilder(page)
    })

    test('player added to slot is visible in read-only mode after Done', async ({ page }) => {
      // Count empty Goals slots before
      const goalsSection = positionSection(page, 'Goals')
      const emptyBefore = await goalsSection.getByText('Empty slot').count()
      expect(emptyBefore).toBeGreaterThan(0)

      // Enter manage mode and add a player to Goals
      await page.getByRole('button', { name: 'Manage' }).click()
      const panel = squadPanel(page)
      await panel.getByRole('button', { name: 'G' }).first().click()

      // Save
      await page.getByRole('button', { name: 'Done' }).click()

      // One fewer empty slot in Goals after saving
      await expect(goalsSection.getByText('Empty slot')).toHaveCount(emptyBefore - 1)
    })

    test('local edits not reset when re-entering manage mode (state retention)', async ({ page }) => {
      // Add a player to an available Kicks slot (Goals is used by other tests in this group)
      await page.getByRole('button', { name: 'Manage' }).click()
      const panel = squadPanel(page)

      // Note the player name about to be added
      const playerName = await panel.locator('.font-medium').first().textContent()
      await panel.getByRole('button', { name: 'K' }).first().click()

      // Save (Done fires the mutation)
      await page.getByRole('button', { name: 'Done' }).click()

      // Immediately re-enter manage mode — player must still be in the Kicks slot
      await page.getByRole('button', { name: 'Manage' }).click()
      const kicksSection = positionSection(page, 'Kicks')
      await expect(kicksSection.getByText(playerName!.trim())).toBeVisible()
    })

    test('two rounds of editing accumulate correctly', async ({ page }) => {
      // Round 1: add to Handballs (Goals/Kicks used by other tests in this group)
      await page.getByRole('button', { name: 'Manage' }).click()
      let panel = squadPanel(page)
      const player1 = await panel.locator('.font-medium').first().textContent()
      await panel.getByRole('button', { name: 'HB' }).first().click()
      await page.getByRole('button', { name: 'Done' }).click()

      // Round 2: add to Kicks
      await page.getByRole('button', { name: 'Manage' }).click()
      panel = squadPanel(page)
      const player2 = await panel.locator('.font-medium').first().textContent()
      await panel.getByRole('button', { name: 'K' }).first().click()
      await page.getByRole('button', { name: 'Done' }).click()

      // Both players visible in read-only mode
      await expect(positionSection(page, 'Handballs').getByText(player1!.trim())).toBeVisible()
      await expect(positionSection(page, 'Kicks').getByText(player2!.trim())).toBeVisible()
    })
  })

  // ── Navigate away and back (server persistence) ───────────────────────────

  test.describe('navigate away and back', () => {
    test('team persists after navigating to Squad view and returning', async ({ page }) => {
      await goToTeamBuilder(page)

      // Add a player to Kicks and save (Goals is filled by state-retention tests)
      await page.getByRole('button', { name: 'Manage' }).click()
      const playerName = await squadPanel(page).locator('.font-medium').first().textContent()
      await squadPanel(page).getByRole('button', { name: 'K' }).first().click()
      await page.getByRole('button', { name: 'Done' }).click()

      // Navigate away to Squad page
      await page.getByRole('link', { name: 'Squad' }).click()
      await page.waitForURL(/\/ffl\/.*\/squad/)

      // Navigate back to Team Builder
      await page.getByRole('link', { name: 'Team Builder' }).click()
      await page.waitForURL(/\/ffl\/.*\/team-builder/)

      // Player still in Kicks (loaded from server)
      await expect(positionSection(page, 'Kicks').getByText(playerName!.trim())).toBeVisible()

      // And still there when entering manage mode
      await page.getByRole('button', { name: 'Manage' }).click()
      await expect(positionSection(page, 'Kicks').getByText(playerName!.trim())).toBeVisible()
    })
  })

  // ── Continue building across sessions ─────────────────────────────────────

  test.describe('continue building on existing team', () => {
    test('existing players present when entering manage mode on partial team', async ({ page }) => {
      await goToTeamBuilder(page)

      // Seed data has some players in the team already
      // Verify they are present in read-only mode
      const goalsSection = positionSection(page, 'Goals')
      // Should have at least 1 filled slot (seed has Jordan Dawson in Goals)
      const emptyCount = await goalsSection.getByText('Empty slot').count()
      expect(emptyCount).toBeLessThan(3)

      // Enter manage and those same players should still be there
      await page.getByRole('button', { name: 'Manage' }).click()
      const stillEmpty = await goalsSection.getByText('Empty slot').count()
      expect(stillEmpty).toBe(emptyCount)
    })

    test('adding to existing team: all players visible after Done', async ({ page }) => {
      await goToTeamBuilder(page)

      // Find players already in Handballs read-only (has capacity from prior tests in the run)
      const hbSection = positionSection(page, 'Handballs')
      const existingNames = await hbSection.locator('.font-medium').allTextContents()

      // Add one more to Handballs
      await page.getByRole('button', { name: 'Manage' }).click()
      const newPlayerName = await squadPanel(page).locator('.font-medium').first().textContent()
      await squadPanel(page).getByRole('button', { name: 'HB' }).first().click()
      await page.getByRole('button', { name: 'Done' }).click()

      // All existing + new player visible
      for (const name of existingNames) {
        await expect(hbSection.getByText(name.trim())).toBeVisible()
      }
      await expect(hbSection.getByText(newPlayerName!.trim())).toBeVisible()

      // And re-entering manage mode retains all of them
      await page.getByRole('button', { name: 'Manage' }).click()
      for (const name of existingNames) {
        await expect(hbSection.getByText(name.trim())).toBeVisible()
      }
      await expect(hbSection.getByText(newPlayerName!.trim())).toBeVisible()
    })
  })

  // ── Bench: backup star ────────────────────────────────────────────────────

  test.describe('bench: backup star', () => {
    test.beforeEach(async ({ page }) => {
      await goToTeamBuilder(page)
      await page.getByRole('button', { name: 'Manage' }).click()
    })

    test('★ button in squad panel adds player to Backup Star row', async ({ page }) => {
      const panel = squadPanel(page)
      const playerName = await panel.locator('.font-medium').first().textContent()

      await panel.getByRole('button', { name: '★' }).first().click()

      const bench = benchSection(page)
      const starRow = bench.locator('.rounded-lg').first()
      await expect(starRow.getByText(playerName!.trim())).toBeVisible()
    })

    test('★ button disabled for all players once Backup Star slot is filled', async ({ page }) => {
      const panel = squadPanel(page)

      // Seed data may or may not have the backup star slot filled
      // Ensure it's empty first (if there's no player there, the ★ buttons should be enabled)
      const bench = benchSection(page)
      const starRow = bench.locator('.rounded-lg').first()
      const hasExistingStarPlayer = await starRow.getByRole('button', { name: 'Remove' }).isVisible()

      if (!hasExistingStarPlayer) {
        // Add a player via ★
        await panel.getByRole('button', { name: '★' }).first().click()
      }

      // Now all ★ buttons should be disabled
      const starButtons = panel.getByRole('button', { name: '★' })
      const count = await starButtons.count()
      for (let i = 0; i < count; i++) {
        await expect(starButtons.nth(i)).toBeDisabled()
      }
    })

    test('removing from Backup Star row re-enables ★ buttons', async ({ page }) => {
      const panel = squadPanel(page)
      const bench = benchSection(page)
      const starRow = bench.locator('.rounded-lg').first()

      // Ensure slot is filled (add if necessary)
      const hasPlayer = await starRow.getByRole('button', { name: 'Remove' }).isVisible()
      if (!hasPlayer) {
        await panel.getByRole('button', { name: '★' }).first().click()
      }

      // Remove from star slot
      await starRow.getByRole('button', { name: 'Remove' }).click()

      // ★ buttons should now be enabled for squad players
      await expect(panel.getByRole('button', { name: '★' }).first()).toBeEnabled()
    })

    test('backup star slot visible after Done → re-open Manage', async ({ page }) => {
      const panel = squadPanel(page)
      const playerName = await panel.locator('.font-medium').first().textContent()

      await panel.getByRole('button', { name: '★' }).first().click()
      await page.getByRole('button', { name: 'Done' }).click()

      // Read-only mode: player name visible in bench section
      await expect(benchSection(page).locator('.rounded-lg').first().getByText(playerName!.trim())).toBeVisible()

      // Re-enter manage mode: still there
      await page.getByRole('button', { name: 'Manage' }).click()
      await expect(benchSection(page).locator('.rounded-lg').first().getByText(playerName!.trim())).toBeVisible()
    })
  })

  // ── Bench: dual-position ──────────────────────────────────────────────────

  test.describe('bench: dual-position', () => {
    test.beforeEach(async ({ page }) => {
      await goToTeamBuilder(page)
      await page.getByRole('button', { name: 'Manage' }).click()
    })

    test('B button in squad panel adds player to first empty dual bench row', async ({ page }) => {
      const panel = squadPanel(page)
      const playerName = await panel.locator('.font-medium').first().textContent()

      // Click B for first available player (adds to next empty dual slot)
      await panel.getByRole('button', { name: 'B' }).first().click()

      // That player name should appear in one of B1/B2/B3 rows
      const bench = benchSection(page)
      const dualRows = bench.locator('.rounded-lg').filter({ hasText: playerName!.trim() })
      await expect(dualRows).toHaveCount(1)
    })

    test('dual bench row shows two position selectors after player is added', async ({ page }) => {
      // Find an empty dual bench row
      const bench = benchSection(page)
      // B1 may already have a player from seed; check B2 (nth 2, 0-indexed)
      const b2Row = bench.locator('.rounded-lg').nth(2)
      const b2HasPlayer = await b2Row.locator('select').count() > 0

      if (!b2HasPlayer) {
        const panel = squadPanel(page)
        // Need to fill B1 first (if not already filled), then B2
        const b1Row = bench.locator('.rounded-lg').nth(1)
        const b1HasPlayer = await b1Row.locator('select').count() > 0
        if (!b1HasPlayer) {
          await panel.getByRole('button', { name: 'B' }).first().click()
        }
        await panel.getByRole('button', { name: 'B' }).first().click()
      }

      // The row now should have 2 select elements
      await expect(b2Row.locator('select')).toHaveCount(2)
    })

    test('selecting a position in one bench row disables it in other rows', async ({ page }) => {
      // Ensure at least 2 dual bench rows have players
      const panel = squadPanel(page)
      const bench = benchSection(page)

      // Fill B1 if empty
      const b1Row = bench.locator('.rounded-lg').nth(1)
      if (!(await b1Row.locator('select').count())) {
        await panel.getByRole('button', { name: 'B' }).first().click()
      }
      // Fill B2
      const b2Row = bench.locator('.rounded-lg').nth(2)
      if (!(await b2Row.locator('select').count())) {
        await panel.getByRole('button', { name: 'B' }).first().click()
      }

      // Select "Goals" in B1's first position selector
      await b1Row.locator('select').first().selectOption('goals')

      // In B2's selectors, the "Goals" option should be disabled
      const b2Select1 = b2Row.locator('select').first()
      const goalsOption = b2Select1.locator('option[value="goals"]')
      await expect(goalsOption).toBeDisabled()
    })

    test('B button disabled once all 3 dual slots are filled', async ({ page }) => {
      const panel = squadPanel(page)
      const bench = benchSection(page)

      // Fill any empty dual slots (there may already be some filled from seed)
      for (let i = 1; i <= 3; i++) {
        const row = bench.locator('.rounded-lg').nth(i)
        const hasPlayer = await row.locator('select').count() > 0
        if (!hasPlayer) {
          await panel.getByRole('button', { name: 'B' }).first().click()
        }
      }

      // All B buttons in squad panel should now be disabled
      const bButtons = panel.getByRole('button', { name: 'B' })
      const count = await bButtons.count()
      for (let i = 0; i < count; i++) {
        await expect(bButtons.nth(i)).toBeDisabled()
      }
    })
  })

  // ── Interchange ───────────────────────────────────────────────────────────

  test.describe('interchange', () => {
    test.beforeEach(async ({ page }) => {
      await goToTeamBuilder(page)
      await page.getByRole('button', { name: 'Manage' }).click()
    })

    test('IC checkbox visible on Backup Star row in manage mode', async ({ page }) => {
      const starRow = benchSection(page).locator('.rounded-lg').first()
      await expect(starRow.getByText('IC')).toBeVisible()
    })

    test('clicking IC on one bench row deactivates IC on another', async ({ page }) => {
      const bench = benchSection(page)
      const panel = squadPanel(page)

      // Ensure B1 has a player (for its IC to appear)
      const b1Row = bench.locator('.rounded-lg').nth(1)
      if (!(await b1Row.locator('select').count())) {
        await panel.getByRole('button', { name: 'B' }).first().click()
      }

      // Activate IC on Backup Star row
      const starIC = bench.locator('.rounded-lg').first().locator('input[type="checkbox"]')
      await starIC.check()
      await expect(starIC).toBeChecked()

      // Activate IC on B1 row → star IC should deactivate
      const b1IC = b1Row.locator('input[type="checkbox"]')
      await b1IC.check()
      await expect(b1IC).toBeChecked()
      await expect(starIC).not.toBeChecked()
    })

    test('interchange state persists through Done → re-open Manage', async ({ page }) => {
      const bench = benchSection(page)

      // Activate IC on Backup Star row
      const starRow = bench.locator('.rounded-lg').first()
      const starIC = starRow.locator('input[type="checkbox"]')
      await starIC.check()

      await page.getByRole('button', { name: 'Done' }).click()
      await page.getByRole('button', { name: 'Manage' }).click()

      // IC should still be checked after reload
      const starICAfter = benchSection(page).locator('.rounded-lg').first().locator('input[type="checkbox"]')
      await expect(starICAfter).toBeChecked()
    })
  })
})
