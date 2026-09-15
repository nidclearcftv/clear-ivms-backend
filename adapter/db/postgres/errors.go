package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// Postgres SQLSTATE codes this adapter cares about.
// https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	pgErrCodeUniqueViolation     = "23505"
	pgErrCodeForeignKeyViolation = "23503"
	// pgErrCodeRestrictViolation is what Postgres raises instead of
	// pgErrCodeForeignKeyViolation when a DELETE/UPDATE on the *referenced*
	// row is blocked by an ON DELETE/UPDATE RESTRICT foreign key — e.g.
	// deleting an organization that still has accounts attached via
	// fk_account_organizations_organization. Distinct SQLSTATE, same
	// ConstraintName reporting, so foreignKeyViolationConstraint treats it
	// the same way.
	pgErrCodeRestrictViolation = "23001"
)

// asPgError unwraps err into a *pgconn.PgError, if it is (or wraps) one.
func asPgError(err error) (*pgconn.PgError, bool) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr, true
	}
	return nil, false
}

func isUniqueViolation(err error) bool {
	pgErr, ok := asPgError(err)
	return ok && pgErr.Code == pgErrCodeUniqueViolation
}

// foreignKeyViolationConstraint returns the violated constraint's name if
// err is a foreign key or restrict violation, and "" otherwise.
func foreignKeyViolationConstraint(err error) string {
	pgErr, ok := asPgError(err)
	if !ok || (pgErr.Code != pgErrCodeForeignKeyViolation && pgErr.Code != pgErrCodeRestrictViolation) {
		return ""
	}
	return pgErr.ConstraintName
}
