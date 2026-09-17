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

var equipmentModelColumns = []string{"id", "name", "description", "type", "public", "organization_id", "created_at", "updated_at"}

// EquipmentModelRepository implements port.EquipmentModelRepository
// against Postgres.
type EquipmentModelRepository struct {
	db *DB
}

func NewEquipmentModelRepository(db *DB) *EquipmentModelRepository {
	return &EquipmentModelRepository{db: db}
}

// Create never sets public — it always defaults to false at the database
// level (see schema.sql); RETURNING reads it back along with the other
// server-generated columns.
func (r *EquipmentModelRepository) Create(ctx context.Context, equipmentModel model.EquipmentModel) (model.EquipmentModel, error) {
	query, args, err := psql.Insert("equipment_models").
		Columns("name", "description", "type", "organization_id").
		Values(equipmentModel.Name, equipmentModel.Description, string(equipmentModel.Type), string(equipmentModel.OrganizationID)).
		Suffix("RETURNING id, public, created_at, updated_at").
		ToSql()
	if err != nil {
		return model.EquipmentModel{}, fmt.Errorf("postgres: failed to build create equipment model query: %w", err)
	}

	var id string
	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(&id, &equipmentModel.Public, &equipmentModel.CreatedAt, &equipmentModel.UpdatedAt)
	if err != nil {
		if foreignKeyViolationConstraint(err) == "fk_equipment_models_organization" {
			return model.EquipmentModel{}, model.NewError(model.ErrCodeOrganizationNotFound, err)
		}
		return model.EquipmentModel{}, fmt.Errorf("postgres: failed to create equipment model: %w", err)
	}

	equipmentModel.ID = model.ID(id)
	return equipmentModel, nil
}

func (r *EquipmentModelRepository) Get(ctx context.Context, id model.ID) (model.EquipmentModel, error) {
	query, args, err := psql.Select(equipmentModelColumns...).
		From("equipment_models").
		Where(sq.Eq{"id": string(id)}).
		ToSql()
	if err != nil {
		return model.EquipmentModel{}, fmt.Errorf("postgres: failed to build get equipment model query: %w", err)
	}

	equipmentModel, err := scanEquipmentModel(r.db.Pool.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.EquipmentModel{}, model.NewError(model.ErrCodeEquipmentModelNotFound, err)
		}
		return model.EquipmentModel{}, fmt.Errorf("postgres: failed to get equipment model: %w", err)
	}

	return equipmentModel, nil
}

// equipmentModelOrderBy translates EquipmentModelFilters' sort fields
// into a safe ORDER BY clause — the column/direction always come from the
// fixed switches below, never straight from request input, so this can't
// be abused for SQL injection. Defaults to created_at DESC when SortBy is
// unset; a set SortBy defaults to ascending when SortDir isn't also set.
func equipmentModelOrderBy(filters model.EquipmentModelFilters) string {
	column := "created_at"
	direction := "DESC"

	switch filters.SortBy {
	case model.EquipmentModelSortByName:
		column = "name"
		direction = "ASC"
	case model.EquipmentModelSortByType:
		column = "type"
		direction = "ASC"
	case model.EquipmentModelSortByCreatedAt:
		column = "created_at"
		direction = "ASC"
	case model.EquipmentModelSortByUpdatedAt:
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

func (r *EquipmentModelRepository) List(ctx context.Context, filters model.EquipmentModelFilters) (model.List[model.EquipmentModel], error) {
	page := max(filters.Page, 1)
	pageSize := filters.PageSize
	if pageSize < 1 {
		pageSize = model.EquipmentModelDefaultPageSize
	}

	builder := applyEquipmentModelFilters(psql.Select(equipmentModelColumns...).From("equipment_models").OrderBy(equipmentModelOrderBy(filters)), filters).
		Limit(uint64(pageSize)).
		Offset(uint64((page - 1) * pageSize))

	query, args, err := builder.ToSql()
	if err != nil {
		return model.List[model.EquipmentModel]{}, fmt.Errorf("postgres: failed to build list equipment models query: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return model.List[model.EquipmentModel]{}, fmt.Errorf("postgres: failed to list equipment models: %w", err)
	}
	defer rows.Close()

	list, err := scanEquipmentModelList(rows)
	if err != nil {
		return model.List[model.EquipmentModel]{}, err
	}

	total, err := r.Count(ctx, filters)
	if err != nil {
		return model.List[model.EquipmentModel]{}, err
	}
	list.Total = total

	return list, nil
}

// Count reports how many equipment models match filters — the same
// filters List accepts. List uses this to fill model.List.Total.
func (r *EquipmentModelRepository) Count(ctx context.Context, filters model.EquipmentModelFilters) (int, error) {
	builder := applyEquipmentModelFilters(psql.Select("COUNT(*)").From("equipment_models"), filters)

	query, args, err := builder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("postgres: failed to build count equipment models query: %w", err)
	}

	var count int
	if err := r.db.Pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("postgres: failed to count equipment models: %w", err)
	}
	return count, nil
}

// applyEquipmentModelFilters applies filters shared by List and Count.
// filters.OrganizationID scopes to that organization's own equipment
// models (public or not) OR any organization's public ones — a listing
// is never limited to strictly one organization's models, since public
// models are meant to be visible everywhere.
func applyEquipmentModelFilters(builder sq.SelectBuilder, filters model.EquipmentModelFilters) sq.SelectBuilder {
	if filters.OrganizationID != "" {
		builder = builder.Where(sq.Or{
			sq.Eq{"organization_id": string(filters.OrganizationID)},
			sq.Eq{"public": true},
		})
	}
	if filters.Type != "" {
		builder = builder.Where(sq.Eq{"type": string(filters.Type)})
	}
	if filters.Search != "" {
		pattern := "%" + filters.Search + "%"
		builder = builder.Where(sq.Or{
			sq.ILike{"name": pattern},
			sq.ILike{"description": pattern},
		})
	}
	return builder
}

// Update never touches public — see SetPublic. RETURNING reads its
// current (unchanged) value back so the returned model reflects the
// database's actual state.
func (r *EquipmentModelRepository) Update(ctx context.Context, equipmentModel model.EquipmentModel) (model.EquipmentModel, error) {
	query, args, err := psql.Update("equipment_models").
		Set("name", equipmentModel.Name).
		Set("description", equipmentModel.Description).
		Set("type", string(equipmentModel.Type)).
		Set("organization_id", string(equipmentModel.OrganizationID)).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": string(equipmentModel.ID)}).
		Suffix("RETURNING public, created_at, updated_at").
		ToSql()
	if err != nil {
		return model.EquipmentModel{}, fmt.Errorf("postgres: failed to build update equipment model query: %w", err)
	}

	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(&equipmentModel.Public, &equipmentModel.CreatedAt, &equipmentModel.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.EquipmentModel{}, model.NewError(model.ErrCodeEquipmentModelNotFound, err)
		}
		if foreignKeyViolationConstraint(err) == "fk_equipment_models_organization" {
			return model.EquipmentModel{}, model.NewError(model.ErrCodeOrganizationNotFound, err)
		}
		return model.EquipmentModel{}, fmt.Errorf("postgres: failed to update equipment model: %w", err)
	}

	return equipmentModel, nil
}

// Delete has no dependents to map: nothing else in the schema references
// equipment_models.id.
func (r *EquipmentModelRepository) Delete(ctx context.Context, id model.ID) error {
	query, args, err := psql.Delete("equipment_models").
		Where(sq.Eq{"id": string(id)}).
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build delete equipment model query: %w", err)
	}

	tag, err := r.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("postgres: failed to delete equipment model: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return model.NewError(model.ErrCodeEquipmentModelNotFound, nil)
	}

	return nil
}

// SetPublic is the only way to change an equipment model's public flag —
// Update deliberately excludes the column so a caller can't overwrite it
// as a side effect of an unrelated field change.
func (r *EquipmentModelRepository) SetPublic(ctx context.Context, id model.ID, public bool) error {
	query, args, err := psql.Update("equipment_models").
		Set("public", public).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": string(id)}).
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build set equipment model public query: %w", err)
	}

	tag, err := r.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("postgres: failed to set equipment model public: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return model.NewError(model.ErrCodeEquipmentModelNotFound, nil)
	}

	return nil
}

func scanEquipmentModel(row scannableRow) (model.EquipmentModel, error) {
	var (
		m              model.EquipmentModel
		id             string
		equipmentType  string
		organizationID string
	)

	err := row.Scan(&id, &m.Name, &m.Description, &equipmentType, &m.Public, &organizationID, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return model.EquipmentModel{}, err
	}

	m.ID = model.ID(id)
	m.Type = model.EquipmentModelType(equipmentType)
	m.OrganizationID = model.ID(organizationID)
	return m, nil
}

func scanEquipmentModelList(rows pgx.Rows) (model.List[model.EquipmentModel], error) {
	equipmentModels := make([]model.EquipmentModel, 0)
	for rows.Next() {
		equipmentModel, err := scanEquipmentModel(rows)
		if err != nil {
			return model.List[model.EquipmentModel]{}, fmt.Errorf("postgres: failed to scan equipment model: %w", err)
		}
		equipmentModels = append(equipmentModels, equipmentModel)
	}
	if err := rows.Err(); err != nil {
		return model.List[model.EquipmentModel]{}, fmt.Errorf("postgres: failed to list equipment models: %w", err)
	}

	return model.List[model.EquipmentModel]{Items: equipmentModels, Total: len(equipmentModels)}, nil
}

var _ port.EquipmentModelRepository = (*EquipmentModelRepository)(nil)
