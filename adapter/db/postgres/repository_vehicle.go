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

var vehicleColumns = []string{"id", "organization_id", "group_id", "ivms_type", "external_id", "plate_number", "created_at", "updated_at"}

// VehicleRepository implements port.VehicleRepository against Postgres.
type VehicleRepository struct {
	db *DB
}

func NewVehicleRepository(db *DB) *VehicleRepository {
	return &VehicleRepository{db: db}
}

func (r *VehicleRepository) Create(ctx context.Context, vehicle model.Vehicle) (model.Vehicle, error) {
	query, args, err := psql.Insert("vehicles").
		Columns("organization_id", "group_id", "ivms_type", "external_id", "plate_number").
		Values(string(vehicle.OrganizationID), idPtrToStringPtr(vehicle.GroupID), vehicle.IVMSType.String(), vehicle.ExternalID, vehicle.PlateNumber).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return model.Vehicle{}, fmt.Errorf("postgres: failed to build create vehicle query: %w", err)
	}

	var id string
	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(&id, &vehicle.CreatedAt, &vehicle.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Vehicle{}, model.NewError(model.ErrCodeVehicleAlreadyExists, err)
		}
		switch foreignKeyViolationConstraint(err) {
		case "fk_vehicles_organization":
			return model.Vehicle{}, model.NewError(model.ErrCodeOrganizationNotFound, err)
		case "fk_vehicles_group":
			return model.Vehicle{}, model.NewError(model.ErrCodeGroupNotFound, err)
		}
		return model.Vehicle{}, fmt.Errorf("postgres: failed to create vehicle: %w", err)
	}

	vehicle.ID = model.ID(id)
	return vehicle, nil
}

func (r *VehicleRepository) Get(ctx context.Context, id model.ID) (model.Vehicle, error) {
	query, args, err := psql.Select(vehicleColumns...).
		From("vehicles").
		Where(sq.Eq{"id": string(id)}).
		ToSql()
	if err != nil {
		return model.Vehicle{}, fmt.Errorf("postgres: failed to build get vehicle query: %w", err)
	}

	vehicle, err := scanVehicle(r.db.Pool.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Vehicle{}, model.NewError(model.ErrCodeVehicleNotFound, err)
		}
		return model.Vehicle{}, fmt.Errorf("postgres: failed to get vehicle: %w", err)
	}

	return vehicle, nil
}

func (r *VehicleRepository) List(ctx context.Context, filters model.VehicleFilters) (model.List[model.Vehicle], error) {
	builder := psql.Select(vehicleColumns...).
		From("vehicles").
		OrderBy("created_at DESC")

	if filters.OrganizationID != "" {
		builder = builder.Where(sq.Eq{"organization_id": string(filters.OrganizationID)})
	}
	if filters.GroupID != "" {
		builder = builder.Where(sq.Eq{"group_id": string(filters.GroupID)})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return model.List[model.Vehicle]{}, fmt.Errorf("postgres: failed to build list vehicles query: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return model.List[model.Vehicle]{}, fmt.Errorf("postgres: failed to list vehicles: %w", err)
	}
	defer rows.Close()

	return scanVehicleList(rows)
}

func (r *VehicleRepository) Update(ctx context.Context, vehicle model.Vehicle) (model.Vehicle, error) {
	query, args, err := psql.Update("vehicles").
		Set("organization_id", string(vehicle.OrganizationID)).
		Set("group_id", idPtrToStringPtr(vehicle.GroupID)).
		Set("ivms_type", vehicle.IVMSType.String()).
		Set("external_id", vehicle.ExternalID).
		Set("plate_number", vehicle.PlateNumber).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": string(vehicle.ID)}).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return model.Vehicle{}, fmt.Errorf("postgres: failed to build update vehicle query: %w", err)
	}

	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(&vehicle.CreatedAt, &vehicle.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Vehicle{}, model.NewError(model.ErrCodeVehicleNotFound, err)
		}
		if isUniqueViolation(err) {
			return model.Vehicle{}, model.NewError(model.ErrCodeVehicleAlreadyExists, err)
		}
		switch foreignKeyViolationConstraint(err) {
		case "fk_vehicles_organization":
			return model.Vehicle{}, model.NewError(model.ErrCodeOrganizationNotFound, err)
		case "fk_vehicles_group":
			return model.Vehicle{}, model.NewError(model.ErrCodeGroupNotFound, err)
		}
		return model.Vehicle{}, fmt.Errorf("postgres: failed to update vehicle: %w", err)
	}

	return vehicle, nil
}

// Delete has no dependents to map: nothing else in the schema references
// vehicles.id.
func (r *VehicleRepository) Delete(ctx context.Context, id model.ID) error {
	query, args, err := psql.Delete("vehicles").
		Where(sq.Eq{"id": string(id)}).
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build delete vehicle query: %w", err)
	}

	tag, err := r.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("postgres: failed to delete vehicle: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return model.NewError(model.ErrCodeVehicleNotFound, nil)
	}

	return nil
}

// idPtrToStringPtr converts a nullable model.ID into a nullable string for
// the driver — pgx encodes a nil *string as SQL NULL, matching vehicles'
// nullable group_id column (and groups' nullable parent_id column).
func idPtrToStringPtr(id *model.ID) *string {
	if id == nil {
		return nil
	}
	s := string(*id)
	return &s
}

func scanVehicle(row scannableRow) (model.Vehicle, error) {
	var (
		v              model.Vehicle
		id             string
		organizationID string
		groupID        *string
		ivmsType       string
	)

	err := row.Scan(&id, &organizationID, &groupID, &ivmsType, &v.ExternalID, &v.PlateNumber, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return model.Vehicle{}, err
	}

	v.ID = model.ID(id)
	v.OrganizationID = model.ID(organizationID)
	if groupID != nil {
		id := model.ID(*groupID)
		v.GroupID = &id
	}
	v.IVMSType = model.IVMSTypeFromString(ivmsType)

	return v, nil
}

func scanVehicleList(rows pgx.Rows) (model.List[model.Vehicle], error) {
	vehicles := make([]model.Vehicle, 0)
	for rows.Next() {
		vehicle, err := scanVehicle(rows)
		if err != nil {
			return model.List[model.Vehicle]{}, fmt.Errorf("postgres: failed to scan vehicle: %w", err)
		}
		vehicles = append(vehicles, vehicle)
	}
	if err := rows.Err(); err != nil {
		return model.List[model.Vehicle]{}, fmt.Errorf("postgres: failed to list vehicles: %w", err)
	}

	return model.List[model.Vehicle]{Items: vehicles, Total: len(vehicles)}, nil
}

var _ port.VehicleRepository = (*VehicleRepository)(nil)
