package model

import "time"

// Vehicle is the domain read model. GroupID is nullable: a vehicle can be
// registered to an organization without yet being assigned to a group.
type Vehicle struct {
	ID             ID
	OrganizationID ID
	GroupID        *ID
	IVMSType       IVMSType
	ExternalID     string
	PlateNumber    string
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
