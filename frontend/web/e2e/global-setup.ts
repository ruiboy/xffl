import { execSync } from 'child_process'

export default async function globalSetup() {
  // Reseed the DB to a known clean state before the test suite runs.
  // `just test-e2e` runs from frontend/web; '../../' reaches the project root.
  execSync('just dev-seed', { cwd: '../../', stdio: 'inherit' })
}
