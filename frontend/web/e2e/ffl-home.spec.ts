import { test, expect } from '@playwright/test'

test.describe('FFL Home', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  test('displays season and round name', async ({ page }) => {
    await expect(page.getByRole('heading', { level: 1 })).toContainText('2024 Season')
    await expect(page.getByRole('paragraph')).toContainText('Round 1')
  })

  test('displays ladder with clubs', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Ladder' })).toBeVisible()
    await expect(page.getByRole('cell', { name: 'Ruiboys' })).toBeVisible()
    await expect(page.getByRole('cell', { name: 'The Howling Cows' })).toBeVisible()
  })

  test('displays ladder with percentage column', async ({ page }) => {
    await expect(page.getByRole('columnheader', { name: '%' })).toBeVisible()
  })

  test('displays match summary', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Matches' })).toBeVisible()
    const matchLink = page.getByRole('link', { name: /Ruiboys.+v.+The Howling Cows/ })
    await expect(matchLink).toBeVisible()
  })

  test('match summary links to FFL match page', async ({ page }) => {
    const matchLink = page.getByRole('link', { name: /Ruiboys.+v.+The Howling Cows/ })
    await matchLink.click()
    await expect(page.getByRole('heading', { level: 1 })).toContainText('Ruiboys')
    await expect(page.getByRole('heading', { level: 1 })).toContainText('The Howling Cows')
  })

  test('displays round navigation', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Rounds' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Round 1' })).toBeVisible()
  })

  test('has Build Team button', async ({ page }) => {
    await expect(page.getByRole('link', { name: 'Build Team' })).toBeVisible()
  })

  test('navbar has FFL and AFL links', async ({ page }) => {
    await expect(page.getByRole('navigation').getByRole('link', { name: 'FFL', exact: true })).toBeVisible()
    await expect(page.getByRole('navigation').getByRole('link', { name: 'AFL', exact: true })).toBeVisible()
  })
})
