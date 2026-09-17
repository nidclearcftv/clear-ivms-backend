package model

import (
	"strings"
	"time"
)

// EquipmentModelType categorizes an equipment model as either the primary
// piece of equipment or an accessory to one.
type EquipmentModelType string

const (
	EquipmentModelTypePrimary   EquipmentModelType = "primary"
	EquipmentModelTypeAccessory EquipmentModelType = "accessory"
)

// EquipmentModelAllowedPictureContentTypes are the only content types an
// equipment model's picture may be declared as — ordinary
// web-displayable images, not arbitrary file uploads. Checked by
// registerEquipmentModelRoutes' PUT /:id/picture against the client's
// declared Content-Type before it's ever handed a presigned upload URL
// for it (see EquipmentModelService.SetPicture). This is the only check
// there is: the actual uploaded bytes go straight to storage, never
// through this backend, so nothing here re-verifies what was truly
// written.
var EquipmentModelAllowedPictureContentTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
	"image/avif": true,
}

// EquipmentModel is the domain read model for an organization's equipment
// catalog entry. Public defaults to false and is deliberately excluded
// from EquipmentModelService.Update — see SetPublic, the only way to
// change it, which (unlike every other operation here) is admin-only.
// PictureObjectKey is nil when no picture has been uploaded; like Public,
// it's excluded from Update — see SetPicture/DeletePicture, the only way
// to change it. It identifies an object in a port.ObjectStorage, not a
// user-facing value — never resolved/interpreted by callers outside
// EquipmentModelService.
type EquipmentModel struct {
	ID               ID
	Name             string
	Description      string
	Manufacturer     string
	Type             EquipmentModelType
	Public           bool
	PictureObjectKey *string
	// ExternalViewURL is an optional link to more information about this
	// equipment model hosted elsewhere (e.g. the manufacturer's own
	// product page) — purely informational, never resolved/fetched by
	// this backend.
	ExternalViewURL string
	// Features is a free-form, caller-defined set of tags describing this
	// equipment model (e.g. "gps", "camera") — never validated against a
	// fixed vocabulary. See EquipmentModelFilters.FeaturesInclude/
	// FeaturesExclude for how it's filtered on.
	Features       []string
	OrganizationID ID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// EquipmentModelSortField is a column EquipmentModelFilters.SortBy can
// order List by.
type EquipmentModelSortField string

const (
	EquipmentModelSortByName      EquipmentModelSortField = "name"
	EquipmentModelSortByType      EquipmentModelSortField = "type"
	EquipmentModelSortByCreatedAt EquipmentModelSortField = "createdAt"
	EquipmentModelSortByUpdatedAt EquipmentModelSortField = "updatedAt"
)

// EquipmentModelDefaultPageSize and EquipmentModelMaxPageSize bound
// EquipmentModelFilters.PageSize: unset (zero) defaults to
// EquipmentModelDefaultPageSize, and any larger value is capped at
// EquipmentModelMaxPageSize.
const (
	EquipmentModelDefaultPageSize = 20
	EquipmentModelMaxPageSize     = 100
)

// EquipmentModelFilters narrows/orders/paginates an equipment model
// listing. OrganizationID is set by EquipmentModelService.List from the
// request's context (see utils.OrganizationID), not by callers directly.
// It scopes the listing to that organization's own equipment models (public
// or not) plus every other organization's public equipment models — not
// exclusively to that organization; see applyEquipmentModelFilters. Search,
// when set, matches equipment models whose name, description, or
// manufacturer contains it (case-insensitive). Type, when set, narrows to
// that single type. FeaturesInclude, when set, narrows to equipment models
// whose Features contains every listed value (AND); FeaturesExclude, when
// set, excludes any equipment model whose Features contains any listed
// value (OR). The two are independent and may both be set at once.
// SortBy defaults to createdAt (descending) when unset; SortDir defaults
// to ascending when SortBy is set but SortDir isn't. Page defaults to 1
// and PageSize to EquipmentModelDefaultPageSize when unset.
type EquipmentModelFilters struct {
	OrganizationID  ID
	Search          string                  `form:"search"`
	Type            EquipmentModelType      `form:"type" binding:"omitempty,oneof=primary accessory"`
	FeaturesInclude []string                `form:"featuresInclude" collection_format:"csv"`
	FeaturesExclude []string                `form:"featuresExclude" collection_format:"csv"`
	SortBy          EquipmentModelSortField `form:"sortBy" binding:"omitempty,oneof=name type createdAt updatedAt"`
	SortDir         SortDirection           `form:"sortDir" binding:"omitempty,oneof=asc desc"`
	Page            int                     `form:"page" binding:"omitempty,min=1"`
	PageSize        int                     `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

func (f *EquipmentModelFilters) String() string {
	return "organization_id:" + string(f.OrganizationID) + ":search:" + f.Search + ":type:" + string(f.Type) + ":features_include:" + strings.Join(f.FeaturesInclude, ",") + ":features_exclude:" + strings.Join(f.FeaturesExclude, ",") + ":sort_by:" + string(f.SortBy) + ":sort_dir:" + string(f.SortDir)
}
