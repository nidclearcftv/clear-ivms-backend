package model

import "time"

// Vehicle is the domain read model. FleetID is nullable: a vehicle can be
// registered to an organization without yet being assigned to a fleet.
type Vehicle struct {
	ID             ID
	OrganizationID ID
	FleetID        *ID
	IVMSType       IVMSType
	ExternalID     string
	PlateNumber    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// VehicleFilters narrows a vehicle listing. OrganizationID is set by
// VehicleService.List from the request's context (see utils.OrganizationID),
// not by callers directly — vehicles are always scoped to the current
// organization. FleetID, if set, additionally narrows to a single fleet.
type VehicleFilters struct {
	OrganizationID ID
	FleetID        ID `form:"fleetId"`
}

func (f *VehicleFilters) String() string {
	return "organization_id:" + string(f.OrganizationID) + ":fleet_id:" + string(f.FleetID)
}
