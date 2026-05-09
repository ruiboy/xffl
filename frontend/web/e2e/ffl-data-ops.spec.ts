import { test, expect } from './fixtures'
import { setupFflSession } from './helpers'

// Minimal Ruiboys-format post: enough lines for the parser to identify the format
// and return some parsed rows (1 goals player) without depending on seeded squad data.
const minimalRuiboysPost = `388
GOALS                15
Jeremy Cameron – Geel            15
BENCH
Hugh McCluggage – Bris            * (INT)    52
`

test.describe('FFL Data Ops', () => {
  test.beforeEach(async ({ page }) => {
    await setupFflSession(page)
    await page.getByRole('link', { name: /Data Ops/i }).click()
    await page.waitForURL('/ffl/data-ops')
  })

  test('AFL Stats tab shows match table with correct columns', async ({ page }) => {
    await page.getByRole('button', { name: 'AFL Stats' }).click()
    await page.waitForLoadState('networkidle')
    await expect(page.getByRole('columnheader', { name: 'Match' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Status' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Score' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Players' })).toBeVisible()
  })

  test.describe('FFL Teams tab', () => {
    test.beforeEach(async ({ page }) => {
      await page.getByRole('button', { name: 'FFL Teams' }).click()
      await page.waitForLoadState('networkidle')
    })

    test('shows the Data Ops heading', async ({ page }) => {
      await expect(page.getByRole('heading', { level: 1 })).toContainText('Data Ops')
    })

    test('club rows are populated', async ({ page }) => {
      await expect(page.getByRole('cell', { name: 'Ruiboys' })).toBeVisible()
    })

    test('round selector defaults to the live round', async ({ page }) => {
      // Clock override puts us in Round 3.
      const roundSelect = page.locator('select').first()
      await expect(roundSelect).toHaveValue(/\d+/)
      const selectedText = await roundSelect.locator('option:checked').textContent()
      expect(selectedText).toContain('Round 3')
    })

    test('import panel opens inside a card with Cancel and Read Team buttons', async ({ page }) => {
      const ruiboysRow = page.getByRole('row').filter({ hasText: 'Ruiboys' }).first()
      await ruiboysRow.getByRole('button', { name: 'Import Team' }).click()

      // Import Team button should be gone once panel is open
      await expect(ruiboysRow.getByRole('button', { name: 'Import Team' })).not.toBeVisible()

      // Card should contain Cancel and Read Team buttons
      await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible()
      await expect(page.getByRole('button', { name: 'Read Team' })).toBeVisible()
    })

    test('cancel button inside card closes the import panel', async ({ page }) => {
      const ruiboysRow = page.getByRole('row').filter({ hasText: 'Ruiboys' }).first()
      await ruiboysRow.getByRole('button', { name: 'Import Team' }).click()

      await expect(page.getByRole('button', { name: 'Cancel' })).toBeVisible()
      await page.getByRole('button', { name: 'Cancel' }).click()

      // Panel should be closed and Import Team button should return
      await expect(ruiboysRow.getByRole('button', { name: 'Import Team' })).toBeVisible()
      await expect(page.getByRole('button', { name: 'Cancel' })).not.toBeVisible()
    })

    test('golden path: read a post and reach the review phase', async ({ page }) => {
      // Open import panel for Ruiboys
      await page.getByRole('row').filter({ hasText: 'Ruiboys' }).first().getByRole('button', { name: 'Import Team' }).click()

      // Select team format
      await page.locator('select').nth(1).selectOption('Ruiboys')

      // Paste post
      await page.locator('textarea').fill(minimalRuiboysPost)

      // Read Team button should be enabled
      const readBtn = page.getByRole('button', { name: 'Read Team' })
      await expect(readBtn).toBeEnabled()
      await readBtn.click()

      // Review phase
      await expect(page.getByRole('button', { name: '← Back' })).toBeVisible({ timeout: 15000 })
      await expect(page.getByText(/\d+ players/)).toBeVisible()

      // Table structure
      await expect(page.getByRole('columnheader', { name: 'Posted' })).toBeVisible()
      await expect(page.getByRole('columnheader', { name: 'Resolved' })).toBeVisible()
      await expect(page.getByRole('columnheader', { name: 'Position' })).toBeVisible()
      await expect(page.getByRole('columnheader', { name: 'Score' }).first()).toBeVisible()
      await expect(page.getByRole('columnheader', { name: 'Confidence' })).toBeVisible()
    })

    test('back button returns to input phase', async ({ page }) => {
      // Open import panel for Ruiboys, select format, paste, read
      await page.getByRole('row').filter({ hasText: 'Ruiboys' }).first().getByRole('button', { name: 'Import Team' }).click()
      await page.locator('select').nth(1).selectOption('Ruiboys')
      await page.locator('textarea').fill(minimalRuiboysPost)
      await page.getByRole('button', { name: 'Read Team' }).click()

      await expect(page.getByRole('button', { name: '← Back' })).toBeVisible({ timeout: 15000 })
      await page.getByRole('button', { name: '← Back' }).click()

      await expect(page.locator('textarea')).toBeVisible()
      await expect(page.getByRole('button', { name: 'Read Team' })).toBeVisible()
    })
  })
})
