package model

import "time"

// AccountSession is a logged-in session for an Account. UserAgent,
// IPAddress and RevokedAt are pointers because they're nullable in the
// schema: a nil RevokedAt means the session is still active.
type AccountSession struct {
	ID        ID
	AccountID ID
	TokenHash string
	UserAgent *string
	IPAddress *string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AccountSessionKey is keyed by the session's token hash, not its ID —
// that's what a request actually has on hand (see AccountService.Authenticate).
func AccountSessionKey(tokenHash string) string {
	return "account_session:" + tokenHash
}
