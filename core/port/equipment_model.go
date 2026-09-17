package port

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// EquipmentModelRepository is the driven (secondary) port for persisting
// and reading equipment model data. It is implemented by outbound
// adapters, e.g. adapter/db/postgres.
type EquipmentModelRepository interface {
	Create(ctx context.Context, equipmentModel model.EquipmentModel) (model.EquipmentModel, error)
	Get(ctx context.Context, id model.ID) (model.EquipmentModel, error)
	List(ctx context.Context, filters model.EquipmentModelFilters) (model.List[model.EquipmentModel], error)
	// Count reports how many equipment models match filters — the same
	// filters List accepts. List uses this to fill model.List.Total.
	Count(ctx context.Context, filters model.EquipmentModelFilters) (int, error)
	// Update never touches public — see SetPublic.
	Update(ctx context.Context, equipmentModel model.EquipmentModel) (model.EquipmentModel, error)
	Delete(ctx context.Context, id model.ID) error
	// SetPublic is the only way to change an equipment model's public
	// flag; Update deliberately excludes it.
	SetPublic(ctx context.Context, id model.ID, public bool) error
}

// EquipmentModelService is the driving (primary) port exposing equipment
// model-related business operations to inbound adapters, e.g.
// adapter/http controllers.
type EquipmentModelService interface {
	Create(ctx context.Context, equipmentModel model.EquipmentModel) (model.EquipmentModel, error)
	Get(ctx context.Context, id model.ID) (model.EquipmentModel, error)
	List(ctx context.Context, filters model.EquipmentModelFilters) (model.List[model.EquipmentModel], error)
	Count(ctx context.Context, filters model.EquipmentModelFilters) (int, error)
	Update(ctx context.Context, equipmentModel model.EquipmentModel) (model.EquipmentModel, error)
	Delete(ctx context.Context, id model.ID) error
	// SetPublic sets whether equipmentModel is public. Exposed via an
	// admin-only route (see registerEquipmentModelRoutes) — unlike
	// everything else here, an org_admin cannot call it.
	SetPublic(ctx context.Context, id model.ID, public bool) error
}
