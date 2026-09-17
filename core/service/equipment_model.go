package service

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
	"github.com/nidclearcftv/clear-ivms-backend/utils"
	"github.com/nidclearcftv/clear-ivms-backend/utils/validate"
)

type EquipmentModelServiceOptions struct {
	Repository port.EquipmentModelRepository `validate:"required"`
}

// EquipmentModelService implements port.EquipmentModelService by
// delegating directly to a port.EquipmentModelRepository.
type EquipmentModelService struct {
	repo port.EquipmentModelRepository
}

func NewEquipmentModelService(opts EquipmentModelServiceOptions) (*EquipmentModelService, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	return &EquipmentModelService{repo: opts.Repository}, nil
}

func (s *EquipmentModelService) Create(ctx context.Context, equipmentModel model.EquipmentModel) (model.EquipmentModel, error) {
	return s.repo.Create(ctx, equipmentModel)
}

// Get fails with ErrCodeEquipmentModelNotFound if id exists but belongs
// to a different organization than the request's — reported the same as
// a nonexistent id, so a caller can't distinguish "not found" from "not
// yours" for another organization's equipment model.
func (s *EquipmentModelService) Get(ctx context.Context, id model.ID) (model.EquipmentModel, error) {
	equipmentModel, err := s.repo.Get(ctx, id)
	if err != nil {
		return model.EquipmentModel{}, err
	}
	if equipmentModel.OrganizationID != utils.OrganizationID(ctx) {
		return model.EquipmentModel{}, model.NewError(model.ErrCodeEquipmentModelNotFound, nil)
	}
	return equipmentModel, nil
}

// List always scopes to the current organization: filters.OrganizationID
// is overwritten from ctx, never trusted from the caller.
func (s *EquipmentModelService) List(ctx context.Context, filters model.EquipmentModelFilters) (model.List[model.EquipmentModel], error) {
	filters.OrganizationID = utils.OrganizationID(ctx)
	return s.repo.List(ctx, filters)
}

// Count scopes the same way List does.
func (s *EquipmentModelService) Count(ctx context.Context, filters model.EquipmentModelFilters) (int, error) {
	filters.OrganizationID = utils.OrganizationID(ctx)
	return s.repo.Count(ctx, filters)
}

// Update fails with ErrCodeEquipmentModelNotFound the same way Get does
// for an equipment model belonging to a different organization.
// equipmentModel.OrganizationID is always overwritten with its existing
// (verified) organization — never trusted from the caller — so this
// can't be used to move an equipment model into a different
// organization. Public is likewise always overwritten with its existing
// value — see SetPublic, the only way to change it.
func (s *EquipmentModelService) Update(ctx context.Context, equipmentModel model.EquipmentModel) (model.EquipmentModel, error) {
	existing, err := s.repo.Get(ctx, equipmentModel.ID)
	if err != nil {
		return model.EquipmentModel{}, err
	}
	if existing.OrganizationID != utils.OrganizationID(ctx) {
		return model.EquipmentModel{}, model.NewError(model.ErrCodeEquipmentModelNotFound, nil)
	}
	equipmentModel.OrganizationID = existing.OrganizationID
	equipmentModel.Public = existing.Public
	return s.repo.Update(ctx, equipmentModel)
}

// Delete fails with ErrCodeEquipmentModelNotFound the same way Get does
// for an equipment model belonging to a different organization.
func (s *EquipmentModelService) Delete(ctx context.Context, id model.ID) error {
	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if existing.OrganizationID != utils.OrganizationID(ctx) {
		return model.NewError(model.ErrCodeEquipmentModelNotFound, nil)
	}
	return s.repo.Delete(ctx, id)
}

// SetPublic fails with ErrCodeEquipmentModelNotFound the same way Get
// does for an equipment model belonging to a different organization than
// the request's — enforced even for the admin-only caller allowed to
// reach this (see registerEquipmentModelRoutes): it's an org-scoping
// consistency check, not a permission check, so a stale/wrong ?orgId=
// still 404s instead of silently acting on the wrong organization.
func (s *EquipmentModelService) SetPublic(ctx context.Context, id model.ID, public bool) error {
	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if existing.OrganizationID != utils.OrganizationID(ctx) {
		return model.NewError(model.ErrCodeEquipmentModelNotFound, nil)
	}
	return s.repo.SetPublic(ctx, id, public)
}

var _ port.EquipmentModelService = (*EquipmentModelService)(nil)
