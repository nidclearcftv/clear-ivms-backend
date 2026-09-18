package model

import "time"

// VehicleEquipment is one physical unit of an equipment model registered on
// a vehicle — the join between vehicles and the organization's equipment
// catalog (see EquipmentModel), plus per-unit data specific to this
// installation.
//
// EquipmentModelType snapshots the referenced EquipmentModel's own Type at
// the moment this row was created; it's never re-derived afterward, even
// if that catalog entry's Type is edited later, so what's recorded here
// always reflects what was actually installed. It backs the "at most one
// primary, at most VehicleEquipmentMaxAccessoriesPerVehicle accessories per
// vehicle" caps — see VehicleEquipmentService.Create, the partial unique
// index on vehicle_equipment (vehicle_id) WHERE equipment_model_type =
// 'primary' in schema.sql, and VehicleEquipmentMaxAccessoriesPerVehicle.
type VehicleEquipment struct {
	ID                 ID
	VehicleID          ID
	EquipmentModelID   ID
	EquipmentModelType EquipmentModelType
	// SerialNumber and Description are both optional (default to "") —
	// free-form data about this specific installed unit, distinct from
	// the catalog-level EquipmentModel it references.
	SerialNumber string
	Description  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// VehicleEquipmentMaxAccessoriesPerVehicle bounds how many accessory-type
// registrations a single vehicle can have — enforced in
// VehicleEquipmentService.Create, since Postgres can't express a
// cross-row COUNT check declaratively the way the primary cap's partial
// unique index does. Combined with the primary cap (at most one), a
// vehicle can have at most 50 equipment registrations total.
const VehicleEquipmentMaxAccessoriesPerVehicle = 49
