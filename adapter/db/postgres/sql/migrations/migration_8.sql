-- Adds equipment_models.manufacturer and .external_view_url. Existing
-- rows are backfilled to '' since there's no better source for either.
ALTER TABLE equipment_models ADD COLUMN manufacturer TEXT NOT NULL DEFAULT '';
ALTER TABLE equipment_models ALTER COLUMN manufacturer DROP DEFAULT;

ALTER TABLE equipment_models ADD COLUMN external_view_url TEXT NOT NULL DEFAULT '';
ALTER TABLE equipment_models ALTER COLUMN external_view_url DROP DEFAULT;
