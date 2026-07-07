import { ref } from 'vue'
import { getCookie, setCookie } from '@/utils/cookie'

export type StatSource = 'form' | 'season'

const COOKIE = 'xffl_stat_source'

// Module-level singleton — one switch shared by every stats surface, so
// flipping it on any page changes them all. Persisted across sessions.
const statSource = ref<StatSource>(getCookie(COOKIE) === 'season' ? 'season' : 'form')

function setStatSource(source: StatSource) {
  statSource.value = source
  setCookie(COOKIE, source, 365)
}

export function useStatSource() {
  return { statSource, setStatSource }
}
