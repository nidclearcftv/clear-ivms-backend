package model

import "time"

// Group is the domain read model. ParentID is nullable: a group can nest
// under another group, or sit at the top level.
type Group struct {
	ID             ID
	Name           string
	OrganizationID ID
	ParentID       *ID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// GroupFilters narrows a group listing. OrganizationID is set by
// GroupService.List from the request's context (see utils.OrganizationID),
// not by callers directly — groups are always scoped to the current
// organization.
type GroupFilters struct {
	OrganizationID ID
}

func (f *GroupFilters) String() string {
	return "organization_id:" + string(f.OrganizationID)
}
