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
        
        <Column field="percentage" header="%" class="text-center w-20">
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
import { ref, computed, onMounted } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import gql from 'graphql-tag'

interface Team {
  id: string
  clubName: string
  played: number
  won: number
  lost: number
  pointsFor: number
  pointsAgainst: number
  percentage: number
}

const FFL_LADDER_QUERY = gql`
  query FFLLadder($seasonId: ID!) {
    fflLadder(seasonId: $seasonId) {
      id
      clubName
      played
      won
      lost
      pointsFor
      pointsAgainst
      percentage
    }
  }
`

// Default to season ID 1 - you can make this configurable later
const seasonId = ref('1')

const { result, loading, error } = useQuery(FFL_LADDER_QUERY, {
  seasonId: seasonId.value
})

// Transform GraphQL result to local interface
const teams = computed(() => {
  if (!result.value?.fflLadder) return []
  
  return result.value.fflLadder.map((clubSeason: any) => ({
    id: clubSeason.id,
    name: clubSeason.clubName,
    played: clubSeason.played,
    won: clubSeason.won,
    lost: clubSeason.lost,
    pointsFor: clubSeason.pointsFor,
    pointsAgainst: clubSeason.pointsAgainst,
    percentage: clubSeason.percentage
  }))
})

// Teams are already sorted by the backend query
const sortedTeams = computed(() => teams.value)
</script>
