// Package port declares the application's ports: interfaces that decouple
// core business logic from the adapters that implement it, per the
// hexagonal/ports-and-adapters pattern.
package port

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// VehicleRepository is the driven (secondary) port for persisting and
// reading vehicle data. It is implemented by outbound adapters, e.g.
// adapter/db/postgres.
type VehicleRepository interface {
	Create(ctx context.Context, vehicle model.Vehicle) (model.Vehicle, error)
	Get(ctx context.Context, id model.ID) (model.Vehicle, error)
	List(ctx context.Context, filters model.VehicleFilters) (model.List[model.Vehicle], error)
	// Count reports how many vehicles match filters — the same filters
	// List accepts. List uses this to fill model.List.Total.
	Count(ctx context.Context, filters model.VehicleFilters) (int, error)
	// ListAll returns every vehicle in organizationID, ordered by plate
	// number — unpaginated; see GroupRepository.ListAll for why.
	ListAll(ctx context.Context, organizationID model.ID) ([]model.Vehicle, error)
	// Update never touches status — see SetStatus.
	Update(ctx context.Context, vehicle model.Vehicle) (model.Vehicle, error)
	Delete(ctx context.Context, id model.ID) error
	// SetStatus is the only way to change a vehicle's status; Update
	// deliberately excludes it.
	SetStatus(ctx context.Context, id model.ID, status model.VehicleStatus) error
	// SetStatusByExternalID is SetStatus keyed by the vehicle's external ID
	// instead of its internal ID — what a vendor status webhook/poller has
	// on hand. external_id is only unique per organization, so if it's
	// reused across organizations this updates every vehicle that matches
	// it.
	SetStatusByExternalID(ctx context.Context, externalID string, status model.VehicleStatus) error
}

// VehicleService is the driving (primary) port exposing vehicle-related
// business operations to inbound adapters, e.g. adapter/http controllers.
type VehicleService interface {
	Create(ctx context.Context, vehicle model.Vehicle) (model.Vehicle, error)
	Get(ctx context.Context, id model.ID) (model.Vehicle, error)
	List(ctx context.Context, filters model.VehicleFilters) (model.List[model.Vehicle], error)
	Count(ctx context.Context, filters model.VehicleFilters) (int, error)
	// ListAll returns every vehicle belonging to the current request's
	// organization (see utils.OrganizationID), unpaginated, ordered by
	// plate number.
	ListAll(ctx context.Context) ([]model.Vehicle, error)
	Update(ctx context.Context, vehicle model.Vehicle) (model.Vehicle, error)
	Delete(ctx context.Context, id model.ID) error
	SetStatus(ctx context.Context, id model.ID, status model.VehicleStatus) error
	SetStatusByExternalID(ctx context.Context, externalID string, status model.VehicleStatus) error
}
