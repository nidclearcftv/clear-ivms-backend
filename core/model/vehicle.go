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
	IVMSType       IVMSType
	ExternalID     string
	PlateNumber    string
	Status         VehicleStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// VehicleFilters narrows a vehicle listing. OrganizationID is set by
// VehicleService.List from the request's context (see utils.OrganizationID),
// not by callers directly — vehicles are always scoped to the current
// organization. GroupID, if set, additionally narrows to a single group.
type VehicleFilters struct {
	OrganizationID ID
	GroupID        ID `form:"groupId"`
}

func (f *VehicleFilters) String() string {
	return "organization_id:" + string(f.OrganizationID) + ":group_id:" + string(f.GroupID)
}
