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

type BrandingFilters struct {
}

func (f *BrandingFilters) String() string {
	return ""
}
