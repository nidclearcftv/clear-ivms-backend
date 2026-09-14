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

// GroupSortField is a column GroupFilters.SortBy can order List by.
type GroupSortField string

const (
	GroupSortByName      GroupSortField = "name"
	GroupSortByCreatedAt GroupSortField = "createdAt"
	GroupSortByUpdatedAt GroupSortField = "updatedAt"
)

// GroupDefaultPageSize and GroupMaxPageSize bound GroupFilters.PageSize:
// unset (zero) defaults to GroupDefaultPageSize, and any larger value is
// capped at GroupMaxPageSize.
const (
	GroupDefaultPageSize = 20
	GroupMaxPageSize     = 100
)

// GroupFilters narrows/orders/paginates a group listing. OrganizationID is
// set by GroupService.List from the request's context (see
// utils.OrganizationID), not by callers directly — groups are always
// scoped to the current organization. Search, when set, matches groups
// whose name contains it (case-insensitive). SortBy defaults to createdAt
// (descending) when unset; SortDir defaults to ascending when SortBy is
// set but SortDir isn't. Page defaults to 1 and PageSize to
// GroupDefaultPageSize when unset.
type GroupFilters struct {
	OrganizationID ID
	Search         string         `form:"search"`
	SortBy         GroupSortField `form:"sortBy" binding:"omitempty,oneof=name createdAt updatedAt"`
	SortDir        SortDirection  `form:"sortDir" binding:"omitempty,oneof=asc desc"`
	Page           int            `form:"page" binding:"omitempty,min=1"`
	PageSize       int            `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

func (f *GroupFilters) String() string {
	return "organization_id:" + string(f.OrganizationID) + ":search:" + f.Search + ":sort_by:" + string(f.SortBy) + ":sort_dir:" + string(f.SortDir)
}
