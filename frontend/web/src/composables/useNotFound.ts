import { computed, type ComputedRef, type Ref } from 'vue'

/**
 * Derives the "no such entity" state from a single-entity GraphQL query.
 *
 * The FFL subgraph returns null (not an error) when an id doesn't resolve, so
 * absence is only meaningful once the query has settled without failing.
 */
export function useNotFound(
  entity: ComputedRef<unknown> | Ref<unknown>,
  loading: Ref<boolean>,
  error: Ref<unknown>,
): ComputedRef<boolean> {
  return computed(() => !loading.value && !error.value && entity.value == null)
}
