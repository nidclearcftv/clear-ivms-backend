package service

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
	"github.com/nidclearcftv/clear-ivms-backend/utils"
	"github.com/nidclearcftv/clear-ivms-backend/utils/validate"
)

type FleetServiceOptions struct {
	Repository port.FleetRepository `validate:"required"`
}

// FleetService implements port.FleetService by delegating directly to a
// port.FleetRepository.
type FleetService struct {
	repo port.FleetRepository
}

func NewFleetService(opts FleetServiceOptions) (*FleetService, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	return &FleetService{repo: opts.Repository}, nil
}

func (s *FleetService) Create(ctx context.Context, fleet model.Fleet) (model.Fleet, error) {
	return s.repo.Create(ctx, fleet)
}

func (s *FleetService) Get(ctx context.Context, id model.ID) (model.Fleet, error) {
	return s.repo.Get(ctx, id)
}

// List always scopes to the current organization: filters.OrganizationID
// is overwritten from ctx, never trusted from the caller.
func (s *FleetService) List(ctx context.Context, filters model.FleetFilters) (model.List[model.Fleet], error) {
	filters.OrganizationID = utils.OrganizationID(ctx)
	return s.repo.List(ctx, filters)
}

func (s *FleetService) Update(ctx context.Context, fleet model.Fleet) (model.Fleet, error) {
	return s.repo.Update(ctx, fleet)
}

func (s *FleetService) Delete(ctx context.Context, id model.ID) error {
	return s.repo.Delete(ctx, id)
}

func (s *FleetService) ListFromTeam(ctx context.Context, teamID model.ID) (model.List[model.Fleet], error) {
	return s.repo.ListFromTeam(ctx, teamID)
}

func (s *FleetService) AddTeam(ctx context.Context, fleetID, teamID model.ID) error {
	return s.repo.AddTeam(ctx, fleetID, teamID)
}

func (s *FleetService) RemoveTeam(ctx context.Context, fleetID, teamID model.ID) error {
	return s.repo.RemoveTeam(ctx, fleetID, teamID)
}

var _ port.FleetService = (*FleetService)(nil)
