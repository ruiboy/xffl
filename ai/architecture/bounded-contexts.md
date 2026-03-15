# Bounded Contexts

## AFL
Real-world Australian Football League data.

**Key concepts:** Club, Season, Round, Match, PlayerMatch (kicks, handballs, marks, hitouts, tackles, goals, behinds)

**Publishes:** `AFL.PlayerMatchUpdated`

## FFL
Fantasy league built on AFL statistics.

**Key concepts:** Club, Player, ClubSeason (ladder: wins/losses/draws/points/premiership points), PlayerMatch (position, status)

**Subscribes:** `AFL.PlayerMatchUpdated` → calculates fantasy score
**Publishes:** `FFL.FantasyScoreCalculated`