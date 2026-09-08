package service

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
	"github.com/nidclearcftv/clear-ivms-backend/utils"
	"github.com/nidclearcftv/clear-ivms-backend/utils/validate"
)

type BrandingServiceOptions struct {
	Repository port.BrandingRepository `validate:"required"`
	Cache      port.Cache              `validate:"required"`
}

// BrandingService implements port.BrandingService by delegating directly
// to a port.BrandingRepository, fronted by a port.Cache (see Get,
// GetByDomain) — GetByDomain in particular backs the public,
// unauthenticated branding lookup route, so caching it keeps that route
// cheap under repeated hits for the same domain.
type BrandingService struct {
	repo  port.BrandingRepository
	cache port.Cache
}

func NewBrandingService(opts BrandingServiceOptions) (*BrandingService, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	return &BrandingService{repo: opts.Repository, cache: opts.Cache}, nil
}

func (s *BrandingService) Create(ctx context.Context, branding model.Branding) (model.Branding, error) {
	return s.repo.Create(ctx, branding)
}

func (s *BrandingService) Get(ctx context.Context, id model.ID) (model.Branding, error) {
	return utils.FetchThrough(ctx, s.cache, model.BrandingKey(id), func() (model.Branding, error) {
		return s.repo.Get(ctx, id)
	})
}

// GetByDomain is cached separately from Get, keyed by domain instead of ID
// — see Update/Delete for how both entries get invalidated together.
func (s *BrandingService) GetByDomain(ctx context.Context, domain string) (model.Branding, error) {
	return utils.FetchThrough(ctx, s.cache, model.BrandingDomainKey(domain), func() (model.Branding, error) {
		return s.repo.GetByDomain(ctx, domain)
	})
}

func (s *BrandingService) List(ctx context.Context, filters model.BrandingFilters) (model.List[model.Branding], error) {
	return s.repo.List(ctx, filters)
}

func (s *BrandingService) Count(ctx context.Context, filters model.BrandingFilters) (int, error) {
	return s.repo.Count(ctx, filters)
}

// Update invalidates both cache entries Get/GetByDomain populate: the
// id-keyed one always, and the domain-keyed one for both the branding's
// old and new domain (in case Domain itself changed) — otherwise a stale
// config could keep being served from whichever domain key wasn't
// invalidated.
func (s *BrandingService) Update(ctx context.Context, branding model.Branding) (model.Branding, error) {
	existing, err := s.repo.Get(ctx, branding.ID)
	if err != nil {
		return model.Branding{}, err
	}

	updated, err := s.repo.Update(ctx, branding)
	if err != nil {
		return model.Branding{}, err
	}

	s.cache.Del(ctx, model.BrandingKey(branding.ID))
	s.cache.Del(ctx, model.BrandingDomainKey(existing.Domain))
	if updated.Domain != existing.Domain {
		s.cache.Del(ctx, model.BrandingDomainKey(updated.Domain))
	}

	return updated, nil
}

// Delete invalidates both cache entries Get/GetByDomain populate; see
// Update.
func (s *BrandingService) Delete(ctx context.Context, id model.ID) error {
	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	s.cache.Del(ctx, model.BrandingKey(id))
	s.cache.Del(ctx, model.BrandingDomainKey(existing.Domain))

	return nil
}

var _ port.BrandingService = (*BrandingService)(nil)
