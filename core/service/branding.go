package service

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
	"github.com/nidclearcftv/clear-ivms-backend/utils/validate"
)

type BrandingServiceOptions struct {
	Repository port.BrandingRepository `validate:"required"`
}

// BrandingService implements port.BrandingService by delegating directly
// to a port.BrandingRepository.
type BrandingService struct {
	repo port.BrandingRepository
}

func NewBrandingService(opts BrandingServiceOptions) (*BrandingService, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	return &BrandingService{repo: opts.Repository}, nil
}

func (s *BrandingService) Create(ctx context.Context, branding model.Branding) (model.Branding, error) {
	return s.repo.Create(ctx, branding)
}

func (s *BrandingService) Get(ctx context.Context, id model.ID) (model.Branding, error) {
	return s.repo.Get(ctx, id)
}

func (s *BrandingService) GetByDomain(ctx context.Context, domain string) (model.Branding, error) {
	return s.repo.GetByDomain(ctx, domain)
}

func (s *BrandingService) List(ctx context.Context, filters model.BrandingFilters) (model.List[model.Branding], error) {
	return s.repo.List(ctx, filters)
}

func (s *BrandingService) Count(ctx context.Context, filters model.BrandingFilters) (int, error) {
	return s.repo.Count(ctx, filters)
}

func (s *BrandingService) Update(ctx context.Context, branding model.Branding) (model.Branding, error) {
	return s.repo.Update(ctx, branding)
}

func (s *BrandingService) Delete(ctx context.Context, id model.ID) error {
	return s.repo.Delete(ctx, id)
}

var _ port.BrandingService = (*BrandingService)(nil)
