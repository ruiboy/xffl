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
    await page.waitForLoadState('networkidle')
  })

  test('shows the Data Ops heading', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Data Ops')
  })

  test('club selector is populated', async ({ page }) => {
    const select = page.locator('select').first()
    const options = await select.locator('option').allTextContents()
    expect(options.some(o => o.includes('Ruiboys'))).toBeTruthy()
  })

  test('round selector defaults to the live round', async ({ page }) => {
    // Clock override puts us in Round 3.
    const roundSelect = page.locator('select').nth(1)
    await expect(roundSelect).toHaveValue(/\d+/)
    const selectedText = await roundSelect.locator('option:checked').textContent()
    expect(selectedText).toContain('Round 3')
  })

  test('golden path: parse a post and reach the review phase', async ({ page }) => {
    // Select Ruiboys
    await page.locator('select').first().selectOption({ label: 'Ruiboys' })

    // Round 3 should already be selected; wait for "no match" warning to disappear
    await expect(page.locator('text=No match found')).not.toBeVisible()

    // Select team format
    await page.locator('select').nth(2).selectOption('Ruiboys')

    // Paste post
    await page.locator('textarea').fill(minimalRuiboysPost)

    // Parse button should be enabled
    const parseBtn = page.getByRole('button', { name: 'Parse' })
    await expect(parseBtn).toBeEnabled()
    await parseBtn.click()

    // Review phase
    await expect(page.getByRole('button', { name: '← Back' })).toBeVisible({ timeout: 15000 })
    await expect(page.getByText(/players parsed/)).toBeVisible()

    // Table structure
    await expect(page.getByRole('columnheader', { name: 'Posted' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Resolved' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Position' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Score' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Confidence' })).toBeVisible()
  })

  test('back button returns to input phase', async ({ page }) => {
    // Select Ruiboys, paste, parse
    await page.locator('select').first().selectOption({ label: 'Ruiboys' })
    await page.locator('select').nth(2).selectOption('Ruiboys')
    await page.locator('textarea').fill(minimalRuiboysPost)
    await page.getByRole('button', { name: 'Parse' }).click()

    await expect(page.getByRole('button', { name: '← Back' })).toBeVisible({ timeout: 15000 })
    await page.getByRole('button', { name: '← Back' }).click()

    await expect(page.locator('textarea')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Parse' })).toBeVisible()
  })
})
