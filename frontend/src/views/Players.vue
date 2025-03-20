<template>
  <div class="container mx-auto p-8">
    <div class="flex justify-between items-center mb-6">
      <div class="w-72">
        <label class="block text-sm font-medium text-gray-700 mb-2">Select Club</label>
        <select
          v-model="selectedClub"
          class="w-full p-2 border rounded"
          :disabled="loading.clubs"
        >
          <option value="">Select a club</option>
          <option v-for="club in clubs" :key="club.id" :value="club.id">
            {{ club.name }}
          </option>
        </select>
      </div>
      <button
        @click="openCreateModal"
        class="bg-blue-500 text-white px-4 py-2 rounded"
        :disabled="!selectedClub"
      >
        Add Player
      </button>
    </div>

    <div v-if="selectedClub" class="overflow-x-auto">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created At</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Updated At</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-for="player in players" :key="player.id">
            <td class="px-6 py-4 whitespace-nowrap">{{ player.name }}</td>
            <td class="px-6 py-4 whitespace-nowrap">{{ formatDate(player.createdAt) }}</td>
            <td class="px-6 py-4 whitespace-nowrap">{{ formatDate(player.updatedAt) }}</td>
            <td class="px-6 py-4 whitespace-nowrap">
              <button
                @click="openEditModal(player)"
                class="text-blue-600 hover:text-blue-900 mr-2"
              >
                Edit
              </button>
              <button
                @click="handleDeletePlayer(player.id)"
                class="text-red-600 hover:text-red-900"
              >
                Delete
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full">
      <div class="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
        <div class="mt-3">
          <h3 class="text-lg font-medium leading-6 text-gray-900 mb-4">
            {{ editingPlayer ? 'Edit Player' : 'Add Player' }}
          </h3>
          <div class="mt-2">
            <label class="block text-sm font-medium text-gray-700 mb-2">Name</label>
            <input
              v-model="playerName"
              type="text"
              class="w-full p-2 border rounded"
              placeholder="Enter player name"
            />
          </div>
          <div class="mt-4 flex justify-end space-x-2">
            <button
              @click="closeModal"
              class="bg-gray-200 text-gray-800 px-4 py-2 rounded"
            >
              Cancel
            </button>
            <button
              @click="editingPlayer ? handleUpdatePlayer() : handleCreatePlayer()"
              class="bg-blue-500 text-white px-4 py-2 rounded"
              :disabled="!playerName"
            >
              {{ editingPlayer ? 'Update' : 'Create' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import gql from 'graphql-tag';

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

const handleDeletePlayer = async (id: string) => {
  if (!confirm('Are you sure you want to delete this player?')) return;

  try {
    await deletePlayer({ id });
    refetchPlayers();
  } catch (error) {
    console.error('Error deleting player:', error);
  }
};
</script> 