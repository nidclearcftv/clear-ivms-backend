package model

import "time"

// VehicleTelemetry is one GPS/positional sample for a vehicle. Unlike most
// domain types in this package, it is not persisted to Postgres — see
// port.VehicleTelemetryRepository — but to a time-series store, since
// telemetry is high-volume, append-only, and always queried by time range
// rather than looked up by its own ID.
//
// Latitude/Longitude/Altitude/Angle are pointers so nil can mean "not
// selected" (VehicleTelemetryFilter.Fields didn't ask for it) as opposed to
// a real value of zero — Query never populates a field that wasn't
// requested, and Insert never writes one left nil.
type VehicleTelemetry struct {
	VehicleID ID
	Latitude  *float64
	Longitude *float64
	Altitude  *float64
	Angle     *float64
	Time      time.Time
}

// VehicleTelemetryAggrFn is a function applied to each field selected by a
// VehicleTelemetryFilter with Aggr set, one value per window.
type VehicleTelemetryAggrFn string

const (
	VehicleTelemetryAggrFnMean  VehicleTelemetryAggrFn = "mean"
	VehicleTelemetryAggrFnMin   VehicleTelemetryAggrFn = "min"
	VehicleTelemetryAggrFnMax   VehicleTelemetryAggrFn = "max"
	VehicleTelemetryAggrFnSum   VehicleTelemetryAggrFn = "sum"
	VehicleTelemetryAggrFnCount VehicleTelemetryAggrFn = "count"
)

// VehicleTelemetryFilter selects one vehicle's telemetry within a time
// range.
type VehicleTelemetryFilter struct {
	VehicleID ID
	// From/To are both inclusive; a zero value leaves that bound open.
	From time.Time
	To   time.Time
	// Limit caps how many rows are returned — or, with Aggr set, how many
	// windows. Required, must be greater than zero.
	Limit int
	// Fields restricts which of latitude/longitude/altitude/angle are
	// returned. Empty means all of them.
	Fields []string
	// Aggr bins samples into fixed windows of this duration instead of
	// returning raw samples, truncated to whole seconds. Zero (the
	// default) disables aggregation and returns raw samples instead.
	// Requires AggrFn.
	Aggr time.Duration
	// AggrFn is the function applied to each of Fields within each Aggr
	// window. Only meaningful, and required, when Aggr is set.
	AggrFn VehicleTelemetryAggrFn
}
