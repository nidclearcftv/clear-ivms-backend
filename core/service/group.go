package service

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
	"github.com/nidclearcftv/clear-ivms-backend/utils"
	"github.com/nidclearcftv/clear-ivms-backend/utils/validate"
)

type GroupServiceOptions struct {
	Repository port.GroupRepository `validate:"required"`

	// Accounts backs AddAccount's account-type check — see AddAccount.
	Accounts port.AccountRepository `validate:"required"`
}

// GroupService implements port.GroupService by delegating directly to a
// port.GroupRepository.
type GroupService struct {
	repo     port.GroupRepository
	accounts port.AccountRepository
}

func NewGroupService(opts GroupServiceOptions) (*GroupService, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	return &GroupService{repo: opts.Repository, accounts: opts.Accounts}, nil
}

func (s *GroupService) Create(ctx context.Context, group model.Group) (model.Group, error) {
	return s.repo.Create(ctx, group)
}

func (s *GroupService) Get(ctx context.Context, id model.ID) (model.Group, error) {
	return s.repo.Get(ctx, id)
}

// List always scopes to the current organization: filters.OrganizationID
// is overwritten from ctx, never trusted from the caller.
func (s *GroupService) List(ctx context.Context, filters model.GroupFilters) (model.List[model.Group], error) {
	filters.OrganizationID = utils.OrganizationID(ctx)
	return s.repo.List(ctx, filters)
}

// Count scopes the same way List does.
func (s *GroupService) Count(ctx context.Context, filters model.GroupFilters) (int, error) {
	filters.OrganizationID = utils.OrganizationID(ctx)
	return s.repo.Count(ctx, filters)
}

func (s *GroupService) Update(ctx context.Context, group model.Group) (model.Group, error) {
	return s.repo.Update(ctx, group)
}

func (s *GroupService) Delete(ctx context.Context, id model.ID) error {
	return s.repo.Delete(ctx, id)
}

func (s *GroupService) ListFromAccount(ctx context.Context, accountID model.ID) (model.List[model.Group], error) {
	return s.repo.ListFromAccount(ctx, accountID)
}

func (s *GroupService) CountFromAccount(ctx context.Context, accountID model.ID) (int, error) {
	return s.repo.CountFromAccount(ctx, accountID)
}

// AddAccount rejects accounts that aren't model.AccountTypeUser with
// ErrCodeAccountTypeNotAllowedInGroup; see
// AccountService.AddGroup for the same rule enforced from the other
// direction.
func (s *GroupService) AddAccount(ctx context.Context, groupID, accountID model.ID) error {
	account, err := s.accounts.Get(ctx, accountID)
	if err != nil {
		return err
	}
	if account.Type != model.AccountTypeUser {
		return model.NewError(model.ErrCodeAccountTypeNotAllowedInGroup, nil)
	}

	return s.repo.AddAccount(ctx, groupID, accountID)
}

func (s *GroupService) RemoveAccount(ctx context.Context, groupID, accountID model.ID) error {
	return s.repo.RemoveAccount(ctx, groupID, accountID)
}

var _ port.GroupService = (*GroupService)(nil)
