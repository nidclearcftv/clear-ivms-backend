package postgres

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
)

var fleetColumns = []string{"id", "name", "organization_id", "created_at", "updated_at"}

// FleetRepository implements port.FleetRepository against Postgres.
type FleetRepository struct {
	db *DB
}

func NewFleetRepository(db *DB) *FleetRepository {
	return &FleetRepository{db: db}
}

func (r *FleetRepository) Create(ctx context.Context, fleet model.Fleet) (model.Fleet, error) {
	query, args, err := psql.Insert("fleets").
		Columns("name", "organization_id").
		Values(fleet.Name, string(fleet.OrganizationID)).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return model.Fleet{}, fmt.Errorf("postgres: failed to build create fleet query: %w", err)
	}

	var id string
	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(&id, &fleet.CreatedAt, &fleet.UpdatedAt)
	if err != nil {
		if foreignKeyViolationConstraint(err) == "fk_fleets_organization" {
			return model.Fleet{}, model.NewError(model.ErrCodeOrganizationNotFound, err)
		}
		return model.Fleet{}, fmt.Errorf("postgres: failed to create fleet: %w", err)
	}

	fleet.ID = model.ID(id)
	return fleet, nil
}

func (r *FleetRepository) Get(ctx context.Context, id model.ID) (model.Fleet, error) {
	query, args, err := psql.Select(fleetColumns...).
		From("fleets").
		Where(sq.Eq{"id": string(id)}).
		ToSql()
	if err != nil {
		return model.Fleet{}, fmt.Errorf("postgres: failed to build get fleet query: %w", err)
	}

	fleet, err := scanFleet(r.db.Pool.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Fleet{}, model.NewError(model.ErrCodeFleetNotFound, err)
		}
		return model.Fleet{}, fmt.Errorf("postgres: failed to get fleet: %w", err)
	}

	return fleet, nil
}

func (r *FleetRepository) List(ctx context.Context, filters model.FleetFilters) (model.List[model.Fleet], error) {
	builder := psql.Select(fleetColumns...).
		From("fleets").
		OrderBy("created_at DESC")

	if filters.OrganizationID != "" {
		builder = builder.Where(sq.Eq{"organization_id": string(filters.OrganizationID)})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return model.List[model.Fleet]{}, fmt.Errorf("postgres: failed to build list fleets query: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return model.List[model.Fleet]{}, fmt.Errorf("postgres: failed to list fleets: %w", err)
	}
	defer rows.Close()

	return scanFleetList(rows)
}

func (r *FleetRepository) Update(ctx context.Context, fleet model.Fleet) (model.Fleet, error) {
	query, args, err := psql.Update("fleets").
		Set("name", fleet.Name).
		Set("organization_id", string(fleet.OrganizationID)).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": string(fleet.ID)}).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return model.Fleet{}, fmt.Errorf("postgres: failed to build update fleet query: %w", err)
	}

	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(&fleet.CreatedAt, &fleet.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Fleet{}, model.NewError(model.ErrCodeFleetNotFound, err)
		}
		if foreignKeyViolationConstraint(err) == "fk_fleets_organization" {
			return model.Fleet{}, model.NewError(model.ErrCodeOrganizationNotFound, err)
		}
		return model.Fleet{}, fmt.Errorf("postgres: failed to update fleet: %w", err)
	}

	return fleet, nil
}

// Delete has no RESTRICT dependents to map: fleet_teams references fleets
// ON DELETE CASCADE and vehicles references fleets ON DELETE SET NULL, so
// neither can block deletion.
func (r *FleetRepository) Delete(ctx context.Context, id model.ID) error {
	query, args, err := psql.Delete("fleets").
		Where(sq.Eq{"id": string(id)}).
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build delete fleet query: %w", err)
	}

	tag, err := r.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("postgres: failed to delete fleet: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return model.NewError(model.ErrCodeFleetNotFound, nil)
	}

	return nil
}

func (r *FleetRepository) ListFromTeam(ctx context.Context, teamID model.ID) (model.List[model.Fleet], error) {
	query, args, err := psql.Select("f.id", "f.name", "f.organization_id", "f.created_at", "f.updated_at").
		From("fleets f").
		Join("fleet_teams ft ON ft.fleet_id = f.id").
		Where(sq.Eq{"ft.team_id": string(teamID)}).
		OrderBy("f.created_at DESC").
		ToSql()
	if err != nil {
		return model.List[model.Fleet]{}, fmt.Errorf("postgres: failed to build list fleets from team query: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return model.List[model.Fleet]{}, fmt.Errorf("postgres: failed to list fleets from team: %w", err)
	}
	defer rows.Close()

	return scanFleetList(rows)
}

// AddTeam is idempotent: granting access a team already has is a no-op,
// not an error.
func (r *FleetRepository) AddTeam(ctx context.Context, fleetID, teamID model.ID) error {
	query, args, err := psql.Insert("fleet_teams").
		Columns("fleet_id", "team_id").
		Values(string(fleetID), string(teamID)).
		Suffix("ON CONFLICT (fleet_id, team_id) DO NOTHING").
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build add team to fleet query: %w", err)
	}

	if _, err := r.db.Pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("postgres: failed to add team to fleet: %w", err)
	}
	return nil
}

// RemoveTeam is idempotent: revoking access a team doesn't have is a
// no-op, not an error.
func (r *FleetRepository) RemoveTeam(ctx context.Context, fleetID, teamID model.ID) error {
	query, args, err := psql.Delete("fleet_teams").
		Where(sq.Eq{"fleet_id": string(fleetID), "team_id": string(teamID)}).
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build remove team from fleet query: %w", err)
	}

	if _, err := r.db.Pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("postgres: failed to remove team from fleet: %w", err)
	}
	return nil
}

func scanFleet(row scannableRow) (model.Fleet, error) {
	var (
		f              model.Fleet
		id             string
		organizationID string
	)

	err := row.Scan(&id, &f.Name, &organizationID, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return model.Fleet{}, err
	}

	f.ID = model.ID(id)
	f.OrganizationID = model.ID(organizationID)
	return f, nil
}

func scanFleetList(rows pgx.Rows) (model.List[model.Fleet], error) {
	fleets := make([]model.Fleet, 0)
	for rows.Next() {
		fleet, err := scanFleet(rows)
		if err != nil {
			return model.List[model.Fleet]{}, fmt.Errorf("postgres: failed to scan fleet: %w", err)
		}
		fleets = append(fleets, fleet)
	}
	if err := rows.Err(); err != nil {
		return model.List[model.Fleet]{}, fmt.Errorf("postgres: failed to list fleets: %w", err)
	}

	return model.List[model.Fleet]{Items: fleets, Total: len(fleets)}, nil
}

var _ port.FleetRepository = (*FleetRepository)(nil)
