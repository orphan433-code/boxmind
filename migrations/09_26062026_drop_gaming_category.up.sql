-- "gaming" is no longer a target category: watching gameplay is entertainment,
-- a game guide is learning, game tools/mods are tools. Reroute legacy rows to
-- the safe default (entertainment) so no stored bookmark keeps the old code.
UPDATE bookmarks
SET category = 'entertainment', updated_at = now()
WHERE category = 'gaming';

UPDATE url_enrichment_cache
SET category = 'entertainment', updated_at = now()
WHERE category = 'gaming';
