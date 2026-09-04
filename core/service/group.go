package service

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
	"github.com/nidclearcftv/clear-ivms-backend/utils"
	"github.com/nidclearcftv/clear-ivms-backend/utils/validate"
)

type GroupServiceOptions struct {
	Repository port.GroupRepository `validate:"required"`
}

// GroupService implements port.GroupService by delegating directly to a
// port.GroupRepository.
type GroupService struct {
	repo port.GroupRepository
}

func NewGroupService(opts GroupServiceOptions) (*GroupService, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	return &GroupService{repo: opts.Repository}, nil
}

func (s *GroupService) Create(ctx context.Context, group model.Group) (model.Group, error) {
	return s.repo.Create(ctx, group)
}

func (s *GroupService) Get(ctx context.Context, id model.ID) (model.Group, error) {
	return s.repo.Get(ctx, id)
}

// List always scopes to the current organization: filters.OrganizationID
// is overwritten from ctx, never trusted from the caller.
func (s *GroupService) List(ctx context.Context, filters model.GroupFilters) (model.List[model.Group], error) {
	filters.OrganizationID = utils.OrganizationID(ctx)
	return s.repo.List(ctx, filters)
}

func (s *GroupService) Update(ctx context.Context, group model.Group) (model.Group, error) {
	return s.repo.Update(ctx, group)
}

func (s *GroupService) Delete(ctx context.Context, id model.ID) error {
	return s.repo.Delete(ctx, id)
}

var _ port.GroupService = (*GroupService)(nil)
