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
	Update(ctx context.Context, vehicle model.Vehicle) (model.Vehicle, error)
	Delete(ctx context.Context, id model.ID) error
}

// VehicleService is the driving (primary) port exposing vehicle-related
// business operations to inbound adapters, e.g. adapter/http controllers.
type VehicleService interface {
	Create(ctx context.Context, vehicle model.Vehicle) (model.Vehicle, error)
	Get(ctx context.Context, id model.ID) (model.Vehicle, error)
	List(ctx context.Context, filters model.VehicleFilters) (model.List[model.Vehicle], error)
	Update(ctx context.Context, vehicle model.Vehicle) (model.Vehicle, error)
	Delete(ctx context.Context, id model.ID) error
}
