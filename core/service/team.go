package service

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
	"github.com/nidclearcftv/clear-ivms-backend/utils"
	"github.com/nidclearcftv/clear-ivms-backend/utils/validate"
)

type TeamServiceOptions struct {
	Repository port.TeamRepository `validate:"required"`
}

// TeamService implements port.TeamService by delegating directly to a
// port.TeamRepository.
type TeamService struct {
	repo port.TeamRepository
}

func NewTeamService(opts TeamServiceOptions) (*TeamService, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	return &TeamService{repo: opts.Repository}, nil
}

func (s *TeamService) Create(ctx context.Context, team model.Team) (model.Team, error) {
	return s.repo.Create(ctx, team)
}

func (s *TeamService) Get(ctx context.Context, id model.ID) (model.Team, error) {
	return s.repo.Get(ctx, id)
}

// List always scopes to the current organization: filters.OrganizationID
// is overwritten from ctx, never trusted from the caller.
func (s *TeamService) List(ctx context.Context, filters model.TeamFilters) (model.List[model.Team], error) {
	filters.OrganizationID = utils.OrganizationID(ctx)
	return s.repo.List(ctx, filters)
}

func (s *TeamService) Update(ctx context.Context, team model.Team) (model.Team, error) {
	return s.repo.Update(ctx, team)
}

func (s *TeamService) Delete(ctx context.Context, id model.ID) error {
	return s.repo.Delete(ctx, id)
}

func (s *TeamService) ListFromAccount(ctx context.Context, accountID model.ID) (model.List[model.Team], error) {
	return s.repo.ListFromAccount(ctx, accountID)
}

func (s *TeamService) AddAccount(ctx context.Context, teamID, accountID model.ID) error {
	return s.repo.AddAccount(ctx, teamID, accountID)
}

func (s *TeamService) RemoveAccount(ctx context.Context, teamID, accountID model.ID) error {
	return s.repo.RemoveAccount(ctx, teamID, accountID)
}

func (s *TeamService) ListFromFleet(ctx context.Context, fleetID model.ID) (model.List[model.Team], error) {
	return s.repo.ListFromFleet(ctx, fleetID)
}

func (s *TeamService) AddFleet(ctx context.Context, teamID, fleetID model.ID) error {
	return s.repo.AddFleet(ctx, teamID, fleetID)
}

func (s *TeamService) RemoveFleet(ctx context.Context, teamID, fleetID model.ID) error {
	return s.repo.RemoveFleet(ctx, teamID, fleetID)
}

var _ port.TeamService = (*TeamService)(nil)
