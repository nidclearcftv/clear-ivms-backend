package model

import "time"

type Organization struct {
	ID        ID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// OrganizationSortField is a column OrganizationFilters.SortBy can order
// List by.
type OrganizationSortField string

const (
	OrganizationSortByName      OrganizationSortField = "name"
	OrganizationSortByCreatedAt OrganizationSortField = "createdAt"
	OrganizationSortByUpdatedAt OrganizationSortField = "updatedAt"
)

// OrganizationDefaultPageSize and OrganizationMaxPageSize bound
// OrganizationFilters.PageSize: unset (zero) defaults to
// OrganizationDefaultPageSize, and any larger value is capped at
// OrganizationMaxPageSize.
const (
	OrganizationDefaultPageSize = 20
	OrganizationMaxPageSize     = 100
)

// OrganizationFilters narrows/orders/paginates an organization listing.
// SortBy defaults to createdAt (descending) when unset; SortDir defaults to
// ascending when SortBy is set but SortDir isn't. Search, when set, matches
// organizations whose name contains it (case-insensitive).
// CreatedFrom/CreatedTo, when set, narrow to organizations created on or
// after / on or before that calendar date (inclusive on both ends). Page
// defaults to 1 and PageSize to OrganizationDefaultPageSize when unset.
type OrganizationFilters struct {
	Search      string                `form:"search"`
	CreatedFrom *time.Time            `form:"createdFrom" time_format:"2006-01-02"`
	CreatedTo   *time.Time            `form:"createdTo" time_format:"2006-01-02"`
	SortBy      OrganizationSortField `form:"sortBy" binding:"omitempty,oneof=name createdAt updatedAt"`
	SortDir     SortDirection         `form:"sortDir" binding:"omitempty,oneof=asc desc"`
	Page        int                   `form:"page" binding:"omitempty,min=1"`
	PageSize    int                   `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

func (f *OrganizationFilters) String() string {
	s := "search:" + f.Search + ":sort_by:" + string(f.SortBy) + ":sort_dir:" + string(f.SortDir)
	if f.CreatedFrom != nil {
		s += ":created_from:" + f.CreatedFrom.Format("2006-01-02")
	}
	if f.CreatedTo != nil {
		s += ":created_to:" + f.CreatedTo.Format("2006-01-02")
	}
	return s
}
