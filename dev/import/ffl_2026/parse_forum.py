#!/usr/bin/env python3
"""
FFL forum post parser → ffl_teams.csv + ffl_scores.csv

Usage:
    python parse_forum.py <round> <input_file>
    python parse_forum.py 1 r1.txt

Outputs (in same directory as input):
    ffl_teams.csv   — one row per player per round/team
    ffl_scores.csv  — one row per team per round

Assumptions encoded from session:
- * / * (INT) / *** *** = interchange player for star position
- R / RUCK = hitouts; HB / H = handballs; HO = hitouts
- Cheetahs shows raw stats for goals(×5), marks(×2), tackles(×4); FFL pts for all others
- (AV) tag = averages annotation, ignored
- "sub score" in starter line = DNP sub in, use score from sub (starter score = 0)
- "xN=score" (THC) = raw stat × multiplier shown explicitly, use the score
- "DNP= / DNP-" = player did not play, score shown is sub's contribution
- Interchange assumed NOT to have happened if no score given for interchange player
- Blank bench slot (T/H -) = only 3 bench players that round
- TDK = Tom De Koning (SK); strip inline nicknames (Journeyman, Mountain Goat)
"""

import re
import csv
import sys
import os
from dataclasses import dataclass, field
from typing import Optional

# ---------------------------------------------------------------------------
# Constants
# ---------------------------------------------------------------------------

POSITION_ALIASES = {
    'GOALS': 'goals', 'GOAL': 'goals',
    'KICKS': 'kicks', 'KICK': 'kicks',
    'HANDBALLS': 'handballs', 'HANDBALL': 'handballs', 'HB': 'handballs', 'HBS': 'handballs',
    'MARKS': 'marks', 'MARK': 'marks',
    'TACKLES': 'tackles', 'TACKLE': 'tackles',
    'HITOUTS': 'hitouts', 'HITOUT': 'hitouts',
    'RUCK': 'hitouts', 'RUCKS': 'hitouts', 'HO': 'hitouts',
    'STAR': 'star',
    'BENCH': 'bench', 'INTERCHANGE': 'bench',
}

# Single/double letter bench codes
BENCH_LETTER = {
    'G': 'goals', 'K': 'kicks',
    'H': 'handballs', 'HB': 'handballs',
    'M': 'marks', 'T': 'tackles',
    'R': 'hitouts', 'HO': 'hitouts',
    'S': 'star',
}

# Cheetahs raw stat multipliers (other teams show FFL pts directly)
CHEETAHS_RAW = {'goals': 5, 'marks': 2, 'tackles': 4}

# Nickname → canonical name (club optional)
NICKNAMES = {
    'TDK':          ('Tom De Koning', 'SK'),
    'T De Koning':  ('Tom De Koning', 'SK'),
}

# Inline nickname words to strip from player names
STRIP_WORDS = ['Journeyman', 'Mountain Goat']

# Lines to discard as forum artifacts
ARTIFACT_RE = re.compile(
    r'^(Quote|Edit|Share|Like|Dislike|Pin\s+Topic|TATLTWDNMTS|Bloody\s+Legend'
    r'|hugs?\s*$|reacted\s+to|likes?\s+this\s+post|\d{1,2}:\d{2}\s*(AM|PM))',
    re.I
)

# Lines that are purely a number (position subtotals) — discard
SUBTOTAL_RE = re.compile(r'^\s*\d+\s*$')

# Position section header — e.g. "GOALS", "GOALS  30", "Goals", "HB"
SECTION_RE = re.compile(
    r'^\s*(?:I/C[–-]\s*)?(GOALS?|KICKS?|HANDBALLS?|HANDBALL|HB|HBS|MARKS?'
    r'|TACKLES?|HITOUTS?|RUCK[S]?|HO|STAR|BENCH|INTERCHANGE)\b[\s\d]*$',
    re.I
)

# THC I/C section header: "I/C- Star"
IC_SECTION_RE = re.compile(r'^\s*I/C[–\-]\s*\w+', re.I)

# ---------------------------------------------------------------------------
# Data classes
# ---------------------------------------------------------------------------

@dataclass
class PlayerRow:
    round: int
    team: str
    player_name: str
    afl_club: str
    position: str
    backup_positions: str
    interchange_position: str
    score: str
    notes: str

@dataclass
class ScoreRow:
    round: int
    team: str
    score: str
    comment: str

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def normalize_position(raw: str) -> str:
    key = raw.strip().upper()
    return POSITION_ALIASES.get(key, raw.lower())


def decode_bench_code(code: str) -> tuple[str, str]:
    """Decode 'T/H' → backup_positions='tackles,handballs', interchange=''
       Decode '*'  → backup_positions='',          interchange='star'
       Decode '* (INT)' → same as '*'
    """
    code = code.strip()
    if code in ('*', '*(INT)', '* (INT)'):
        return '', 'star'
    # Strip asterisks from interchange-marked names (handled upstream)
    parts = re.split(r'/', code.upper())
    positions = []
    for p in parts:
        p = p.strip()
        pos = BENCH_LETTER.get(p)
        if pos:
            positions.append(pos)
        else:
            positions.append(p.lower())
    return ','.join(positions), ''


def strip_inline_nicknames(name: str) -> str:
    for word in STRIP_WORDS:
        name = name.replace(word, '').strip()
    return re.sub(r'\s{2,}', ' ', name)


def resolve_nickname(name: str, club: str) -> tuple[str, str]:
    for nick, (canonical, canon_club) in NICKNAMES.items():
        if nick.lower() in name.lower():
            return canonical, canon_club or club
    return strip_inline_nicknames(name), club


def is_artifact(line: str) -> bool:
    line = line.strip()
    if not line:
        return True
    if ARTIFACT_RE.match(line):
        return True
    if SUBTOTAL_RE.match(line):
        return True
    return False


def extract_score(s: str) -> str:
    """Extract numeric score string, return '' if none."""
    m = re.search(r'\b(\d+)\b', s)
    return m.group(1) if m else ''

# ---------------------------------------------------------------------------
# Team block splitter
# ---------------------------------------------------------------------------

TEAM_HEADER = re.compile(
    r'(?:'
    r'(?P<ruiboys>^R\d+\s+\d+|^\d{3,}\s*$)'   # Ruiboys: "R1  404" or bare score
    r'|(?P<cheetahs>CHEETAHS\s+\d+)'
    r'|(?P<slashers>TOTAL\s*:\s*\d+|^(?:Slashers?.*)?$)'   # detected from TOTAL: line
    r'|(?P<thc>THC[-–\s]+\d+|THC\s+\d+)'
    r')',
    re.I | re.M
)

def identify_team(lines: list[str]) -> Optional[str]:
    """Return team name from block header lines."""
    text = '\n'.join(lines[:5])
    if re.search(r'\bTHC\b', text, re.I):
        return 'THC'
    if re.search(r'\bCHEETAHS?\b', text, re.I):
        return 'Cheetahs'
    # Slashers identified by TOTAL: line or known username pattern
    for line in lines:
        if re.search(r'TOTAL\s*:', line, re.I):
            return 'Slashers'
    # Ruiboys identified by dash-format player lines
    for line in lines:
        if re.search(r'–', line):
            return 'Ruiboys'
    return None

# ---------------------------------------------------------------------------
# Per-team player line parsers
# ---------------------------------------------------------------------------

def parse_ruiboys_player(line: str, position: str, team: str, round_: int) -> Optional[PlayerRow]:
    """
    Formats:
      Full Name – Club  score
      Full Name – Club  raw  score       (shows raw stat + FFL pts)
      Full Name – Club  raw  sub  score  (interchange happened)
      Full Name – Club  * (INT)  score
      Full Name – Club  T/H
      Full Name – Club  K/M  26\13       (raw stats, no score)
    """
    line = line.strip()
    # Split on dash/endash separator
    m = re.match(r'^(.+?)\s*[–\-]\s*([A-Z][a-zA-Z]+)\s+(.*)', line)
    if not m:
        return None

    raw_name, club, rest = m.group(1).strip(), m.group(2).strip(), m.group(3).strip()
    name, club = resolve_nickname(raw_name, club)

    notes = ''
    backup_positions = ''
    interchange_position = ''
    score = ''

    # Bench line: position is already 'bench', rest contains bench code
    if position == 'bench':
        # e.g. "* (INT)  60" or "T/H" or "K/M  26\13"
        bench_code_m = re.match(r'^(\*\s*\(?\s*INT\s*\)?|\*|[A-Z]+/[A-Z]+)\s*(.*)', rest, re.I)
        if bench_code_m:
            code, tail = bench_code_m.group(1).strip(), bench_code_m.group(2).strip()
            backup_positions, interchange_position = decode_bench_code(code)
            # Extract score if present (ignore raw stat pairs like 26\13 or 0\10)
            if tail and not re.search(r'\d+[\\|]\d+', tail):
                score = extract_score(tail)
            elif tail and re.search(r'\d+[\\|]\d+', tail):
                notes = f'raw stats shown: {tail}'
        return PlayerRow(round_, team, name, club, 'bench',
                         backup_positions, interchange_position, score, notes)

    # Starter line: detect "sub" pattern
    sub_m = re.search(r'(\d+)\s+sub\s+(\d+)', rest, re.I)
    if sub_m:
        notes = f'starter score {sub_m.group(1)}; interchange/sub used, slot score = {sub_m.group(2)}'
        score = sub_m.group(2)
    else:
        # "raw_stat  ffl_score" — two numbers: use the last
        nums = re.findall(r'\d+', rest)
        if nums:
            score = nums[-1]

    return PlayerRow(round_, team, name, club, position,
                     backup_positions, interchange_position, score, notes)


def parse_slashers_player(line: str, position: str, team: str, round_: int) -> Optional[PlayerRow]:
    """
    Formats:
      Initial Surname (Club)  score
      Initial Surname (Club)  dnp - interchange Name score
      ***Name (Club)***  score           (interchange player in INTERCHANGE section)
      K/G - Name (Club)                  (dual-position bench)
    """
    line = line.strip()

    # Bench code prefix: "K/G - Name" or "R/M - TDK"
    bench_prefix_m = re.match(r'^([A-Z]+/[A-Z]+)\s*[-–]\s*(.+)', line, re.I)
    if bench_prefix_m and position == 'bench':
        code, rest = bench_prefix_m.group(1), bench_prefix_m.group(2).strip()
        backup_positions, interchange_position = decode_bench_code(code)
        # Parse player from rest
        name, club = _slashers_name_club(rest)
        name, club = resolve_nickname(name, club)
        score = extract_score(rest.split('(')[-1]) if '(' in rest else ''
        return PlayerRow(round_, team, name, club, 'bench',
                         backup_positions, interchange_position, score, '')

    # Interchange player: ***Name (Club)***
    ic_m = re.match(r'^\*{2,3}\s*(.+?)\s*\*{2,3}\s*(.*)', line)
    if ic_m and position == 'bench':
        inner, tail = ic_m.group(1), ic_m.group(2)
        name, club = _slashers_name_club(inner)
        name, club = resolve_nickname(name, club)
        score = extract_score(tail)
        notes = 'interchange player' + (f'; score {score}' if score else '; no score (interchange not assumed to have occurred)')
        return PlayerRow(round_, team, name, club, 'bench',
                         '', 'star', score, notes)

    # DNP line: "E Richards (WB) dnp - interchange Merrett 15"
    dnp_m = re.match(r'^(.+?)\s+dnp\s*[-–]\s*interchange\s+(\w+)\s+(\d+)', line, re.I)
    if dnp_m:
        name_part, sub_name, sub_score = dnp_m.group(1), dnp_m.group(2), dnp_m.group(3)
        name, club = _slashers_name_club(name_part)
        notes = f'DNP; {sub_name} subbed in for {sub_score} pts'
        return PlayerRow(round_, team, name, club, position,
                         '', '', '0', notes)

    # Normal player line
    name, club = _slashers_name_club(line)
    if not name:
        return None
    name, club = resolve_nickname(name, club)
    # Score: last number after club
    score_m = re.search(r'\)\s+(\d+)', line)
    score = score_m.group(1) if score_m else ''

    # Interchange annotation on STAR line: "M Holmes 68 - interchange Brayshaw 72"
    ic_ann = re.search(r'[-–]\s*interchange[d]?\s+(?:with\s+)?(\w+)\s+(\d+)', line, re.I)
    notes = ''
    if ic_ann:
        ic_name, ic_score = ic_ann.group(1), ic_ann.group(2)
        if int(ic_score) > int(score):
            notes = f'interchange occurred: {ic_name} {ic_score} > starter {score}; slot score = {ic_score}'
            score = ic_score
        else:
            notes = f'interchange NOT occurred: {ic_name} {ic_score} <= starter {score}'

    return PlayerRow(round_, team, name, club, position,
                     '', '', score, notes)


def _slashers_name_club(s: str) -> tuple[str, str]:
    """Extract 'Initial Surname (Club)' → (name, club). Returns ('','') if no match."""
    s = s.strip()
    m = re.match(r'^([A-Z][A-Za-z\s\'\-]+?)\s*\(([A-Z][a-zA-Z]+)\)', s)
    if m:
        return m.group(1).strip(), m.group(2).strip()
    # Fallback: no club in parens
    m2 = re.match(r'^([A-Z][A-Za-z\s\'\-]+)', s)
    if m2:
        return m2.group(1).strip(), ''
    return '', ''


def parse_cheetahs_player(line: str, position: str, team: str, round_: int,
                           raw_positions: set) -> Optional[PlayerRow]:
    """
    Formats:
      First Last (Club) score        — score is raw stat for goals/marks/tackles
      Name (Club) *                  — interchange bench
      Name (Club) T/K                — dual-position bench
      Name (Club) * (INT)            — same as *
      Isaac Heeney (Syd) G/K 4,16   — bench with raw stats (ignore stats)
    """
    line = line.strip()
    m = re.match(r'^(.+?)\s*\(([A-Z][a-zA-Z]+)\)\s*(.*)', line)
    if not m:
        return None

    raw_name, club, rest = m.group(1).strip(), m.group(2).strip(), m.group(3).strip()
    name = strip_inline_nicknames(raw_name)
    name, club = resolve_nickname(name, club)

    notes = ''
    backup_positions = ''
    interchange_position = ''
    score = ''

    if position == 'bench':
        # Bench code is first token of rest
        tokens = rest.split()
        if tokens:
            code = tokens[0].rstrip(',')
            backup_positions, interchange_position = decode_bench_code(code)
            # Remaining might be raw stats like "4,16" — ignore for score
            tail = ' '.join(tokens[1:])
            if tail and not re.search(r'\d+[,]\d+', tail):
                score = extract_score(tail)
            elif tail:
                notes = f'raw stats shown for bench player: {tail}'
        return PlayerRow(round_, team, name, club, 'bench',
                         backup_positions, interchange_position, score, notes)

    # Starter: rest is the score (raw stat for some positions)
    raw_score = extract_score(rest)
    if raw_score:
        if position in raw_positions:
            mult = CHEETAHS_RAW.get(position, 1)
            ffl_score = str(int(raw_score) * mult)
            notes = f'raw {position} stat: {raw_score} × {mult} = {ffl_score}'
            score = ffl_score
        else:
            score = raw_score

    return PlayerRow(round_, team, name, club, position,
                     backup_positions, interchange_position, score, notes)


def parse_thc_player(line: str, position: str, team: str, round_: int,
                     in_ic: bool) -> Optional[PlayerRow]:
    """
    Formats:
      Name CLUB= score
      Name CLUB- score
      Name CLUB x3=15           (explicit raw×mult)
      Name CLUB DNP= score      (or DNP- score)
      Name CLUB DNP- 16HO       (DNP with sub score + stat suffix)
      *= Name CLUB= score       (THC interchange player)
      K/HB- Name CLUB           (dual-position bench in I/C section)
      Star- Name CLUB= score    (interchange player declaration)
      HO/T- Name CLUB- score    (dual-pos bench with raw stat)
    """
    line = line.strip()

    # I/C section: "Star- Name CLUB= score" → interchange player
    star_ic_m = re.match(r'^Star[-–]\s*(.+)', line, re.I)
    if star_ic_m and in_ic:
        rest = star_ic_m.group(1)
        name, club, score, notes = _thc_name_score(rest)
        ic_note = 'interchange player (star)'
        if score and notes:
            notes = f'{ic_note}; {notes}'
        else:
            notes = ic_note
        return PlayerRow(round_, team, name, club, 'bench', '', 'star', score, notes)

    # I/C section: "*= Name CLUB= score" → same
    star_ic2_m = re.match(r'^\*\s*=\s*(.+)', line)
    if star_ic2_m and in_ic:
        rest = star_ic2_m.group(1)
        name, club, score, notes = _thc_name_score(rest)
        notes = ('interchange player (star); ' + notes).rstrip('; ')
        return PlayerRow(round_, team, name, club, 'bench', '', 'star', score, notes)

    # I/C section: "K/HB- Name" or "G/M= Name" or "T/HO- Name"
    bench_code_m = re.match(r'^([A-Z]+/[A-Z]+)\s*[-=]\s*(.+)', line, re.I)
    if bench_code_m and in_ic:
        code, rest = bench_code_m.group(1), bench_code_m.group(2)
        backup_positions, interchange_position = decode_bench_code(code)
        name, club, score, notes = _thc_name_score(rest)
        return PlayerRow(round_, team, name, club, 'bench',
                         backup_positions, interchange_position, score, notes)

    # Normal starter line
    name, club, score, notes = _thc_name_score(line)
    if not name:
        return None

    return PlayerRow(round_, team, name, club, position,
                     '', '', score, notes)


def _thc_name_score(s: str) -> tuple[str, str, str, str]:
    """Parse 'Name CLUB= score' or 'Name CLUB x3=15' etc. Returns (name, club, score, notes)."""
    s = s.strip()
    notes = ''

    # Strip (AV) annotation
    if re.search(r'\(AV\)', s, re.I):
        s = re.sub(r'\s*\(AV\)', '', s, flags=re.I).strip()
        notes = 'AV (averages) annotation stripped'

    # DNP with stat suffix: "J Sweet PA DNP- 16HO"
    dnp_stat_m = re.match(r'^(.+?)\s+([A-Z]+)\s+DNP[-=]\s*(\d+)[A-Za-z]+', s, re.I)
    if dnp_stat_m:
        name = strip_inline_nicknames(dnp_stat_m.group(1).strip())
        club = dnp_stat_m.group(2)
        sub_score = dnp_stat_m.group(3)
        return name, club, '0', f'DNP; sub contributed {sub_score} pts'

    # DNP plain: "Wines PA-DNP=11" or "Wines PA DNP= 11"
    dnp_m = re.match(r'^(.+?)\s+([A-Z]+)\s*[-–]?\s*DNP[-=]\s*(\d+)', s, re.I)
    if dnp_m:
        name = strip_inline_nicknames(dnp_m.group(1).strip())
        club = dnp_m.group(2)
        sub_score = dnp_m.group(3)
        return name, club, '0', f'DNP; sub contributed {sub_score} pts'

    # Explicit multiplier: "Gunston HAW x4=20"
    mult_m = re.match(r'^(.+?)\s+([A-Z]+)\s+x(\d+)\s*=\s*(\d+)', s, re.I)
    if mult_m:
        name = strip_inline_nicknames(mult_m.group(1).strip())
        club = mult_m.group(2)
        score = mult_m.group(4)
        raw = mult_m.group(3)
        return name, club, score, f'explicit multiplier x{raw} shown'

    # Tackles with explicit "xN=score": "Graham WCE x9=36"
    # (same pattern, covered above)

    # Score with stat suffix stripped: e.g. "16HO" → we already handle DNP case above
    # Normal: "Name CLUB= score" or "Name CLUB- score"
    score_m = re.match(r'^(.+?)\s+([A-Z]{2,4})\s*[-=]\s*(\d+)', s)
    if score_m:
        name = strip_inline_nicknames(score_m.group(1).strip())
        club = score_m.group(2)
        score = score_m.group(3)
        name, club = resolve_nickname(name, club)
        return name, club, score, notes

    # No score: "Name CLUB"
    no_score_m = re.match(r'^(.+?)\s+([A-Z]{2,4})\s*$', s)
    if no_score_m:
        name = strip_inline_nicknames(no_score_m.group(1).strip())
        club = no_score_m.group(2)
        name, club = resolve_nickname(name, club)
        return name, club, '', notes

    # Fallback
    name, club = resolve_nickname(strip_inline_nicknames(s), '')
    return name, club, '', notes

# ---------------------------------------------------------------------------
# Block parser
# ---------------------------------------------------------------------------

def parse_block(round_: int, team: str, lines: list[str]) -> tuple[list[PlayerRow], ScoreRow]:
    players: list[PlayerRow] = []
    score_total = ''
    comment_lines = []
    current_position = ''
    in_ic = False  # inside THC I/C section

    is_cheetahs = (team == 'Cheetahs')
    is_thc = (team == 'THC')
    is_slashers = (team == 'Slashers')
    is_ruiboys = (team == 'Ruiboys')

    raw_positions = set(CHEETAHS_RAW.keys()) if is_cheetahs else set()

    for line in lines:
        line = line.strip()
        if not line:
            continue
        if is_artifact(line):
            continue

        # Score/header lines
        # Ruiboys: "R6  0" or "355"
        if is_ruiboys:
            r_header = re.match(r'^R\d+\s+(\d+)$', line)
            if r_header:
                score_total = r_header.group(1)
                continue
            bare_score = re.match(r'^(\d{3,})\s*$', line)
            if bare_score and not current_position:
                score_total = bare_score.group(1)
                continue

        # Cheetahs: "CHEETAHS 385"
        if is_cheetahs:
            ch_header = re.match(r'^CHEETAHS\s+(\d+)', line, re.I)
            if ch_header:
                score_total = ch_header.group(1)
                continue

        # THC: "THC- 332" or "THC 332"
        if is_thc:
            thc_header = re.match(r'^THC[-–\s]+(\d+)', line, re.I)
            if thc_header:
                score_total = thc_header.group(1)
                continue

        # Slashers: "TOTAL: 386" or bare score
        if is_slashers:
            total_m = re.match(r'^TOTAL\s*:\s*(\d+)', line, re.I)
            if total_m:
                score_total = total_m.group(1)
                continue
            bare_score = re.match(r'^(\d{3,}),?\s*', line)
            if bare_score and not current_position:
                score_total = bare_score.group(1)
                continue
            # "366, a leap year" style
            score_comment_m = re.match(r'^(\d{3,}),?\s+(.+)', line)
            if score_comment_m and not current_position:
                score_total = score_comment_m.group(1)
                comment_lines.append(score_comment_m.group(2))
                continue

        # THC I/C section header
        if is_thc and IC_SECTION_RE.match(line):
            in_ic = True
            current_position = 'bench'
            continue

        # "Interchange = *" (Cheetahs) — informational, skip
        if re.match(r'^Interchange\s*=\s*\*', line, re.I):
            continue

        # Position section header
        sec_m = SECTION_RE.match(line)
        if sec_m:
            raw_pos = sec_m.group(1)
            current_position = normalize_position(raw_pos)
            in_ic = False
            continue

        # Skip lines with no useful content
        if not current_position:
            # Could be a comment before the first position
            comment_lines.append(line)
            continue

        # Parse player line
        row = None
        if is_ruiboys:
            row = parse_ruiboys_player(line, current_position, team, round_)
        elif is_slashers:
            row = parse_slashers_player(line, current_position, team, round_)
        elif is_cheetahs:
            row = parse_cheetahs_player(line, current_position, team, round_, raw_positions)
        elif is_thc:
            row = parse_thc_player(line, current_position, team, round_, in_ic)

        if row:
            players.append(row)
        else:
            # Unrecognised line — record as comment
            comment_lines.append(f'[UNRECOGNISED: {line}]')

    comment = ' | '.join(c for c in comment_lines if c and not re.match(r'^\[UNRECOGNISED', c))
    score_row = ScoreRow(round_, team, score_total, comment)
    return players, score_row

# ---------------------------------------------------------------------------
# Top-level: split raw text into team blocks
# ---------------------------------------------------------------------------

TEAM_SPLIT_RE = re.compile(
    r'(?:^|\n)(?='
    r'(?:R\d+\s+\d+)'           # Ruiboys R header
    r'|(?:CHEETAHS\s+\d+)'      # Cheetahs
    r'|(?:THC[-–\s]+\d+)'       # THC
    r'|(?:\d{3,}\s*\n.*GOALS)'  # Slashers bare score followed by GOALS
    r')',
    re.I | re.M
)

def split_blocks(text: str) -> list[str]:
    """Heuristically split forum text into per-team blocks."""
    # Normalise line endings
    text = text.replace('\r\n', '\n').replace('\r', '\n')

    # Try splitting on known team headers
    parts = []
    current = []
    lines = text.split('\n')
    i = 0
    while i < len(lines):
        line = lines[i]
        # New block signals
        is_ruiboys_header = bool(re.match(r'^R\d+\s+\d+\s*$', line.strip()))
        is_cheetahs_header = bool(re.match(r'^CHEETAHS\s+\d+', line.strip(), re.I))
        is_thc_header = bool(re.match(r'^THC[-–\s]+\d+', line.strip(), re.I))
        # Slashers: bare 3-digit score followed within 3 lines by GOALS
        is_slashers_header = False
        if re.match(r'^\d{3}\s*$', line.strip()):
            lookahead = '\n'.join(lines[i:i+5])
            if re.search(r'\bGOALS?\b', lookahead, re.I):
                is_slashers_header = True

        if (is_ruiboys_header or is_cheetahs_header or is_thc_header or is_slashers_header) and current:
            parts.append('\n'.join(current))
            current = []

        current.append(line)
        i += 1

    if current:
        parts.append('\n'.join(current))

    return [p for p in parts if p.strip()]


def detect_team(block: str) -> Optional[str]:
    lines = block.strip().split('\n')
    for line in lines[:8]:
        if re.search(r'\bTHC\b', line, re.I):
            return 'THC'
        if re.search(r'\bCHEETAHS?\b', line, re.I):
            return 'Cheetahs'
    for line in lines:
        if re.search(r'TOTAL\s*:', line, re.I):
            return 'Slashers'
    for line in lines:
        if '–' in line or re.search(r'\bTATLTWDNMTS\b', line):
            return 'Ruiboys'
    return None

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main():
    if len(sys.argv) < 3:
        print(f'Usage: {sys.argv[0]} <round_number> <input_file>', file=sys.stderr)
        sys.exit(1)

    round_ = int(sys.argv[1])
    input_path = sys.argv[2]

    with open(input_path, 'r', encoding='utf-8') as f:
        text = f.read()

    blocks = split_blocks(text)
    all_players: list[PlayerRow] = []
    all_scores: list[ScoreRow] = []

    for block in blocks:
        team = detect_team(block)
        if not team:
            print(f'[WARN] Could not identify team for block:\n{block[:120]}\n', file=sys.stderr)
            continue

        lines = block.strip().split('\n')
        players, score_row = parse_block(round_, team, lines)
        all_players.extend(players)
        all_scores.append(score_row)

    out_dir = os.path.dirname(os.path.abspath(input_path))
    teams_path = os.path.join(out_dir, 'ffl_teams.csv')
    scores_path = os.path.join(out_dir, 'ffl_scores.csv')

    # Append mode so multiple rounds accumulate
    teams_exists = os.path.exists(teams_path)
    scores_exists = os.path.exists(scores_path)

    with open(teams_path, 'a', newline='', encoding='utf-8') as f:
        w = csv.writer(f)
        if not teams_exists:
            w.writerow(['round', 'team', 'player_name', 'afl_club', 'position',
                        'backup_positions', 'interchange_position', 'score', 'notes'])
        for r in all_players:
            w.writerow([r.round, r.team, r.player_name, r.afl_club, r.position,
                        r.backup_positions, r.interchange_position, r.score, r.notes])

    with open(scores_path, 'a', newline='', encoding='utf-8') as f:
        w = csv.writer(f)
        if not scores_exists:
            w.writerow(['round', 'team', 'score', 'comment'])
        for r in all_scores:
            w.writerow([r.round, r.team, r.score, r.comment])

    print(f'Round {round_}: {len(all_players)} player rows, {len(all_scores)} score rows')
    print(f'  → {teams_path}')
    print(f'  → {scores_path}')

    # Print any unrecognised lines to stderr for review
    for r in all_players:
        if 'UNRECOGNISED' in r.notes:
            print(f'[REVIEW] {r}', file=sys.stderr)


if __name__ == '__main__':
    main()
