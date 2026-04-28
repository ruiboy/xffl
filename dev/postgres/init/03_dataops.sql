-- Data Ops tables (per ADR-016: no FK references to core schema tables)

-- ACL identity mapping: external source match IDs → afl.match IDs
CREATE TABLE IF NOT EXISTS afl.dataops_match_source (
    source      TEXT NOT NULL,
    external_id TEXT NOT NULL,
    match_id    INTEGER NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (source, external_id),
    CONSTRAINT uni_afl_dataops_match_source_match UNIQUE (source, match_id)
);
