package port

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// VehicleTelemetryRepository is the driven (secondary) port for a
// vehicle's raw positional telemetry (GPS fixes). Implemented by
// adapter/influxdb rather than adapter/db/postgres — see
// model.VehicleTelemetry's doc comment for why.
type VehicleTelemetryRepository interface {
	// Insert appends one telemetry sample.
	Insert(ctx context.Context, telemetry model.VehicleTelemetry) error
	// Query returns filter.VehicleID's samples within [filter.From,
	// filter.To], ordered by time ascending and capped at filter.Limit
	// rows. With filter.Aggr set, rows are windowed averages/min/max/etc.
	// (per filter.AggrFn) instead of raw samples — see
	// model.VehicleTelemetryFilter's doc comment.
	Query(ctx context.Context, filter model.VehicleTelemetryFilter) ([]model.VehicleTelemetry, error)
}
