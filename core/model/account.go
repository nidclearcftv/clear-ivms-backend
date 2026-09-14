package model

import "time"

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

// AccountFilters narrows/paginates an account listing. Search, when set,
// matches accounts whose name or email contains it (case-insensitive).
// Page defaults to 1 and PageSize to AccountDefaultPageSize when unset.
type AccountFilters struct {
	Search   string `form:"search"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

func (f *AccountFilters) String() string {
	return "search:" + f.Search
}

func AccountKey(id ID) string {
	return "account:" + string(id)
}

// AccountOrganizationMemberKey caches whether an account belongs to an
// organization (see AccountService.IsMemberOfOrganization).
func AccountOrganizationMemberKey(accountID, organizationID ID) string {
	return "account_organization_member:" + string(accountID) + ":" + string(organizationID)
}
