import { test as base } from '@playwright/test'
import { resetDb } from './helpers/reset-db'

// Auto fixture: runs before every test in every spec that imports `test` from
// this module, giving each test a fresh seeded DB. Pattern: replace
// `import { test, expect } from '@playwright/test'` with
// `import { test, expect } from './fixtures'`.
export const test = base.extend<{ resetSeed: void }>({
  resetSeed: [
    async ({}, use) => {
      await resetDb()
      await use()
    },
    { auto: true },
  ],
})

export { expect } from '@playwright/test'
