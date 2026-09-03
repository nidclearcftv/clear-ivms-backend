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

var teamColumns = []string{"id", "name", "organization_id", "created_at", "updated_at"}

// TeamRepository implements port.TeamRepository against Postgres.
type TeamRepository struct {
	db *DB
}

func NewTeamRepository(db *DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Create(ctx context.Context, team model.Team) (model.Team, error) {
	query, args, err := psql.Insert("teams").
		Columns("name", "organization_id").
		Values(team.Name, string(team.OrganizationID)).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return model.Team{}, fmt.Errorf("postgres: failed to build create team query: %w", err)
	}

	var id string
	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(&id, &team.CreatedAt, &team.UpdatedAt)
	if err != nil {
		if foreignKeyViolationConstraint(err) == "fk_teams_organization" {
			return model.Team{}, model.NewError(model.ErrCodeOrganizationNotFound, err)
		}
		return model.Team{}, fmt.Errorf("postgres: failed to create team: %w", err)
	}

	team.ID = model.ID(id)
	return team, nil
}

func (r *TeamRepository) Get(ctx context.Context, id model.ID) (model.Team, error) {
	query, args, err := psql.Select(teamColumns...).
		From("teams").
		Where(sq.Eq{"id": string(id)}).
		ToSql()
	if err != nil {
		return model.Team{}, fmt.Errorf("postgres: failed to build get team query: %w", err)
	}

	team, err := scanTeam(r.db.Pool.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Team{}, model.NewError(model.ErrCodeTeamNotFound, err)
		}
		return model.Team{}, fmt.Errorf("postgres: failed to get team: %w", err)
	}

	return team, nil
}

func (r *TeamRepository) List(ctx context.Context, filters model.TeamFilters) (model.List[model.Team], error) {
	builder := psql.Select(teamColumns...).
		From("teams").
		OrderBy("created_at DESC")

	if filters.OrganizationID != "" {
		builder = builder.Where(sq.Eq{"organization_id": string(filters.OrganizationID)})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return model.List[model.Team]{}, fmt.Errorf("postgres: failed to build list teams query: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return model.List[model.Team]{}, fmt.Errorf("postgres: failed to list teams: %w", err)
	}
	defer rows.Close()

	return scanTeamList(rows)
}

func (r *TeamRepository) Update(ctx context.Context, team model.Team) (model.Team, error) {
	query, args, err := psql.Update("teams").
		Set("name", team.Name).
		Set("organization_id", string(team.OrganizationID)).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": string(team.ID)}).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return model.Team{}, fmt.Errorf("postgres: failed to build update team query: %w", err)
	}

	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(&team.CreatedAt, &team.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Team{}, model.NewError(model.ErrCodeTeamNotFound, err)
		}
		if foreignKeyViolationConstraint(err) == "fk_teams_organization" {
			return model.Team{}, model.NewError(model.ErrCodeOrganizationNotFound, err)
		}
		return model.Team{}, fmt.Errorf("postgres: failed to update team: %w", err)
	}

	return team, nil
}

// Delete has no RESTRICT dependents to map: account_teams and fleet_teams
// both reference teams ON DELETE CASCADE, so their rows are cleaned up
// automatically.
func (r *TeamRepository) Delete(ctx context.Context, id model.ID) error {
	query, args, err := psql.Delete("teams").
		Where(sq.Eq{"id": string(id)}).
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build delete team query: %w", err)
	}

	tag, err := r.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("postgres: failed to delete team: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return model.NewError(model.ErrCodeTeamNotFound, nil)
	}

	return nil
}

func (r *TeamRepository) ListFromAccount(ctx context.Context, accountID model.ID) (model.List[model.Team], error) {
	query, args, err := psql.Select("t.id", "t.name", "t.organization_id", "t.created_at", "t.updated_at").
		From("teams t").
		Join("account_teams act ON act.team_id = t.id").
		Where(sq.Eq{"act.account_id": string(accountID)}).
		OrderBy("t.created_at DESC").
		ToSql()
	if err != nil {
		return model.List[model.Team]{}, fmt.Errorf("postgres: failed to build list teams from account query: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return model.List[model.Team]{}, fmt.Errorf("postgres: failed to list teams from account: %w", err)
	}
	defer rows.Close()

	return scanTeamList(rows)
}

// AddAccount is idempotent: adding an account that's already a member is a
// no-op, not an error.
func (r *TeamRepository) AddAccount(ctx context.Context, teamID, accountID model.ID) error {
	query, args, err := psql.Insert("account_teams").
		Columns("team_id", "account_id").
		Values(string(teamID), string(accountID)).
		Suffix("ON CONFLICT (account_id, team_id) DO NOTHING").
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build add account to team query: %w", err)
	}

	if _, err := r.db.Pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("postgres: failed to add account to team: %w", err)
	}
	return nil
}

// RemoveAccount is idempotent: removing a membership that doesn't exist is
// a no-op, not an error.
func (r *TeamRepository) RemoveAccount(ctx context.Context, teamID, accountID model.ID) error {
	query, args, err := psql.Delete("account_teams").
		Where(sq.Eq{"team_id": string(teamID), "account_id": string(accountID)}).
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build remove account from team query: %w", err)
	}

	if _, err := r.db.Pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("postgres: failed to remove account from team: %w", err)
	}
	return nil
}

func (r *TeamRepository) ListFromFleet(ctx context.Context, fleetID model.ID) (model.List[model.Team], error) {
	query, args, err := psql.Select("t.id", "t.name", "t.organization_id", "t.created_at", "t.updated_at").
		From("teams t").
		Join("fleet_teams ft ON ft.team_id = t.id").
		Where(sq.Eq{"ft.fleet_id": string(fleetID)}).
		OrderBy("t.created_at DESC").
		ToSql()
	if err != nil {
		return model.List[model.Team]{}, fmt.Errorf("postgres: failed to build list teams from fleet query: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return model.List[model.Team]{}, fmt.Errorf("postgres: failed to list teams from fleet: %w", err)
	}
	defer rows.Close()

	return scanTeamList(rows)
}

// AddFleet is idempotent: granting access the team already has is a
// no-op, not an error.
func (r *TeamRepository) AddFleet(ctx context.Context, teamID, fleetID model.ID) error {
	query, args, err := psql.Insert("fleet_teams").
		Columns("team_id", "fleet_id").
		Values(string(teamID), string(fleetID)).
		Suffix("ON CONFLICT (fleet_id, team_id) DO NOTHING").
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build add fleet to team query: %w", err)
	}

	if _, err := r.db.Pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("postgres: failed to add fleet to team: %w", err)
	}
	return nil
}

// RemoveFleet is idempotent: revoking access the team doesn't have is a
// no-op, not an error.
func (r *TeamRepository) RemoveFleet(ctx context.Context, teamID, fleetID model.ID) error {
	query, args, err := psql.Delete("fleet_teams").
		Where(sq.Eq{"team_id": string(teamID), "fleet_id": string(fleetID)}).
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build remove fleet from team query: %w", err)
	}

	if _, err := r.db.Pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("postgres: failed to remove fleet from team: %w", err)
	}
	return nil
}

func scanTeam(row scannableRow) (model.Team, error) {
	var (
		t              model.Team
		id             string
		organizationID string
	)

	err := row.Scan(&id, &t.Name, &organizationID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return model.Team{}, err
	}

	t.ID = model.ID(id)
	t.OrganizationID = model.ID(organizationID)
	return t, nil
}

func scanTeamList(rows pgx.Rows) (model.List[model.Team], error) {
	teams := make([]model.Team, 0)
	for rows.Next() {
		team, err := scanTeam(rows)
		if err != nil {
			return model.List[model.Team]{}, fmt.Errorf("postgres: failed to scan team: %w", err)
		}
		teams = append(teams, team)
	}
	if err := rows.Err(); err != nil {
		return model.List[model.Team]{}, fmt.Errorf("postgres: failed to list teams: %w", err)
	}

	return model.List[model.Team]{Items: teams, Total: len(teams)}, nil
}

var _ port.TeamRepository = (*TeamRepository)(nil)
