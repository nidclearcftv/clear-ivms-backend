// Package influxdb is the InfluxDB adapter: a shared client connection
// (this file) plus entity-specific repositories built on top of it, e.g.
// repository_vehicle_telemetry.go — mirrors adapter/db/postgres's split
// between db.go and its own repository_*.go files.
package influxdb

import (
	"fmt"

	"github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"
	"go.uber.org/zap"

	"github.com/nidclearcftv/clear-ivms-backend/utils/validate"
)

type Options struct {
	Logger *zap.SugaredLogger

	// Host is the InfluxDB 3 Core server's base URL, e.g.
	// "http://localhost:8181".
	Host string `validate:"required"`
	// Token authenticates against Host — an admin or database token; see
	// https://docs.influxdata.com/influxdb3/core/admin/tokens/.
	Token string `validate:"required"`
	// Database is the target database name. InfluxDB 3 Core sets
	// retention per-database at creation time (not per-table), so this
	// must name a database already created with the desired retention —
	// see repository_vehicle_telemetry.go's doc comment.
	Database string `validate:"required"`
}

// Client wraps an influxdb3.Client scoped to a single database, configured
// via Options.
type Client struct {
	raw  *influxdb3.Client
	opts Options
	log  *zap.SugaredLogger
}

func NewClient(opts Options) (*Client, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	raw, err := influxdb3.New(influxdb3.ClientConfig{
		Host:     opts.Host,
		Token:    opts.Token,
		Database: opts.Database,
	})
	if err != nil {
		return nil, fmt.Errorf("influxdb: failed to create client: %w", err)
	}

	log := opts.Logger
	if log == nil {
		log = zap.NewNop().Sugar()
	}
	log = log.With("adapter", "influxdb")
	log.Infow("configured influxdb client", "database", opts.Database)

	return &Client{raw: raw, opts: opts, log: log}, nil
}

// Close releases the underlying client's idle connections.
func (c *Client) Close() error {
	return c.raw.Close()
}
