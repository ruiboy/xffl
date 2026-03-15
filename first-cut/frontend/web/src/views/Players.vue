<template>
  <div class="max-w-6xl mx-auto">
    <div class="bg-surface-0 dark:bg-surface-800 rounded-lg shadow-lg p-6">
      <div class="mb-6">
        <h1 class="text-surface-900 dark:text-surface-0 text-3xl font-bold mb-2">Player Management</h1>
        <p class="text-surface-600 dark:text-surface-400">Manage players across different clubs</p>
      </div>
      
      <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center mb-6 gap-4">
        <div class="w-full sm:w-72">
          <label class="block text-sm font-medium text-surface-700 dark:text-surface-300 mb-2">Select Club</label>
          <Dropdown
            v-model="selectedClub"
            :options="clubs"
            optionLabel="name"
            optionValue="id"
            placeholder="Select a club"
            :loading="loading.clubs"
            class="w-full"
          />
        </div>
        <Button
          @click="openCreateModal"
          icon="pi pi-plus"
          label="Add Player"
          :disabled="!selectedClub"
          severity="primary"
          class="w-full sm:w-auto"
        />
      </div>

      <div v-if="selectedClub">
        <DataTable
          :value="players"
          :loading="loading.players"
          stripedRows
          paginator
          :rows="10"
          :rowsPerPageOptions="[5, 10, 20]"
          responsiveLayout="scroll"
          class="p-datatable-sm"
        >
          <Column field="name" header="Name" sortable class="font-semibold"></Column>
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
          <Column header="Actions" class="w-32">
            <template #body="slotProps">
              <div class="flex gap-2">
                <Button
                  icon="pi pi-pencil"
                  severity="secondary"
                  outlined
                  size="small"
                  @click="openEditModal(slotProps.data)"
                />
                <Button
                  icon="pi pi-trash"
                  severity="danger"
                  outlined
                  size="small"
                  @click="handleDeletePlayer(slotProps.data.id)"
                />
              </div>
            </template>
          </Column>
        </DataTable>
      </div>
    </div>

    <Dialog
      v-model:visible="showModal"
      :header="editingPlayer ? 'Edit Player' : 'Add Player'"
      :modal="true"
      :style="{ width: '450px' }"
    >
      <div class="p-6">
        <div class="mb-4">
          <label for="playerName" class="block text-sm font-medium text-surface-700 dark:text-surface-300 mb-2">Name</label>
          <InputText
            id="playerName"
            v-model="playerName"
            placeholder="Enter player name"
            class="w-full"
            :invalid="submitted && !playerName"
          />
          <small class="text-red-500 text-sm mt-1 block" v-if="submitted && !playerName">Name is required.</small>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <Button
            label="Cancel"
            icon="pi pi-times"
            severity="secondary"
            outlined
            @click="closeModal"
          />
          <Button
            :label="editingPlayer ? 'Update' : 'Create'"
            icon="pi pi-check"
            severity="primary"
            @click="editingPlayer ? handleUpdatePlayer() : handleCreatePlayer()"
            :disabled="!playerName"
          />
        </div>
      </template>
    </Dialog>

    <ConfirmDialog></ConfirmDialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import gql from 'graphql-tag';
import Button from 'primevue/button';
import DataTable from 'primevue/datatable';
import Column from 'primevue/column';
import Dialog from 'primevue/dialog';
import InputText from 'primevue/inputtext';
import Dropdown from 'primevue/dropdown';
import ConfirmDialog from 'primevue/confirmdialog';
import { useConfirm } from 'primevue/useconfirm';

const GET_CLUBS = gql`
  query GetClubs {
    fflClubs {
      id
      name
    }
  }
`;

const GET_PLAYERS = gql`
  query GetPlayers($clubId: ID) {
    fflPlayers(clubId: $clubId) {
      id
      name
      createdAt
      updatedAt
      deletedAt
    }
  }
`;

const CREATE_PLAYER = gql`
  mutation CreatePlayer($input: CreateFFLPlayerInput!) {
    createFFLPlayer(input: $input) {
      id
      name
      createdAt
      updatedAt
    }
  }
`;

const UPDATE_PLAYER = gql`
  mutation UpdatePlayer($input: UpdateFFLPlayerInput!) {
    updateFFLPlayer(input: $input) {
      id
      name
      createdAt
      updatedAt
    }
  }
`;

const DELETE_PLAYER = gql`
  mutation DeletePlayer($id: ID!) {
    deleteFFLPlayer(id: $id)
  }
`;

interface Player {
  id: string;
  name: string;
  createdAt: string;
  updatedAt: string | null;
  deletedAt: string | null;
}

interface Club {
  id: string;
  name: string;
}

const selectedClub = ref('');
const editingPlayer = ref<Player | null>(null);
const playerName = ref('');
const showModal = ref(false);
const loading = ref({
  clubs: false,
  players: false,
});

const { result: clubsResult, loading: clubsLoading } = useQuery(GET_CLUBS);
const { result: playersResult, loading: playersLoading, refetch: refetchPlayers } = useQuery(
  GET_PLAYERS,
  () => ({
    clubId: selectedClub.value || undefined,
  }),
  () => ({
    enabled: !!selectedClub.value,
  })
);

const { mutate: createPlayer } = useMutation(CREATE_PLAYER);
const { mutate: updatePlayer } = useMutation(UPDATE_PLAYER);
const { mutate: deletePlayer } = useMutation(DELETE_PLAYER);

const clubs = computed(() => clubsResult.value?.fflClubs || []);
const players = computed(() => playersResult.value?.fflPlayers || []);

watch(clubsLoading, (newValue) => {
  loading.value.clubs = newValue;
});

watch(playersLoading, (newValue) => {
  loading.value.players = newValue;
});

const formatDate = (date: string | null) => {
  if (!date) return '-';
  try {
    // Remove the timezone offset part and parse the date
    const dateWithoutTz = date.split(' ').slice(0, 2).join(' ');
    const parsedDate = new Date(dateWithoutTz);
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

const openCreateModal = () => {
  editingPlayer.value = null;
  playerName.value = '';
  showModal.value = true;
};

const openEditModal = (player: Player) => {
  editingPlayer.value = player;
  playerName.value = player.name;
  showModal.value = true;
};

const closeModal = () => {
  showModal.value = false;
  editingPlayer.value = null;
  playerName.value = '';
};

const handleCreatePlayer = async () => {
  if (!selectedClub.value || !playerName.value) return;

  try {
    await createPlayer({
      input: {
        name: playerName.value,
        clubId: selectedClub.value,
      },
    });
    closeModal();
    refetchPlayers();
  } catch (error) {
    console.error('Error creating player:', error);
  }
};

const handleUpdatePlayer = async () => {
  if (!editingPlayer.value || !playerName.value) return;

  try {
    await updatePlayer({
      input: {
        id: editingPlayer.value.id,
        name: playerName.value,
      },
    });
    closeModal();
    refetchPlayers();
  } catch (error) {
    console.error('Error updating player:', error);
  }
};

const confirm = useConfirm();

const handleDeletePlayer = (id: string) => {
  confirm.require({
    message: 'Are you sure you want to delete this player?',
    header: 'Delete Confirmation',
    icon: 'pi pi-exclamation-triangle',
    acceptClass: 'p-button-danger',
    accept: () => {
      deletePlayer({ id });
    },
  });
};
const submitted = ref(false);
</script>
