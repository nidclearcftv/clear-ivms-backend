package port

import (
	"context"

	"github.com/nidclearcftv/clear-ivms-backend/core/model"
)

// AccountRepository is the driven (secondary) port for persisting and
// reading account data, including its sessions (account_sessions) — there's
// no separate session port; it's part of the account's own lifecycle.
// It is implemented by outbound adapters, e.g. adapter/db/postgres.
//
// Relation methods only ever return/accept Account — e.g.
// ListFromOrganization returns the accounts belonging to an organization,
// not the organization itself.
//
// The password hash never appears on model.Account, so it's threaded
// through explicitly: passwordHash on Create, GetPassword to read it back
// (e.g. to verify a login attempt), and SetPassword to change it.
type AccountRepository interface {
	Create(ctx context.Context, account model.Account, passwordHash string) (model.Account, error)
	Get(ctx context.Context, id model.ID) (model.Account, error)
	List(ctx context.Context, filters model.AccountFilters) (model.List[model.Account], error)
	Update(ctx context.Context, account model.Account) (model.Account, error)
	Delete(ctx context.Context, id model.ID) error
	GetPassword(ctx context.Context, id model.ID) (string, error)
	SetPassword(ctx context.Context, id model.ID, passwordHash string) error
	// GetByEmailWithPassword looks up an account by its (unique) email,
	// returning its password hash alongside it — the one query a login
	// needs both halves of.
	GetByEmailWithPassword(ctx context.Context, email string) (model.Account, string, error)

	CreateSession(ctx context.Context, session model.AccountSession) (model.AccountSession, error)
	GetSession(ctx context.Context, tokenHash string) (model.AccountSession, error)
	ListSessions(ctx context.Context, accountID model.ID) (model.List[model.AccountSession], error)
	// RevokeSession returns the revoked session's token hash (or "" if it
	// was already revoked — nothing new happened), so AccountService can
	// invalidate whatever it caches sessions under (see
	// AccountService.Authenticate) without needing a separate lookup.
	RevokeSession(ctx context.Context, id model.ID) (tokenHash string, err error)
	// RevokeSessionByToken is RevokeSession keyed by the session's token
	// hash instead of its ID — what a logout request actually has on hand.
	// It has no need to return the hash: the caller already has it.
	RevokeSessionByToken(ctx context.Context, tokenHash string) error
	// RevokeAllSessions returns the revoked sessions' token hashes, same
	// reasoning as RevokeSession.
	RevokeAllSessions(ctx context.Context, accountID model.ID) (tokenHashes []string, err error)

	ListFromOrganization(ctx context.Context, organizationID model.ID) (model.List[model.Account], error)
	AddOrganization(ctx context.Context, accountID, organizationID model.ID) error
	RemoveOrganization(ctx context.Context, accountID, organizationID model.ID) error

	ListFromTeam(ctx context.Context, teamID model.ID) (model.List[model.Account], error)
	AddTeam(ctx context.Context, accountID, teamID model.ID) error
	RemoveTeam(ctx context.Context, accountID, teamID model.ID) error
}

// AccountService is the driving (primary) port exposing account-related
// business operations to inbound adapters, e.g. adapter/http controllers.
type AccountService interface {
	Create(ctx context.Context, account model.Account, passwordHash string) (model.Account, error)
	Get(ctx context.Context, id model.ID) (model.Account, error)
	List(ctx context.Context, filters model.AccountFilters) (model.List[model.Account], error)
	Update(ctx context.Context, account model.Account) (model.Account, error)
	Delete(ctx context.Context, id model.ID) error
	GetPassword(ctx context.Context, id model.ID) (string, error)
	SetPassword(ctx context.Context, id model.ID, passwordHash string) error

	// Login verifies email/password and, on success, starts a new session,
	// returning the account and the session's raw (unhashed) token — the
	// value to hand back to the client (e.g. as a cookie); only its hash is
	// ever persisted. Fails with ErrCodeInvalidCredentials for either an
	// unknown email or a wrong password (never distinguished, to avoid
	// leaking which emails are registered), or ErrCodeAccountBlocked.
	Login(ctx context.Context, email, password string) (model.Account, string, error)
	// Logout revokes the session identified by its raw token (as returned
	// by Login).
	Logout(ctx context.Context, token string) error
	// Authenticate validates a raw session token (as returned by Login) and
	// returns the account it belongs to. Fails with
	// ErrCodeInvalidCredentials if the token doesn't correspond to an
	// active (unexpired, unrevoked) session, or ErrCodeAccountBlocked if
	// the account has since been blocked.
	Authenticate(ctx context.Context, token string) (model.Account, error)

	CreateSession(ctx context.Context, session model.AccountSession) (model.AccountSession, error)
	GetSession(ctx context.Context, tokenHash string) (model.AccountSession, error)
	ListSessions(ctx context.Context, accountID model.ID) (model.List[model.AccountSession], error)
	RevokeSession(ctx context.Context, id model.ID) error
	RevokeAllSessions(ctx context.Context, accountID model.ID) error

	ListFromOrganization(ctx context.Context, organizationID model.ID) (model.List[model.Account], error)
	AddOrganization(ctx context.Context, accountID, organizationID model.ID) error
	RemoveOrganization(ctx context.Context, accountID, organizationID model.ID) error

	ListFromTeam(ctx context.Context, teamID model.ID) (model.List[model.Account], error)
	AddTeam(ctx context.Context, accountID, teamID model.ID) error
	RemoveTeam(ctx context.Context, accountID, teamID model.ID) error
}
