// Package postgres is the Postgres adapter: connection handling plus
// schema/migration bootstrapping for adapter/db/postgres/sql.
package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/nidclearcftv/clear-ivms-backend/utils/validate"
)

// currentVersion is the schema version this codebase expects. schema.sql
// bootstraps a fresh database directly to this version (its `version` row
// starts at 1); every version after that needs a matching
// migrations/migration_<version>.sql. Bump this whenever a new migration
// file is added.
const currentVersion = 1

type Options struct {
	Logger *zap.SugaredLogger

	ConnectionString string `validate:"required"`

	// SchemaPath is the base schema.sql applied to a brand-new database
	// (one with no `version` table yet).
	SchemaPath string `validate:"required"`

	// MigrationsPath holds migration_<n>.sql files, applied in order to
	// bring an existing database from its current version up to
	// currentVersion.
	MigrationsPath string `validate:"required"`
}

type DB struct {
	Pool *pgxpool.Pool

	opts Options
	log  *zap.SugaredLogger
}

func NewDB(ctx context.Context, opts Options) (*DB, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	pool, err := pgxpool.New(ctx, opts.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to create connection pool: %w", err)
	}

	log := opts.Logger
	if log == nil {
		log = zap.NewNop().Sugar()
	}
	log = log.With("adapter", "postgres")

	return &DB{Pool: pool, opts: opts, log: log}, nil
}

// Close releases the connection pool. Safe to call even if NewDB failed to
// fully initialize.
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// Initialize brings the database up to currentVersion: a brand-new
// database gets the full schema applied once; an existing one gets every
// pending migration applied in order.
func (db *DB) Initialize(ctx context.Context) error {
	exists, err := db.versionTableExists(ctx)
	if err != nil {
		return err
	}

	if !exists {
		return db.runSchema(ctx)
	}
	return db.runPendingMigrations(ctx)
}

func (db *DB) versionTableExists(ctx context.Context) (bool, error) {
	var exists bool
	err := db.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM pg_tables
			WHERE schemaname = 'public' AND tablename = 'version'
		)
	`).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("postgres: failed to check for version table: %w", err)
	}

	return exists, nil
}

func (db *DB) runSchema(ctx context.Context) error {
	schema, err := os.ReadFile(db.opts.SchemaPath)
	if err != nil {
		return fmt.Errorf("postgres: failed to read schema file: %w", err)
	}

	// The simple protocol is required here: schema.sql contains many
	// statements in one string, which the default extended protocol
	// (prepared statements) doesn't support.
	if _, err := db.Pool.Exec(ctx, string(schema), pgx.QueryExecModeSimpleProtocol); err != nil {
		return fmt.Errorf("postgres: failed to apply schema: %w", err)
	}

	db.log.Infow("applied schema", "version", currentVersion)
	return nil
}

func (db *DB) runPendingMigrations(ctx context.Context) error {
	var version int
	if err := db.Pool.QueryRow(ctx, `SELECT version FROM version LIMIT 1`).Scan(&version); err != nil {
		return fmt.Errorf("postgres: failed to read current schema version: %w", err)
	}

	for version < currentVersion {
		version++

		path := filepath.Join(db.opts.MigrationsPath, fmt.Sprintf("migration_%d.sql", version))
		migration, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("postgres: failed to read migration %d: %w", version, err)
		}

		if _, err := db.Pool.Exec(ctx, string(migration), pgx.QueryExecModeSimpleProtocol); err != nil {
			return fmt.Errorf("postgres: failed to apply migration %d: %w", version, err)
		}

		if _, err := db.Pool.Exec(ctx, `UPDATE version SET version = $1, updated_at = NOW()`, version); err != nil {
			return fmt.Errorf("postgres: failed to record migration %d: %w", version, err)
		}

		db.log.Infow("applied migration", "version", version)
	}

	return nil
}
