package model

import "time"

// EquipmentModelType categorizes an equipment model as either the primary
// piece of equipment or an accessory to one.
type EquipmentModelType string

const (
	EquipmentModelTypePrimary   EquipmentModelType = "primary"
	EquipmentModelTypeAccessory EquipmentModelType = "accessory"
)

// EquipmentModel is the domain read model for an organization's equipment
// catalog entry. Public defaults to false and is deliberately excluded
// from EquipmentModelService.Update — see SetPublic, the only way to
// change it, which (unlike every other operation here) is admin-only.
type EquipmentModel struct {
	ID             ID
	Name           string
	Description    string
	Type           EquipmentModelType
	Public         bool
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
// when set, matches equipment models whose name or description contains
// it (case-insensitive). Type, when set, narrows to that single type.
// SortBy defaults to createdAt (descending) when unset; SortDir defaults
// to ascending when SortBy is set but SortDir isn't. Page defaults to 1
// and PageSize to EquipmentModelDefaultPageSize when unset.
type EquipmentModelFilters struct {
	OrganizationID ID
	Search         string                  `form:"search"`
	Type           EquipmentModelType      `form:"type" binding:"omitempty,oneof=primary accessory"`
	SortBy         EquipmentModelSortField `form:"sortBy" binding:"omitempty,oneof=name type createdAt updatedAt"`
	SortDir        SortDirection           `form:"sortDir" binding:"omitempty,oneof=asc desc"`
	Page           int                     `form:"page" binding:"omitempty,min=1"`
	PageSize       int                     `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

func (f *EquipmentModelFilters) String() string {
	return "organization_id:" + string(f.OrganizationID) + ":search:" + f.Search + ":type:" + string(f.Type) + ":sort_by:" + string(f.SortBy) + ":sort_dir:" + string(f.SortDir)
}
