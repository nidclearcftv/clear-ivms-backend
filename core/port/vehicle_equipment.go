package port

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// VehicleEquipmentRepository is the driven (secondary) port for persisting
// and reading vehicle equipment registrations. It is implemented by
// outbound adapters, e.g. adapter/db/postgres.
type VehicleEquipmentRepository interface {
	Create(ctx context.Context, ve model.VehicleEquipment) (model.VehicleEquipment, error)
	Get(ctx context.Context, id model.ID) (model.VehicleEquipment, error)
	// List returns every equipment registration for vehicleID, primary
	// first (there is ever at most one) then accessories ordered by
	// created_at — unpaginated; see
	// model.VehicleEquipmentMaxAccessoriesPerVehicle for why that's safe.
	List(ctx context.Context, vehicleID model.ID) ([]model.VehicleEquipment, error)
	// CountByType reports how many of vehicleID's registrations have the
	// given equipmentModelType — used to enforce
	// model.VehicleEquipmentMaxAccessoriesPerVehicle before an accessory
	// insert.
	CountByType(ctx context.Context, vehicleID model.ID, equipmentModelType model.EquipmentModelType) (int, error)
	// Update never touches vehicle_id, equipment_model_id or
	// equipment_model_type — see VehicleEquipmentService.Update.
	Update(ctx context.Context, ve model.VehicleEquipment) (model.VehicleEquipment, error)
	Delete(ctx context.Context, id model.ID) error
}

// VehicleEquipmentService is the driving (primary) port exposing vehicle
// equipment-related business operations to inbound adapters, e.g.
// adapter/http controllers.
type VehicleEquipmentService interface {
	// Create registers a new equipment unit on ve.VehicleID, referencing
	// ve.EquipmentModelID — ve.EquipmentModelType is always overwritten
	// from that equipment model's own Type, never trusted from the
	// caller. Fails with ErrCodeVehicleEquipmentPrimaryAlreadyExists if
	// the referenced equipment model is primary-type and the vehicle
	// already has one registered, or with
	// ErrCodeVehicleEquipmentAccessoryLimitReached if it's accessory-type
	// and the vehicle already has model.VehicleEquipmentMaxAccessoriesPerVehicle
	// of them.
	Create(ctx context.Context, ve model.VehicleEquipment) (model.VehicleEquipment, error)
	Get(ctx context.Context, id model.ID) (model.VehicleEquipment, error)
	// List returns every equipment registration for vehicleID. Fails with
	// ErrCodeVehicleNotFound if vehicleID doesn't belong to the request's
	// organization.
	List(ctx context.Context, vehicleID model.ID) ([]model.VehicleEquipment, error)
	// Update only ever changes SerialNumber/Description — see
	// VehicleEquipmentRepository.Update.
	Update(ctx context.Context, ve model.VehicleEquipment) (model.VehicleEquipment, error)
	Delete(ctx context.Context, id model.ID) error
}
