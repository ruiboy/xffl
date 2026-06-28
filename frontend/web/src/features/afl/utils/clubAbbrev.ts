const clubAbbrevMap: Record<string, string> = {
  'Adelaide Crows':                  'Adel',
  'Brisbane Lions':                  'Bris',
  'Carlton Blues':                   'Carl',
  'Collingwood Magpies':             'Coll',
  'Essendon Bombers':                'Ess',
  'Fremantle Dockers':               'Fre',
  'Geelong Cats':                    'Geel',
  'Gold Coast Suns':                 'GCS',
  'Greater Western Sydney Giants':   'GWS',
  'Hawthorn Hawks':                  'Haw',
  'Melbourne Demons':                'Melb',
  'North Melbourne Kangaroos':       'NM',
  'Port Adelaide Power':             'Port',
  'Richmond Tigers':                 'Rich',
  'St Kilda Saints':                 'SK',
  'Sydney Swans':                    'Syd',
  'West Coast Eagles':               'WCE',
  'Western Bulldogs':                'WB',
}

export function clubAbbrev(name: string | null | undefined): string {
  if (!name) return ''
  return clubAbbrevMap[name] ?? name
}
