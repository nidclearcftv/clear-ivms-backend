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

var vehicleEquipmentColumns = []string{"id", "vehicle_id", "equipment_model_id", "equipment_model_type", "serial_number", "description", "created_at", "updated_at"}

// VehicleEquipmentRepository implements port.VehicleEquipmentRepository
// against Postgres.
type VehicleEquipmentRepository struct {
	db *DB
}

func NewVehicleEquipmentRepository(db *DB) *VehicleEquipmentRepository {
	return &VehicleEquipmentRepository{db: db}
}

// Create relies on idx_vehicle_equipment_one_primary_per_vehicle (see
// schema.sql) to reject a second primary registration for the same
// vehicle — there is no pre-check for it in Go, only this constraint's
// violation mapped below.
func (r *VehicleEquipmentRepository) Create(ctx context.Context, ve model.VehicleEquipment) (model.VehicleEquipment, error) {
	query, args, err := psql.Insert("vehicle_equipment").
		Columns("vehicle_id", "equipment_model_id", "equipment_model_type", "serial_number", "description").
		Values(string(ve.VehicleID), string(ve.EquipmentModelID), string(ve.EquipmentModelType), ve.SerialNumber, ve.Description).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()
	if err != nil {
		return model.VehicleEquipment{}, fmt.Errorf("postgres: failed to build create vehicle equipment query: %w", err)
	}

	var id string
	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(&id, &ve.CreatedAt, &ve.UpdatedAt)
	if err != nil {
		if uniqueViolationConstraint(err) == "idx_vehicle_equipment_one_primary_per_vehicle" {
			return model.VehicleEquipment{}, model.NewError(model.ErrCodeVehicleEquipmentPrimaryAlreadyExists, err)
		}
		switch foreignKeyViolationConstraint(err) {
		case "fk_vehicle_equipment_vehicle":
			return model.VehicleEquipment{}, model.NewError(model.ErrCodeVehicleNotFound, err)
		case "fk_vehicle_equipment_equipment_model":
			return model.VehicleEquipment{}, model.NewError(model.ErrCodeEquipmentModelNotFound, err)
		}
		return model.VehicleEquipment{}, fmt.Errorf("postgres: failed to create vehicle equipment: %w", err)
	}

	ve.ID = model.ID(id)
	return ve, nil
}

func (r *VehicleEquipmentRepository) Get(ctx context.Context, id model.ID) (model.VehicleEquipment, error) {
	query, args, err := psql.Select(vehicleEquipmentColumns...).
		From("vehicle_equipment").
		Where(sq.Eq{"id": string(id)}).
		ToSql()
	if err != nil {
		return model.VehicleEquipment{}, fmt.Errorf("postgres: failed to build get vehicle equipment query: %w", err)
	}

	ve, err := scanVehicleEquipment(r.db.Pool.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.VehicleEquipment{}, model.NewError(model.ErrCodeVehicleEquipmentNotFound, err)
		}
		return model.VehicleEquipment{}, fmt.Errorf("postgres: failed to get vehicle equipment: %w", err)
	}

	return ve, nil
}

// List returns every equipment registration for vehicleID, primary first
// (there is ever at most one) then accessories ordered by created_at —
// unpaginated, see port.VehicleEquipmentRepository.
func (r *VehicleEquipmentRepository) List(ctx context.Context, vehicleID model.ID) ([]model.VehicleEquipment, error) {
	query, args, err := psql.Select(vehicleEquipmentColumns...).
		From("vehicle_equipment").
		Where(sq.Eq{"vehicle_id": string(vehicleID)}).
		OrderBy("equipment_model_type DESC", "created_at ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to build list vehicle equipment query: %w", err)
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to list vehicle equipment: %w", err)
	}
	defer rows.Close()

	items := make([]model.VehicleEquipment, 0)
	for rows.Next() {
		ve, err := scanVehicleEquipment(rows)
		if err != nil {
			return nil, fmt.Errorf("postgres: failed to scan vehicle equipment: %w", err)
		}
		items = append(items, ve)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: failed to list vehicle equipment: %w", err)
	}

	return items, nil
}

// CountByType reports how many of vehicleID's registrations have
// equipmentModelType — used to enforce
// model.VehicleEquipmentMaxAccessoriesPerVehicle before an accessory
// insert.
func (r *VehicleEquipmentRepository) CountByType(ctx context.Context, vehicleID model.ID, equipmentModelType model.EquipmentModelType) (int, error) {
	query, args, err := psql.Select("COUNT(*)").
		From("vehicle_equipment").
		Where(sq.Eq{"vehicle_id": string(vehicleID), "equipment_model_type": string(equipmentModelType)}).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("postgres: failed to build count vehicle equipment query: %w", err)
	}

	var count int
	if err := r.db.Pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("postgres: failed to count vehicle equipment: %w", err)
	}
	return count, nil
}

// Update never touches vehicle_id, equipment_model_id or
// equipment_model_type — see VehicleEquipmentService.Update. RETURNING
// reads their current (unchanged) values back so the returned model
// reflects the database's actual state.
func (r *VehicleEquipmentRepository) Update(ctx context.Context, ve model.VehicleEquipment) (model.VehicleEquipment, error) {
	query, args, err := psql.Update("vehicle_equipment").
		Set("serial_number", ve.SerialNumber).
		Set("description", ve.Description).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": string(ve.ID)}).
		Suffix("RETURNING vehicle_id, equipment_model_id, equipment_model_type, created_at, updated_at").
		ToSql()
	if err != nil {
		return model.VehicleEquipment{}, fmt.Errorf("postgres: failed to build update vehicle equipment query: %w", err)
	}

	var vehicleID, equipmentModelID, equipmentModelType string
	err = r.db.Pool.QueryRow(ctx, query, args...).Scan(&vehicleID, &equipmentModelID, &equipmentModelType, &ve.CreatedAt, &ve.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.VehicleEquipment{}, model.NewError(model.ErrCodeVehicleEquipmentNotFound, err)
		}
		return model.VehicleEquipment{}, fmt.Errorf("postgres: failed to update vehicle equipment: %w", err)
	}

	ve.VehicleID = model.ID(vehicleID)
	ve.EquipmentModelID = model.ID(equipmentModelID)
	ve.EquipmentModelType = model.EquipmentModelType(equipmentModelType)
	return ve, nil
}

func (r *VehicleEquipmentRepository) Delete(ctx context.Context, id model.ID) error {
	query, args, err := psql.Delete("vehicle_equipment").
		Where(sq.Eq{"id": string(id)}).
		ToSql()
	if err != nil {
		return fmt.Errorf("postgres: failed to build delete vehicle equipment query: %w", err)
	}

	tag, err := r.db.Pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("postgres: failed to delete vehicle equipment: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return model.NewError(model.ErrCodeVehicleEquipmentNotFound, nil)
	}

	return nil
}

func scanVehicleEquipment(row scannableRow) (model.VehicleEquipment, error) {
	var (
		ve                 model.VehicleEquipment
		id                 string
		vehicleID          string
		equipmentModelID   string
		equipmentModelType string
	)

	err := row.Scan(&id, &vehicleID, &equipmentModelID, &equipmentModelType, &ve.SerialNumber, &ve.Description, &ve.CreatedAt, &ve.UpdatedAt)
	if err != nil {
		return model.VehicleEquipment{}, err
	}

	ve.ID = model.ID(id)
	ve.VehicleID = model.ID(vehicleID)
	ve.EquipmentModelID = model.ID(equipmentModelID)
	ve.EquipmentModelType = model.EquipmentModelType(equipmentModelType)
	return ve, nil
}

var _ port.VehicleEquipmentRepository = (*VehicleEquipmentRepository)(nil)
