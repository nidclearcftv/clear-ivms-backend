// Package port declares the application's ports: interfaces that decouple
// core business logic from the adapters that implement it, per the
// hexagonal/ports-and-adapters pattern.
package port

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// VehicleRepository is the driven (secondary) port for reading vehicle
// data from a data source. It is implemented by outbound adapters, e.g.
// adapter/cmsv6.
type VehicleRepository interface {
	ListVehicles(ctx context.Context, filters model.VehicleFilters) ([]model.Vehicle, error)
}

// VehicleService is the driving (primary) port exposing vehicle-related
// business operations to inbound adapters, e.g. adapter/http controllers.
type VehicleService interface {
	ListVehicles(ctx context.Context, filters model.VehicleFilters) ([]model.Vehicle, error)
}
