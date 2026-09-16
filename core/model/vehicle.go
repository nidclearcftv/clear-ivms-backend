package model

import "time"

// VehicleStatus mirrors the vehicles.status CHECK constraint in the
// Postgres schema.
type VehicleStatus string

const (
	VehicleStatusOnline  VehicleStatus = "online"
	VehicleStatusOffline VehicleStatus = "offline"
)

// Vehicle is the domain read model. GroupID is nullable: a vehicle can be
// registered to an organization without yet being assigned to a group.
// Status defaults to VehicleStatusOffline and can only be changed via
// VehicleRepository.SetStatus/VehicleService.SetStatus — never through
// Update, so callers can't accidentally overwrite it as a side effect of
// an unrelated field change.
type Vehicle struct {
	ID             ID
	OrganizationID ID
	GroupID        *ID
	Name           string
	IVMSType       IVMSType
	ExternalID     string
	PlateNumber    string
	Status         VehicleStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// VehicleSortField is a column VehicleFilters.SortBy can order List by.
type VehicleSortField string

const (
	VehicleSortByName        VehicleSortField = "name"
	VehicleSortByPlateNumber VehicleSortField = "plateNumber"
	VehicleSortByStatus      VehicleSortField = "status"
	VehicleSortByCreatedAt   VehicleSortField = "createdAt"
	VehicleSortByUpdatedAt   VehicleSortField = "updatedAt"
)

// VehicleDefaultPageSize and VehicleMaxPageSize bound
// VehicleFilters.PageSize: unset (zero) defaults to
// VehicleDefaultPageSize, and any larger value is capped at
// VehicleMaxPageSize.
const (
	VehicleDefaultPageSize = 20
	VehicleMaxPageSize     = 100
)

// VehicleFilters narrows/orders/paginates a vehicle listing. OrganizationID
// is set by VehicleService.List from the request's context (see
// utils.OrganizationID), not by callers directly — vehicles are always
// scoped to the current organization. GroupID, if set, additionally
// narrows to a single group. Search, when set, matches vehicles whose name
// or plate number contains it (case-insensitive). SortBy defaults to
// createdAt (descending) when unset; SortDir defaults to ascending when
// SortBy is set but SortDir isn't. Page defaults to 1 and PageSize to
// VehicleDefaultPageSize when unset.
type VehicleFilters struct {
	OrganizationID ID
	GroupID        ID               `form:"groupId"`
	Search         string           `form:"search"`
	SortBy         VehicleSortField `form:"sortBy" binding:"omitempty,oneof=name plateNumber status createdAt updatedAt"`
	SortDir        SortDirection    `form:"sortDir" binding:"omitempty,oneof=asc desc"`
	Page           int              `form:"page" binding:"omitempty,min=1"`
	PageSize       int              `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

func (f *VehicleFilters) String() string {
	return "organization_id:" + string(f.OrganizationID) + ":group_id:" + string(f.GroupID) + ":search:" + f.Search + ":sort_by:" + string(f.SortBy) + ":sort_dir:" + string(f.SortDir)
}
