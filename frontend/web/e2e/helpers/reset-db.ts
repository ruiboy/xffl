import { execFile } from 'node:child_process'
import { promisify } from 'node:util'

const execFileP = promisify(execFile)

const CONTAINER = 'xffl-postgres-test'
const SEED_FILES = [
  '/docker-entrypoint-initdb.d/10_afl_seed.sql',
  '/docker-entrypoint-initdb.d/11_ffl_seed.sql',
]

// Resets the e2e Postgres to its seeded baseline. Truncates with RESTART IDENTITY
// and replays the seed files so IDs are stable across re-runs. Wired in via the
// Playwright fixture in e2e/fixtures.ts so every test gets isolation automatically.
export async function resetDb(): Promise<void> {
  const args = ['exec', '-i', CONTAINER, 'psql', '-U', 'postgres', '-d', 'xffl', '-v', 'ON_ERROR_STOP=1', '-q']
  for (const file of SEED_FILES) {
    args.push('-f', file)
  }
  try {
    await execFileP('docker', args)
  } catch (err) {
    const e = err as { stderr?: string; stdout?: string; message: string }
    const detail = (e.stderr || e.stdout || '').trim()
    throw new Error(`resetDb failed: ${e.message}${detail ? `\n${detail}` : ''}`)
  }
}
