#!/usr/bin/env python3
"""
resolve_squads.py — Match 2026_squads.txt players to afl.player records.

Usage:
  python resolve_squads.py                       # resolve → squads_review.csv
  python resolve_squads.py --sql squads_review.csv  # generate SQL to stdout
"""

import argparse
import csv
import difflib
import re
import sys
import psycopg2

DB_DSN = "host=localhost port=5432 dbname=xffl user=postgres password=postgres"

CLUB_CODE = {
    'Adel': 'Adelaide Crows',
    'Bris': 'Brisbane Lions',
    'Carl': 'Carlton Blues',
    'Col':  'Collingwood Magpies',
    'Coll': 'Collingwood Magpies',
    'Ess':  'Essendon Bombers',
    'Freo': 'Fremantle Dockers',
    'Geel': 'Geelong Cats',
    'GC':   'Gold Coast Suns',
    'GWS':  'Greater Western Sydney Giants',
    'Haw':  'Hawthorn Hawks',
    'Melb': 'Melbourne Demons',
    'NM':   'North Melbourne Kangaroos',
    'Port': 'Port Adelaide Power',
    'PA':   'Port Adelaide Power',
    'Rich': 'Richmond Tigers',
    'SK':   'St Kilda Saints',
    'StK':  'St Kilda Saints',
    'Syd':  'Sydney Swans',
    'WB':   'Western Bulldogs',
    'WC':   'West Coast Eagles',
}

FFL_SECTION = {
    'CHEETAHS': 'Cheetahs',
    'THC':      'The Howling Cows',
    'RUIBOYS':  'Ruiboys',
    'SLASHERS': 'Slashers',
}

PLAYER_LINE = re.compile(
    r'^(.+?)\s+[\d.]+\s+([A-Za-z]+)\s*$'
)


def parse_squads(path: str) -> list[dict]:
    players = []
    current_club = None
    with open(path) as f:
        for raw in f:
            line = raw.strip()
            if not line:
                continue
            if line.upper() in FFL_SECTION:
                current_club = FFL_SECTION[line.upper()]
                continue
            m = PLAYER_LINE.match(line)
            if m and current_club:
                name = m.group(1).strip()
                code = m.group(2).strip()
                afl_club = CLUB_CODE.get(code)
                players.append({
                    'ffl_club': current_club,
                    'squad_name': name,
                    'afl_club_code': code,
                    'afl_club': afl_club or f'UNKNOWN:{code}',
                })
    return players


def fetch_afl_players(conn) -> list[dict]:
    """All players with a 2026 AFL season, with their club name."""
    with conn.cursor() as cur:
        cur.execute("""
            SELECT ap.id, ap.name, ac.name AS club
            FROM afl.player ap
            JOIN afl.player_season aps ON aps.player_id = ap.id
            JOIN afl.club_season acs ON aps.club_season_id = acs.id
            JOIN afl.club ac ON acs.club_id = ac.id
            JOIN afl.season asn ON acs.season_id = asn.id
            WHERE asn.name = 'AFL 2026'
        """)
        return [{'id': r[0], 'name': r[1], 'club': r[2]} for r in cur.fetchall()]


def resolve(squad_player: dict, afl_players: list[dict]) -> dict:
    name = squad_player['squad_name']
    want_club = squad_player['afl_club']

    # 1. Exact name + club
    for p in afl_players:
        if p['name'] == name and p['club'] == want_club:
            return {**squad_player, 'matched_name': p['name'], 'afl_player_id': p['id'], 'confidence': 'exact'}

    # 2. Exact name, any club
    for p in afl_players:
        if p['name'] == name:
            return {**squad_player, 'matched_name': p['name'], 'afl_player_id': p['id'], 'confidence': 'club_mismatch'}

    # 3. Fuzzy name within same club
    club_players = [p for p in afl_players if p['club'] == want_club]
    names = [p['name'] for p in club_players]
    matches = difflib.get_close_matches(name, names, n=1, cutoff=0.6)
    if matches:
        p = next(p for p in club_players if p['name'] == matches[0])
        return {**squad_player, 'matched_name': p['name'], 'afl_player_id': p['id'], 'confidence': 'fuzzy'}

    # 4. Fuzzy name across all clubs
    all_names = [p['name'] for p in afl_players]
    matches = difflib.get_close_matches(name, all_names, n=1, cutoff=0.6)
    if matches:
        p = next(p for p in afl_players if p['name'] == matches[0])
        return {**squad_player, 'matched_name': p['name'], 'afl_player_id': p['id'], 'confidence': 'fuzzy_any_club'}

    return {**squad_player, 'matched_name': '', 'afl_player_id': '', 'confidence': 'NO_MATCH'}


CSV_FIELDS = ['ffl_club', 'squad_name', 'afl_club_code', 'afl_club', 'matched_name', 'afl_player_id', 'confidence']


def cmd_resolve(squads_path: str, out_path: str):
    players = parse_squads(squads_path)
    conn = psycopg2.connect(DB_DSN)
    afl_players = fetch_afl_players(conn)
    conn.close()

    results = [resolve(p, afl_players) for p in players]

    with open(out_path, 'w', newline='') as f:
        w = csv.DictWriter(f, fieldnames=CSV_FIELDS)
        w.writeheader()
        w.writerows(results)

    no_match = [r for r in results if r['confidence'] == 'NO_MATCH']
    fuzzy = [r for r in results if 'fuzzy' in r['confidence']]

    print(f"Resolved {len(results)} players → {out_path}")
    print(f"  exact:          {sum(1 for r in results if r['confidence'] == 'exact')}")
    print(f"  club_mismatch:  {sum(1 for r in results if r['confidence'] == 'club_mismatch')}")
    print(f"  fuzzy:          {len(fuzzy)}")
    print(f"  NO_MATCH:       {len(no_match)}")
    if fuzzy:
        print("\nFuzzy matches (verify):")
        for r in fuzzy:
            print(f"  [{r['ffl_club']}] '{r['squad_name']}' → '{r['matched_name']}' ({r['confidence']})")
    if no_match:
        print("\nNo matches (fix manually in CSV):")
        for r in no_match:
            print(f"  [{r['ffl_club']}] '{r['squad_name']}' @ {r['afl_club_code']}")


def cmd_sql(review_path: str):
    with open(review_path) as f:
        rows = list(csv.DictReader(f))

    bad = [r for r in rows if not r['matched_name'] or r['confidence'] == 'NO_MATCH']
    if bad:
        print("ERROR: unresolved players in CSV:", file=sys.stderr)
        for r in bad:
            print(f"  {r['ffl_club']} / {r['squad_name']}", file=sys.stderr)
        sys.exit(1)

    # Emit SQL — all lookups by name, no hardcoded IDs
    print("-- Auto-generated by resolve_squads.py -- do not edit manually")
    print("-- Source: dev/import/ffl/squads_review.csv")
    print("BEGIN;")
    print()

    clubs_seen: dict[str, list] = {}
    for r in rows:
        clubs_seen.setdefault(r['ffl_club'], []).append(r)

    for ffl_club, players in clubs_seen.items():
        fc = ffl_club.replace("'", "''")
        print(f"-- {ffl_club}")
        print()

        # ffl.player: look up afl.player by name — no hardcoded afl_player_id
        for p in players:
            name = p['matched_name'].replace("'", "''")
            print(f"INSERT INTO ffl.player (afl_player_id, drv_name)")
            print(f"SELECT ap.id, '{name}' FROM afl.player ap WHERE ap.name = '{name}'")
            print(f"AND NOT EXISTS (SELECT 1 FROM ffl.player WHERE afl_player_id = ap.id);")

        print()

        # ffl.player_season: join by player name — no hardcoded player_season_id
        print(f"INSERT INTO ffl.player_season (player_id, club_season_id, from_round_id, afl_player_season_id)")
        print(f"SELECT")
        print(f"    fp.id,")
        print(f"    (SELECT cs.id FROM ffl.club_season cs JOIN ffl.club c ON cs.club_id = c.id JOIN ffl.season s ON cs.season_id = s.id WHERE c.name = '{fc}' AND s.name = 'FFL 2026'),")
        print(f"    (SELECT r.id FROM ffl.round r JOIN ffl.season s ON r.season_id = s.id WHERE r.name = 'Round 1' AND s.name = 'FFL 2026'),")
        print(f"    (SELECT aps.id FROM afl.player_season aps JOIN afl.club_season acs ON aps.club_season_id = acs.id JOIN afl.season asn ON acs.season_id = asn.id WHERE aps.player_id = fp.afl_player_id AND asn.name = 'AFL 2026')")
        print(f"FROM (VALUES")

        for i, p in enumerate(players):
            name = p['matched_name'].replace("'", "''")
            comma = '' if i == len(players) - 1 else ','
            print(f"    ('{name}'){comma}")

        print(f") AS v(afl_player_name)")
        print(f"JOIN afl.player ap ON ap.name = v.afl_player_name")
        print(f"JOIN ffl.player fp ON fp.afl_player_id = ap.id")
        print(f"ON CONFLICT (player_id, club_season_id) DO NOTHING;")
        print()

    print("COMMIT;")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--sql', metavar='review.csv', help='Generate SQL from reviewed CSV')
    parser.add_argument('--squads', default='2026_squads.txt')
    parser.add_argument('--out', default='squads_review.csv')
    args = parser.parse_args()

    if args.sql:
        cmd_sql(args.sql)
    else:
        cmd_resolve(args.squads, args.out)


if __name__ == '__main__':
    main()
