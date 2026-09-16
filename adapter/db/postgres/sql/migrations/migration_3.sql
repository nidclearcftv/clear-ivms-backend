-- Removes vehicles.ivms_type: vehicles no longer track which IVMS vendor
-- sourced them. external_id keeps its own UNIQUE constraint, which already
-- made the composite constraint below redundant.
ALTER TABLE vehicles DROP CONSTRAINT uq_vehicles_ivms_type_external_id;
ALTER TABLE vehicles DROP COLUMN ivms_type;
