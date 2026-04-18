import { test, expect } from '@playwright/test'

// Helpers
async function goToTeamBuilder(page: import('@playwright/test').Page) {
  await page.goto('/ffl')
  await page.locator('main nav').last().getByRole('link', { name: '1', exact: true }).click()
  await page.getByTitle('Team Builder').click()
  await page.waitForURL(/\/ffl\/.*\/team-builder/)
}

function positionSection(page: import('@playwright/test').Page, name: string) {
  return page.locator('div.mb-6').filter({ has: page.locator('h3').filter({ hasText: name }) })
}

function benchSection(page: import('@playwright/test').Page) {
  return page.locator('div.mb-6').filter({ has: page.locator('h3').filter({ hasText: 'Bench' }) })
}

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

    test('Manage button visible; Save Team not visible', async ({ page }) => {
      await expect(page.getByRole('button', { name: 'Manage' })).toBeVisible()
      await expect(page.getByRole('button', { name: 'Save Team' })).not.toBeVisible()
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

    test('bench has 4 uniform empty slots', async ({ page }) => {
      const bench = benchSection(page)
      await expect(bench.locator('.rounded-lg')).toHaveCount(4)
      await expect(bench.getByText('Backup Star')).not.toBeVisible()
      await expect(bench.getByText('B1')).not.toBeVisible()
      await expect(bench.getByText('Empty slot').first()).toBeVisible()
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

    test('Save Team and Cancel visible; Manage button gone', async ({ page }) => {
      await expect(page.getByRole('button', { name: 'Save Team' })).toBeVisible()
      await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible()
      await expect(page.getByRole('button', { name: 'Manage' })).not.toBeVisible()
    })

    test('Save Team disabled until a change is made', async ({ page }) => {
      await expect(page.getByRole('button', { name: 'Save Team' })).toBeDisabled()
      await page.getByRole('button', { name: 'Remove' }).first().click()
      await expect(page.getByRole('button', { name: 'Save Team' })).toBeEnabled()
    })

    test('Cancel resets changes and exits manage mode', async ({ page }) => {
      const goalsSection = positionSection(page, 'Goals')
      const filledBefore = await goalsSection.locator('.rounded-lg').filter({ hasNot: page.getByText('Empty slot') }).count()
      await page.getByRole('button', { name: 'Remove' }).first().click()
      await page.getByRole('button', { name: 'Cancel' }).click()
      await expect(page.getByRole('button', { name: 'Manage' })).toBeVisible()
      const filledAfter = await goalsSection.locator('.rounded-lg').filter({ hasNot: page.getByText('Empty slot') }).count()
      expect(filledAfter).toBe(filledBefore)
    })

    test('squad panel shows player count', async ({ page }) => {
      await expect(page.getByRole('heading', { name: /Squad \(/ })).toBeVisible()
    })

    test('Remove buttons visible on filled starter slots', async ({ page }) => {
      await expect(page.getByRole('button', { name: 'Remove' }).first()).toBeVisible()
    })

    test('B button appears in squad panel', async ({ page }) => {
      const panel = squadPanel(page)
      await expect(panel.getByRole('button', { name: 'B' }).first()).toBeVisible()
    })

    test('interchange dropdown visible in manage mode', async ({ page }) => {
      await expect(page.getByLabel('Interchange')).toBeVisible()
    })

    test('Save Team saves and returns to read-only mode', async ({ page }) => {
      await page.getByRole('button', { name: 'Remove' }).first().click()
      await page.getByRole('button', { name: 'Save Team' }).click()
      await expect(page.getByRole('button', { name: 'Manage' })).toBeVisible()
      await expect(page.getByRole('button', { name: 'Save Team' })).not.toBeVisible()
      await expect(page.getByRole('heading', { name: /Squad \(/ })).not.toBeVisible()
    })
  })

  // ── Bench: dual-position slots ────────────────────────────────────────────

  test.describe('bench: dual-position slots', () => {
    test.beforeEach(async ({ page }) => {
      await goToTeamBuilder(page)
      await page.getByRole('button', { name: 'Manage' }).click()
    })

    test('B button adds player to a bench slot', async ({ page }) => {
      const panel = squadPanel(page)
      const playerName = await panel.locator('.font-medium').first().textContent()
      await panel.getByRole('button', { name: 'B' }).first().click()
      await expect(benchSection(page).getByText(playerName!.trim())).toBeVisible()
    })

    test('position selectors appear on filled bench slot', async ({ page }) => {
      await squadPanel(page).getByRole('button', { name: 'B' }).first().click()
      await expect(page.getByLabel('Position 1')).toBeVisible()
      await expect(page.getByLabel('Position 2')).toBeVisible()
    })

    test('selecting Star as position 1 hides position 2 selector', async ({ page }) => {
      await squadPanel(page).getByRole('button', { name: 'B' }).first().click()
      await page.getByLabel('Position 1').selectOption('star')
      await expect(page.getByLabel('Position 2')).not.toBeVisible()
    })

    test('removing bench player clears slot back to empty', async ({ page }) => {
      const bench = benchSection(page)
      await squadPanel(page).getByRole('button', { name: 'B' }).first().click()
      await bench.getByRole('button', { name: 'Remove' }).first().click()
      await expect(bench.getByText('Empty slot').first()).toBeVisible()
    })
  })

  // ── Bench: validation ─────────────────────────────────────────────────────

  test.describe('bench: validation', () => {
    test.beforeEach(async ({ page }) => {
      await goToTeamBuilder(page)
      await page.getByRole('button', { name: 'Manage' }).click()
    })

    test('save blocked when bench player has no position assigned', async ({ page }) => {
      await squadPanel(page).getByRole('button', { name: 'B' }).first().click()
      await expect(page.getByRole('button', { name: 'Save Team' })).toBeDisabled()
      await expect(page.getByText('Each bench player must have a position assigned')).toBeVisible()
    })

    test('save unblocked when bench player has valid star position', async ({ page }) => {
      await squadPanel(page).getByRole('button', { name: 'B' }).first().click()
      await page.getByLabel('Position 1').selectOption('star')
      await expect(page.getByRole('button', { name: 'Save Team' })).toBeEnabled()
    })

    test('save blocked when two bench players set but no interchange chosen', async ({ page }) => {
      const panel = squadPanel(page)
      await panel.getByRole('button', { name: 'B' }).nth(0).click()
      await panel.getByRole('button', { name: 'B' }).nth(0).click()
      // Assign valid positions to both
      await page.getByLabel('Position 1').nth(0).selectOption('star')
      await page.getByLabel('Position 1').nth(1).selectOption('goals')
      await page.getByLabel('Position 2').nth(0).selectOption('kicks')
      await expect(page.getByRole('button', { name: 'Save Team' })).toBeDisabled()
      await expect(page.getByText('Choose an interchange position')).toBeVisible()
    })
  })

  // ── Interchange ───────────────────────────────────────────────────────────

  test.describe('interchange', () => {
    test.beforeEach(async ({ page }) => {
      await goToTeamBuilder(page)
      await page.getByRole('button', { name: 'Manage' }).click()
    })

    test('interchange dropdown lists all 7 positions', async ({ page }) => {
      const ic = page.getByLabel('Interchange')
      // blank + 7 positions
      await expect(ic.locator('option')).toHaveCount(8)
      for (const value of ['goals', 'kicks', 'handballs', 'marks', 'tackles', 'hitouts', 'star']) {
        await expect(ic.locator(`option[value="${value}"]`)).toBeAttached()
      }
    })

    test('interchange selection persists through Save → re-open Manage', async ({ page }) => {
      // Add bench player with star position (valid, single bench = no IC required)
      await squadPanel(page).getByRole('button', { name: 'B' }).first().click()
      await page.getByLabel('Position 1').selectOption('star')
      await page.getByLabel('Interchange').selectOption('star')

      await page.getByRole('button', { name: 'Save Team' }).click()

      // Read-only: check Int label visible in bench slot pill
      await expect(benchSection(page).getByText(/·\s*Int/)).toBeVisible()

      // Re-enter manage — interchange dropdown still set
      await page.getByRole('button', { name: 'Manage' }).click()
      await expect(page.getByLabel('Interchange')).toHaveValue('star')
    })
  })

  // ── Partial edit → view → re-edit (state retention) ──────────────────────

  test.describe('state retention across Manage/Save cycles', () => {
    test.beforeEach(async ({ page }) => {
      await goToTeamBuilder(page)
    })

    test('local edits not reset when re-entering manage mode (state retention)', async ({ page }) => {
      await page.getByRole('button', { name: 'Manage' }).click()
      const panel = squadPanel(page)

      const playerName = await panel.locator('.font-medium').first().textContent()
      await panel.getByRole('button', { name: 'K' }).first().click()

      await page.getByRole('button', { name: 'Save Team' }).click()

      await page.getByRole('button', { name: 'Manage' }).click()
      const kicksSection = positionSection(page, 'Kicks')
      await expect(kicksSection.getByText(playerName!.trim())).toBeVisible()
    })

    test('two rounds of editing accumulate correctly', async ({ page }) => {
      await page.getByRole('button', { name: 'Manage' }).click()
      let panel = squadPanel(page)
      const player1 = await panel.locator('.font-medium').first().textContent()
      await panel.getByRole('button', { name: 'H' }).first().click()
      await page.getByRole('button', { name: 'Save Team' }).click()

      await page.getByRole('button', { name: 'Manage' }).click()
      panel = squadPanel(page)
      const player2 = await panel.locator('.font-medium').first().textContent()
      await panel.getByRole('button', { name: 'K' }).first().click()
      await page.getByRole('button', { name: 'Save Team' }).click()

      await expect(positionSection(page, 'Handballs').getByText(player1!.trim())).toBeVisible()
      await expect(positionSection(page, 'Kicks').getByText(player2!.trim())).toBeVisible()
    })
  })

  // ── Navigate away and back (server persistence) ───────────────────────────

  test.describe('navigate away and back', () => {
  })

  // ── Continue building across sessions ─────────────────────────────────────

  test.describe('continue building on existing team', () => {
    test('adding to existing team: all players visible after Save', async ({ page }) => {
      await goToTeamBuilder(page)

      const hbSection = positionSection(page, 'Handballs')
      const existingNames = await hbSection.locator('.font-medium').allTextContents()

      await page.getByRole('button', { name: 'Manage' }).click()
      const newPlayerName = await squadPanel(page).locator('.font-medium').first().textContent()
      await squadPanel(page).getByRole('button', { name: 'H' }).first().click()
      await page.getByRole('button', { name: 'Save Team' }).click()

      for (const name of existingNames) {
        await expect(hbSection.getByText(name.trim())).toBeVisible()
      }
      await expect(hbSection.getByText(newPlayerName!.trim())).toBeVisible()

      await page.getByRole('button', { name: 'Manage' }).click()
      for (const name of existingNames) {
        await expect(hbSection.getByText(name.trim())).toBeVisible()
      }
      await expect(hbSection.getByText(newPlayerName!.trim())).toBeVisible()
    })
  })

  // ── Header ────────────────────────────────────────────────────────────────

  test.describe('header', () => {
    test.beforeEach(async ({ page }) => {
      await goToTeamBuilder(page)
    })

    test('shows club name in heading', async ({ page }) => {
      await expect(page.getByRole('heading', { level: 1 })).toContainText('The Howling Cows')
    })

    test('shows round name in breadcrumb', async ({ page }) => {
      await expect(page.locator('main').getByRole('link', { name: 'Round 1' })).toBeVisible()
    })

    test('shows season name in breadcrumb', async ({ page }) => {
      await expect(page.locator('main').getByRole('link', { name: 'FFL 2026' })).toBeVisible()
    })
  })

  // ── Manage layout ─────────────────────────────────────────────────────────

  test.describe('manage layout', () => {
    test('squad panel visible alongside team in manage mode', async ({ page }) => {
      await goToTeamBuilder(page)
      await page.getByRole('button', { name: 'Manage' }).click()
      await expect(page.getByRole('heading', { name: /Squad \(/ })).toBeVisible()
    })
  })
})
