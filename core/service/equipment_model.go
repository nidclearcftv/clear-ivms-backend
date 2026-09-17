package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
	"github.com/nidclearcftv/clear-ivms-backend/core/port"
	"github.com/nidclearcftv/clear-ivms-backend/utils"
	"github.com/nidclearcftv/clear-ivms-backend/utils/validate"
)

type EquipmentModelServiceOptions struct {
	Repository port.EquipmentModelRepository `validate:"required"`
	// Storage backs SetPicture/GetPictureURL/DeletePicture.
	Storage port.ObjectStorage `validate:"required"`
}

// EquipmentModelService implements port.EquipmentModelService by
// delegating directly to a port.EquipmentModelRepository, plus a
// port.ObjectStorage for picture uploads.
type EquipmentModelService struct {
	repo    port.EquipmentModelRepository
	storage port.ObjectStorage
}

func NewEquipmentModelService(opts EquipmentModelServiceOptions) (*EquipmentModelService, error) {
	if err := validate.Struct(opts); err != nil {
		return nil, err
	}

	return &EquipmentModelService{repo: opts.Repository, storage: opts.Storage}, nil
}

// newObjectKey generates a random, unpredictable object storage key — the
// same crypto/rand + hex construction AccountService.newSessionToken uses
// for session tokens, appropriate here for the same reason: it needs to be
// unique and non-guessable, not human-meaningful.
func newObjectKey() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

func (s *EquipmentModelService) Create(ctx context.Context, equipmentModel model.EquipmentModel) (model.EquipmentModel, error) {
	return s.repo.Create(ctx, equipmentModel)
}

// Get fails with ErrCodeEquipmentModelNotFound if id exists but belongs
// to a different organization than the request's AND isn't public —
// reported the same as a nonexistent id, so a caller can't distinguish
// "not found" from "not yours" for another organization's private
// equipment model. A public equipment model is readable from any
// organization (see List), but this does not make it editable — Update,
// Delete and SetPublic each re-check ownership independently against the
// repository directly, not through this relaxed check.
func (s *EquipmentModelService) Get(ctx context.Context, id model.ID) (model.EquipmentModel, error) {
	equipmentModel, err := s.repo.Get(ctx, id)
	if err != nil {
		return model.EquipmentModel{}, err
	}
	if equipmentModel.OrganizationID != utils.OrganizationID(ctx) && !equipmentModel.Public {
		return model.EquipmentModel{}, model.NewError(model.ErrCodeEquipmentModelNotFound, nil)
	}
	return equipmentModel, nil
}

// List scopes to the current organization's own equipment models (public
// or not) plus every other organization's public equipment models —
// filters.OrganizationID is overwritten from ctx, never trusted from the
// caller, but the repository treats it as "at least this organization",
// not "only this organization" (see applyEquipmentModelFilters).
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
// organization. Public and PictureObjectKey are likewise always
// overwritten with their existing values — see SetPublic and
// SetPicture/DeletePicture, the only ways to change them.
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
	equipmentModel.PictureObjectKey = existing.PictureObjectKey
	return s.repo.Update(ctx, equipmentModel)
}

// Delete fails with ErrCodeEquipmentModelNotFound the same way Get does
// for an equipment model belonging to a different organization. Its
// picture, if any, is deleted from storage on a best-effort basis — see
// SetPicture for why a storage failure here doesn't fail the request.
func (s *EquipmentModelService) Delete(ctx context.Context, id model.ID) error {
	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if existing.OrganizationID != utils.OrganizationID(ctx) {
		return model.NewError(model.ErrCodeEquipmentModelNotFound, nil)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	if existing.PictureObjectKey != nil {
		_ = s.storage.Delete(ctx, *existing.PictureObjectKey)
	}
	return nil
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

// SetPicture fails with ErrCodeEquipmentModelNotFound the same way
// SetPublic does for an equipment model belonging to a different
// organization. It generates a fresh key, asks storage.PutURL for a URL
// the client can upload directly to, and — unlike a typical presigned
// upload flow — records that key on the equipment model immediately,
// before any bytes have actually been uploaded: there is no confirmation
// step (see registerEquipmentModelRoutes), so this is the only point at
// which the equipment model can be updated at all. A client that requests
// this URL and never follows through (or whose upload fails) leaves the
// equipment model pointed at an object that doesn't exist — GetPictureURL
// will still return a URL for it, which will 404 when fetched, until
// SetPicture or DeletePicture is called again. The old picture, if any,
// is deleted from storage on a best-effort basis once the new reference
// is safely recorded — a failure to clean it up is a harmless storage
// leak, not worth failing an otherwise-successful request over.
func (s *EquipmentModelService) SetPicture(ctx context.Context, id model.ID, contentType string) (string, error) {
	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return "", err
	}
	if existing.OrganizationID != utils.OrganizationID(ctx) {
		return "", model.NewError(model.ErrCodeEquipmentModelNotFound, nil)
	}

	key, err := newObjectKey()
	if err != nil {
		return "", fmt.Errorf("service: failed to generate object key: %w", err)
	}

	url, err := s.storage.PutURL(ctx, key, contentType)
	if err != nil {
		return "", err
	}

	if err := s.repo.SetPictureObjectKey(ctx, id, &key); err != nil {
		return "", err
	}

	if existing.PictureObjectKey != nil {
		_ = s.storage.Delete(ctx, *existing.PictureObjectKey)
	}

	return url, nil
}

// GetPictureURL fails with ErrCodeEquipmentModelPictureNotFound if the
// equipment model has no picture set. It resolves the equipment model via
// Get, not the stricter check SetPicture/DeletePicture use, so a public
// equipment model's picture is readable from any organization the same
// way its other fields are.
func (s *EquipmentModelService) GetPictureURL(ctx context.Context, id model.ID) (string, error) {
	equipmentModel, err := s.Get(ctx, id)
	if err != nil {
		return "", err
	}
	if equipmentModel.PictureObjectKey == nil {
		return "", model.NewError(model.ErrCodeEquipmentModelPictureNotFound, nil)
	}
	return s.storage.GetURL(ctx, *equipmentModel.PictureObjectKey)
}

// DeletePicture fails with ErrCodeEquipmentModelNotFound the same way
// SetPicture does for an equipment model belonging to a different
// organization. Deleting when no picture is set is a no-op, not an
// error. The stored object's deletion is best-effort — see SetPicture.
func (s *EquipmentModelService) DeletePicture(ctx context.Context, id model.ID) (model.EquipmentModel, error) {
	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return model.EquipmentModel{}, err
	}
	if existing.OrganizationID != utils.OrganizationID(ctx) {
		return model.EquipmentModel{}, model.NewError(model.ErrCodeEquipmentModelNotFound, nil)
	}
	if existing.PictureObjectKey == nil {
		return existing, nil
	}

	if err := s.repo.SetPictureObjectKey(ctx, id, nil); err != nil {
		return model.EquipmentModel{}, err
	}
	_ = s.storage.Delete(ctx, *existing.PictureObjectKey)

	existing.PictureObjectKey = nil
	return existing, nil
}

var _ port.EquipmentModelService = (*EquipmentModelService)(nil)
