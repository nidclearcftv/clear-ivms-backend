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
// different organization. If group.ParentID is set, it's also validated by
// checkParentValid so a group can never become its own ancestor.
func (s *GroupService) Update(ctx context.Context, group model.Group) (model.Group, error) {
	existing, err := s.repo.Get(ctx, group.ID)
	if err != nil {
		return model.Group{}, err
	}
	if existing.OrganizationID != utils.OrganizationID(ctx) {
		return model.Group{}, model.NewError(model.ErrCodeGroupNotFound, nil)
	}
	group.OrganizationID = existing.OrganizationID

	if group.ParentID != nil {
		if err := s.checkParentValid(ctx, group.ID, *group.ParentID); err != nil {
			return model.Group{}, err
		}
	}

	return s.repo.Update(ctx, group)
}

// checkParentValid rejects a proposed new parent (candidateParentID) for
// groupID with ErrCodeGroupInvalidParent if candidateParentID is groupID
// itself, or one of groupID's own descendants — either would make groupID
// an ancestor of its own ancestor, i.e. a cycle. A candidateParentID from a
// different organization is rejected the same way Get does
// (ErrCodeGroupNotFound — cross-org is indistinguishable from
// nonexistent).
//
// Implemented with one Get (to validate/org-check candidateParentID) plus
// one ListAll (the same whole-org fetch GetTree uses) and an in-memory walk
// of the parent_id chain — a fixed two queries no matter how deep the
// hierarchy is, instead of one repo.Get per ancestor.
func (s *GroupService) checkParentValid(ctx context.Context, groupID, candidateParentID model.ID) error {
	orgID := utils.OrganizationID(ctx)

	parent, err := s.repo.Get(ctx, candidateParentID)
	if err != nil {
		return err
	}
	if parent.OrganizationID != orgID {
		return model.NewError(model.ErrCodeGroupNotFound, nil)
	}

	groups, err := s.repo.ListAll(ctx, orgID)
	if err != nil {
		return err
	}
	parentByID := make(map[model.ID]*model.ID, len(groups))
	for _, g := range groups {
		parentByID[g.ID] = g.ParentID
	}

	visited := map[model.ID]bool{}
	currentID := &candidateParentID
	for currentID != nil {
		if *currentID == groupID {
			return model.NewError(model.ErrCodeGroupInvalidParent, nil)
		}
		if visited[*currentID] {
			// Defensive: an unrelated pre-existing cycle. Shouldn't be
			// reachable given this same guard runs on every Update.
			return model.NewError(model.ErrCodeGroupInvalidParent, nil)
		}
		visited[*currentID] = true
		currentID = parentByID[*currentID]
	}
	return nil
}

// GetTree assembles every group and vehicle in the current request's
// organization into a nested hierarchy. A group whose ParentID doesn't
// resolve to another group in this organization (shouldn't happen given
// checkParentValid above, but defends against it) is treated as a root
// rather than dropped; a vehicle whose GroupID doesn't resolve is treated
// as unassigned the same way.
func (s *GroupService) GetTree(ctx context.Context) (model.GroupTree, error) {
	orgID := utils.OrganizationID(ctx)

	groups, err := s.repo.ListAll(ctx, orgID)
	if err != nil {
		return model.GroupTree{}, err
	}
	vehicles, err := s.vehicles.ListAll(ctx, orgID)
	if err != nil {
		return model.GroupTree{}, err
	}

	groupsByID := make(map[model.ID]model.Group, len(groups))
	for _, g := range groups {
		groupsByID[g.ID] = g
	}

	childIDsByParent := make(map[model.ID][]model.ID, len(groups))
	var rootIDs []model.ID
	for _, g := range groups {
		if g.ParentID != nil {
			if _, ok := groupsByID[*g.ParentID]; ok {
				childIDsByParent[*g.ParentID] = append(childIDsByParent[*g.ParentID], g.ID)
				continue
			}
		}
		rootIDs = append(rootIDs, g.ID)
	}

	vehiclesByGroup := make(map[model.ID][]model.Vehicle, len(vehicles))
	var unassigned []model.Vehicle
	for _, v := range vehicles {
		if v.GroupID == nil {
			unassigned = append(unassigned, v)
			continue
		}
		if _, ok := groupsByID[*v.GroupID]; !ok {
			unassigned = append(unassigned, v)
			continue
		}
		vehiclesByGroup[*v.GroupID] = append(vehiclesByGroup[*v.GroupID], v)
	}

	// visited bounds the walk against a corrupted/cyclic parent_id chain —
	// checkParentValid is what actually prevents cycles from ever being
	// written; this is a pure defensive backstop so GetTree itself can
	// never recurse forever or duplicate a node even if one slipped
	// through (e.g. a direct DB edit bypassing the app).
	visited := make(map[model.ID]bool, len(groups))
	var build func(id model.ID) model.GroupTreeNode
	build = func(id model.ID) model.GroupTreeNode {
		visited[id] = true
		node := model.GroupTreeNode{Group: groupsByID[id], Vehicles: vehiclesByGroup[id]}
		for _, childID := range childIDsByParent[id] {
			if visited[childID] {
				continue
			}
			node.Children = append(node.Children, build(childID))
		}
		return node
	}

	tree := model.GroupTree{UnassignedVehicles: unassigned}
	for _, id := range rootIDs {
		if visited[id] {
			continue
		}
		tree.Roots = append(tree.Roots, build(id))
	}
	return tree, nil
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
