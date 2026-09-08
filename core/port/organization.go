package port

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// OrganizationRepository is the driven (secondary) port for persisting and
// reading organization data. It is implemented by outbound adapters, e.g.
// adapter/db/postgres.
//
// Relation methods only ever return/accept Organization — e.g.
// ListFromAccount returns the organizations an account belongs to, not the
// account itself.
type OrganizationRepository interface {
	Create(ctx context.Context, organization model.Organization) (model.Organization, error)
	Get(ctx context.Context, id model.ID) (model.Organization, error)
	List(ctx context.Context, filters model.OrganizationFilters) (model.List[model.Organization], error)
	// Count reports how many organizations match filters — the same
	// filters List accepts. List uses this to fill model.List.Total.
	Count(ctx context.Context, filters model.OrganizationFilters) (int, error)
	Update(ctx context.Context, organization model.Organization) (model.Organization, error)
	Delete(ctx context.Context, id model.ID) error

	ListFromAccount(ctx context.Context, accountID model.ID) (model.List[model.Organization], error)
	// CountFromAccount is to ListFromAccount what Count is to List.
	CountFromAccount(ctx context.Context, accountID model.ID) (int, error)
	AddAccount(ctx context.Context, organizationID, accountID model.ID) error
	RemoveAccount(ctx context.Context, organizationID, accountID model.ID) error
}

// OrganizationService is the driving (primary) port exposing
// organization-related business operations to inbound adapters, e.g.
// adapter/http controllers.
type OrganizationService interface {
	Create(ctx context.Context, organization model.Organization) (model.Organization, error)
	Get(ctx context.Context, id model.ID) (model.Organization, error)
	List(ctx context.Context, filters model.OrganizationFilters) (model.List[model.Organization], error)
	Count(ctx context.Context, filters model.OrganizationFilters) (int, error)
	Update(ctx context.Context, organization model.Organization) (model.Organization, error)
	Delete(ctx context.Context, id model.ID) error

	ListFromAccount(ctx context.Context, accountID model.ID) (model.List[model.Organization], error)
	CountFromAccount(ctx context.Context, accountID model.ID) (int, error)
	AddAccount(ctx context.Context, organizationID, accountID model.ID) error
	RemoveAccount(ctx context.Context, organizationID, accountID model.ID) error
}
