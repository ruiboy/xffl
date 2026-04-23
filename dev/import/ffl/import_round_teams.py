#!/usr/bin/env python3
"""
import_round_teams.py — Generate SQL from parsed round team CSVs.

Usage:
  python3 import_round_teams.py            # resolve + generate SQL → stdout
  python3 import_round_teams.py --apply    # resolve + apply directly to DB
"""

import argparse
import csv
import difflib
import re
import sys
import psycopg2

DB_DSN = "host=localhost port=5432 dbname=xffl user=postgres password=postgres"

ROUNDS = [1, 2, 3, 4, 5]

TEAM_TO_CLUB = {
    'Ruiboys':  'Ruiboys',
    'Slashers': 'Slashers',
    'Cheetahs': 'Cheetahs',
    'THC':      'The Howling Cows',
}

IMPORT_DIR = __file__.replace('import_round_teams.py', '')


# ---------------------------------------------------------------------------
# DB helpers
# ---------------------------------------------------------------------------

def fetch_squad(conn) -> dict[str, list[dict]]:
    """Return {ffl_club_name: [{player_season_id, drv_name}]} for FFL 2026."""
    with conn.cursor() as cur:
        cur.execute("""
            SELECT c.name, fps.id, fp.drv_name
            FROM ffl.player_season fps
            JOIN ffl.player fp ON fp.id = fps.player_id
            JOIN ffl.club_season cs ON fps.club_season_id = cs.id
            JOIN ffl.club c ON cs.club_id = c.id
            JOIN ffl.season s ON cs.season_id = s.id
            WHERE s.name = 'FFL 2026'
            ORDER BY c.name, fp.drv_name
        """)
        result: dict[str, list] = {}
        for club, ps_id, name in cur.fetchall():
            result.setdefault(club, []).append({'player_season_id': ps_id, 'name': name})
    return result


def _club_match_subquery(round_name: str, ffl_club: str) -> str:
    """SQL subquery that resolves club_match_id by name — no hardcoded IDs."""
    return (
        f"(SELECT cm.id FROM ffl.club_match cm"
        f" JOIN ffl.club_season cs ON cm.club_season_id = cs.id"
        f" JOIN ffl.club c ON cs.club_id = c.id"
        f" JOIN ffl.match m ON cm.match_id = m.id"
        f" JOIN ffl.round r ON m.round_id = r.id"
        f" JOIN ffl.season s ON r.season_id = s.id"
        f" WHERE r.name = '{round_name}' AND c.name = '{ffl_club}' AND s.name = 'FFL 2026')"
    )


def _player_season_subquery(drv_name: str, ffl_club: str) -> str:
    """SQL subquery that resolves player_season_id by drv_name — no hardcoded IDs.
    Uses ffl.player.drv_name directly to avoid afl.player name duplicates."""
    n = drv_name.replace("'", "''")
    c = ffl_club.replace("'", "''")
    return (
        f"(SELECT fps.id FROM ffl.player_season fps"
        f" JOIN ffl.player fp ON fp.id = fps.player_id"
        f" JOIN ffl.club_season cs ON fps.club_season_id = cs.id"
        f" JOIN ffl.club c ON cs.club_id = c.id"
        f" JOIN ffl.season s ON cs.season_id = s.id"
        f" WHERE fp.drv_name = '{n}' AND c.name = '{c}' AND s.name = 'FFL 2026')"
    )


# ---------------------------------------------------------------------------
# Name matching
# ---------------------------------------------------------------------------

def _last_name(name: str) -> str:
    parts = name.strip().split()
    return parts[-1].lower() if parts else ''


def _initials(name: str) -> str:
    parts = name.strip().split()
    return parts[0][0].lower() if parts else ''


def match_player(csv_name: str, squad: list[dict]) -> dict | None:
    """Match a CSV player name to a squad entry. Returns the entry or None."""
    names = [p['name'] for p in squad]

    # 1. Exact match
    for p in squad:
        if p['name'] == csv_name:
            return p

    # 2. Case-insensitive exact
    for p in squad:
        if p['name'].lower() == csv_name.lower():
            return p

    # 3. "Initial Lastname" pattern — e.g. "H McKay" → "Harry McKay"
    tokens = csv_name.strip().split()
    if len(tokens) >= 2 and len(tokens[0]) == 1:
        initial = tokens[0].lower()
        last = ' '.join(tokens[1:]).lower()
        candidates = [p for p in squad
                      if _last_name(p['name']) == last and _initials(p['name']) == initial]
        if len(candidates) == 1:
            return candidates[0]
        if len(candidates) > 1:
            # Multiple matches — try fuzzy on full name
            pass

    # 4. Last-name-only match (rare fallback)
    last = _last_name(csv_name)
    candidates = [p for p in squad if _last_name(p['name']) == last]
    if len(candidates) == 1:
        return candidates[0]

    # 5. Fuzzy on full name
    matches = difflib.get_close_matches(csv_name, names, n=1, cutoff=0.6)
    if matches:
        return next(p for p in squad if p['name'] == matches[0])

    return None


# ---------------------------------------------------------------------------
# CSV reading
# ---------------------------------------------------------------------------

def read_teams(round_num: int) -> list[dict]:
    path = f"{IMPORT_DIR}2026_{round_num}_teams.csv"
    with open(path) as f:
        return list(csv.DictReader(f))


def read_scores(round_num: int) -> dict[str, int | None]:
    """Return {team_label: score_or_None}."""
    path = f"{IMPORT_DIR}2026_{round_num}_scores.csv"
    result = {}
    with open(path) as f:
        for row in csv.DictReader(f):
            team = row['team'].strip()
            raw = row['score'].strip()
            result[team] = int(raw) if raw else None
    return result


# ---------------------------------------------------------------------------
# SQL generation
# ---------------------------------------------------------------------------

def generate_sql(warnings: list[str], squad: dict) -> list[str]:
    """
    Returns list of SQL statements. All IDs resolved by name — no hardcoded integers.
    Appends warnings to the provided list for unresolved players.
    """
    stmts = []

    for rn in ROUNDS:
        rows = read_teams(rn)
        scores = read_scores(rn)
        round_name = f"Round {rn}"
        stmts.append(f"-- Round {rn}")

        # club_match score updates — subquery by round + club name
        for team_label, score in scores.items():
            ffl_club = TEAM_TO_CLUB.get(team_label)
            if not ffl_club:
                warnings.append(f"WARN R{rn}: unknown team label '{team_label}' in scores")
                continue
            if score is None:
                continue
            cm_sq = _club_match_subquery(round_name, ffl_club.replace("'", "''"))
            stmts.append(f"UPDATE ffl.club_match SET drv_score = {score} WHERE id = {cm_sq};")

        # player_match inserts — subqueries for both club_match_id and player_season_id
        for row in rows:
            if int(row['round']) != rn:
                continue
            team_label = row['team'].strip()
            ffl_club = TEAM_TO_CLUB.get(team_label)
            if not ffl_club:
                warnings.append(f"WARN R{rn}: unknown team '{team_label}'")
                continue

            csv_name = row['player_name'].strip()
            matched = match_player(csv_name, squad.get(ffl_club, []))
            if not matched:
                warnings.append(f"WARN R{rn} [{ffl_club}]: no match for '{csv_name}'")
                continue

            afl_name = matched['name']
            position = row.get('position', '').strip() or None
            backup = row.get('backup_positions', '').strip() or None
            interchange = row.get('interchange_position', '').strip() or None
            score_raw = row.get('score', '').strip()
            score = int(score_raw) if score_raw else 0

            pos_sql = f"'{position}'" if position else 'NULL'
            bak_sql = f"'{backup}'" if backup else 'NULL'
            ich_sql = f"'{interchange}'" if interchange else 'NULL'

            cm_sq = _club_match_subquery(round_name, ffl_club.replace("'", "''"))
            ps_sq = _player_season_subquery(afl_name, ffl_club)

            stmts.append(
                f"INSERT INTO ffl.player_match"
                f" (club_match_id, player_season_id, status, position, backup_positions, interchange_position, drv_score)"
                f" VALUES ({cm_sq}, {ps_sq}, 'played', {pos_sql}, {bak_sql}, {ich_sql}, {score})"
                f" ON CONFLICT (player_season_id, club_match_id) DO UPDATE SET"
                f" status='played', position={pos_sql}, backup_positions={bak_sql},"
                f" interchange_position={ich_sql}, drv_score={score};"
            )

        stmts.append("")

    return stmts


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--apply', action='store_true', help='Apply SQL directly to DB')
    args = parser.parse_args()

    conn = psycopg2.connect(DB_DSN)
    squad = fetch_squad(conn)
    conn.close()

    warnings: list[str] = []
    stmts = generate_sql(warnings, squad)

    if warnings:
        print("WARNINGS:", file=sys.stderr)
        for w in warnings:
            print(f"  {w}", file=sys.stderr)

    if args.apply:
        conn2 = psycopg2.connect(DB_DSN)
        with conn2.cursor() as cur:
            cur.execute("BEGIN;")
            for stmt in stmts:
                s = stmt.strip()
                if s and not s.startswith('--'):
                    cur.execute(s)
            cur.execute("COMMIT;")
        conn2.close()
        n = len([s for s in stmts if s.strip().startswith('INSERT') or s.strip().startswith('UPDATE')])
        print(f"Applied {n} statements.")
    else:
        conn.close()
        print("BEGIN;")
        for stmt in stmts:
            if stmt.strip():
                print(stmt)
        print("COMMIT;")


if __name__ == '__main__':
    main()
