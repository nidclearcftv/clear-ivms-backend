package port

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// BrandingRepository is the driven (secondary) port for persisting and
// reading branding data. It is implemented by outbound adapters, e.g.
// adapter/db/postgres.
type BrandingRepository interface {
	Create(ctx context.Context, branding model.Branding) (model.Branding, error)
	Get(ctx context.Context, id model.ID) (model.Branding, error)
	// GetByDomain looks up a branding by its (unique) domain — what the
	// public branding lookup route has on hand.
	GetByDomain(ctx context.Context, domain string) (model.Branding, error)
	List(ctx context.Context, filters model.BrandingFilters) (model.List[model.Branding], error)
	// Count reports how many brandings match filters — the same filters
	// List accepts. List uses this to fill model.List.Total.
	Count(ctx context.Context, filters model.BrandingFilters) (int, error)
	Update(ctx context.Context, branding model.Branding) (model.Branding, error)
	Delete(ctx context.Context, id model.ID) error
}

// BrandingService is the driving (primary) port exposing branding-related
// business operations to inbound adapters, e.g. adapter/http controllers.
type BrandingService interface {
	Create(ctx context.Context, branding model.Branding) (model.Branding, error)
	Get(ctx context.Context, id model.ID) (model.Branding, error)
	GetByDomain(ctx context.Context, domain string) (model.Branding, error)
	List(ctx context.Context, filters model.BrandingFilters) (model.List[model.Branding], error)
	Count(ctx context.Context, filters model.BrandingFilters) (int, error)
	Update(ctx context.Context, branding model.Branding) (model.Branding, error)
	Delete(ctx context.Context, id model.ID) error
}
