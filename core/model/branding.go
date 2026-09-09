package model

import "time"

// Branding is a per-domain white-label config (theme, logo, copy, ...),
// looked up by the domain a client is served from. Config is opaque JSON —
// the schema of what it holds is entirely up to callers.
type Branding struct {
	ID        ID
	Name      string
	Domain    string
	Config    JSON
	CreatedAt time.Time
	UpdatedAt time.Time
}

// BrandingSortField is a column BrandingFilters.SortBy can order List by.
type BrandingSortField string

const (
	BrandingSortByName      BrandingSortField = "name"
	BrandingSortByDomain    BrandingSortField = "domain"
	BrandingSortByCreatedAt BrandingSortField = "createdAt"
)

type SortDirection string

const (
	SortDirectionAsc  SortDirection = "asc"
	SortDirectionDesc SortDirection = "desc"
)

// BrandingFilters narrows/orders a branding listing. SortBy defaults to
// createdAt (descending) when unset; SortDir defaults to ascending when
// SortBy is set but SortDir isn't.
type BrandingFilters struct {
	SortBy  BrandingSortField `form:"sortBy" binding:"omitempty,oneof=name domain createdAt"`
	SortDir SortDirection     `form:"sortDir" binding:"omitempty,oneof=asc desc"`
}

func (f *BrandingFilters) String() string {
	return "sort_by:" + string(f.SortBy) + ":sort_dir:" + string(f.SortDir)
}

func BrandingKey(id ID) string {
	return "branding:" + string(id)
}

// BrandingDomainKey caches a branding by domain (see
// BrandingService.GetByDomain), separately from BrandingKey — the public
// branding lookup route only ever has the domain, not the ID.
func BrandingDomainKey(domain string) string {
	return "branding_domain:" + domain
}
