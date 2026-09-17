-- Speeds up the cross-organization "OR public = true" lookup added to
-- equipment model listings. Partial: only public models are ever looked
-- up by this column, so indexing just the true rows keeps it small
-- regardless of how many private models exist.
CREATE INDEX idx_equipment_models_public ON equipment_models (public) WHERE public;
