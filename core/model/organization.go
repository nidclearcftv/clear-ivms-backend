package model

import "time"

type Organization struct {
	ID        ID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type OrganizationFilters struct {
}

func (f *OrganizationFilters) String() string {
	return ""
}
