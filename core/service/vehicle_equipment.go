package service

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
	"github.com/nidclearcftv/clear-ivms-backend/utils"
	"github.com/nidclearcftv/clear-ivms-backend/utils/validate"
)

type VehicleEquipmentServiceOptions struct {
	Repository port.VehicleEquipmentRepository `validate:"required"`
	// Vehicles backs the ownership checks every method here does — this
	// entity has no organization_id of its own, so its organization is
	// always derived from the vehicle it belongs to.
	Vehicles port.VehicleRepository `validate:"required"`
	// EquipmentModels backs Create's lookup of the equipment model being
	// registered — used to snapshot its Type and to check it's visible
	// to the caller (own organization or public), the same relaxed check
	// EquipmentModelService.Get does.
	EquipmentModels port.EquipmentModelRepository `validate:"required"`
}

// VehicleEquipmentService implements port.VehicleEquipmentService by
// delegating to a port.VehicleEquipmentRepository, with ownership checks
// backed directly by port.VehicleRepository and port.EquipmentModelRepository
// — mirroring how GroupService depends on VehicleRepository directly for
// its own AddVehicle/RemoveVehicle checks, rather than on VehicleService.
type VehicleEquipmentService struct {
	repo            port.VehicleEquipmentRepository
	vehicles        port.VehicleRepository
	equipmentModels port.EquipmentModelRepository
}

func NewVehicleEquipmentService(opts VehicleEquipmentServiceOptions) (*VehicleEquipmentService, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	return &VehicleEquipmentService{
		repo:            opts.Repository,
		vehicles:        opts.Vehicles,
		equipmentModels: opts.EquipmentModels,
	}, nil
}

// checkVehicleOwnership fails with ErrCodeVehicleEquipmentNotFound if
// vehicleID doesn't belong to the request's organization — used by every
// method here that operates on an existing vehicle_equipment row, so a
// caller can't distinguish "this row doesn't exist" from "it belongs to
// another organization's vehicle." Create uses its own, separate check
// instead (see below) since it's creating a new reference to the vehicle,
// not looking up an existing registration.
func (s *VehicleEquipmentService) checkVehicleOwnership(ctx context.Context, vehicleID model.ID) error {
	vehicle, err := s.vehicles.Get(ctx, vehicleID)
	if err != nil {
		return err
	}
	if vehicle.OrganizationID != utils.OrganizationID(ctx) {
		return model.NewError(model.ErrCodeVehicleEquipmentNotFound, nil)
	}
	return nil
}

// Create fails with ErrCodeVehicleNotFound if ve.VehicleID doesn't belong
// to the request's organization, or ErrCodeEquipmentModelNotFound if
// ve.EquipmentModelID doesn't exist or isn't visible to it (own
// organization or public). ve.EquipmentModelType is always overwritten
// from the referenced equipment model's own Type — see
// port.VehicleEquipmentService.Create.
func (s *VehicleEquipmentService) Create(ctx context.Context, ve model.VehicleEquipment) (model.VehicleEquipment, error) {
	vehicle, err := s.vehicles.Get(ctx, ve.VehicleID)
	if err != nil {
		return model.VehicleEquipment{}, err
	}
	if vehicle.OrganizationID != utils.OrganizationID(ctx) {
		return model.VehicleEquipment{}, model.NewError(model.ErrCodeVehicleNotFound, nil)
	}

	equipmentModel, err := s.equipmentModels.Get(ctx, ve.EquipmentModelID)
	if err != nil {
		return model.VehicleEquipment{}, err
	}
	if equipmentModel.OrganizationID != utils.OrganizationID(ctx) && !equipmentModel.Public {
		return model.VehicleEquipment{}, model.NewError(model.ErrCodeEquipmentModelNotFound, nil)
	}
	ve.EquipmentModelType = equipmentModel.Type

	if equipmentModel.Type == model.EquipmentModelTypeAccessory {
		count, err := s.repo.CountByType(ctx, ve.VehicleID, model.EquipmentModelTypeAccessory)
		if err != nil {
			return model.VehicleEquipment{}, err
		}
		if count >= model.VehicleEquipmentMaxAccessoriesPerVehicle {
			return model.VehicleEquipment{}, model.NewError(model.ErrCodeVehicleEquipmentAccessoryLimitReached, nil)
		}
	}
	// The primary cap (at most one) is enforced by a DB constraint
	// instead of a pre-check here — see
	// idx_vehicle_equipment_one_primary_per_vehicle in schema.sql and
	// VehicleEquipmentRepository.Create's handling of its violation.

	return s.repo.Create(ctx, ve)
}

// Get fails with ErrCodeVehicleEquipmentNotFound if id doesn't exist, or
// exists but belongs to a vehicle outside the request's organization —
// reported the same way, so a caller can't distinguish the two.
func (s *VehicleEquipmentService) Get(ctx context.Context, id model.ID) (model.VehicleEquipment, error) {
	ve, err := s.repo.Get(ctx, id)
	if err != nil {
		return model.VehicleEquipment{}, err
	}
	if err := s.checkVehicleOwnership(ctx, ve.VehicleID); err != nil {
		return model.VehicleEquipment{}, err
	}
	return ve, nil
}

// List fails with ErrCodeVehicleNotFound if vehicleID doesn't belong to
// the request's organization.
func (s *VehicleEquipmentService) List(ctx context.Context, vehicleID model.ID) ([]model.VehicleEquipment, error) {
	vehicle, err := s.vehicles.Get(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	if vehicle.OrganizationID != utils.OrganizationID(ctx) {
		return nil, model.NewError(model.ErrCodeVehicleNotFound, nil)
	}
	return s.repo.List(ctx, vehicleID)
}

// Update fails with ErrCodeVehicleEquipmentNotFound the same way Get does.
// ve.VehicleID, ve.EquipmentModelID and ve.EquipmentModelType are always
// overwritten with their existing values — only SerialNumber/Description
// are actually editable; see port.VehicleEquipmentService.Update.
func (s *VehicleEquipmentService) Update(ctx context.Context, ve model.VehicleEquipment) (model.VehicleEquipment, error) {
	existing, err := s.repo.Get(ctx, ve.ID)
	if err != nil {
		return model.VehicleEquipment{}, err
	}
	if err := s.checkVehicleOwnership(ctx, existing.VehicleID); err != nil {
		return model.VehicleEquipment{}, err
	}

	ve.VehicleID = existing.VehicleID
	ve.EquipmentModelID = existing.EquipmentModelID
	ve.EquipmentModelType = existing.EquipmentModelType
	return s.repo.Update(ctx, ve)
}

// Delete fails with ErrCodeVehicleEquipmentNotFound the same way Get does.
func (s *VehicleEquipmentService) Delete(ctx context.Context, id model.ID) error {
	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.checkVehicleOwnership(ctx, existing.VehicleID); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

var _ port.VehicleEquipmentService = (*VehicleEquipmentService)(nil)
