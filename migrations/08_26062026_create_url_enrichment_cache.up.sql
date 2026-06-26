CREATE TABLE IF NOT EXISTS url_enrichment_cache (
    canonical_url TEXT        PRIMARY KEY,
    title         TEXT        NOT NULL DEFAULT '',
    description   TEXT        NOT NULL DEFAULT '',
    category      TEXT        NOT NULL DEFAULT 'other',
    tags          TEXT[]      NOT NULL DEFAULT '{}',
    image_url     TEXT        NOT NULL DEFAULT '',
    enriched_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_url_enrichment_cache_enriched_at
    ON url_enrichment_cache (enriched_at DESC);
