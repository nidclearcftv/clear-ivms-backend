-- Adds equipment_models.features, a caller-defined tag array used for
-- FeaturesInclude/FeaturesExclude filtering (see applyEquipmentModelFilters).
-- Existing rows backfill to '{}' — the same default new rows get.
ALTER TABLE equipment_models ADD COLUMN features TEXT[] NOT NULL DEFAULT '{}';

CREATE INDEX idx_equipment_models_features ON equipment_models USING GIN (features);
