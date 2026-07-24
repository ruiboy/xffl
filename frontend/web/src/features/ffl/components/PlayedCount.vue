<!--
  The shared "on-field players who have played" indicator, e.g. (8). Rendered next
  to a team's score everywhere it appears (round row, match header, readonly
  club-match summary). Shows nothing once the side is final or when there's nothing
  to count. The full "8 of 18 played" is on hover; the /18 is dropped inline so it
  doesn't compete with the score. Font size / spacing are inherited from the parent.
-->
<template>
  <span
    v-if="count"
    class="font-normal text-text-faint tabular-nums"
    :title="`${count} of ${TEAM_SIZE} played`"
  >({{ count }})</span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { progressCount, TEAM_SIZE, type CountableClubMatch } from '../utils/teamCount'

const props = defineProps<{ clubMatch: CountableClubMatch | null | undefined }>()

const count = computed(() => progressCount(props.clubMatch))
</script>
