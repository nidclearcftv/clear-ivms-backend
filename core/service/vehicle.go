// Package service implements the application's driving ports (core/port)
// on top of driven ports, containing the business logic that sits between
// inbound adapters (e.g. HTTP) and outbound adapters (e.g. cmsv6).
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
	Cache      port.Cache             `validate:"required"`
}

// VehicleService implements port.VehicleService by delegating to a
// port.VehicleRepository, fronted by a port.Cache and request coalescing
// (see utils.FetchThrough) so concurrent identical requests share one
// fetch. It's kept as its own layer, distinct from the repository, so
// business logic like this can be added without changing what inbound
// adapters depend on.
type VehicleService struct {
	repo  port.VehicleRepository
	cache port.Cache
}

func NewVehicleService(opts VehicleServiceOptions) (*VehicleService, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	return &VehicleService{repo: opts.Repository, cache: opts.Cache}, nil
}

// Vehicle request -> singleflight -> cache -> repository.
func (s *VehicleService) ListVehicles(ctx context.Context, filters model.VehicleFilters) ([]model.Vehicle, error) {
	agentID := utils.AgentID(ctx)
	key := model.VechicleListKey(agentID, filters)

	return utils.FetchThrough(ctx, s.cache, key, func() ([]model.Vehicle, error) {
		return s.repo.ListVehicles(ctx, filters)
	})
}

var _ port.VehicleService = (*VehicleService)(nil)
