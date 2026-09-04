package port

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// GroupRepository is the driven (secondary) port for persisting and
// reading group data. It is implemented by outbound adapters, e.g.
// adapter/db/postgres.
type GroupRepository interface {
	Create(ctx context.Context, group model.Group) (model.Group, error)
	Get(ctx context.Context, id model.ID) (model.Group, error)
	List(ctx context.Context, filters model.GroupFilters) (model.List[model.Group], error)
	Update(ctx context.Context, group model.Group) (model.Group, error)
	Delete(ctx context.Context, id model.ID) error
}

// GroupService is the driving (primary) port exposing group-related
// business operations to inbound adapters, e.g. adapter/http controllers.
type GroupService interface {
	Create(ctx context.Context, group model.Group) (model.Group, error)
	Get(ctx context.Context, id model.ID) (model.Group, error)
	List(ctx context.Context, filters model.GroupFilters) (model.List[model.Group], error)
	Update(ctx context.Context, group model.Group) (model.Group, error)
	Delete(ctx context.Context, id model.ID) error
}
