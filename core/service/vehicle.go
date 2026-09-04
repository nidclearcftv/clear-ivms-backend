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

func (s *VehicleService) Get(ctx context.Context, id model.ID) (model.Vehicle, error) {
	return s.repo.Get(ctx, id)
}

// List always scopes to the current organization: filters.OrganizationID
// is overwritten from ctx, never trusted from the caller.
func (s *VehicleService) List(ctx context.Context, filters model.VehicleFilters) (model.List[model.Vehicle], error) {
	filters.OrganizationID = utils.OrganizationID(ctx)
	return s.repo.List(ctx, filters)
}

func (s *VehicleService) Update(ctx context.Context, vehicle model.Vehicle) (model.Vehicle, error) {
	return s.repo.Update(ctx, vehicle)
}

func (s *VehicleService) Delete(ctx context.Context, id model.ID) error {
	return s.repo.Delete(ctx, id)
}

var _ port.VehicleService = (*VehicleService)(nil)
