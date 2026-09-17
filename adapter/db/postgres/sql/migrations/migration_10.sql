-- Adds equipment_models.kind, a free-form, optional classification of the
-- equipment (e.g. "dvr", "cameraip") capped at 50 characters by the HTTP
-- layer. Existing rows backfill to ''.
ALTER TABLE equipment_models ADD COLUMN kind TEXT NOT NULL DEFAULT '';
ALTER TABLE equipment_models ALTER COLUMN kind DROP DEFAULT;
