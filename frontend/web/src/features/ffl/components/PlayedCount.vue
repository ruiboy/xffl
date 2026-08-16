<!--
  The shared "on-field players who have played" indicator, e.g. (8), with a ★ when
  the team's star is among them. Rendered next to a team's score everywhere it
  appears (round row, match header, readonly club-match summary). Shows nothing once
  the side is final or when there's nothing to count. The full "8 of 18 played"
  (", including star") is on hover; the /18 is dropped inline so it doesn't compete
  with the score. Font size / spacing are inherited from the parent.
-->
<template>
  <span
    v-if="count"
    class="font-normal text-text-faint tabular-nums"
    :title="title"
  >({{ count }}/{{ TEAM_SIZE }}<span v-if="star" aria-hidden="true">&nbsp;★</span>)</span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { progressCount, hasStar, TEAM_SIZE, type CountableClubMatch } from '../utils/teamCount'

const props = defineProps<{ clubMatch: CountableClubMatch | null | undefined }>()

const count = computed(() => progressCount(props.clubMatch))
const star = computed(() => hasStar(props.clubMatch))
const title = computed(() => `${count.value} of ${TEAM_SIZE} played${star.value ? ', including star' : ''}`)
</script>
