// Reference scores are stashed in club match notes by the importers as
// "posted:NN" (forum) and "spreadsheet:NN" (fixture sheet). They are references,
// not the truth — the evaluated drv_score is. This parses whichever are present
// so the delta view can compare evaluated vs each reference.

export type ReferenceScores = {
  posted: number | null
  spreadsheet: number | null
}

function pick(notes: string, source: string): number | null {
  const m = notes.match(new RegExp(`${source}:(-?\\d+)`))
  return m ? parseInt(m[1], 10) : null
}

export function parseReferenceScores(notes: string | null | undefined): ReferenceScores {
  if (!notes) return { posted: null, spreadsheet: null }
  return { posted: pick(notes, 'posted'), spreadsheet: pick(notes, 'spreadsheet') }
}
