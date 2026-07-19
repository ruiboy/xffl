import { test, expect } from './fixtures'
import { setupFflSession, setupAflSession } from './helpers'

test.describe('Navigation — Phase 21', () => {
  test.describe('NAV-1: Header links navigate to live round', () => {
    test('FFL header link navigates to the live FFL round', async ({ page }) => {
      await setupFflSession(page)
      const topNav = page.getByRole('navigation').first()
      await topNav.getByRole('link', { name: 'FFL', exact: true }).filter({ hasText: 'FFL' }).click()
      await expect(page).toHaveURL(/\/ffl\/rounds\//)
    })

    test('AFL header link navigates to the live AFL round', async ({ page }) => {
      await setupFflSession(page)
      const topNav = page.getByRole('navigation').first()
      await topNav.getByRole('link', { name: 'AFL', exact: true }).filter({ hasText: 'AFL' }).click()
      await expect(page).toHaveURL(/\/afl\/rounds\//)
    })
  })

  test.describe('NAV-3: Ladder is the home; RoundNav has no ladder pill', () => {
    test('AFL home shows the ladder', async ({ page }) => {
      await setupAflSession(page)
      await expect(page).toHaveURL(/\/afl\/ladder$/)
      await expect(page.getByRole('heading', { name: 'Ladder' })).toBeVisible()
    })

    test('FFL home shows the ladder', async ({ page }) => {
      await setupFflSession(page)
      await expect(page).toHaveURL(/\/ffl\/ladder$/)
      await expect(page.getByRole('heading', { name: 'Ladder' })).toBeVisible()
    })

    test('RoundNav no longer has a ladder pill', async ({ page }) => {
      await setupAflSession(page)
      await expect(page.getByTitle('Ladder')).toHaveCount(0)
    })

    test('breadcrumb returns to the ladder from an AFL round', async ({ page }) => {
      await setupAflSession(page)
      await page.locator('main nav').last().getByRole('link', { name: '1', exact: true }).click()
      await page.waitForURL(/\/afl\/rounds\//)
      await page.getByRole('link', { name: 'AFL 2026', exact: true }).click()
      await expect(page).toHaveURL(/\/afl\/ladder$/)
      await expect(page.getByRole('heading', { name: 'Ladder' })).toBeVisible()
    })
  })

  // Data Ops moved out of the header and into the Settings menu when the nav was
  // decluttered — it is reachable from every route, but behind one click.
  test.describe('NAV-4: Data Ops is reachable from the Settings menu', () => {
    test('Data Ops is not a bare header icon', async ({ page }) => {
      await setupFflSession(page)
      await expect(page.getByRole('link', { name: 'Data Ops' })).not.toBeVisible()
    })

    test('Settings menu offers Data Ops on FFL home', async ({ page }) => {
      await setupFflSession(page)
      await page.getByTitle('Settings').click()
      await expect(page.getByRole('link', { name: 'Data Ops' })).toBeVisible()
    })

    test('Settings menu offers Data Ops on AFL home', async ({ page }) => {
      await setupAflSession(page)
      await page.getByTitle('Settings').click()
      await expect(page.getByRole('link', { name: 'Data Ops' })).toBeVisible()
    })

    test('Data Ops navigates to the data-ops page', async ({ page }) => {
      await setupFflSession(page)
      await page.getByTitle('Settings').click()
      await page.getByRole('link', { name: 'Data Ops' }).click()
      await expect(page).toHaveURL(/\/ffl\/data-ops/)
    })
  })

  test.describe('NAV-6: Free Agents icon visible on FFL routes only', () => {
    test('Free Agents icon is visible in header on FFL home', async ({ page }) => {
      await setupFflSession(page)
      await expect(page.getByTitle('Free Agents')).toBeVisible()
    })

    test('Free Agents icon is not visible on AFL routes', async ({ page }) => {
      await setupAflSession(page)
      await expect(page.getByTitle('Free Agents')).not.toBeVisible()
    })

    test('Free Agents icon href points to /ffl/free-agents', async ({ page }) => {
      await setupFflSession(page)
      await expect(page.getByTitle('Free Agents')).toHaveAttribute('href', /\/ffl\/free-agents/)
    })
  })

  test.describe('NAV-5: DataOps uses RoundNav instead of dropdowns', () => {
    test.beforeEach(async ({ page }) => {
      await setupFflSession(page)
      await page.goto('/ffl/data-ops')
    })

    test('FFL Teams tab shows round nav pills instead of a select', async ({ page }) => {
      await page.getByRole('button', { name: 'FFL Teams' }).click()
      await page.waitForLoadState('networkidle')
      // Should have no round select dropdown
      await expect(page.locator('select').filter({ hasText: /Round/ })).not.toBeAttached()
      // Should have round nav with numbered pills
      const roundNav = page.locator('main nav').last()
      await expect(roundNav.getByRole('link', { name: '1', exact: true })).toBeVisible()
    })

    test('AFL Stats tab shows round nav pills instead of a select', async ({ page }) => {
      await page.getByRole('button', { name: 'AFL Stats' }).click()
      await page.waitForLoadState('networkidle')
      const roundNav = page.locator('main nav').last()
      await expect(roundNav.getByRole('link', { name: '1', exact: true })).toBeVisible()
    })

    test('clicking a round pill in FFL Teams tab updates the displayed round', async ({ page }) => {
      await page.getByRole('button', { name: 'FFL Teams' }).click()
      await page.waitForLoadState('networkidle')
      // Click round 1
      await page.locator('main nav').last().getByRole('link', { name: '1', exact: true }).click()
      await page.waitForLoadState('networkidle')
      // Round 1 pill should now be active
      const round1 = page.locator('main nav').last().getByRole('link', { name: '1', exact: true })
      await expect(round1).toHaveClass(/bg-active/)
    })
  })
})
