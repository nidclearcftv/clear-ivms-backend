-- Adds equipment_models.picture_object_key: a reference to the picture's
-- object in whichever port.ObjectStorage adapter is configured (see
-- adapter/storage/local). Nullable — most rows won't have a picture set.
ALTER TABLE equipment_models ADD COLUMN picture_object_key TEXT;
