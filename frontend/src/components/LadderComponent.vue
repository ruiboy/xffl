<template>
  <div class="ladder">
    <h1>League Ladder</h1>
    <div class="ladder-table">
      <div class="table-header">
        <div class="pos">Pos</div>
        <div class="team">Team</div>
        <div class="played">P</div>
        <div class="won">W</div>
        <div class="lost">L</div>
        <div class="percentage">%</div>
        <div class="points-for">PF</div>
        <div class="points-against">PA</div>
      </div>
      <div 
        v-for="(team, index) in sortedTeams" 
        :key="team.id"
        class="table-row"
        :class="{ 'even': index % 2 === 0 }"
      >
        <div class="pos">{{ index + 1 }}</div>
        <div class="team">{{ team.name }}</div>
        <div class="played">{{ team.played }}</div>
        <div class="won">{{ team.won }}</div>
        <div class="lost">{{ team.lost }}</div>
        <div class="percentage">{{ team.percentage.toFixed(1) }}%</div>
        <div class="points-for">{{ team.pointsFor }}</div>
        <div class="points-against">{{ team.pointsAgainst }}</div>
      </div>
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

<style scoped>
.ladder {
  max-width: 1000px;
  margin: 0 auto;
}

h1 {
  text-align: center;
  color: #2c3e50;
  margin-bottom: 2rem;
}

.ladder-table {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.table-header,
.table-row {
  display: grid;
  grid-template-columns: 50px 2fr 50px 50px 50px 80px 80px 80px;
  align-items: center;
  padding: 12px 16px;
  gap: 8px;
}

.table-header {
  background-color: #f8f9fa;
  font-weight: bold;
  color: #2c3e50;
  border-bottom: 2px solid #e9ecef;
}

.table-row {
  border-bottom: 1px solid #e9ecef;
  transition: background-color 0.2s;
}

.table-row:hover {
  background-color: #f8f9fa;
}

.table-row.even {
  background-color: #fdfdfd;
}

.table-row:last-child {
  border-bottom: none;
}

.pos {
  text-align: center;
  font-weight: bold;
  color: #42b983;
}

.team {
  font-weight: 600;
  color: #2c3e50;
}

.played,
.won,
.lost,
.percentage,
.points-for,
.points-against {
  text-align: center;
  font-family: 'Courier New', monospace;
}

.percentage {
  font-weight: bold;
  color: #42b983;
}

.points-for {
  color: #28a745;
}

.points-against {
  color: #dc3545;
}

@media (max-width: 768px) {
  .table-header,
  .table-row {
    grid-template-columns: 40px 1fr 40px 40px 40px 60px;
    font-size: 14px;
  }
  
  .points-for,
  .points-against {
    display: none;
  }
}
</style>