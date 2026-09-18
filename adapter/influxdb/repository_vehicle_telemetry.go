package influxdb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// vehicleTelemetryFields lists every valid model.VehicleTelemetryFilter.
// Fields value — also the allow-list that keeps Query's field selection
// safe to interpolate directly into SQL as column identifiers, since
// identifiers (unlike values) can't be passed as bind parameters.
var vehicleTelemetryFields = map[string]bool{
	"latitude":  true,
	"longitude": true,
	"altitude":  true,
	"angle":     true,
}

// vehicleTelemetryAggrFns maps model.VehicleTelemetryAggrFn to the SQL
// aggregate function name it runs — also an allow-list, for the same
// identifier-interpolation reason as vehicleTelemetryFields.
var vehicleTelemetryAggrFns = map[model.VehicleTelemetryAggrFn]string{
	model.VehicleTelemetryAggrFnMean:  "AVG",
	model.VehicleTelemetryAggrFnMin:   "MIN",
	model.VehicleTelemetryAggrFnMax:   "MAX",
	model.VehicleTelemetryAggrFnSum:   "SUM",
	model.VehicleTelemetryAggrFnCount: "COUNT",
}

// vehicleTelemetryMeasurement is the InfluxDB table (measurement) storing
// model.VehicleTelemetry: tag vehicle_id; fields latitude, longitude,
// altitude, angle; timestamp time. It lives in a database that must be
// created ahead of time with a 5-year retention period — InfluxDB 3 Core
// only supports retention per-database, not per-table, and retention can't
// be changed after creation:
//
//	influxdb3 create database --retention-period 5y <database>
//
// (or the HTTP equivalent, POST /api/v3/configure/database — see
// https://docs.influxdata.com/influxdb3/core/admin/databases/).
const vehicleTelemetryMeasurement = "vehicle_telemetry"

// VehicleTelemetryRepository implements port.VehicleTelemetryRepository
// against InfluxDB.
type VehicleTelemetryRepository struct {
	client *Client
}

func NewVehicleTelemetryRepository(client *Client) *VehicleTelemetryRepository {
	return &VehicleTelemetryRepository{client: client}
}

func (r *VehicleTelemetryRepository) Insert(ctx context.Context, telemetry model.VehicleTelemetry) error {
	fields := map[string]any{}
	if telemetry.Latitude != nil {
		fields["latitude"] = *telemetry.Latitude
	}
	if telemetry.Longitude != nil {
		fields["longitude"] = *telemetry.Longitude
	}
	if telemetry.Altitude != nil {
		fields["altitude"] = *telemetry.Altitude
	}
	if telemetry.Angle != nil {
		fields["angle"] = *telemetry.Angle
	}
	// A point with zero fields is rejected by InfluxDB (tags alone aren't
	// enough) — caught here with a clearer message than that remote error.
	if len(fields) == 0 {
		return fmt.Errorf("influxdb: telemetry must set at least one field")
	}

	point := influxdb3.NewPoint(
		vehicleTelemetryMeasurement,
		map[string]string{
			"vehicle_id": string(telemetry.VehicleID),
		},
		fields,
		telemetry.Time,
	)

	if err := r.client.raw.WritePoints(ctx, []*influxdb3.Point{point}); err != nil {
		return fmt.Errorf("influxdb: failed to insert vehicle telemetry: %w", err)
	}
	return nil
}

func (r *VehicleTelemetryRepository) Query(ctx context.Context, filter model.VehicleTelemetryFilter) ([]model.VehicleTelemetry, error) {
	if filter.Limit <= 0 {
		return nil, fmt.Errorf("influxdb: limit must be greater than zero")
	}

	fields := filter.Fields
	if len(fields) == 0 {
		fields = []string{"latitude", "longitude", "altitude", "angle"}
	}
	for _, field := range fields {
		if !vehicleTelemetryFields[field] {
			return nil, fmt.Errorf("influxdb: unknown vehicle telemetry field %q", field)
		}
	}

	selectColumns := []string{"vehicle_id"}
	var groupBy string
	if filter.Aggr > 0 {
		sqlFn, ok := vehicleTelemetryAggrFns[filter.AggrFn]
		if !ok {
			return nil, fmt.Errorf("influxdb: unknown vehicle telemetry aggregation function %q", filter.AggrFn)
		}

		seconds := int64(filter.Aggr.Seconds())
		if seconds <= 0 {
			return nil, fmt.Errorf("influxdb: aggr must be at least one second")
		}

		// DATE_BIN buckets each row's time into fixed windows — the
		// standard SQL downsampling pattern for InfluxDB 3 (see
		// https://docs.influxdata.com/influxdb3/core/query-data/sql/aggregate-select/).
		// It's listed first so the positional "GROUP BY 1" below refers
		// to it regardless of how many fields follow.
		selectColumns[0] = fmt.Sprintf("DATE_BIN(INTERVAL '%d seconds', time) AS time", seconds)
		selectColumns = append(selectColumns, "vehicle_id")
		for _, field := range fields {
			selectColumns = append(selectColumns, fmt.Sprintf("%s(%s) AS %s", sqlFn, field, field))
		}
		groupBy = " GROUP BY 1, vehicle_id"
	} else {
		selectColumns = append(selectColumns, fields...)
		selectColumns = append(selectColumns, "time")
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE vehicle_id = $vehicle_id",
		strings.Join(selectColumns, ", "),
		vehicleTelemetryMeasurement,
	)
	parameters := influxdb3.QueryParameters{
		"vehicle_id": string(filter.VehicleID),
	}

	// From/To are embedded as quoted RFC3339 literals rather than bind
	// parameters: InfluxDB's SQL dialect accepts timestamp comparisons
	// written this way, and both values come from parsed time.Time, never
	// raw user text, so there's no injection risk in formatting them
	// straight into the query string.
	if !filter.From.IsZero() {
		query += fmt.Sprintf(" AND time >= '%s'", filter.From.Format(time.RFC3339Nano))
	}
	if !filter.To.IsZero() {
		query += fmt.Sprintf(" AND time <= '%s'", filter.To.Format(time.RFC3339Nano))
	}
	query += groupBy
	// Limit is a validated positive int formatted with %d, never
	// attacker-controlled text, so interpolating it here carries the same
	// no-injection-risk reasoning as From/To above.
	query += fmt.Sprintf(" ORDER BY time ASC LIMIT %d", filter.Limit)

	iterator, err := r.client.raw.QueryWithParameters(ctx, query, parameters)
	if err != nil {
		return nil, fmt.Errorf("influxdb: failed to query vehicle telemetry: %w", err)
	}

	var results []model.VehicleTelemetry
	for iterator.Next() {
		telemetry, err := vehicleTelemetryFromRow(iterator.Value())
		if err != nil {
			return nil, fmt.Errorf("influxdb: failed to read vehicle telemetry row: %w", err)
		}
		results = append(results, telemetry)
	}

	return results, nil
}

// vehicleTelemetryFromRow decodes one query result row. latitude,
// longitude, altitude and angle are all optional — Query only ever selects
// the subset requested via VehicleTelemetryFilter.Fields, so any of them
// may simply be absent from row rather than present-but-wrong-typed; the
// resulting model.VehicleTelemetry just leaves that field at its zero
// value.
func vehicleTelemetryFromRow(row map[string]any) (model.VehicleTelemetry, error) {
	vehicleID, ok := row["vehicle_id"].(string)
	if !ok {
		return model.VehicleTelemetry{}, fmt.Errorf("vehicle_id: expected string, got %T", row["vehicle_id"])
	}

	timestamp, ok := row["time"].(time.Time)
	if !ok {
		return model.VehicleTelemetry{}, fmt.Errorf("time: expected time.Time, got %T", row["time"])
	}

	telemetry := model.VehicleTelemetry{VehicleID: model.ID(vehicleID), Time: timestamp}

	var err error
	if telemetry.Latitude, err = vehicleTelemetryRowNumber(row, "latitude"); err != nil {
		return model.VehicleTelemetry{}, err
	}
	if telemetry.Longitude, err = vehicleTelemetryRowNumber(row, "longitude"); err != nil {
		return model.VehicleTelemetry{}, err
	}
	if telemetry.Altitude, err = vehicleTelemetryRowNumber(row, "altitude"); err != nil {
		return model.VehicleTelemetry{}, err
	}
	if telemetry.Angle, err = vehicleTelemetryRowNumber(row, "angle"); err != nil {
		return model.VehicleTelemetry{}, err
	}

	return telemetry, nil
}

// vehicleTelemetryRowNumber reads one optional numeric column, returning
// nil if it wasn't selected. Aggregated queries can produce int64 (e.g.
// COUNT) as well as float64 (e.g. AVG), so both are accepted.
func vehicleTelemetryRowNumber(row map[string]any, key string) (*float64, error) {
	value, ok := row[key]
	if !ok {
		return nil, nil
	}

	switch v := value.(type) {
	case float64:
		return &v, nil
	case int64:
		f := float64(v)
		return &f, nil
	default:
		return nil, fmt.Errorf("%s: expected a numeric value, got %T", key, value)
	}
}
