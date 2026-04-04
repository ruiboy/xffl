const clubLogoMap: Record<string, string> = {
  'Ruiboys': 'ruiboys',
  'The Howling Cows': 'thc',
  'Slashers': 'slashers',
  'Cheetahs': 'cheetahs',
}

export function clubLogoUrl(clubName: string): string {
  const file = clubLogoMap[clubName]
  return file ? `/images/ffl-clubs/${file}.png` : ''
}
