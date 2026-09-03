package port

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// TeamRepository is the driven (secondary) port for persisting and reading
// team data. It is implemented by outbound adapters, e.g.
// adapter/db/postgres.
//
// Relation methods only ever return/accept Team — e.g. ListFromAccount
// returns the teams an account belongs to, not the account itself.
type TeamRepository interface {
	Create(ctx context.Context, team model.Team) (model.Team, error)
	Get(ctx context.Context, id model.ID) (model.Team, error)
	List(ctx context.Context, filters model.TeamFilters) (model.List[model.Team], error)
	Update(ctx context.Context, team model.Team) (model.Team, error)
	Delete(ctx context.Context, id model.ID) error

	ListFromAccount(ctx context.Context, accountID model.ID) (model.List[model.Team], error)
	AddAccount(ctx context.Context, teamID, accountID model.ID) error
	RemoveAccount(ctx context.Context, teamID, accountID model.ID) error

	ListFromFleet(ctx context.Context, fleetID model.ID) (model.List[model.Team], error)
	AddFleet(ctx context.Context, teamID, fleetID model.ID) error
	RemoveFleet(ctx context.Context, teamID, fleetID model.ID) error
}

// TeamService is the driving (primary) port exposing team-related business
// operations to inbound adapters, e.g. adapter/http controllers.
type TeamService interface {
	Create(ctx context.Context, team model.Team) (model.Team, error)
	Get(ctx context.Context, id model.ID) (model.Team, error)
	List(ctx context.Context, filters model.TeamFilters) (model.List[model.Team], error)
	Update(ctx context.Context, team model.Team) (model.Team, error)
	Delete(ctx context.Context, id model.ID) error

	ListFromAccount(ctx context.Context, accountID model.ID) (model.List[model.Team], error)
	AddAccount(ctx context.Context, teamID, accountID model.ID) error
	RemoveAccount(ctx context.Context, teamID, accountID model.ID) error

	ListFromFleet(ctx context.Context, fleetID model.ID) (model.List[model.Team], error)
	AddFleet(ctx context.Context, teamID, fleetID model.ID) error
	RemoveFleet(ctx context.Context, teamID, fleetID model.ID) error
}
