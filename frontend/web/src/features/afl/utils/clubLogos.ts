const clubLogoMap: Record<string, string> = {
  'Adelaide Crows': 'Adelaide',
  'Brisbane Lions': 'Brisbane',
  'Carlton Blues': 'Carlton',
  'Collingwood Magpies': 'Collingwood',
  'Essendon Bombers': 'Essendon',
  'Fremantle Dockers': 'Fremantle',
  'Geelong Cats': 'Geelong',
  'Gold Coast Suns': 'GoldCoast',
  'Greater Western Sydney Giants': 'Giants',
  'Hawthorn Hawks': 'Hawthorn',
  'Melbourne Demons': 'Melbourne',
  'North Melbourne Kangaroos': 'NorthMelbourne',
  'Port Adelaide Power': 'PortAdelaide',
  'Richmond Tigers': 'Richmond',
  'St Kilda Saints': 'StKilda',
  'Sydney Swans': 'Sydney',
  'West Coast Eagles': 'WestCoast',
  'Western Bulldogs': 'Bulldogs',
}

export function clubLogoUrl(clubName: string): string {
  const file = clubLogoMap[clubName]
  return file ? `/images/clubs/${file}.png` : ''
}
