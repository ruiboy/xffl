<template>
  <div class="max-w-6xl mx-auto">
    <div class="bg-surface-0 dark:bg-surface-800 rounded-lg shadow-lg p-6">
      <div class="mb-6">
        <h1 class="text-surface-900 dark:text-surface-0 text-3xl font-bold mb-2">AFL Clubs</h1>
        <p class="text-surface-600 dark:text-surface-400">Australian Football League clubs and teams</p>
      </div>
      
      <DataTable
        :value="clubs"
        :loading="loading"
        stripedRows
        paginator
        :rows="10"
        :rowsPerPageOptions="[5, 10, 20]"
        responsiveLayout="scroll"
        class="p-datatable-sm"
      >
        <Column field="name" header="Club Name" sortable class="font-semibold">
          <template #body="slotProps">
            <div class="flex items-center">
              <span class="text-surface-900 dark:text-surface-0">{{ slotProps.data.name }}</span>
            </div>
          </template>
        </Column>
        <Column field="id" header="ID" sortable>
          <template #body="slotProps">
            <Tag 
              :value="slotProps.data.id" 
              severity="info" 
              class="text-sm font-semibold"
            />
          </template>
        </Column>
        <Column field="createdAt" header="Created At" sortable>
          <template #body="slotProps">
            <span class="text-surface-600 dark:text-surface-400">{{ formatDate(slotProps.data.createdAt) }}</span>
          </template>
        </Column>
        <Column field="updatedAt" header="Updated At" sortable>
          <template #body="slotProps">
            <span class="text-surface-600 dark:text-surface-400">{{ formatDate(slotProps.data.updatedAt) }}</span>
          </template>
        </Column>
      </DataTable>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useQuery } from '@vue/apollo-composable';
import gql from 'graphql-tag';
import DataTable from 'primevue/datatable';
import Column from 'primevue/column';
import Tag from 'primevue/tag';

const GET_AFL_CLUBS = gql`
  query GetAFLClubs {
    aflClubs {
      id
      name
      createdAt
      updatedAt
    }
  }
`;

interface AFLClub {
  id: string;
  name: string;
  createdAt: string;
  updatedAt: string;
}

const { result, loading } = useQuery(GET_AFL_CLUBS);

const clubs = computed(() => result.value?.aflClubs || []);

const formatDate = (date: string | null) => {
  if (!date) return '-';
  try {
    const parsedDate = new Date(date);
    if (isNaN(parsedDate.getTime())) {
      console.error('Invalid date:', date);
      return '-';
    }
    return parsedDate.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  } catch (error) {
    console.error('Error formatting date:', date, error);
    return '-';
  }
};
</script>