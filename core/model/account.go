package model

import (
	"strconv"
	"time"
)

// AccountType mirrors the accounts.type CHECK constraint in the Postgres
// schema.
type AccountType string

const (
	AccountTypeAdmin    AccountType = "admin"
	AccountTypeOrgAdmin AccountType = "org_admin"
	AccountTypeUser     AccountType = "user"
)

// Account is the domain read model. It deliberately excludes the password
// hash — that's a write-only credential, never something read back out
// through the domain layer. See AccountRepository.Create and SetPassword.
type Account struct {
	ID          ID
	Name        string
	Email       string
	PhoneNumber string
	Type        AccountType
	Blocked     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// AccountDefaultPageSize and AccountMaxPageSize bound
// AccountFilters.PageSize: unset (zero) defaults to
// AccountDefaultPageSize, and any larger value is capped at
// AccountMaxPageSize.
const (
	AccountDefaultPageSize = 20
	AccountMaxPageSize     = 100
)

// AccountSortField is a column AccountFilters.SortBy can order List by.
type AccountSortField string

const (
	AccountSortByName      AccountSortField = "name"
	AccountSortByEmail     AccountSortField = "email"
	AccountSortByCreatedAt AccountSortField = "createdAt"
	AccountSortByUpdatedAt AccountSortField = "updatedAt"
)

// AccountFilters narrows/orders/paginates an account listing. Search, when
// set, matches accounts whose name or email contains it (case-insensitive).
// Type, when set, narrows to that exact account type. Blocked, when set,
// narrows to accounts with that exact blocked status. CreatedFrom/CreatedTo,
// when set, narrow to accounts created on or after / on or before that
// calendar date (inclusive on both ends). SortBy defaults to name
// (ascending) when unset; SortDir defaults to ascending when SortBy is set
// but SortDir isn't. Page defaults to 1 and PageSize to
// AccountDefaultPageSize when unset.
type AccountFilters struct {
	Search      string           `form:"search"`
	Type        AccountType      `form:"type" binding:"omitempty,oneof=admin org_admin user"`
	Blocked     *bool            `form:"blocked"`
	CreatedFrom *time.Time       `form:"createdFrom" time_format:"2006-01-02"`
	CreatedTo   *time.Time       `form:"createdTo" time_format:"2006-01-02"`
	SortBy      AccountSortField `form:"sortBy" binding:"omitempty,oneof=name email createdAt updatedAt"`
	SortDir     SortDirection    `form:"sortDir" binding:"omitempty,oneof=asc desc"`
	Page        int              `form:"page" binding:"omitempty,min=1"`
	PageSize    int              `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

func (f *AccountFilters) String() string {
	s := "search:" + f.Search + ":type:" + string(f.Type) + ":sort_by:" + string(f.SortBy) + ":sort_dir:" + string(f.SortDir)
	if f.Blocked != nil {
		s += ":blocked:" + strconv.FormatBool(*f.Blocked)
	}
	if f.CreatedFrom != nil {
		s += ":created_from:" + f.CreatedFrom.Format("2006-01-02")
	}
	if f.CreatedTo != nil {
		s += ":created_to:" + f.CreatedTo.Format("2006-01-02")
	}
	return s
}

func AccountKey(id ID) string {
	return "account:" + string(id)
}

// AccountOrganizationMemberKey caches whether an account belongs to an
// organization (see AccountService.IsMemberOfOrganization).
func AccountOrganizationMemberKey(accountID, organizationID ID) string {
	return "account_organization_member:" + string(accountID) + ":" + string(organizationID)
}
