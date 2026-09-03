package port

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// FleetRepository is the driven (secondary) port for persisting and
// reading fleet data. It is implemented by outbound adapters, e.g.
// adapter/db/postgres.
//
// Relation methods only ever return/accept Fleet — e.g. ListFromTeam
// returns the fleets a team has access to, not the team itself.
type FleetRepository interface {
	Create(ctx context.Context, fleet model.Fleet) (model.Fleet, error)
	Get(ctx context.Context, id model.ID) (model.Fleet, error)
	List(ctx context.Context, filters model.FleetFilters) (model.List[model.Fleet], error)
	Update(ctx context.Context, fleet model.Fleet) (model.Fleet, error)
	Delete(ctx context.Context, id model.ID) error

	ListFromTeam(ctx context.Context, teamID model.ID) (model.List[model.Fleet], error)
	AddTeam(ctx context.Context, fleetID, teamID model.ID) error
	RemoveTeam(ctx context.Context, fleetID, teamID model.ID) error
}

// FleetService is the driving (primary) port exposing fleet-related
// business operations to inbound adapters, e.g. adapter/http controllers.
type FleetService interface {
	Create(ctx context.Context, fleet model.Fleet) (model.Fleet, error)
	Get(ctx context.Context, id model.ID) (model.Fleet, error)
	List(ctx context.Context, filters model.FleetFilters) (model.List[model.Fleet], error)
	Update(ctx context.Context, fleet model.Fleet) (model.Fleet, error)
	Delete(ctx context.Context, id model.ID) error

	ListFromTeam(ctx context.Context, teamID model.ID) (model.List[model.Fleet], error)
	AddTeam(ctx context.Context, fleetID, teamID model.ID) error
	RemoveTeam(ctx context.Context, fleetID, teamID model.ID) error
}
