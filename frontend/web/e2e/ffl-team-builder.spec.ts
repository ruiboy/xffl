import { test, expect } from './fixtures'
import { setupFflSession } from './helpers'

// Helpers
async function goToTeamBuilder(page: import('@playwright/test').Page) {
  await page.goto('/ffl')
  await page.locator('main nav').last().getByRole('link', { name: '1', exact: true }).click()
  // Wait for round data to load; use main-scoped selector to avoid matching the global nav link
  await expect(page.getByRole('heading', { name: 'Matches' })).toBeVisible({ timeout: 15000 })
  await page.locator('main').getByTitle('Team Builder').click()
  await page.waitForURL(/\/ffl\/club-matches\/.*\/edit/)
}

function positionSection(page: import('@playwright/test').Page, name: string) {
  return page.locator('div.mb-6').filter({ has: page.locator('h3').filter({ hasText: name }) })
}

function benchSection(page: import('@playwright/test').Page) {
  return page.locator('div.mb-6').filter({ has: page.locator('h3').filter({ hasText: 'Bench' }) })
}

function squadPanel(page: import('@playwright/test').Page) {
  // The heading sits inside a flex header row; the panel is that row's parent.
  return page.getByRole('heading', { name: /Squad \(/ }).locator('../..')
}

// Adds the first available squad player via the "+" popup menu.
// `option` matches the menu entry, e.g. /^Kicks/ or /^Bench/.
async function addFromSquad(page: import('@playwright/test').Page, option: RegExp) {
  const panel = squadPanel(page)
  await panel.getByRole('button', { name: 'Add to team' }).first().click()
  await panel.getByRole('button', { name: option }).click()
}

// Removes the first filled starter via the row's "−" popup menu.
async function removeFirstStarter(page: import('@playwright/test').Page) {
  await page.getByRole('button', { name: 'Player actions' }).first().click()
  await page.getByRole('button', { name: 'Remove' }).click()
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

    test('Build Team button visible; Save Team not visible', async ({ page }) => {
      await expect(page.getByRole('button', { name: 'Build Team' })).toBeVisible()
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
      await page.getByRole('button', { name: 'Build Team' }).click()
    })

    test('Save Team and Cancel visible; Build Team button gone', async ({ page }) => {
      await expect(page.getByRole('button', { name: 'Save Team' })).toBeVisible()
      await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible()
      await expect(page.getByRole('button', { name: 'Build Team' })).not.toBeVisible()
    })

    test('Save Team disabled until a change is made', async ({ page }) => {
      await expect(page.getByRole('button', { name: 'Save Team' })).toBeDisabled()
      await removeFirstStarter(page)
      await expect(page.getByRole('button', { name: 'Save Team' })).toBeEnabled()
    })

    test('Cancel resets changes and exits manage mode', async ({ page }) => {
      const goalsSection = positionSection(page, 'Goals')
      const filledBefore = await goalsSection.locator('.rounded-lg').filter({ hasNot: page.getByText('Empty slot') }).count()
      await removeFirstStarter(page)
      await page.getByRole('button', { name: 'Cancel' }).click()
      await expect(page.getByRole('button', { name: 'Build Team' })).toBeVisible()
      const filledAfter = await goalsSection.locator('.rounded-lg').filter({ hasNot: page.getByText('Empty slot') }).count()
      expect(filledAfter).toBe(filledBefore)
    })

    test('squad panel shows player count', async ({ page }) => {
      await expect(page.getByRole('heading', { name: /Squad \(/ })).toBeVisible()
    })

    test('starter menu offers Remove on filled starter slots', async ({ page }) => {
      await page.getByRole('button', { name: 'Player actions' }).first().click()
      await expect(page.getByRole('button', { name: 'Remove' })).toBeVisible()
    })

    test('add menu in squad panel offers all positions and Bench', async ({ page }) => {
      const panel = squadPanel(page)
      await panel.getByRole('button', { name: 'Add to team' }).first().click()
      for (const name of ['Goals', 'Kicks', 'Handballs', 'Marks', 'Tackles', 'Hitouts', 'Star', 'Bench']) {
        await expect(panel.getByRole('button', { name: new RegExp(`^${name}`) })).toBeVisible()
      }
    })

    test('squad panel shows stat summary columns', async ({ page }) => {
      await expect(page.locator('[title="Last 5 form averages"]').first()).toBeVisible()
    })

    test('form/season toggle switches the stat source', async ({ page }) => {
      await expect(page.locator('[title="Last 5 form averages"]').first()).toBeVisible()
      await page.getByRole('button', { name: 'Season', exact: true }).click()
      await expect(page.locator('[title="Season averages"]').first()).toBeVisible()
    })

    test('summary bar shows projected team total', async ({ page }) => {
      await expect(page.getByText(/Projected ~\d+/)).toBeVisible()
    })

    test('clicking a stat header sorts squad cards by that stat', async ({ page }) => {
      const panel = squadPanel(page)
      const cards = panel.locator('[draggable="true"]')
      await expect(cards.first()).toBeVisible()
      await panel.getByRole('button', { name: 'K', exact: true }).click()
      // K is the first stat cell on each card; values must be non-increasing
      const kVals = await cards.evaluateAll(els =>
        els.map(c => {
          const v = parseFloat(c.querySelectorAll('span.w-9')[0]?.textContent ?? '')
          return Number.isNaN(v) ? -1 : v
        }),
      )
      expect(kVals.length).toBeGreaterThan(1)
      expect(kVals).toEqual([...kVals].sort((a, b) => b - a))
    })

    test('interchange dropdown visible in manage mode', async ({ page }) => {
      await expect(page.getByLabel('Interchange')).toBeVisible()
    })

    test('Save Team saves and returns to read-only mode', async ({ page }) => {
      await removeFirstStarter(page)
      await page.getByRole('button', { name: 'Save Team' }).click()
      await expect(page.getByRole('button', { name: 'Build Team' })).toBeVisible()
      await expect(page.getByRole('button', { name: 'Save Team' })).not.toBeVisible()
      await expect(page.getByRole('heading', { name: /Squad \(/ })).not.toBeVisible()
    })
  })

  // ── Bench: dual-position slots ────────────────────────────────────────────

  test.describe('bench: dual-position slots', () => {
    test.beforeEach(async ({ page }) => {
      await goToTeamBuilder(page)
      await page.getByRole('button', { name: 'Build Team' }).click()
    })

    test('Bench menu option adds player to a bench slot', async ({ page }) => {
      const panel = squadPanel(page)
      const playerName = await panel.locator('.font-medium').first().textContent()
      await addFromSquad(page, /^Bench/)
      await expect(benchSection(page).getByText(playerName!.trim())).toBeVisible()
    })

    test('position selectors appear on filled bench slot', async ({ page }) => {
      await addFromSquad(page, /^Bench/)
      await expect(page.getByLabel('Position 1')).toBeVisible()
      await expect(page.getByLabel('Position 2')).toBeVisible()
    })

    test('selecting Star as position 1 hides position 2 selector', async ({ page }) => {
      await addFromSquad(page, /^Bench/)
      await page.getByLabel('Position 1').selectOption('star')
      await expect(page.getByLabel('Position 2')).not.toBeVisible()
    })

    test('removing bench player via menu clears slot back to empty', async ({ page }) => {
      const bench = benchSection(page)
      await addFromSquad(page, /^Bench/)
      await bench.getByRole('button', { name: 'Bench player actions' }).first().click()
      await page.getByRole('button', { name: 'Remove' }).click()
      await expect(bench.getByText('Empty slot').first()).toBeVisible()
    })

    test('bench menu moves player to a starter position', async ({ page }) => {
      const bench = benchSection(page)
      const panel = squadPanel(page)
      const playerName = (await panel.locator('.font-medium').first().textContent())!.trim()
      await addFromSquad(page, /^Bench/)
      await bench.getByRole('button', { name: 'Bench player actions' }).first().click()
      await page.getByRole('button', { name: /^Kicks/ }).click()
      await expect(positionSection(page, 'Kicks').getByText(playerName)).toBeVisible()
      await expect(bench.getByText(playerName)).not.toBeVisible()
    })
  })

  // ── Bench: validation ─────────────────────────────────────────────────────

  test.describe('bench: validation', () => {
    test.beforeEach(async ({ page }) => {
      await goToTeamBuilder(page)
      await page.getByRole('button', { name: 'Build Team' }).click()
    })

    test('save blocked when bench player has no position assigned', async ({ page }) => {
      await addFromSquad(page, /^Bench/)
      await expect(page.getByRole('button', { name: 'Save Team' })).toBeDisabled()
      await expect(page.getByText('Each bench player must have a position assigned')).toBeVisible()
    })

    test('save unblocked when bench player has valid star position', async ({ page }) => {
      await addFromSquad(page, /^Bench/)
      await page.getByLabel('Position 1').selectOption('star')
      await expect(page.getByRole('button', { name: 'Save Team' })).toBeEnabled()
    })

    test('save blocked when two bench players set but no interchange chosen', async ({ page }) => {
      await addFromSquad(page, /^Bench/)
      await addFromSquad(page, /^Bench/)
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
      await page.getByRole('button', { name: 'Build Team' }).click()
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
      await addFromSquad(page, /^Bench/)
      await page.getByLabel('Position 1').selectOption('star')
      await page.getByLabel('Interchange').selectOption('star')

      await page.getByRole('button', { name: 'Save Team' }).click()

      // Read-only: check Int label visible in bench slot pill
      await expect(benchSection(page).getByText(/·\s*Int/)).toBeVisible()

      // Re-enter manage — interchange dropdown still set
      await page.getByRole('button', { name: 'Build Team' }).click()
      await expect(page.getByLabel('Interchange')).toHaveValue('star')
    })
  })

  // ── Partial edit → view → re-edit (state retention) ──────────────────────

  test.describe('state retention across Manage/Save cycles', () => {
    test.beforeEach(async ({ page }) => {
      await goToTeamBuilder(page)
    })

    test('local edits not reset when re-entering manage mode (state retention)', async ({ page }) => {
      await page.getByRole('button', { name: 'Build Team' }).click()
      const panel = squadPanel(page)

      const playerName = await panel.locator('.font-medium').first().textContent()
      await addFromSquad(page, /^Kicks/)

      await page.getByRole('button', { name: 'Save Team' }).click()

      await page.getByRole('button', { name: 'Build Team' }).click()
      const kicksSection = positionSection(page, 'Kicks')
      await expect(kicksSection.getByText(playerName!.trim())).toBeVisible()
    })

    test('two rounds of editing accumulate correctly', async ({ page }) => {
      await page.getByRole('button', { name: 'Build Team' }).click()
      let panel = squadPanel(page)
      const player1 = await panel.locator('.font-medium').first().textContent()
      await addFromSquad(page, /^Handballs/)
      await page.getByRole('button', { name: 'Save Team' }).click()

      await page.getByRole('button', { name: 'Build Team' }).click()
      panel = squadPanel(page)
      const player2 = await panel.locator('.font-medium').first().textContent()
      await addFromSquad(page, /^Kicks/)
      await page.getByRole('button', { name: 'Save Team' }).click()

      await expect(positionSection(page, 'Handballs').getByText(player1!.trim())).toBeVisible()
      await expect(positionSection(page, 'Kicks').getByText(player2!.trim())).toBeVisible()
    })
  })

  // ── Starter popup menu ────────────────────────────────────────────────────

  test.describe('starter popup menu', () => {
    test.beforeEach(async ({ page }) => {
      await goToTeamBuilder(page)
      await page.getByRole('button', { name: 'Build Team' }).click()
    })

    test('− menu moves starter to another position', async ({ page }) => {
      const goals = positionSection(page, 'Goals')
      const row = goals.locator('.rounded-lg').filter({ hasNot: page.getByText('Empty slot') }).first()
      const name = await row.locator('.font-medium').first().textContent()
      await row.getByRole('button', { name: 'Player actions' }).click()
      await page.getByRole('button', { name: /^Kicks/ }).click()
      await expect(positionSection(page, 'Kicks').getByText(name!.trim())).toBeVisible()
    })

    test('− menu moves starter to bench', async ({ page }) => {
      const goals = positionSection(page, 'Goals')
      const row = goals.locator('.rounded-lg').filter({ hasNot: page.getByText('Empty slot') }).first()
      const name = await row.locator('.font-medium').first().textContent()
      await row.getByRole('button', { name: 'Player actions' }).click()
      await page.getByRole('button', { name: /^Bench/ }).click()
      await expect(benchSection(page).getByText(name!.trim())).toBeVisible()
    })

    test('− menu removes starter from the team', async ({ page }) => {
      const goals = positionSection(page, 'Goals')
      const row = goals.locator('.rounded-lg').filter({ hasNot: page.getByText('Empty slot') }).first()
      const name = (await row.locator('.font-medium').first().textContent())!.trim()
      await row.getByRole('button', { name: 'Player actions' }).click()
      await page.getByRole('button', { name: 'Remove' }).click()
      await expect(goals.getByText(name)).not.toBeVisible()
      await expect(squadPanel(page).getByText(name)).toBeVisible()
    })
  })

  // ── Drag and drop ─────────────────────────────────────────────────────────

  test.describe('drag and drop', () => {
    // Tall viewport: native HTML5 drag breaks if Playwright has to scroll mid-drag,
    // so keep source and target both on screen.
    test.use({ viewport: { width: 1280, height: 2400 } })

    test.beforeEach(async ({ page }) => {
      await goToTeamBuilder(page)
      await page.getByRole('button', { name: 'Build Team' }).click()
    })

    test('drag squad player onto an empty starter slot assigns them', async ({ page }) => {
      const panel = squadPanel(page)
      const card = panel.locator('[draggable="true"]').first()
      const name = await card.locator('.font-medium').first().textContent()
      const target = positionSection(page, 'Kicks').locator('.rounded-lg').filter({ hasText: 'Empty slot' }).first()
      await card.dragTo(target)
      await expect(positionSection(page, 'Kicks').getByText(name!.trim())).toBeVisible()
    })

    test('drag squad player onto an empty bench slot assigns them', async ({ page }) => {
      const panel = squadPanel(page)
      const card = panel.locator('[draggable="true"]').first()
      const name = await card.locator('.font-medium').first().textContent()
      const target = benchSection(page).locator('.rounded-lg').filter({ hasText: 'Empty slot' }).first()
      await card.dragTo(target)
      await expect(benchSection(page).getByText(name!.trim())).toBeVisible()
    })

    test('drag starter back to squad panel removes them from the team', async ({ page }) => {
      const goals = positionSection(page, 'Goals')
      const row = goals.locator('.rounded-lg').filter({ hasNot: page.getByText('Empty slot') }).first()
      const name = (await row.locator('.font-medium').first().textContent())!.trim()
      const panel = squadPanel(page)
      await row.dragTo(panel.locator('[draggable="true"]').first())
      await expect(goals.getByText(name)).not.toBeVisible()
      await expect(panel.getByText(name)).toBeVisible()
    })
  })

  // ── Subs mode — substitution ─────────────────────────────────────────────
  //
  // Round 2, The Howling Cows (club_match id=4):
  //   Henry Smith   — goals, played, 48  (active starter)
  //   Hugh McCluggage — kicks, DNP, 0    (candidate to sub out)
  //   Brock Thunder — bench kicks, played, 20  (will sub in)

  test.describe('subs mode: substitution', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/ffl/club-matches/4/edit')
      await expect(page.getByRole('heading', { level: 1 })).toBeVisible({ timeout: 15000 })
    })

    test('Substitutions button appears when AFL match has started', async ({ page }) => {
      await expect(page.getByRole('button', { name: 'Substitutions' })).toBeVisible()
    })

    test('entering subs mode shows Cancel and Save Subs buttons', async ({ page }) => {
      await page.getByRole('button', { name: 'Substitutions' }).click()
      await expect(page.getByRole('button', { name: 'Save Subs' })).toBeVisible()
      await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible()
      await expect(page.getByRole('button', { name: 'Substitutions' })).not.toBeVisible()
    })

    test('DNP starter has amber highlight in subs mode', async ({ page }) => {
      await page.getByRole('button', { name: 'Substitutions' }).click()
      const kicksSection = positionSection(page, 'Kicks')
      const hughRow = kicksSection.locator('.rounded-lg').filter({ hasText: 'Hugh McCluggage' })
      await expect(hughRow).toHaveClass(/border-amber-600/)
    })

    test('clicking DNP starter toggles to sky selection border', async ({ page }) => {
      await page.getByRole('button', { name: 'Substitutions' }).click()
      const kicksSection = positionSection(page, 'Kicks')
      const hughRow = kicksSection.locator('.rounded-lg').filter({ hasText: 'Hugh McCluggage' })
      await hughRow.click()
      await expect(hughRow).toHaveClass(/border-sky-500/)
    })

    test('clicking DNP starter again deselects it', async ({ page }) => {
      await page.getByRole('button', { name: 'Substitutions' }).click()
      const kicksSection = positionSection(page, 'Kicks')
      const hughRow = kicksSection.locator('.rounded-lg').filter({ hasText: 'Hugh McCluggage' })
      await hughRow.click()
      await hughRow.click()
      await expect(hughRow).toHaveClass(/border-amber-600/)
    })

    test('Cancel exits subs mode without saving', async ({ page }) => {
      await page.getByRole('button', { name: 'Substitutions' }).click()
      await page.getByRole('button', { name: 'Cancel' }).click()
      await expect(page.getByRole('button', { name: 'Substitutions' })).toBeVisible()
      // No badges should be changed
      const kicksSection = positionSection(page, 'Kicks')
      await expect(kicksSection.getByText('Subbed')).not.toBeVisible()
    })

    test('Save Subs persists subbed_out and subbed_in statuses', async ({ page }) => {
      await page.getByRole('button', { name: 'Substitutions' }).click()
      const kicksSection = positionSection(page, 'Kicks')
      await kicksSection.locator('.rounded-lg').filter({ hasText: 'Hugh McCluggage' }).click()
      await page.getByRole('button', { name: 'Save Subs' }).click()

      // Wait for normal mode to return
      await expect(page.getByRole('button', { name: 'Substitutions' })).toBeVisible({ timeout: 10000 })

      // Hugh shows Subbed badge
      await expect(kicksSection.getByText('Subbed')).toBeVisible()
      // Brock shows Sub In badge in bench
      await expect(benchSection(page).getByText('Sub In')).toBeVisible()
    })

    test('after sub saved, bench shows covering arrow for the DNP starter', async ({ page }) => {
      await page.getByRole('button', { name: 'Substitutions' }).click()
      await positionSection(page, 'Kicks').locator('.rounded-lg').filter({ hasText: 'Hugh McCluggage' }).click()
      await page.getByRole('button', { name: 'Save Subs' }).click()
      await expect(page.getByRole('button', { name: 'Substitutions' })).toBeVisible({ timeout: 10000 })

      // Brock's bench row shows "↑ Brock Thunder" (highlighted as covering)
      await expect(benchSection(page).getByText(/↑.*Brock Thunder/)).toBeVisible()
    })

    test('after sub saved, starter row shows covering bench player', async ({ page }) => {
      await page.getByRole('button', { name: 'Substitutions' }).click()
      await positionSection(page, 'Kicks').locator('.rounded-lg').filter({ hasText: 'Hugh McCluggage' }).click()
      await page.getByRole('button', { name: 'Save Subs' }).click()
      await expect(page.getByRole('button', { name: 'Substitutions' })).toBeVisible({ timeout: 10000 })

      // Hugh's starter row shows "↑ Brock Thunder"
      await expect(positionSection(page, 'Kicks').getByText(/↑.*Brock Thunder/)).toBeVisible()
    })

    test('re-declaring with empty subs resets statuses to named', async ({ page }) => {
      // First declare a sub
      await page.getByRole('button', { name: 'Substitutions' }).click()
      await positionSection(page, 'Kicks').locator('.rounded-lg').filter({ hasText: 'Hugh McCluggage' }).click()
      await page.getByRole('button', { name: 'Save Subs' }).click()
      await expect(page.getByRole('button', { name: 'Substitutions' })).toBeVisible({ timeout: 10000 })

      // Re-enter subs mode and save with no subs selected
      await page.getByRole('button', { name: 'Substitutions' }).click()
      await positionSection(page, 'Kicks').locator('.rounded-lg').filter({ hasText: 'Hugh McCluggage' }).click()
      await page.getByRole('button', { name: 'Save Subs' }).click()
      await expect(page.getByRole('button', { name: 'Substitutions' })).toBeVisible({ timeout: 10000 })

      // Badges should be gone
      await expect(positionSection(page, 'Kicks').getByText('Subbed')).not.toBeVisible()
      await expect(benchSection(page).getByText('Sub In')).not.toBeVisible()
    })

    test('sub badges visible in match SquadTable after saving', async ({ page }) => {
      await page.getByRole('button', { name: 'Substitutions' }).click()
      await positionSection(page, 'Kicks').locator('.rounded-lg').filter({ hasText: 'Hugh McCluggage' }).click()
      await page.getByRole('button', { name: 'Save Subs' }).click()
      await expect(page.getByRole('button', { name: 'Substitutions' })).toBeVisible({ timeout: 10000 })

      // Navigate to Round 2 match view (match id=2)
      await page.goto('/ffl/matches/2')
      await page.waitForLoadState('networkidle')

      // SquadTable: Hugh shows Subbed, Brock shows Sub In
      await expect(page.getByText('Subbed').first()).toBeVisible()
      await expect(page.getByText('Sub In').first()).toBeVisible()
    })

    test('SquadTable shows covering arrow after sub saved', async ({ page }) => {
      await page.getByRole('button', { name: 'Substitutions' }).click()
      await positionSection(page, 'Kicks').locator('.rounded-lg').filter({ hasText: 'Hugh McCluggage' }).click()
      await page.getByRole('button', { name: 'Save Subs' }).click()
      await expect(page.getByRole('button', { name: 'Substitutions' })).toBeVisible({ timeout: 10000 })

      await page.goto('/ffl/matches/2')
      await page.waitForLoadState('networkidle')

      // SquadTable: covering arrow visible for Brock's bench row
      await expect(page.getByText(/↑.*Brock Thunder/).first()).toBeVisible()
    })
  })

  // ── Subs mode — interchange ───────────────────────────────────────────────
  //
  // Round 4, The Howling Cows (club_match id=8):
  //   Henry Smith     — goals, played, 48  (higher score, stays)
  //   Hugh McCluggage — goals, played, 30  (lower score, displaced by interchange)
  //   Brock Thunder   — bench goals/IC, played, 50  (interchange bench, outscores Hugh)

  test.describe('subs mode: interchange', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/ffl/club-matches/8/edit')
      await expect(page.getByRole('heading', { level: 1 })).toBeVisible({ timeout: 15000 })
    })

    test('Substitutions button appears for a round with AFL data', async ({ page }) => {
      await expect(page.getByRole('button', { name: 'Substitutions' })).toBeVisible()
    })

    test('interchange bench row is shown in subs mode', async ({ page }) => {
      await page.getByRole('button', { name: 'Substitutions' }).click()
      // Brock is the interchange bench player — his row must be visible
      await expect(benchSection(page).locator('.rounded-lg').filter({ hasText: 'Brock Thunder' })).toBeVisible()
    })

    test('interchange starts not applied (amber) even when bench outscores starter', async ({ page }) => {
      await page.getByRole('button', { name: 'Substitutions' }).click()
      // Brock (50) outscores Hugh (30) → interchange beneficial, but NOT auto-applied;
      // user must explicitly click to apply (phase 23 UX change)
      const brockRow = benchSection(page).locator('.rounded-lg').filter({ hasText: 'Brock Thunder' })
      await expect(brockRow).toHaveClass(/border-amber-600/)
    })

    test('clicking interchange row toggles it on (sky border)', async ({ page }) => {
      await page.getByRole('button', { name: 'Substitutions' }).click()
      const brockRow = benchSection(page).locator('.rounded-lg').filter({ hasText: 'Brock Thunder' })
      await brockRow.click()
      await expect(brockRow).toHaveClass(/border-sky-500/)
    })

    test('Save Subs persists interchanged_out and interchanged_in statuses', async ({ page }) => {
      await page.getByRole('button', { name: 'Substitutions' }).click()
      // Interchange is NOT auto-applied — user must click Brock to apply it
      await benchSection(page).locator('.rounded-lg').filter({ hasText: 'Brock Thunder' }).click()
      await page.getByRole('button', { name: 'Save Subs' }).click()
      await expect(page.getByRole('button', { name: 'Substitutions' })).toBeVisible({ timeout: 10000 })

      // Hugh (displaced) shows IC'ed; Brock (interchange) shows IC In
      await expect(positionSection(page, 'Goals').getByText('IC\'ed')).toBeVisible()
      await expect(benchSection(page).getByText('IC In')).toBeVisible()
      // Henry remains with normal AFL status (Played)
      await expect(positionSection(page, 'Goals').getByText('Played').first()).toBeVisible()
    })

    test('saving without applying interchange leaves no statuses changed', async ({ page }) => {
      await page.getByRole('button', { name: 'Substitutions' }).click()
      // Interchange starts NOT applied (amber) — save immediately without clicking
      await page.getByRole('button', { name: 'Save Subs' }).click()
      await expect(page.getByRole('button', { name: 'Substitutions' })).toBeVisible({ timeout: 10000 })

      await expect(positionSection(page, 'Goals').getByText('IC\'ed')).not.toBeVisible()
      await expect(benchSection(page).getByText('IC In')).not.toBeVisible()
    })
  })

  // ── Bye badge ─────────────────────────────────────────────────────────────
  //
  // Round 5, The Howling Cows (club_match id=10):
  //   Henry Smith    — goals, bye, 38  (Brisbane Lions on AFL bye)
  //   Hugh McCluggage — kicks, bye, 14

  test.describe('bye badge', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/ffl/club-matches/10/edit')
      await expect(page.getByRole('heading', { level: 1 })).toBeVisible({ timeout: 15000 })
    })

    test('bye players show Bye badge', async ({ page }) => {
      await expect(positionSection(page, 'Goals').getByText('Bye')).toBeVisible()
      await expect(positionSection(page, 'Kicks').getByText('Bye')).toBeVisible()
    })

    test('bye starters display their pre-computed average score', async ({ page }) => {
      await expect(positionSection(page, 'Goals').getByText('38', { exact: true })).toBeVisible()
      await expect(positionSection(page, 'Kicks').getByText('14', { exact: true })).toBeVisible()
    })

    test('Substitutions button appears in a bye round', async ({ page }) => {
      await expect(page.getByRole('button', { name: 'Substitutions' })).toBeVisible()
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

      await page.getByRole('button', { name: 'Build Team' }).click()
      const newPlayerName = await squadPanel(page).locator('.font-medium').first().textContent()
      await addFromSquad(page, /^Handballs/)
      await page.getByRole('button', { name: 'Save Team' }).click()

      for (const name of existingNames) {
        await expect(hbSection.getByText(name.trim())).toBeVisible()
      }
      await expect(hbSection.getByText(newPlayerName!.trim())).toBeVisible()

      await page.getByRole('button', { name: 'Build Team' }).click()
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

    test('shows current round name in header nav', async ({ page }) => {
      // Round name is shown in the prev/next round navigator in the header
      await expect(page.locator('main').getByText('Round 1').first()).toBeVisible()
    })

    test('header has Squad link', async ({ page }) => {
      await expect(page.locator('main').getByRole('link', { name: 'Squad' })).toBeVisible()
    })

    test('Squad link in header navigates to squad page', async ({ page }) => {
      await page.locator('main').getByRole('link', { name: 'Squad' }).click()
      await expect(page).toHaveURL(/\/ffl\/club-seasons\//)
    })
  })

  // ── Manage layout ─────────────────────────────────────────────────────────

  test.describe('manage layout', () => {
    test('squad panel visible alongside team in manage mode', async ({ page }) => {
      await goToTeamBuilder(page)
      await page.getByRole('button', { name: 'Build Team' }).click()
      await expect(page.getByRole('heading', { name: /Squad \(/ })).toBeVisible()
    })
  })

  // ── Traded players in available pool ─────────────────────────────────────
  // After trading a player out from the Squad, they should appear muted under
  // the "Traded" toggle in the Team Builder available pane (manage mode only).

  test.describe('traded players', () => {
    test('a traded player surfaces under Traded in the available pool', async ({ page }) => {
      // Step 1: Trade Henry Smith out via the Squad page.
      await setupFflSession(page)
      await page.getByRole('link', { name: 'Squad' }).click()
      await page.waitForURL(/\/ffl\/club-seasons\//)
      await page.waitForLoadState('networkidle')

      await page.getByRole('button', { name: 'Manage' }).click()
      const henrysRow = page.getByRole('row', { name: /Henry Smith/ })
      await henrysRow.getByRole('button', { name: 'Remove' }).click()
      const removeDialog = page.getByRole('heading', { name: 'Remove Player' }).locator('..')
      await removeDialog.getByRole('button', { name: 'Remove' }).click()
      await expect(page.getByText('Saved')).toBeVisible()

      // Step 2: Open Round 3 Team Builder. Round 3 has no FFL player_matches
      // in the seed, so Henry isn't in any slot — `tradedPlayers` will surface
      // him. (Round 1/2 still have him assigned from seed data, where the
      // Traded section is intentionally suppressed.) IDs are stable thanks to
      // the RESTART IDENTITY reset in resetDb: The Howling Cows club_match for
      // Round 3 always has id=6 (insert order: R1-Ruiboys=1, R1-Cows=2,
      // R2-Ruiboys=3, R2-Cows=4, R3-Ruiboys=5, R3-Cows=6).
      await page.goto('/ffl/club-matches/6/edit')
      await page.waitForLoadState('networkidle')
      await page.getByRole('button', { name: 'Build Team' }).click()

      // Traded toggle is visible with a count.
      const tradedToggle = page.getByRole('button', { name: /Traded \(\d+\)/ })
      await expect(tradedToggle).toBeVisible()
      await tradedToggle.click()

      // Henry Smith now appears in the available pool's Traded section.
      const squadPanel = page.getByRole('heading', { name: /Squad \(/ }).locator('../..')
      await expect(squadPanel.getByText('Henry Smith')).toBeVisible()
    })
  })

  // ── Improve your score pill: interchange ─────────────────────────────────
  //
  // Round 4, The Howling Cows (club_match id=8):
  //   Brock Thunder — interchange bench (position=NULL, backup_positions='goals',
  //                   interchange_position='goals'), score=50
  //   Hugh McCluggage — goals starter, played, score=30
  //   → Brock outscores Hugh → interchange suggestion fires → pill visible

  test.describe('improve your score pill', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/ffl/club-matches/8/edit')
      await expect(page.getByRole('heading', { level: 1 })).toBeVisible({ timeout: 15000 })
    })

    test('pill visible in default mode when interchange suggestion exists', async ({ page }) => {
      await expect(page.getByText('Improve your score:')).toBeVisible()
    })

    test('pill shows Interchange wording with player name', async ({ page }) => {
      await expect(page.getByText(/Interchange:.*Brock Thunder/)).toBeVisible()
    })

    test('pill visible in subs mode', async ({ page }) => {
      await page.getByRole('button', { name: 'Substitutions' }).click()
      await expect(page.getByText('Improve your score:')).toBeVisible()
    })

    test('pill hidden in manage mode', async ({ page }) => {
      await page.getByRole('button', { name: 'Build Team' }).click()
      await expect(page.getByText('Improve your score:')).not.toBeVisible()
    })
  })

  // ── Improve your score pill: sub ─────────────────────────────────────────
  //
  // Round 2, The Howling Cows (club_match id=4):
  //   Hugh McCluggage — kicks starter, DNP
  //   Brock Thunder   — bench covering kicks, played, score=20
  //   → Hugh is DNP + Brock covers kicks → sub suggestion fires → pill visible

  test.describe('improve your score pill — sub suggestion', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/ffl/club-matches/4/edit')
      await expect(page.getByRole('heading', { level: 1 })).toBeVisible({ timeout: 15000 })
    })

    test('pill shows Sub wording for DNP starter', async ({ page }) => {
      await expect(page.getByText(/Sub:.*Brock Thunder/)).toBeVisible()
    })
  })

  // ── Replicate and Clear buttons ───────────────────────────────────────────
  //
  // Round 2 (club_match id=4): prevRound is Round 1, which has Howling Cows
  // data → Replicate Round 1 button visible in manage mode.

  test.describe('replicate and clear buttons', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/ffl/club-matches/4/edit')
      await expect(page.getByRole('heading', { level: 1 })).toBeVisible({ timeout: 15000 })
      await page.getByRole('button', { name: 'Build Team' }).click()
    })

    test('Replicate Round button visible in manage mode when previous round exists', async ({ page }) => {
      await expect(page.getByRole('button', { name: /Replicate Round/ })).toBeVisible()
    })

    test('Clear button visible in manage mode', async ({ page }) => {
      await expect(page.getByRole('button', { name: 'Clear' })).toBeVisible()
    })
  })
})
