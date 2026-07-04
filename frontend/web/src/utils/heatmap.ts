// dark:  slate-500 → yellow-400
// light: slate-300 → purple-700
export function heatStyle(value: number, min: number, max: number, isDark: boolean): Record<string, string> {
  if (max === 0 || max === min) return {}
  const t = (value - min) / (max - min)
  const [lr, lg, lb] = isDark ? [100, 116, 139] : [203, 213, 225]
  const [hr, hg, hb] = isDark ? [250, 204, 21]  : [126, 34,  206]
  return { color: `rgb(${Math.round(lr + t * (hr - lr))},${Math.round(lg + t * (hg - lg))},${Math.round(lb + t * (hb - lb))})` }
}
