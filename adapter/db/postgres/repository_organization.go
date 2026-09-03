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

var organizationColumns = []string{"id", "name", "created_at", "updated_at"}

// OrganizationRepository implements port.OrganizationRepository against
// Postgres.
type OrganizationRepository struct {
	db *DB
}

func NewOrganizationRepository(db *DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) Create(ctx context.Context, organization model.Organization) (model.Organization, error) {
	query, args, err := psql.Insert("organizations").
		Columns("name").
		Values(organization.Name).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return model.Organization{}, fmt.Errorf("postgres: failed to build create organization query: %w", err)
	}

	var id string
	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(&id, &organization.CreatedAt, &organization.UpdatedAt)
	if err != nil {
		return model.Organization{}, fmt.Errorf("postgres: failed to create organization: %w", err)
	}

	organization.ID = model.ID(id)
	return organization, nil
}

func (r *OrganizationRepository) Get(ctx context.Context, id model.ID) (model.Organization, error) {
	query, args, err := psql.Select(organizationColumns...).
		From("organizations").
		Where(sq.Eq{"id": string(id)}).
		ToSql()
	if err != nil {
		return model.Organization{}, fmt.Errorf("postgres: failed to build get organization query: %w", err)
	}

	organization, err := scanOrganization(r.db.Pool.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Organization{}, model.NewError(model.ErrCodeOrganizationNotFound, err)
		}
		return model.Organization{}, fmt.Errorf("postgres: failed to get organization: %w", err)
	}

	return organization, nil
}

// List ignores filters for now: model.OrganizationFilters carries no
// fields yet.
func (r *OrganizationRepository) List(ctx context.Context, filters model.OrganizationFilters) (model.List[model.Organization], error) {
	query, args, err := psql.Select(organizationColumns...).
		From("organizations").
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return model.List[model.Organization]{}, fmt.Errorf("postgres: failed to build list organizations query: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return model.List[model.Organization]{}, fmt.Errorf("postgres: failed to list organizations: %w", err)
	}
	defer rows.Close()

	return scanOrganizationList(rows)
}

func (r *OrganizationRepository) Update(ctx context.Context, organization model.Organization) (model.Organization, error) {
	query, args, err := psql.Update("organizations").
		Set("name", organization.Name).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": string(organization.ID)}).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return model.Organization{}, fmt.Errorf("postgres: failed to build update organization query: %w", err)
	}

	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(&organization.CreatedAt, &organization.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Organization{}, model.NewError(model.ErrCodeOrganizationNotFound, err)
		}
		return model.Organization{}, fmt.Errorf("postgres: failed to update organization: %w", err)
	}

	return organization, nil
}

// Delete maps each RESTRICT foreign key that can point at an organization
// (teams, fleets, account_organizations) to its own error, so callers know
// exactly what's still attached.
func (r *OrganizationRepository) Delete(ctx context.Context, id model.ID) error {
	query, args, err := psql.Delete("organizations").
		Where(sq.Eq{"id": string(id)}).
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build delete organization query: %w", err)
	}

	tag, err := r.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		switch foreignKeyViolationConstraint(err) {
		case "fk_teams_organization":
			return model.NewError(model.ErrCodeOrganizationHasTeams, err)
		case "fk_fleets_organization":
			return model.NewError(model.ErrCodeOrganizationHasFleets, err)
		case "fk_account_organizations_organization":
			return model.NewError(model.ErrCodeOrganizationHasAccounts, err)
		}
		return fmt.Errorf("postgres: failed to delete organization: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return model.NewError(model.ErrCodeOrganizationNotFound, nil)
	}

	return nil
}

func (r *OrganizationRepository) ListFromAccount(ctx context.Context, accountID model.ID) (model.List[model.Organization], error) {
	query, args, err := psql.Select("o.id", "o.name", "o.created_at", "o.updated_at").
		From("organizations o").
		Join("account_organizations ao ON ao.organization_id = o.id").
		Where(sq.Eq{"ao.account_id": string(accountID)}).
		OrderBy("o.created_at DESC").
		ToSql()
	if err != nil {
		return model.List[model.Organization]{}, fmt.Errorf("postgres: failed to build list organizations from account query: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return model.List[model.Organization]{}, fmt.Errorf("postgres: failed to list organizations from account: %w", err)
	}
	defer rows.Close()

	return scanOrganizationList(rows)
}

// AddAccount is idempotent: adding an account that already belongs to the
// organization is a no-op, not an error.
func (r *OrganizationRepository) AddAccount(ctx context.Context, organizationID, accountID model.ID) error {
	query, args, err := psql.Insert("account_organizations").
		Columns("organization_id", "account_id").
		Values(string(organizationID), string(accountID)).
		Suffix("ON CONFLICT (account_id, organization_id) DO NOTHING").
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build add account to organization query: %w", err)
	}

	if _, err := r.db.Pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("postgres: failed to add account to organization: %w", err)
	}
	return nil
}

// RemoveAccount is idempotent: removing a membership that doesn't exist is
// a no-op, not an error.
func (r *OrganizationRepository) RemoveAccount(ctx context.Context, organizationID, accountID model.ID) error {
	query, args, err := psql.Delete("account_organizations").
		Where(sq.Eq{"organization_id": string(organizationID), "account_id": string(accountID)}).
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build remove account from organization query: %w", err)
	}

	if _, err := r.db.Pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("postgres: failed to remove account from organization: %w", err)
	}
	return nil
}

func scanOrganization(row scannableRow) (model.Organization, error) {
	var (
		o  model.Organization
		id string
	)

	err := row.Scan(&id, &o.Name, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return model.Organization{}, err
	}

	o.ID = model.ID(id)
	return o, nil
}

func scanOrganizationList(rows pgx.Rows) (model.List[model.Organization], error) {
	organizations := make([]model.Organization, 0)
	for rows.Next() {
		organization, err := scanOrganization(rows)
		if err != nil {
			return model.List[model.Organization]{}, fmt.Errorf("postgres: failed to scan organization: %w", err)
		}
		organizations = append(organizations, organization)
	}
	if err := rows.Err(); err != nil {
		return model.List[model.Organization]{}, fmt.Errorf("postgres: failed to list organizations: %w", err)
	}

	return model.List[model.Organization]{Items: organizations, Total: len(organizations)}, nil
}

var _ port.OrganizationRepository = (*OrganizationRepository)(nil)
