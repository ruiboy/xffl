<template>
  <div class="overflow-x-auto">
    <table class="w-full text-sm">
      <thead>
        <tr class="border-b border-border text-left text-text-muted">
          <th class="py-2 pr-4 font-medium">Player</th>
          <th class="py-2 px-2"></th>
          <th class="py-2 px-2 font-medium">Status</th>
          <th class="py-2 px-2 font-medium text-right">Score</th>
        </tr>
      </thead>
      <tbody>
        <template v-for="(group, i) in starterGroups" :key="group.position">
          <tr>
            <td colspan="4" class="pb-1 text-xs font-semibold text-text-faint" :class="i === 0 ? 'pt-3' : 'pt-5'">
              <span>{{ group.label }}</span>
              <span class="pl-3">({{ group.total }})</span>
            </td>
          </tr>
          <tr
            v-for="pm in group.players"
            :key="pm.id"
            class="border-b border-border-subtle hover:bg-surface-hover"
          >
            <td class="py-2 pr-4">
              <div>
                <span class="font-medium">{{ pm.player.aflPlayer.name }}</span>
                <span v-if="pmAflClub(pm)" class="ml-2 text-xs text-text-muted">{{ pmAflClub(pm) }}</span>
              </div>
              <div v-if="coveringMap.get(pm.id)" class="text-xs text-sky-400">
                ↑ {{ coveringMap.get(pm.id)!.player.aflPlayer.name }}
              </div>
            </td>
            <td class="py-2 px-2"></td>
            <td class="py-2 px-2"><StatusBadge :status="pmStatus(pm)" /></td>
            <td class="py-2 px-2 text-right tabular-nums font-semibold">
              {{ pmShowScore(pm) ? pm.score : (coveringScoreForStarter(pm) ?? '') }}
            </td>
          </tr>
        </template>

        <template v-if="bench.length > 0">
          <tr>
            <td colspan="4" class="pt-5 pb-2 text-xs font-semibold text-text-faint">Bench</td>
          </tr>
          <tr
            v-for="pm in bench"
            :key="pm.id"
            class="border-b border-border-subtle hover:bg-surface-hover"
          >
            <td class="py-2 pr-4">
              <div v-if="coveredStarterMap.get(pm.id)" class="text-xs text-sky-400">
                ↑ {{ coveredStarterMap.get(pm.id)!.player.aflPlayer.name }}
              </div>
              <span class="font-medium text-text-muted">{{ pm.player.aflPlayer.name }}</span>
              <span v-if="pmAflClub(pm)" class="ml-2 text-xs text-text-muted">{{ pmAflClub(pm) }}</span>
            </td>
            <td class="py-2 px-2">
              <div class="flex items-center gap-1">
                <span
                  v-for="pos in pmBackupPositions(pm)"
                  :key="pos"
                  :class="pm.interchangePosition === pos
                    ? 'text-xs rounded px-1.5 py-0.5 bg-sky-500/10 text-sky-400 underline'
                    : coveredStarterMap.get(pm.id)?.position === pos
                      ? 'text-xs bg-control rounded px-1.5 py-0.5 text-text-muted ring-1 ring-sky-400/60'
                      : 'text-xs bg-control rounded px-1.5 py-0.5 text-text-muted'"
                >{{ positionShort(pos) }}</span>
              </div>
            </td>
            <td class="py-2 px-2"><StatusBadge :status="pmStatus(pm)" /></td>
            <td class="py-2 px-2 text-right tabular-nums">
              {{ benchScoreDisplay(pm) }}
            </td>
          </tr>
        </template>
      </tbody>
      <tfoot>
        <tr class="border-t border-border font-semibold">
          <td class="py-2 pr-4">Total</td>
          <td></td>
          <td></td>
          <td class="py-2 px-2 text-right tabular-nums">{{ total }}</td>
        </tr>
      </tfoot>
    </table>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import StatusBadge from './StatusBadge.vue'

interface PlayerMatch {
  id: string
  player: { aflPlayer: { name: string } }
  position: string | null
  status: string | null
  aflStatus: string | null
  backupPositions: string | null
  interchangePosition: string | null
  score: number
  playerSeason?: {
    aflPlayerSeason?: {
      clubSeason?: { club?: { name: string } | null } | null
    } | null
  } | null
  aflPlayerMatch?: {
    goals: number; kicks: number; handballs: number
    marks: number; tackles: number; hitouts: number
  } | null
}

const props = defineProps<{
  playerMatches: PlayerMatch[]
}>()

const POSITION_ORDER = ['goals', 'kicks', 'handballs', 'marks', 'tackles', 'hitouts', 'star'] as const
type PositionKey = typeof POSITION_ORDER[number]

const POSITION_LABELS: Record<PositionKey, string> = {
  goals:     'Goals',
  kicks:     'Kicks',
  handballs: 'Handballs',
  marks:     'Marks',
  tackles:   'Tackles',
  hitouts:   'Hitouts',
  star:      'Star',
}

const isBench = (pm: PlayerMatch) => pm.backupPositions != null || pm.interchangePosition != null
const starters = computed(() => props.playerMatches.filter(pm => !isBench(pm)))
const bench    = computed(() => props.playerMatches.filter(pm => isBench(pm)))

// For each subbed-out starter, find the first bench player whose backupPositions covers that starter's position.
const coveringMap = computed(() => {
  const map = new Map<string, PlayerMatch>() // starter pm.id → covering bench PlayerMatch
  for (const starter of starters.value) {
    if (starter.status !== 'subbed' || !starter.position) continue
    const covering = bench.value.find(bp => {
      if (!bp.backupPositions) return false
      const bps = bp.backupPositions === 'star' ? ['star'] : bp.backupPositions.split(',').map(p => p.trim())
      return bps.includes(starter.position!)
    })
    if (covering) map.set(starter.id, covering)
  }
  return map
})

// Maps bench pm.id → the starter PlayerMatch they are subbing for.
const coveredStarterMap = computed(() => {
  const map = new Map<string, PlayerMatch>()
  for (const [starterId, cp] of coveringMap.value) {
    const starter = starters.value.find(s => s.id === starterId)
    if (starter) map.set(cp.id, starter)
  }
  return map
})

function coveringScoreForStarter(pm: PlayerMatch): number | null {
  const cp = coveringMap.value.get(pm.id)
  if (!cp || !pm.position) return null
  return benchPositionScore(cp, pm.position)
}

function pmAflClub(pm: PlayerMatch): string | null {
  return pm.playerSeason?.aflPlayerSeason?.clubSeason?.club?.name ?? null
}

function pmStatus(pm: PlayerMatch): string | null {
  if (pm.status === 'subbed') return 'subbed'
  if (pm.status === 'interchanged') return 'interchanged'
  return pm.aflStatus
}

function pmShowScore(pm: PlayerMatch): boolean {
  return pm.aflStatus === 'played' || pm.aflStatus === 'playing'
}

const starterGroups = computed(() => {
  const grouped: Partial<Record<string, PlayerMatch[]>> = {}
  for (const pm of starters.value) {
    const key = pm.position ?? 'unknown'
    ;(grouped[key] ??= []).push(pm)
  }
  return POSITION_ORDER
    .filter(pos => grouped[pos]?.length)
    .map(pos => ({
      position: pos,
      label: POSITION_LABELS[pos],
      players: grouped[pos]!,
      total: grouped[pos]!.reduce((sum, pm) => sum + (pmShowScore(pm) ? pm.score : 0), 0),
    }))
})

const POSITION_LETTERS: Record<string, string> = {
  goals: 'G', kicks: 'K', handballs: 'H', marks: 'M', tackles: 'T', hitouts: 'R', star: '★',
}

function positionShort(key: string): string {
  return POSITION_LETTERS[key] ?? key
}

function pmBackupPositions(pm: PlayerMatch): string[] {
  if (!pm.backupPositions) return []
  return pm.backupPositions === 'star' ? ['star'] : pm.backupPositions.split(',').map(p => p.trim())
}

const POSITION_MULTIPLIERS: Record<string, number> = {
  goals: 5, kicks: 1, handballs: 1, marks: 2, tackles: 4, hitouts: 1, star: 1,
}

function benchPositionScore(pm: PlayerMatch, pos: string): number | null {
  const s = pm.aflPlayerMatch
  if (!s) return null
  if (pos === 'star') {
    return s.goals * 5 + s.kicks + s.handballs + s.marks * 2 + s.tackles * 4
  }
  const stat = ({ goals: s.goals, kicks: s.kicks, handballs: s.handballs, marks: s.marks, tackles: s.tackles, hitouts: s.hitouts } as Record<string, number>)[pos] ?? null
  return stat !== null ? stat * (POSITION_MULTIPLIERS[pos] ?? 1) : null
}

function benchScoreDisplay(pm: PlayerMatch): string {
  if (!pmShowScore(pm) || !pm.backupPositions) return ''
  const positions = pm.backupPositions === 'star' ? ['star'] : pm.backupPositions.split(',').map(p => p.trim())
  return positions.map(pos => { const s = benchPositionScore(pm, pos); return s !== null ? String(s) : '?' }).join('/')
}

const total = computed(() =>
  props.playerMatches.reduce((sum, pm) => sum + (pmShowScore(pm) ? pm.score : 0), 0)
)
</script>
