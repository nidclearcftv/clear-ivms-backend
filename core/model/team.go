package model

import "time"

type Team struct {
	ID             ID
	Name           string
	OrganizationID ID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TeamFilters narrows a team listing. OrganizationID is set by
// TeamService.List from the request's context (see utils.OrganizationID),
// not by callers directly — teams are always scoped to the current
// organization.
type TeamFilters struct {
	OrganizationID ID
}

func (f *TeamFilters) String() string {
	return "organization_id:" + string(f.OrganizationID)
}
