<template>
  <div class="max-w-6xl mx-auto">
    <div class="text-center mb-8">
      <h1 class="text-surface-900 dark:text-surface-0 text-4xl font-bold mb-2">League Ladder</h1>
      <p class="text-surface-600 dark:text-surface-400">Current team standings and statistics</p>
    </div>
    
    <div class="bg-surface-0 dark:bg-surface-800 rounded-lg shadow-lg">
      <DataTable 
        :value="sortedTeams" 
        :paginator="false"
        responsiveLayout="scroll"
        class="p-datatable-sm"
        stripedRows
      >
        <Column field="position" header="Pos" class="text-center font-bold w-16">
          <template #body="{ index }">
            <div class="text-primary-500 font-bold">{{ index + 1 }}</div>
          </template>
        </Column>
        
        <Column field="name" header="Team" class="font-semibold min-w-48">
          <template #body="{ data }">
            <div class="font-semibold text-surface-900 dark:text-surface-0">{{ data.name }}</div>
          </template>
        </Column>
        
        <Column field="played" header="P" class="text-center w-12">
          <template #body="{ data }">
            <div class="font-mono">{{ data.played }}</div>
          </template>
        </Column>
        
        <Column field="won" header="W" class="text-center w-12">
          <template #body="{ data }">
            <div class="font-mono text-green-600 dark:text-green-400">{{ data.won }}</div>
          </template>
        </Column>
        
        <Column field="lost" header="L" class="text-center w-12">
          <template #body="{ data }">
            <div class="font-mono text-red-600 dark:text-red-400">{{ data.lost }}</div>
          </template>
        </Column>
        
        <Column field="percentage" header="Win %" class="text-center w-20">
          <template #body="{ data }">
            <div class="font-mono font-bold text-primary-500">{{ data.percentage.toFixed(1) }}%</div>
          </template>
        </Column>
        
        <Column field="pointsFor" header="PF" class="text-center w-20 hidden md:table-cell">
          <template #body="{ data }">
            <div class="font-mono text-green-600 dark:text-green-400">{{ data.pointsFor }}</div>
          </template>
        </Column>
        
        <Column field="pointsAgainst" header="PA" class="text-center w-20 hidden md:table-cell">
          <template #body="{ data }">
            <div class="font-mono text-red-600 dark:text-red-400">{{ data.pointsAgainst }}</div>
          </template>
        </Column>
      </DataTable>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

interface Team {
  id: number
  name: string
  played: number
  won: number
  lost: number
  pointsFor: number
  pointsAgainst: number
  percentage: number
}

// Mock data for teams
const teams = ref<Team[]>([
  {
    id: 1,
    name: "Thunder Bolts",
    played: 8,
    won: 7,
    lost: 1,
    pointsFor: 1284,
    pointsAgainst: 956,
    percentage: 87.5
  },
  {
    id: 2,
    name: "Fire Dragons",
    played: 8,
    won: 6,
    lost: 2,
    pointsFor: 1198,
    pointsAgainst: 1023,
    percentage: 75.0
  },
  {
    id: 3,
    name: "Steel Eagles",
    played: 8,
    won: 5,
    lost: 3,
    pointsFor: 1145,
    pointsAgainst: 1087,
    percentage: 62.5
  },
  {
    id: 4,
    name: "Night Hawks",
    played: 8,
    won: 5,
    lost: 3,
    pointsFor: 1134,
    pointsAgainst: 1098,
    percentage: 62.5
  },
  {
    id: 5,
    name: "Storm Riders",
    played: 7,
    won: 4,
    lost: 3,
    pointsFor: 987,
    pointsAgainst: 934,
    percentage: 57.1
  },
  {
    id: 6,
    name: "Ice Wolves",
    played: 8,
    won: 4,
    lost: 4,
    pointsFor: 1067,
    pointsAgainst: 1145,
    percentage: 50.0
  },
  {
    id: 7,
    name: "Lightning Cats",
    played: 7,
    won: 3,
    lost: 4,
    pointsFor: 845,
    pointsAgainst: 987,
    percentage: 42.9
  },
  {
    id: 8,
    name: "Shadow Panthers",
    played: 8,
    won: 2,
    lost: 6,
    pointsFor: 934,
    pointsAgainst: 1234,
    percentage: 25.0
  },
  {
    id: 9,
    name: "Frost Giants",
    played: 8,
    won: 1,
    lost: 7,
    pointsFor: 798,
    pointsAgainst: 1345,
    percentage: 12.5
  },
  {
    id: 10,
    name: "Wind Runners",
    played: 6,
    won: 1,
    lost: 5,
    pointsFor: 623,
    pointsAgainst: 856,
    percentage: 16.7
  }
])

// Sort teams by percentage (descending), then by points for (descending)
const sortedTeams = computed(() => {
  return [...teams.value].sort((a, b) => {
    if (a.percentage !== b.percentage) {
      return b.percentage - a.percentage
    }
    return b.pointsFor - a.pointsFor
  })
})
</script>
