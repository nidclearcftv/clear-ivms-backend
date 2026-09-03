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

type AccountFilters struct {
}

func (f *AccountFilters) String() string {
	return ""
}

func AccountKey(id ID) string {
	return "account:" + string(id)
}
