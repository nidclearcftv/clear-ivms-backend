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
	// Update never touches public or pictureObjectKey — see SetPublic and
	// SetPictureObjectKey.
	Update(ctx context.Context, equipmentModel model.EquipmentModel) (model.EquipmentModel, error)
	Delete(ctx context.Context, id model.ID) error
	// SetPublic is the only way to change an equipment model's public
	// flag; Update deliberately excludes it.
	SetPublic(ctx context.Context, id model.ID, public bool) error
	// SetPictureObjectKey is the only way to change an equipment model's
	// picture reference; Update deliberately excludes it. key is nil to
	// clear it (see EquipmentModelService.DeletePicture).
	SetPictureObjectKey(ctx context.Context, id model.ID, key *string) error
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
	// SetPicture generates a new picture reference for the equipment
	// model and returns a URL the caller must redirect the client to so
	// it can PUT the actual bytes (as contentType) directly — this
	// backend's bytes never pass through it. The equipment model is
	// pointed at the new reference immediately, before that upload
	// happens (there is no confirmation step — see
	// registerEquipmentModelRoutes), so a client that requests this URL
	// and never follows through leaves the equipment model referencing an
	// object that was never written. The old picture, if any, is deleted
	// from storage on a best-effort basis once the new reference is
	// recorded.
	SetPicture(ctx context.Context, id model.ID, contentType string) (url string, err error)
	// GetPictureURL returns a URL the caller must redirect the client to
	// for the equipment model's current picture. Uses the same relaxed
	// visibility Get does: a public equipment model's picture is
	// readable from any organization. Fails with
	// ErrCodeEquipmentModelPictureNotFound if no picture is set.
	GetPictureURL(ctx context.Context, id model.ID) (url string, err error)
	// DeletePicture clears the equipment model's picture, if any set.
	// Deleting when none is set is not an error.
	DeletePicture(ctx context.Context, id model.ID) (model.EquipmentModel, error)
}
