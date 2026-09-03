package service

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
	"github.com/nidclearcftv/clear-ivms-backend/utils/validate"
)

type OrganizationServiceOptions struct {
	Repository port.OrganizationRepository `validate:"required"`
}

// OrganizationService implements port.OrganizationService by delegating
// directly to a port.OrganizationRepository.
type OrganizationService struct {
	repo port.OrganizationRepository
}

func NewOrganizationService(opts OrganizationServiceOptions) (*OrganizationService, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	return &OrganizationService{repo: opts.Repository}, nil
}

func (s *OrganizationService) Create(ctx context.Context, organization model.Organization) (model.Organization, error) {
	return s.repo.Create(ctx, organization)
}

func (s *OrganizationService) Get(ctx context.Context, id model.ID) (model.Organization, error) {
	return s.repo.Get(ctx, id)
}

func (s *OrganizationService) List(ctx context.Context, filters model.OrganizationFilters) (model.List[model.Organization], error) {
	return s.repo.List(ctx, filters)
}

func (s *OrganizationService) Update(ctx context.Context, organization model.Organization) (model.Organization, error) {
	return s.repo.Update(ctx, organization)
}

func (s *OrganizationService) Delete(ctx context.Context, id model.ID) error {
	return s.repo.Delete(ctx, id)
}

func (s *OrganizationService) ListFromAccount(ctx context.Context, accountID model.ID) (model.List[model.Organization], error) {
	return s.repo.ListFromAccount(ctx, accountID)
}

func (s *OrganizationService) AddAccount(ctx context.Context, organizationID, accountID model.ID) error {
	return s.repo.AddAccount(ctx, organizationID, accountID)
}

func (s *OrganizationService) RemoveAccount(ctx context.Context, organizationID, accountID model.ID) error {
	return s.repo.RemoveAccount(ctx, organizationID, accountID)
}

var _ port.OrganizationService = (*OrganizationService)(nil)
