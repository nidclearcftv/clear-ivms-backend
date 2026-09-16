-- Vehicles' external_id was globally unique; it's now only required to be
-- unique within its own organization, so two different organizations can
-- each register a device under the same vendor-side ID.
ALTER TABLE vehicles DROP CONSTRAINT vehicles_external_id_key;
ALTER TABLE vehicles ADD CONSTRAINT uq_vehicles_organization_external_id UNIQUE (organization_id, external_id);
