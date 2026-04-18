-- Integration xref tables for the AFL service (ACL pattern, see ADR-016).
-- These tables are owned by their respective adapters.
-- No foreign keys to core schema tables; referential integrity is enforced in application code.
-- Deleting a domain entity does not cascade to xref rows.

CREATE TABLE IF NOT EXISTS afl.xref_afltables_player (
    external_id TEXT    NOT NULL,  -- afltables' own player identifier
    player_id   INTEGER NOT NULL,  -- afl.player.id
    PRIMARY KEY (external_id)
);

CREATE INDEX IF NOT EXISTS idx_afl_xref_afltables_player_player_id ON afl.xref_afltables_player(player_id);
