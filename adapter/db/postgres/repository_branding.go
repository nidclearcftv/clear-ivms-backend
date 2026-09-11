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

var brandingColumns = []string{"id", "name", "domain", "config", "created_at", "updated_at"}

// BrandingRepository implements port.BrandingRepository against Postgres.
type BrandingRepository struct {
	db *DB
}

func NewBrandingRepository(db *DB) *BrandingRepository {
	return &BrandingRepository{db: db}
}

// brandingOrderBy translates BrandingFilters' sort fields into a safe
// ORDER BY clause — the column/direction always come from the fixed
// switches below, never straight from request input, so this can't be
// abused for SQL injection. Defaults to created_at DESC when SortBy is
// unset, matching List's pre-sorting behavior; a set SortBy defaults to
// ascending when SortDir isn't also set.
func brandingOrderBy(filters model.BrandingFilters) string {
	column := "created_at"
	direction := "DESC"

	switch filters.SortBy {
	case model.BrandingSortByName:
		column = "name"
		direction = "ASC"
	case model.BrandingSortByDomain:
		column = "domain"
		direction = "ASC"
	case model.BrandingSortByCreatedAt:
		column = "created_at"
		direction = "ASC"
	case model.BrandingSortByUpdatedAt:
		column = "updated_at"
		direction = "ASC"
	}

	switch filters.SortDir {
	case model.SortDirectionAsc:
		direction = "ASC"
	case model.SortDirectionDesc:
		direction = "DESC"
	}

	return column + " " + direction
}

// brandingSearchFilter builds a WHERE clause matching filters.Search
// case-insensitively against name or domain, or nil when Search is empty.
func brandingSearchFilter(filters model.BrandingFilters) sq.Sqlizer {
	if filters.Search == "" {
		return nil
	}
	pattern := "%" + filters.Search + "%"
	return sq.Or{
		sq.ILike{"name": pattern},
		sq.ILike{"domain": pattern},
	}
}

func (r *BrandingRepository) Create(ctx context.Context, branding model.Branding) (model.Branding, error) {
	if branding.Config == nil {
		branding.Config = model.JSON{}
	}

	query, args, err := psql.Insert("brandings").
		Columns("name", "domain", "config").
		Values(branding.Name, branding.Domain, branding.Config).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return model.Branding{}, fmt.Errorf("postgres: failed to build create branding query: %w", err)
	}

	var id string
	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(&id, &branding.CreatedAt, &branding.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Branding{}, model.NewError(model.ErrCodeBrandingDomainAlreadyExists, err)
		}
		return model.Branding{}, fmt.Errorf("postgres: failed to create branding: %w", err)
	}

	branding.ID = model.ID(id)
	return branding, nil
}

func (r *BrandingRepository) Get(ctx context.Context, id model.ID) (model.Branding, error) {
	query, args, err := psql.Select(brandingColumns...).
		From("brandings").
		Where(sq.Eq{"id": string(id)}).
		ToSql()
	if err != nil {
		return model.Branding{}, fmt.Errorf("postgres: failed to build get branding query: %w", err)
	}

	branding, err := scanBranding(r.db.Pool.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Branding{}, model.NewError(model.ErrCodeBrandingNotFound, err)
		}
		return model.Branding{}, fmt.Errorf("postgres: failed to get branding: %w", err)
	}

	return branding, nil
}

// GetByDomain looks up a branding by its (unique) domain — what the public
// branding lookup route has on hand.
func (r *BrandingRepository) GetByDomain(ctx context.Context, domain string) (model.Branding, error) {
	query, args, err := psql.Select(brandingColumns...).
		From("brandings").
		Where(sq.Eq{"domain": domain}).
		ToSql()
	if err != nil {
		return model.Branding{}, fmt.Errorf("postgres: failed to build get branding by domain query: %w", err)
	}

	branding, err := scanBranding(r.db.Pool.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Branding{}, model.NewError(model.ErrCodeBrandingNotFound, err)
		}
		return model.Branding{}, fmt.Errorf("postgres: failed to get branding by domain: %w", err)
	}

	return branding, nil
}

func (r *BrandingRepository) List(ctx context.Context, filters model.BrandingFilters) (model.List[model.Branding], error) {
	builder := psql.Select(brandingColumns...).
		From("brandings").
		OrderBy(brandingOrderBy(filters))
	if search := brandingSearchFilter(filters); search != nil {
		builder = builder.Where(search)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return model.List[model.Branding]{}, fmt.Errorf("postgres: failed to build list brandings query: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return model.List[model.Branding]{}, fmt.Errorf("postgres: failed to list brandings: %w", err)
	}
	defer rows.Close()

	list, err := scanBrandingList(rows)
	if err != nil {
		return model.List[model.Branding]{}, err
	}

	total, err := r.Count(ctx, filters)
	if err != nil {
		return model.List[model.Branding]{}, err
	}
	list.Total = total

	return list, nil
}

func (r *BrandingRepository) Count(ctx context.Context, filters model.BrandingFilters) (int, error) {
	builder := psql.Select("COUNT(*)").From("brandings")
	if search := brandingSearchFilter(filters); search != nil {
		builder = builder.Where(search)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("postgres: failed to build count brandings query: %w", err)
	}

	var count int
	if err := r.db.Pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("postgres: failed to count brandings: %w", err)
	}
	return count, nil
}

func (r *BrandingRepository) Update(ctx context.Context, branding model.Branding) (model.Branding, error) {
	if branding.Config == nil {
		branding.Config = model.JSON{}
	}

	query, args, err := psql.Update("brandings").
		Set("name", branding.Name).
		Set("domain", branding.Domain).
		Set("config", branding.Config).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": string(branding.ID)}).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return model.Branding{}, fmt.Errorf("postgres: failed to build update branding query: %w", err)
	}

	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(&branding.CreatedAt, &branding.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Branding{}, model.NewError(model.ErrCodeBrandingNotFound, err)
		}
		if isUniqueViolation(err) {
			return model.Branding{}, model.NewError(model.ErrCodeBrandingDomainAlreadyExists, err)
		}
		return model.Branding{}, fmt.Errorf("postgres: failed to update branding: %w", err)
	}

	return branding, nil
}

// Delete has no dependents to map: nothing else in the schema references
// brandings.id.
func (r *BrandingRepository) Delete(ctx context.Context, id model.ID) error {
	query, args, err := psql.Delete("brandings").
		Where(sq.Eq{"id": string(id)}).
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build delete branding query: %w", err)
	}

	tag, err := r.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("postgres: failed to delete branding: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return model.NewError(model.ErrCodeBrandingNotFound, nil)
	}

	return nil
}

func scanBranding(row scannableRow) (model.Branding, error) {
	var (
		b  model.Branding
		id string
	)

	err := row.Scan(&id, &b.Name, &b.Domain, &b.Config, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return model.Branding{}, err
	}

	b.ID = model.ID(id)
	return b, nil
}

func scanBrandingList(rows pgx.Rows) (model.List[model.Branding], error) {
	brandings := make([]model.Branding, 0)
	for rows.Next() {
		branding, err := scanBranding(rows)
		if err != nil {
			return model.List[model.Branding]{}, fmt.Errorf("postgres: failed to scan branding: %w", err)
		}
		brandings = append(brandings, branding)
	}
	if err := rows.Err(); err != nil {
		return model.List[model.Branding]{}, fmt.Errorf("postgres: failed to list brandings: %w", err)
	}

	return model.List[model.Branding]{Items: brandings, Total: len(brandings)}, nil
}

var _ port.BrandingRepository = (*BrandingRepository)(nil)
