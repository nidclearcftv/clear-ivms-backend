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

var groupColumns = []string{"id", "name", "organization_id", "parent_id", "created_at", "updated_at"}

// GroupRepository implements port.GroupRepository against Postgres.
type GroupRepository struct {
	db *DB
}

func NewGroupRepository(db *DB) *GroupRepository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) Create(ctx context.Context, group model.Group) (model.Group, error) {
	query, args, err := psql.Insert("groups").
		Columns("name", "organization_id", "parent_id").
		Values(group.Name, string(group.OrganizationID), idPtrToStringPtr(group.ParentID)).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return model.Group{}, fmt.Errorf("postgres: failed to build create group query: %w", err)
	}

	var id string
	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(&id, &group.CreatedAt, &group.UpdatedAt)
	if err != nil {
		switch foreignKeyViolationConstraint(err) {
		case "fk_groups_organization":
			return model.Group{}, model.NewError(model.ErrCodeOrganizationNotFound, err)
		case "fk_groups_parent":
			return model.Group{}, model.NewError(model.ErrCodeGroupNotFound, err)
		}
		return model.Group{}, fmt.Errorf("postgres: failed to create group: %w", err)
	}

	group.ID = model.ID(id)
	return group, nil
}

func (r *GroupRepository) Get(ctx context.Context, id model.ID) (model.Group, error) {
	query, args, err := psql.Select(groupColumns...).
		From("groups").
		Where(sq.Eq{"id": string(id)}).
		ToSql()
	if err != nil {
		return model.Group{}, fmt.Errorf("postgres: failed to build get group query: %w", err)
	}

	group, err := scanGroup(r.db.Pool.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Group{}, model.NewError(model.ErrCodeGroupNotFound, err)
		}
		return model.Group{}, fmt.Errorf("postgres: failed to get group: %w", err)
	}

	return group, nil
}

func (r *GroupRepository) List(ctx context.Context, filters model.GroupFilters) (model.List[model.Group], error) {
	builder := psql.Select(groupColumns...).
		From("groups").
		OrderBy("created_at DESC")

	if filters.OrganizationID != "" {
		builder = builder.Where(sq.Eq{"organization_id": string(filters.OrganizationID)})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return model.List[model.Group]{}, fmt.Errorf("postgres: failed to build list groups query: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return model.List[model.Group]{}, fmt.Errorf("postgres: failed to list groups: %w", err)
	}
	defer rows.Close()

	return scanGroupList(rows)
}

func (r *GroupRepository) Update(ctx context.Context, group model.Group) (model.Group, error) {
	query, args, err := psql.Update("groups").
		Set("name", group.Name).
		Set("organization_id", string(group.OrganizationID)).
		Set("parent_id", idPtrToStringPtr(group.ParentID)).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": string(group.ID)}).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return model.Group{}, fmt.Errorf("postgres: failed to build update group query: %w", err)
	}

	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(&group.CreatedAt, &group.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Group{}, model.NewError(model.ErrCodeGroupNotFound, err)
		}
		switch foreignKeyViolationConstraint(err) {
		case "fk_groups_organization":
			return model.Group{}, model.NewError(model.ErrCodeOrganizationNotFound, err)
		case "fk_groups_parent":
			return model.Group{}, model.NewError(model.ErrCodeGroupNotFound, err)
		}
		return model.Group{}, fmt.Errorf("postgres: failed to update group: %w", err)
	}

	return group, nil
}

// Delete has no RESTRICT dependents to map: vehicles references groups ON
// DELETE SET NULL, and a child group's parent_id references groups ON
// DELETE SET NULL too, so neither can block deletion.
func (r *GroupRepository) Delete(ctx context.Context, id model.ID) error {
	query, args, err := psql.Delete("groups").
		Where(sq.Eq{"id": string(id)}).
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build delete group query: %w", err)
	}

	tag, err := r.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("postgres: failed to delete group: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return model.NewError(model.ErrCodeGroupNotFound, nil)
	}

	return nil
}

func scanGroup(row scannableRow) (model.Group, error) {
	var (
		g              model.Group
		id             string
		organizationID string
		parentID       *string
	)

	err := row.Scan(&id, &g.Name, &organizationID, &parentID, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return model.Group{}, err
	}

	g.ID = model.ID(id)
	g.OrganizationID = model.ID(organizationID)
	if parentID != nil {
		id := model.ID(*parentID)
		g.ParentID = &id
	}
	return g, nil
}

func scanGroupList(rows pgx.Rows) (model.List[model.Group], error) {
	groups := make([]model.Group, 0)
	for rows.Next() {
		group, err := scanGroup(rows)
		if err != nil {
			return model.List[model.Group]{}, fmt.Errorf("postgres: failed to scan group: %w", err)
		}
		groups = append(groups, group)
	}
	if err := rows.Err(); err != nil {
		return model.List[model.Group]{}, fmt.Errorf("postgres: failed to list groups: %w", err)
	}

	return model.List[model.Group]{Items: groups, Total: len(groups)}, nil
}

var _ port.GroupRepository = (*GroupRepository)(nil)
