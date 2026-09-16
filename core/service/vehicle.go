// Package service implements the application's driving ports (core/port)
// on top of driven ports, containing the business logic that sits between
// inbound adapters (e.g. HTTP) and outbound adapters (e.g. adapter/db/postgres).
package service

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
	"github.com/nidclearcftv/clear-ivms-backend/utils"
	"github.com/nidclearcftv/clear-ivms-backend/utils/validate"
)

type VehicleServiceOptions struct {
	Repository port.VehicleRepository `validate:"required"`
}

// VehicleService implements port.VehicleService by delegating directly to a
// port.VehicleRepository.
type VehicleService struct {
	repo port.VehicleRepository
}

func NewVehicleService(opts VehicleServiceOptions) (*VehicleService, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	return &VehicleService{repo: opts.Repository}, nil
}

func (s *VehicleService) Create(ctx context.Context, vehicle model.Vehicle) (model.Vehicle, error) {
	return s.repo.Create(ctx, vehicle)
}

// Get fails with ErrCodeVehicleNotFound if id exists but belongs to a
// different organization than the request's — reported the same as a
// nonexistent id, so a caller can't distinguish "not found" from
// "not yours" for another organization's vehicle.
func (s *VehicleService) Get(ctx context.Context, id model.ID) (model.Vehicle, error) {
	vehicle, err := s.repo.Get(ctx, id)
	if err != nil {
		return model.Vehicle{}, err
	}
	if vehicle.OrganizationID != utils.OrganizationID(ctx) {
		return model.Vehicle{}, model.NewError(model.ErrCodeVehicleNotFound, nil)
	}
	return vehicle, nil
}

// List always scopes to the current organization: filters.OrganizationID
// is overwritten from ctx, never trusted from the caller.
func (s *VehicleService) List(ctx context.Context, filters model.VehicleFilters) (model.List[model.Vehicle], error) {
	filters.OrganizationID = utils.OrganizationID(ctx)
	return s.repo.List(ctx, filters)
}

// Count scopes the same way List does.
func (s *VehicleService) Count(ctx context.Context, filters model.VehicleFilters) (int, error) {
	filters.OrganizationID = utils.OrganizationID(ctx)
	return s.repo.Count(ctx, filters)
}

// ListAll returns every vehicle belonging to the current request's
// organization, unpaginated, ordered by plate number.
func (s *VehicleService) ListAll(ctx context.Context) ([]model.Vehicle, error) {
	return s.repo.ListAll(ctx, utils.OrganizationID(ctx))
}

// Update fails with ErrCodeVehicleNotFound the same way Get does for a
// vehicle belonging to a different organization. vehicle.OrganizationID is
// always overwritten with the vehicle's existing (verified) organization —
// never trusted from the caller — so this can't be used to move a vehicle
// into a different organization.
func (s *VehicleService) Update(ctx context.Context, vehicle model.Vehicle) (model.Vehicle, error) {
	existing, err := s.repo.Get(ctx, vehicle.ID)
	if err != nil {
		return model.Vehicle{}, err
	}
	if existing.OrganizationID != utils.OrganizationID(ctx) {
		return model.Vehicle{}, model.NewError(model.ErrCodeVehicleNotFound, nil)
	}
	vehicle.OrganizationID = existing.OrganizationID
	return s.repo.Update(ctx, vehicle)
}

// Delete fails with ErrCodeVehicleNotFound the same way Get does for a
// vehicle belonging to a different organization.
func (s *VehicleService) Delete(ctx context.Context, id model.ID) error {
	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if existing.OrganizationID != utils.OrganizationID(ctx) {
		return model.NewError(model.ErrCodeVehicleNotFound, nil)
	}
	return s.repo.Delete(ctx, id)
}

func (s *VehicleService) SetStatus(ctx context.Context, id model.ID, status model.VehicleStatus) error {
	return s.repo.SetStatus(ctx, id, status)
}

func (s *VehicleService) SetStatusByExternalID(ctx context.Context, externalID string, status model.VehicleStatus) error {
	return s.repo.SetStatusByExternalID(ctx, externalID, status)
}

var _ port.VehicleService = (*VehicleService)(nil)
