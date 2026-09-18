-- Adds vehicle_equipment, the relation between vehicles and the
-- organization's equipment catalog (equipment_models): one row per
-- physical unit registered on a vehicle. equipment_model_type is
-- snapshotted from equipment_models.type at registration time and never
-- re-derived — see core/model/vehicle_equipment.go.
CREATE TABLE vehicle_equipment (
    id                   UUID        PRIMARY KEY DEFAULT uuidv7(),
    vehicle_id           UUID        NOT NULL,
    equipment_model_id   UUID        NOT NULL,
    equipment_model_type TEXT        NOT NULL CHECK (equipment_model_type IN ('primary', 'accessory')),
    serial_number        TEXT        NOT NULL,
    description          TEXT        NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_vehicle_equipment_vehicle FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE,
    CONSTRAINT fk_vehicle_equipment_equipment_model FOREIGN KEY (equipment_model_id) REFERENCES equipment_models(id) ON DELETE RESTRICT
);

CREATE INDEX idx_vehicle_equipment_vehicle ON vehicle_equipment (vehicle_id);
CREATE INDEX idx_vehicle_equipment_equipment_model ON vehicle_equipment (equipment_model_id);
-- Enforces "at most one primary equipment per vehicle" at the database
-- level — see VehicleEquipmentService.Create.
CREATE UNIQUE INDEX idx_vehicle_equipment_one_primary_per_vehicle ON vehicle_equipment (vehicle_id) WHERE equipment_model_type = 'primary';
