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
	// Vehicles backs AddVehicle's cross-organization check — see AddVehicle.
	Vehicles port.VehicleRepository `validate:"required"`
}

// GroupService implements port.GroupService by delegating directly to a
// port.GroupRepository.
type GroupService struct {
	repo     port.GroupRepository
	accounts port.AccountRepository
	vehicles port.VehicleRepository
}

func NewGroupService(opts GroupServiceOptions) (*GroupService, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	return &GroupService{repo: opts.Repository, accounts: opts.Accounts, vehicles: opts.Vehicles}, nil
}

func (s *GroupService) Create(ctx context.Context, group model.Group) (model.Group, error) {
	return s.repo.Create(ctx, group)
}

// Get fails with ErrCodeGroupNotFound if id exists but belongs to a
// different organization than the request's — reported the same as a
// nonexistent id, so a caller can't distinguish "not found" from
// "not yours" for another organization's group.
func (s *GroupService) Get(ctx context.Context, id model.ID) (model.Group, error) {
	group, err := s.repo.Get(ctx, id)
	if err != nil {
		return model.Group{}, err
	}
	if group.OrganizationID != utils.OrganizationID(ctx) {
		return model.Group{}, model.NewError(model.ErrCodeGroupNotFound, nil)
	}
	return group, nil
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

// Update fails with ErrCodeGroupNotFound the same way Get does for a group
// belonging to a different organization. group.OrganizationID is always
// overwritten with the group's existing (verified) organization — never
// trusted from the caller — so this can't be used to move a group into a
// different organization.
func (s *GroupService) Update(ctx context.Context, group model.Group) (model.Group, error) {
	existing, err := s.repo.Get(ctx, group.ID)
	if err != nil {
		return model.Group{}, err
	}
	if existing.OrganizationID != utils.OrganizationID(ctx) {
		return model.Group{}, model.NewError(model.ErrCodeGroupNotFound, nil)
	}
	group.OrganizationID = existing.OrganizationID
	return s.repo.Update(ctx, group)
}

// Delete fails with ErrCodeGroupNotFound the same way Get does for a group
// belonging to a different organization.
func (s *GroupService) Delete(ctx context.Context, id model.ID) error {
	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if existing.OrganizationID != utils.OrganizationID(ctx) {
		return model.NewError(model.ErrCodeGroupNotFound, nil)
	}
	return s.repo.Delete(ctx, id)
}

func (s *GroupService) ListFromAccount(ctx context.Context, accountID model.ID) (model.List[model.Group], error) {
	return s.repo.ListFromAccount(ctx, accountID)
}

func (s *GroupService) CountFromAccount(ctx context.Context, accountID model.ID) (int, error) {
	return s.repo.CountFromAccount(ctx, accountID)
}

// AddAccount fails with ErrCodeGroupNotFound for a group belonging to a
// different organization than the request's (same reasoning as Get), and
// rejects accounts that aren't model.AccountTypeUser with
// ErrCodeAccountTypeNotAllowedInGroup; see AccountService.AddGroup for the
// same account-type rule enforced from the other direction.
func (s *GroupService) AddAccount(ctx context.Context, groupID, accountID model.ID) error {
	group, err := s.repo.Get(ctx, groupID)
	if err != nil {
		return err
	}
	if group.OrganizationID != utils.OrganizationID(ctx) {
		return model.NewError(model.ErrCodeGroupNotFound, nil)
	}

	account, err := s.accounts.Get(ctx, accountID)
	if err != nil {
		return err
	}
	if account.Type != model.AccountTypeUser {
		return model.NewError(model.ErrCodeAccountTypeNotAllowedInGroup, nil)
	}

	return s.repo.AddAccount(ctx, groupID, accountID)
}

// RemoveAccount fails with ErrCodeGroupNotFound the same way AddAccount
// does for a group belonging to a different organization.
func (s *GroupService) RemoveAccount(ctx context.Context, groupID, accountID model.ID) error {
	group, err := s.repo.Get(ctx, groupID)
	if err != nil {
		return err
	}
	if group.OrganizationID != utils.OrganizationID(ctx) {
		return model.NewError(model.ErrCodeGroupNotFound, nil)
	}

	return s.repo.RemoveAccount(ctx, groupID, accountID)
}

// AddVehicle fails with ErrCodeGroupNotFound for a group belonging to a
// different organization than the request's (same reasoning as Get), and
// with ErrCodeVehicleNotFound the same way for a vehicle belonging to a
// different organization — a vehicle can't be moved into a group outside
// its own organization.
func (s *GroupService) AddVehicle(ctx context.Context, groupID, vehicleID model.ID) error {
	group, err := s.repo.Get(ctx, groupID)
	if err != nil {
		return err
	}
	if group.OrganizationID != utils.OrganizationID(ctx) {
		return model.NewError(model.ErrCodeGroupNotFound, nil)
	}

	vehicle, err := s.vehicles.Get(ctx, vehicleID)
	if err != nil {
		return err
	}
	if vehicle.OrganizationID != utils.OrganizationID(ctx) {
		return model.NewError(model.ErrCodeVehicleNotFound, nil)
	}

	return s.repo.AddVehicle(ctx, groupID, vehicleID)
}

// RemoveVehicle fails with ErrCodeGroupNotFound the same way AddVehicle
// does for a group belonging to a different organization.
func (s *GroupService) RemoveVehicle(ctx context.Context, groupID, vehicleID model.ID) error {
	group, err := s.repo.Get(ctx, groupID)
	if err != nil {
		return err
	}
	if group.OrganizationID != utils.OrganizationID(ctx) {
		return model.NewError(model.ErrCodeGroupNotFound, nil)
	}

	return s.repo.RemoveVehicle(ctx, groupID, vehicleID)
}

var _ port.GroupService = (*GroupService)(nil)
