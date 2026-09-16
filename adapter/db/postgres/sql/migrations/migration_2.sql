-- Adds vehicles.name and allows ivms_type = 'none' (a vehicle with no IVMS
-- vendor integration). Existing rows are backfilled from plate_number
-- since there's no better source for a display name.
ALTER TABLE vehicles ADD COLUMN name TEXT NOT NULL DEFAULT '';
UPDATE vehicles SET name = plate_number WHERE name = '';
ALTER TABLE vehicles ALTER COLUMN name DROP DEFAULT;

ALTER TABLE vehicles DROP CONSTRAINT vehicles_ivms_type_check;
ALTER TABLE vehicles ADD CONSTRAINT vehicles_ivms_type_check CHECK (ivms_type IN ('cmsv6', 'none'));
