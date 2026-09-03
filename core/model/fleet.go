package model

import "time"

type Fleet struct {
	ID             ID
	Name           string
	OrganizationID ID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// FleetFilters narrows a fleet listing. OrganizationID is set by
// FleetService.List from the request's context (see utils.OrganizationID),
// not by callers directly — fleets are always scoped to the current
// organization.
type FleetFilters struct {
	OrganizationID ID
}

func (f *FleetFilters) String() string {
	return "organization_id:" + string(f.OrganizationID)
}
