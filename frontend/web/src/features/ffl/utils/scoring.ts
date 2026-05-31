export function isScoring(aflStatus: string | null): boolean {
  return aflStatus === 'played' || aflStatus === 'playing' || aflStatus === 'bye'
}
