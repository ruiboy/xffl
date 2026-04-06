import { test, expect } from '@playwright/test'

test.describe('FFL Round', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
    // Click round circle "1" in the round selector (main nav)
    await page.locator('main nav').getByRole('link', { name: '1', exact: true }).click()
  })

  test('displays round and season name inline in heading', async ({ page }) => {
    const heading = page.getByRole('heading', { level: 1 })
    await expect(heading).toContainText('Round 1')
    await expect(heading).toContainText('FFL 2026')
  })

  test('displays round selector above matches', async ({ page }) => {
    const roundNav = page.locator('main nav')
    await expect(roundNav).toBeVisible()
    await expect(roundNav.getByTitle('Ladder')).toBeVisible()
    await expect(roundNav.getByRole('link', { name: '1', exact: true })).toBeVisible()
  })

  test('ladder icon navigates back to home', async ({ page }) => {
    await page.locator('main nav').getByTitle('Ladder').click()
    await expect(page).toHaveURL('/ffl')
  })

  test('displays match summaries', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Matches' })).toBeVisible()
    const matchLink = page.getByRole('link', { name: /Ruiboys.+v.+The Howling Cows/ })
    await expect(matchLink).toBeVisible()
  })

  test('displays top fantasy scorers', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Top Fantasy Scorers' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Player' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Score' })).toBeVisible()
  })

  test('shows Build Team link for selected club match', async ({ page }) => {
    await expect(page.getByRole('link', { name: 'Build Team →' })).toBeVisible()
  })

  test('Build Team link navigates to team builder', async ({ page }) => {
    await page.getByRole('link', { name: 'Build Team →' }).click()
    await expect(page).toHaveURL(/\/ffl\/.*\/team-builder/)
  })

  test('rounds display in numeric order', async ({ page }) => {
    const roundLinks = await page.locator('main nav').getByRole('link').allTextContents()
    const roundNumbers = roundLinks.map(t => parseInt(t.trim())).filter(n => !isNaN(n))
    expect(roundNumbers).toEqual([...roundNumbers].sort((a, b) => a - b))
  })
})
