package port

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// GroupRepository is the driven (secondary) port for persisting and
// reading group data. It is implemented by outbound adapters, e.g.
// adapter/db/postgres.
//
// Relation methods only ever return/accept Group — e.g. ListFromAccount
// returns the groups an account belongs to, not the account itself.
type GroupRepository interface {
	Create(ctx context.Context, group model.Group) (model.Group, error)
	Get(ctx context.Context, id model.ID) (model.Group, error)
	List(ctx context.Context, filters model.GroupFilters) (model.List[model.Group], error)
	Update(ctx context.Context, group model.Group) (model.Group, error)
	Delete(ctx context.Context, id model.ID) error

	ListFromAccount(ctx context.Context, accountID model.ID) (model.List[model.Group], error)
	AddAccount(ctx context.Context, groupID, accountID model.ID) error
	RemoveAccount(ctx context.Context, groupID, accountID model.ID) error
}

// GroupService is the driving (primary) port exposing group-related
// business operations to inbound adapters, e.g. adapter/http controllers.
type GroupService interface {
	Create(ctx context.Context, group model.Group) (model.Group, error)
	Get(ctx context.Context, id model.ID) (model.Group, error)
	List(ctx context.Context, filters model.GroupFilters) (model.List[model.Group], error)
	Update(ctx context.Context, group model.Group) (model.Group, error)
	Delete(ctx context.Context, id model.ID) error

	ListFromAccount(ctx context.Context, accountID model.ID) (model.List[model.Group], error)
	// AddAccount fails with ErrCodeAccountTypeNotAllowedInGroup unless
	// accountID's type is model.AccountTypeUser — admins and org_admins
	// already have broader access and aren't meant to be scoped to a
	// group.
	AddAccount(ctx context.Context, groupID, accountID model.ID) error
	RemoveAccount(ctx context.Context, groupID, accountID model.ID) error
}
